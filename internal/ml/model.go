package ml

import (
	"math/rand/v2"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/config"
)

type Model struct {
	FeatureCount int       `json:"feature_count" bson:"feature_count"`
	NumClasses   int       `json:"num_classes" bson:"num_classes"`
	Weights      []float32 `json:"weights" bson:"weights"`
	Biases       []float32 `json:"biases" bson:"biases"`
	Trained      bool      `json:"trained" bson:"trained"`
}

func NewModel() *Model {
	rng := rand.New(rand.NewPCG(uint64(config.GlobalSeed), uint64(config.GlobalSeed)))

	totalWeights := config.TotalFeatures * config.NumClasses
	weights := make([]float32, totalWeights)
	for i := range totalWeights {
		weights[i] = (rng.Float32() - 0.5) * 2 * weightInitScale
	}

	return &Model{
		FeatureCount: config.TotalFeatures,
		NumClasses:   config.NumClasses,
		Weights:      weights,
		Biases:       make([]float32, config.NumClasses),
		Trained:      false,
	}
}

func (m *Model) ClassOffset(class int) int {
	return class * m.FeatureCount
}

func (m *Model) Snapshot() (weights, biases []float32) {
	weights = make([]float32, len(m.Weights))
	copy(weights, m.Weights)
	biases = make([]float32, len(m.Biases))
	copy(biases, m.Biases)
	return
}
