package dataset

import (
	"encoding/csv"
	"io"
	"os"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/config"
)

func readBatches(path string, rowBatchesCh chan<- [][]string) error {
	defer close(rowBatchesCh)

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	r := csv.NewReader(f)

	if _, err := r.Read(); err != nil {
		return err
	}

	batch := make([][]string, 0, config.BatchSize)

	for {
		record, err := r.Read()
		if err == io.EOF {
			if len(batch) > 0 {
				rowBatchesCh <- batch
			}
			return nil
		}
		if err != nil {
			return err
		}

		batch = append(batch, record)

		if len(batch) >= config.BatchSize {
			rowBatchesCh <- batch
			batch = make([][]string, 0, config.BatchSize)
		}
	}
}
