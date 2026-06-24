package protocol

import (
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/config"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/dataset"
)

type InferenceRequest struct {
	Records []PredictRecord `json:"records"`
}

type PredictRecord struct {
	Features [config.FeaturesPerRecord]dataset.SparseFeature `json:"features"`
}

type InferenceResponse struct {
	Predictions []PredictionResult `json:"predictions"`
	LatencyMs   float64            `json:"latency_ms"`
}

type PredictionResult struct {
	Class         int       `json:"class"`
	Confidence    float32   `json:"confidence"`
	Probabilities []float32 `json:"probabilities"`
	LatencyMs     float64   `json:"latency_ms"`
}
