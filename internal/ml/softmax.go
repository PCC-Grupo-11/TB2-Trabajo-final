package ml

import "math"

func Softmax(logits []float32) []float32 {
	if len(logits) == 0 {
		return nil
	}

	max := float32(-math.MaxFloat32)
	for _, v := range logits {
		if v > max {
			max = v
		}
	}

	probs := make([]float32, len(logits))
	var sum float32

	for i, v := range logits {
		probs[i] = float32(math.Exp(float64(v - max)))
		sum += probs[i]
	}

	for i := range probs {
		probs[i] /= sum
	}

	return probs
}
