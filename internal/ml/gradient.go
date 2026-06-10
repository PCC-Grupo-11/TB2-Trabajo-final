package ml

import (
	"sync"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/config"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/dataset"
)

type Gradient struct {
	WeightGrad []float32
	BiasGrad   []float32
}

var gradientPool = sync.Pool{
	New: func() any {
		totalWeights := config.TotalFeatures * config.NumClasses
		return &Gradient{
			WeightGrad: make([]float32, totalWeights),
			BiasGrad:   make([]float32, config.NumClasses),
		}
	},
}

func getGradient() *Gradient {
	return gradientPool.Get().(*Gradient)
}

func putGradient(g *Gradient) {
	clear(g.WeightGrad)
	clear(g.BiasGrad)
	gradientPool.Put(g)
}

func computeGradientInto(batch []dataset.Record, model *Model, grad *Gradient, logits, probs []float32) {
	for _, sample := range batch {
		model.ComputeProbs(&sample, logits, probs)

		y := int(sample.Y)

		for c := range model.NumClasses {
			indicator := float32(0)
			if c == y {
				indicator = 1
			}
			delta := probs[c] - indicator
			base := model.ClassOffset(c)

			for _, f := range sample.Features {
				grad.WeightGrad[base+int(f.Index)] += delta * f.Value
			}
			grad.BiasGrad[c] += delta
		}
	}
}
