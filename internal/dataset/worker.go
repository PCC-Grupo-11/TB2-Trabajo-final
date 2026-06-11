package dataset

import (
	"strconv"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/config"
)

func worker(
	rowBatchesCh <-chan [][]string,
	recordBatchesCh chan<- []Record,
) {
	for rowBatch := range rowBatchesCh {
		records := make([]Record, len(rowBatch))

		for i, row := range rowBatch {
			records[i] = parseRow(row)
		}

		recordBatchesCh <- records
	}
}

func parseRow(row []string) Record {
	y, _ := strconv.ParseUint(row[0], 10, 8)

	timeSin, _ := strconv.ParseFloat(row[1], 32)
	timeCos, _ := strconv.ParseFloat(row[2], 32)
	weekSin, _ := strconv.ParseFloat(row[3], 32)
	weekCos, _ := strconv.ParseFloat(row[4], 32)
	isWeekend, _ := strconv.ParseFloat(row[5], 32)
	latZ, _ := strconv.ParseFloat(row[6], 32)
	lonZ, _ := strconv.ParseFloat(row[7], 32)

	agencyID, _ := strconv.Atoi(row[8])
	complaintID, _ := strconv.Atoi(row[9])
	descriptorID, _ := strconv.Atoi(row[10])
	locationID, _ := strconv.Atoi(row[11])
	boroughID, _ := strconv.Atoi(row[12])
	comboID, _ := strconv.Atoi(row[13])

	return Record{
		Y: uint8(y),
		Features: [config.FeaturesPerRecord]SparseFeature{
			{0, float32(timeSin)},
			{1, float32(timeCos)},
			{2, float32(weekSin)},
			{3, float32(weekCos)},
			{4, float32(isWeekend)},
			{5, float32(latZ)},
			{6, float32(lonZ)},
			{uint32(config.AgencyOffset + agencyID), 1},
			{uint32(config.ComplaintOffset + complaintID), 1},
			{uint32(config.DescriptorOffset + descriptorID), 1},
			{uint32(config.LocationOffset + locationID), 1},
			{uint32(config.BoroughOffset + boroughID), 1},
			{uint32(config.ComboOffset + comboID), 1},
		},
	}
}
