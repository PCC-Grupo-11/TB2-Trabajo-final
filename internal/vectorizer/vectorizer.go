package vectorizer

import (
	"fmt"
	"math"
	"runtime"
	"sync"
	"time"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/config"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/dataset"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/protocol"
	"github.com/uber/h3-go/v4"
)

type ValidationError string

func (e ValidationError) Error() string { return string(e) }

type Vectorizer struct {
	loader *Loader
	loc    *time.Location
}

func New(mappingsDir string) (*Vectorizer, error) {
	l, err := NewLoader(mappingsDir)
	if err != nil {
		return nil, err
	}
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		return nil, fmt.Errorf("failed to load timezone: %w", err)
	}
	return &Vectorizer{loader: l, loc: loc}, nil
}

func (v *Vectorizer) Vectorize(input *protocol.PredictRequest) (protocol.PredictRecord, error) {
	if err := v.validateCategoricals(
		&input.Agency, &input.ComplaintType,
		&input.Descriptor, &input.LocationType,
	); err != nil {
		return protocol.PredictRecord{}, err
	}
	if _, ok := v.loader.BoroughMap[input.Borough]; !ok {
		if f := catchAll(v.loader.BoroughMap); f != "" {
			input.Borough = f
		}
	}

	t := time.Unix(input.Timestamp, 0).In(v.loc)
	features := buildFeatures(
		t,
		input.Latitude,
		input.Longitude,
		input.Agency,
		input.ComplaintType,
		input.Descriptor,
		input.LocationType,
		input.Borough,
		v.loader,
	)

	return protocol.PredictRecord{Features: features}, nil
}

type HexRecord struct {
	ParentHex string
	Record    protocol.PredictRecord
}

func (v *Vectorizer) VectorizeBulk(hexes []string, boroughs map[string]string, categories *protocol.BulkPredictRequest) ([]HexRecord, error) {
	if err := v.validateCategoricals(
		&categories.Agency, &categories.ComplaintType,
		&categories.Descriptor, &categories.LocationType,
	); err != nil {
		return nil, err
	}

	t := time.Unix(categories.Timestamp, 0).In(v.loc)

	if len(hexes) == 0 {
		return nil, nil
	}

	const childrenPerHex = 7

	results := make([]HexRecord, 0, len(hexes)*childrenPerHex)
	type hexResult struct {
		records []HexRecord
		err     error
	}
	hexResults := make([]hexResult, len(hexes))

	type workItem struct {
		idx int
		hex string
	}

	workers := min(runtime.NumCPU()*4, len(hexes))
	jobs := make(chan workItem, workers)
	defaultBorough := catchAll(v.loader.BoroughMap)

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range jobs {
				borough := boroughs[item.hex]
				if _, ok := v.loader.BoroughMap[borough]; !ok && defaultBorough != "" {
					borough = defaultBorough
				}

				cell := h3.CellFromString(item.hex)
				if !cell.IsValid() {
					hexResults[item.idx] = hexResult{err: ValidationError(fmt.Sprintf("invalid hex %q", item.hex))}
					continue
				}

				res := cell.Resolution()
				children, err := cell.Children(res + 1)
				if err != nil {
					hexResults[item.idx] = hexResult{err: fmt.Errorf("failed to get children for hex %q: %w", item.hex, err)}
					continue
				}

				local := make([]HexRecord, 0, len(children))
				var childErr error
				for _, child := range children {
					childHex := child.String()
					centroid, err := h3.CellToLatLng(child)
					if err != nil {
						childErr = fmt.Errorf("failed to get centroid for child hex %q: %w", childHex, err)
						break
					}
					features := buildFeatures(
						t,
						centroid.Lat,
						centroid.Lng,
						categories.Agency,
						categories.ComplaintType,
						categories.Descriptor,
						categories.LocationType,
						borough,
						v.loader,
					)
					local = append(local, HexRecord{
						ParentHex: item.hex,
						Record:    protocol.PredictRecord{Features: features},
					})
				}
				if childErr != nil {
					hexResults[item.idx] = hexResult{err: childErr}
				} else {
					hexResults[item.idx] = hexResult{records: local}
				}
			}
		}()
	}

	for i, hex := range hexes {
		jobs <- workItem{idx: i, hex: hex}
	}
	close(jobs)
	wg.Wait()

	for _, hr := range hexResults {
		if hr.err != nil {
			return nil, hr.err
		}
		results = append(results, hr.records...)
	}

	return results, nil
}

func (v *Vectorizer) validateCategoricals(agency, complaint, descriptor, location *string) error {
	if _, ok := v.loader.AgencyMap[*agency]; !ok {
		return ValidationError(fmt.Sprintf("unknown agency: %q", *agency))
	}
	if _, ok := v.loader.ComplaintMap[*complaint]; !ok {
		return ValidationError(fmt.Sprintf("unknown complaint_type: %q", *complaint))
	}
	if _, ok := v.loader.DescriptorMap[*descriptor]; !ok {
		if f := catchAll(v.loader.DescriptorMap); f != "" {
			*descriptor = f
		}
	}
	if _, ok := v.loader.LocationMap[*location]; !ok {
		if f := catchAll(v.loader.LocationMap); f != "" {
			*location = f
		}
	}
	return nil
}

func catchAll(m map[string]int) string {
	for _, k := range []string{"Unspecified", "OTHER", "UNKNOWN"} {
		if _, ok := m[k]; ok {
			return k
		}
	}
	return ""
}

func buildFeatures(t time.Time, lat, lon float64, agency, complaintType, descriptor, locationType, borough string, loader *Loader) [config.FeaturesPerRecord]dataset.SparseFeature {

	secs := t.Hour()*3600 + t.Minute()*60 + t.Second()
	dayFrac := float64(secs) / 86400
	day := (t.Weekday() + 6) % 7
	weekSecs := float64(day)*86400 + float64(secs)
	weekFrac := weekSecs / (7 * 86400)

	timeSin := float32(math.Round(math.Sin(2*math.Pi*dayFrac)*1e4) / 1e4)
	timeCos := float32(math.Round(math.Cos(2*math.Pi*dayFrac)*1e4) / 1e4)
	weekSin := float32(math.Round(math.Sin(2*math.Pi*weekFrac)*1e4) / 1e4)
	weekCos := float32(math.Round(math.Cos(2*math.Pi*weekFrac)*1e4) / 1e4)
	isWeekend := float32(0)
	if day == 5 || day == 6 {
		isWeekend = 1
	}

	latZ := float32(math.Round((lat-loader.LatMean)/loader.LatStd*1e6) / 1e6)
	lonZ := float32(math.Round((lon-loader.LonMean)/loader.LonStd*1e6) / 1e6)

	agencyID := loader.AgencyMap[agency]
	complaintID := loader.ComplaintMap[complaintType]
	descriptorID := loader.DescriptorMap[descriptor]
	locationID := loader.LocationMap[locationType]
	boroughID := loader.BoroughMap[borough]

	comboKey := complaintType + "|" + descriptor + "|" + locationType
	comboID := loader.ComboMap[comboKey]

	return [config.FeaturesPerRecord]dataset.SparseFeature{
		{Index: 0, Value: timeSin},
		{Index: 1, Value: timeCos},
		{Index: 2, Value: weekSin},
		{Index: 3, Value: weekCos},
		{Index: 4, Value: isWeekend},
		{Index: 5, Value: latZ},
		{Index: 6, Value: lonZ},
		{Index: uint32(config.AgencyOffset + agencyID), Value: 1},
		{Index: uint32(config.ComplaintOffset + complaintID), Value: 1},
		{Index: uint32(config.DescriptorOffset + descriptorID), Value: 1},
		{Index: uint32(config.LocationOffset + locationID), Value: 1},
		{Index: uint32(config.BoroughOffset + boroughID), Value: 1},
		{Index: uint32(config.ComboOffset + comboID), Value: 1},
	}
}
