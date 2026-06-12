package ml

import (
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/dataset"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/logger"
)

func (m *Model) ComputeProbs(record *dataset.Record, logits, probs []float32) {
	for c := range m.NumClasses {
		var sum float32 = m.Biases[c]
		base := m.ClassOffset(c)
		for _, f := range record.Features {
			sum += m.Weights[base+int(f.Index)] * f.Value
		}
		logits[c] = sum
	}
	Softmax(logits, probs)
}

func (m *Model) Predict(record *dataset.Record) (int, float32) {
	if !m.Trained {
		logger.Warn("predicting with untrained model")
	}

	logits, probs := allocForward(m.NumClasses)
	m.ComputeProbs(record, logits, probs)

	return argmax(probs)
}
