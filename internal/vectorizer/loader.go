package vectorizer

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

type ScalerMetadata struct {
	LatMean float64 `json:"latitude_mean"`
	LatStd  float64 `json:"latitude_std"`
	LonMean float64 `json:"longitude_mean"`
	LonStd  float64 `json:"longitude_std"`
}

type Loader struct {
	AgencyMap     map[string]int
	ComplaintMap  map[string]int
	DescriptorMap map[string]int
	LocationMap   map[string]int
	BoroughMap    map[string]int
	ComboMap      map[string]int
	ScalerMetadata
}

func NewLoader(mappingsDir string) (*Loader, error) {
	agencyMap, err := loadReverseMap(mappingsDir + "/agency_map.json")
	if err != nil {
		return nil, fmt.Errorf("agency_map: %w", err)
	}

	complaintMap, err := loadReverseMap(mappingsDir + "/complaint_map.json")
	if err != nil {
		return nil, fmt.Errorf("complaint_map: %w", err)
	}

	descriptorMap, err := loadReverseMap(mappingsDir + "/descriptor_map.json")
	if err != nil {
		return nil, fmt.Errorf("descriptor_map: %w", err)
	}

	locationMap, err := loadReverseMap(mappingsDir + "/location_map.json")
	if err != nil {
		return nil, fmt.Errorf("location_map: %w", err)
	}

	boroughMap, err := loadReverseMap(mappingsDir + "/borough_map.json")
	if err != nil {
		return nil, fmt.Errorf("borough_map: %w", err)
	}

	comboMap, err := loadReverseMap(mappingsDir + "/combo_map.json")
	if err != nil {
		return nil, fmt.Errorf("combo_map: %w", err)
	}

	scaler, err := loadScaler(mappingsDir + "/scaler_metadata.json")
	if err != nil {
		return nil, fmt.Errorf("scaler_metadata: %w", err)
	}

	return &Loader{
		AgencyMap:      agencyMap,
		ComplaintMap:   complaintMap,
		DescriptorMap:  descriptorMap,
		LocationMap:    locationMap,
		BoroughMap:     boroughMap,
		ComboMap:       comboMap,
		ScalerMetadata: *scaler,
	}, nil
}

func loadReverseMap(path string) (map[string]int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var forward map[string]string
	if err := json.Unmarshal(data, &forward); err != nil {
		return nil, err
	}

	reverse := make(map[string]int, len(forward))
	for id, label := range forward {
		n, err := strconv.Atoi(id)
		if err != nil {
			return nil, fmt.Errorf("invalid key %q: %w", id, err)
		}
		reverse[label] = n
	}
	return reverse, nil
}

func loadScaler(path string) (*ScalerMetadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s ScalerMetadata
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}
