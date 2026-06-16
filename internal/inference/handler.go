package inference

import (
	"net"
	"time"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/dataset"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/logger"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/ml"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/protocol"
)

func HandleConnection(conn net.Conn, model *ml.Model) {
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(30 * time.Second))

	var req protocol.InferenceRequest
	if err := protocol.ReadMessage(conn, &req); err != nil {
		logger.Error("failed to read request", "error", err)
		return
	}

	logger.Info("processing request",
		"record_count", len(req.Records),
	)

	start := time.Now()

	results := make([]protocol.PredictionResult, len(req.Records))

	for i := range req.Records {
		rec := dataset.Record{Features: req.Records[i].Features}
		class, confidence, probs := model.Predict(&rec)

		results[i] = protocol.PredictionResult{
			Class:         class,
			Confidence:    confidence,
			Probabilities: probs,
		}
	}

	latencyMs := float64(time.Since(start).Microseconds()) / 1000.0

	resp := protocol.InferenceResponse{
		Predictions: results,
		LatencyMs:   latencyMs,
	}

	if err := protocol.WriteMessage(conn, &resp); err != nil {
		logger.Error("failed to write response", "error", err)
		return
	}

	logger.Info("request completed",
		"latency_ms", latencyMs,
		"records", len(req.Records),
	)
}
