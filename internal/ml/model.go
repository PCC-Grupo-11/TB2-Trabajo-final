package ml

import (
	"math/rand"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/config"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/schema"
)

const weightInitScale = 0.05

type Model struct {
	FeatureCount int       `json:"feature_count"`
	NumClasses   int       `json:"num_classes"`
	Weights      []float32 `json:"weights"`
	Biases       []float32 `json:"biases"`
	Trained      bool      `json:"trained"`
}

func NewModel() *Model {
	rng := rand.New(rand.NewSource(config.GlobalSeed))

	totalWeights := schema.TotalFeatures * schema.NumClasses
	weights := make([]float32, totalWeights)
	for i := range totalWeights {
		weights[i] = (rng.Float32() - 0.5) * 2 * weightInitScale
	}

	return &Model{
		FeatureCount: schema.TotalFeatures,
		NumClasses:   schema.NumClasses,
		Weights:      weights,
		Biases:       make([]float32, schema.NumClasses),
		Trained:      false,
	}
}

func (m *Model) ClassOffset(class int) int {
	return class * m.FeatureCount
}
