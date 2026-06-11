package dataset

import (
	"sync"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/config"
)

func Load(path string) (*Dataset, error) {
	rowBatchesCh := make(chan [][]string, channelBufferSize)
	recordBatchesCh := make(chan []Record, channelBufferSize)
	errCh := make(chan error, errChanBufferSize)

	go func() {
		defer close(errCh)

		if err := readBatches(path, rowBatchesCh); err != nil {
			errCh <- err
		}
	}()

	var wg sync.WaitGroup
	wg.Add(config.NumWorkers)
	for range config.NumWorkers {
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
	ds.Shuffle(config.GlobalSeed)
	return ds, nil
}
