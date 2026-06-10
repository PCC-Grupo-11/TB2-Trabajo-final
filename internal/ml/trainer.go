package ml

import (
	"math"
	"sync"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/config"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/dataset"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/logger"
)

type TrainingReport struct {
	FinalValidationLoss float32 `json:"final_validation_loss"`
	Accuracy            float64 `json:"accuracy"`
	MAE                 float64 `json:"mae"`
	AccuracyAt1         float64 `json:"accuracy_at_1"`
	EpochsTrained       int     `json:"epochs_trained"`
	EarlyStopped        bool    `json:"early_stopped"`
	TotalSamples        int     `json:"total_samples"`
	FeatureCount        int     `json:"feature_count"`
	NumClasses          int     `json:"num_classes"`
	TrainingTimeSeconds float64 `json:"training_time_seconds"`
	LearningRate        float32 `json:"learning_rate"`
	ShuffleSeed         int64   `json:"shuffle_seed"`
	ValidationSplit     float64 `json:"validation_split"`
	ConfusionMatrix     [][]int `json:"confusion_matrix"`
}

func (m *Model) Train(trainSet, valSet *dataset.Dataset) *TrainingReport {
	totalWeights := m.NumClasses * m.FeatureCount

	bestValLoss := float32(math.MaxFloat32)
	var bestWeights []float32
	var bestBiases []float32
	patienceCounter := 0
	epochsTrained := 0
	earlyStopped := false

	trainSamples := trainSet.Records
	numTrain := len(trainSamples)

	logger.Info("starting training",
		"train_samples", numTrain,
		"val_samples", valSet.Len(),
		"workers", config.NumWorkers,
		"learning_rate", LearningRate,
		"epochs", Epochs,
		"patience", EarlyStoppingPatience,
		"l2_lambda", L2Lambda,
		"batch_size", config.BatchSize,
	)

	for epoch := range Epochs {
		epochsTrained = epoch + 1

		trainSet.Shuffle(config.GlobalSeed + int64(epoch))

		batchCh := make(chan []dataset.Record, 5*config.NumWorkers)
		gradCh := make(chan *Gradient, 5*config.NumWorkers)

		// Batch producer
		go func() {
			defer close(batchCh)
			for i := 0; i < numTrain; i += config.BatchSize {
				end := i + config.BatchSize
				if end > numTrain {
					end = numTrain
				}
				batchCh <- trainSamples[i:end]
			}
		}()

		var wg sync.WaitGroup
		wg.Add(config.NumWorkers)
		for range config.NumWorkers {
			go func() {
				defer wg.Done()
				logits := make([]float32, m.NumClasses)
				probs := make([]float32, m.NumClasses)

				for batch := range batchCh {
					grad := getGradient()
					computeGradientInto(batch, m, grad, logits, probs)
					gradCh <- grad
				}
			}()
		}

		go func() {
			wg.Wait()
			close(gradCh)
		}()

		totalGrad := &Gradient{
			WeightGrad: make([]float32, totalWeights),
			BiasGrad:   make([]float32, m.NumClasses),
		}

		for grad := range gradCh {
			for i := range totalGrad.WeightGrad {
				totalGrad.WeightGrad[i] += grad.WeightGrad[i]
			}
			for c := range totalGrad.BiasGrad {
				totalGrad.BiasGrad[c] += grad.BiasGrad[c]
			}
			putGradient(grad)
		}

		invN := 1.0 / float32(numTrain)

		// Average gradients and add L2 regularization
		for i := range totalGrad.WeightGrad {
			totalGrad.WeightGrad[i] *= invN
			totalGrad.WeightGrad[i] += 2 * L2Lambda * m.Weights[i] // L2
		}
		for c := range totalGrad.BiasGrad {
			totalGrad.BiasGrad[c] *= invN
		}

		// Update parameters
		for i := range m.Weights {
			m.Weights[i] -= LearningRate * totalGrad.WeightGrad[i]
		}
		for c := range m.Biases {
			m.Biases[c] -= LearningRate * totalGrad.BiasGrad[c]
		}

		// Loss calculation
		valLoss := computeLoss(m, valSet)

		logger.Info("epoch complete",
			"epoch", epoch+1,
			"val_loss", valLoss,
		)

		// Early stopping
		if bestValLoss-valLoss > MinImprovement {
			bestValLoss = valLoss
			bestWeights = copyWeights(m.Weights)
			bestBiases = make([]float32, len(m.Biases))
			copy(bestBiases, m.Biases)
			patienceCounter = 0
		} else {
			patienceCounter++
			if patienceCounter >= EarlyStoppingPatience {
				logger.Info("early stopping triggered",
					"epoch", epoch+1,
					"best_val_loss", bestValLoss,
				)
				earlyStopped = true
				break
			}
		}
	}

	if bestWeights != nil {
		m.Weights = bestWeights
		m.Biases = bestBiases
	}

	m.Trained = true

	valLoss, accuracy, mae, accAt1, confusionMatrix := computeValidation(m, valSet)

	return &TrainingReport{
		FinalValidationLoss: valLoss,
		Accuracy:            accuracy,
		MAE:                 mae,
		AccuracyAt1:         accAt1,
		EpochsTrained:       epochsTrained,
		EarlyStopped:        earlyStopped,
		TotalSamples:        numTrain + valSet.Len(),
		FeatureCount:        config.TotalFeatures,
		NumClasses:          config.NumClasses,
		LearningRate:        LearningRate,
		ShuffleSeed:         int64(config.GlobalSeed),
		ValidationSplit:     config.ValidationSplit,
		ConfusionMatrix:     confusionMatrix,
	}
}

func copyWeights(src []float32) []float32 {
	dst := make([]float32, len(src))
	copy(dst, src)
	return dst
}
