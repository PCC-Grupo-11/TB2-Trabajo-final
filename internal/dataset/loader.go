package dataset

import (
	"runtime"
	"sync"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/config"
)

func Load(path string) (*Dataset, error) {
	rowBatchesCh := make(chan [][]string, 16)
	recordBatchesCh := make(chan []Record, 16)
	errCh := make(chan error, 1)

	numWorkers := runtime.NumCPU()

	go func() {
		defer close(errCh)

		if err := readBatches(path, rowBatchesCh); err != nil {
			errCh <- err
		}
	}()

	var wg sync.WaitGroup
	wg.Add(numWorkers)
	for range numWorkers {
		go func() {
			defer wg.Done()
			worker(rowBatchesCh, recordBatchesCh)
		}()
	}

	go func() {
		wg.Wait()
		close(recordBatchesCh)
	}()

	var records []Record
	for recordBatch := range recordBatchesCh {
		records = append(records, recordBatch...)
	}

	err, ok := <-errCh
	if ok && err != nil {
		return nil, err
	}

	ds := &Dataset{Records: records}
	ds.Shuffle(int64(config.GlobalSeed))
	return ds, nil
}
