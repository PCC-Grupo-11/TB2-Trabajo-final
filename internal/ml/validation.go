package ml

import (
	"math"
	"sync"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/config"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/dataset"
)

type lossResult struct {
	loss float64
	n    int
}

func sampleLogLoss(model *Model, sample *dataset.Record, logits, probs []float32) float64 {
	model.ComputeProbs(sample, logits, probs)
	p := float64(probs[sample.Y])
	if p < probFloor {
		p = probFloor
	}
	return -math.Log(p)
}

func computeLoss(model *Model, valSet *dataset.Dataset) float32 {
	samples := valSet.Records
	if len(samples) == 0 {
		return 0
	}

	chunkSize := (len(samples) + config.NumWorkers - 1) / config.NumWorkers

	results := make(chan lossResult, config.NumWorkers)

	var wg sync.WaitGroup
	for w := range config.NumWorkers {
		start := w * chunkSize
		if start >= len(samples) {
			break
		}
		end := min(start+chunkSize, len(samples))

		wg.Add(1)
		go func(chunk []dataset.Record) {
			defer wg.Done()
			logits := make([]float32, model.NumClasses)
			probs := make([]float32, model.NumClasses)
			var loss float64

			for _, sample := range chunk {
				loss += sampleLogLoss(model, &sample, logits, probs)
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
	confusion        []int
}

func computeValidation(model *Model, valSet *dataset.Dataset) (loss float32, accuracy, mae, accAt1 float64, confusionMatrix [][]int) {
	samples := valSet.Records
	n := len(samples)
	if n == 0 {
		loss, accuracy, mae, accAt1 = 0, 0, 0, 0
		return
	}

	chunkSize := (n + config.NumWorkers - 1) / config.NumWorkers

	results := make(chan valResult, config.NumWorkers)

	var wg sync.WaitGroup
	for w := range config.NumWorkers {
		start := w * chunkSize
		if start >= n {
			break
		}
		end := min(start+chunkSize, n)

		wg.Add(1)
		go func(chunk []dataset.Record) {
			defer wg.Done()
			logits := make([]float32, model.NumClasses)
			probs := make([]float32, model.NumClasses)
			nc := model.NumClasses

			var res valResult
			res.confusion = make([]int, nc*nc)
			for _, sample := range chunk {
				res.loss += sampleLogLoss(model, &sample, logits, probs)

				bestClass, _ := argmax(probs)

				trueClass := int(sample.Y)
				res.confusion[trueClass*nc+bestClass]++
				if bestClass == trueClass {
					res.correct++
				}

				diff := math.Abs(float64(bestClass - trueClass))
				res.totalMAE += diff
				if int(diff) <= accuracyAt1Threshold {
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

	nc := model.NumClasses
	confusionMatrix = make([][]int, nc)
	for i := range confusionMatrix {
		confusionMatrix[i] = make([]int, nc)
	}

	for r := range results {
		totalLoss += r.loss
		totalCorrect += r.correct
		totalWithin1 += r.within1
		totalMAE += r.totalMAE
		totalN += r.n

		for i := range nc {
			for j := range nc {
				confusionMatrix[i][j] += r.confusion[i*nc+j]
			}
		}
	}

	loss = float32(totalLoss / float64(totalN))
	accuracy = float64(totalCorrect) / float64(totalN)
	mae = totalMAE / float64(totalN)
	accAt1 = float64(totalWithin1) / float64(totalN)

	return
}
