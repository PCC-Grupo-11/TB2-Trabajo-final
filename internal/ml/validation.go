package ml

import (
	"math"
	"runtime"
	"sync"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/dataset"
)

type lossResult struct {
	loss float64
	n    int
}

func computeLoss(model *Model, valSet *dataset.Dataset) float32 {
	samples := valSet.Records
	if len(samples) == 0 {
		return 0
	}

	numWorkers := runtime.NumCPU()
	chunkSize := (len(samples) + numWorkers - 1) / numWorkers

	results := make(chan lossResult, numWorkers)

	var wg sync.WaitGroup
	for w := range numWorkers {
		start := w * chunkSize
		if start >= len(samples) {
			break
		}
		end := start + chunkSize
		if end > len(samples) {
			end = len(samples)
		}

		wg.Add(1)
		go func(chunk []dataset.Record) {
			defer wg.Done()
			logits := make([]float32, model.NumClasses)
			probs := make([]float32, model.NumClasses)
			var loss float64

			for _, sample := range chunk {
				model.ComputeProbs(&sample, logits, probs)
				p := float64(probs[sample.Y])
				if p < 1e-15 {
					p = 1e-15
				}
				loss -= math.Log(p)
			}
			results <- lossResult{loss: loss, n: len(chunk)}
		}(samples[start:end])
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var totalLoss float64
	var totalN int
	for r := range results {
		totalLoss += r.loss
		totalN += r.n
	}
	return float32(totalLoss / float64(totalN))
}

type valResult struct {
	loss             float64
	correct, within1 int
	totalMAE         float64
	n                int
}

func computeValidation(model *Model, valSet *dataset.Dataset) (loss float32, accuracy, mae, accAt1 float64) {
	samples := valSet.Records
	n := len(samples)
	if n == 0 {
		return 0, 0, 0, 0
	}

	numWorkers := runtime.NumCPU()
	chunkSize := (n + numWorkers - 1) / numWorkers

	results := make(chan valResult, numWorkers)

	var wg sync.WaitGroup
	for w := range numWorkers {
		start := w * chunkSize
		if start >= n {
			break
		}
		end := start + chunkSize
		if end > n {
			end = n
		}

		wg.Add(1)
		go func(chunk []dataset.Record) {
			defer wg.Done()
			logits := make([]float32, model.NumClasses)
			probs := make([]float32, model.NumClasses)

			var res valResult
			for _, sample := range chunk {
				model.ComputeProbs(&sample, logits, probs)

				p := float64(probs[sample.Y])
				if p < 1e-15 {
					p = 1e-15
				}
				res.loss -= math.Log(p)

				bestClass := 0
				bestProb := float32(0)
				for c, p := range probs {
					if p > bestProb {
						bestProb = p
						bestClass = c
					}
				}

				trueClass := int(sample.Y)
				if bestClass == trueClass {
					res.correct++
				}

				diff := math.Abs(float64(bestClass - trueClass))
				res.totalMAE += diff
				if diff <= 1 {
					res.within1++
				}
			}
			res.n = len(chunk)
			results <- res
		}(samples[start:end])
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var totalLoss, totalMAE float64
	var totalCorrect, totalWithin1, totalN int

	for r := range results {
		totalLoss += r.loss
		totalCorrect += r.correct
		totalWithin1 += r.within1
		totalMAE += r.totalMAE
		totalN += r.n
	}

	loss = float32(totalLoss / float64(totalN))
	accuracy = float64(totalCorrect) / float64(totalN)
	mae = totalMAE / float64(totalN)
	accAt1 = float64(totalWithin1) / float64(totalN)

	return
}
