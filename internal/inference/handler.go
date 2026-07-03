package inference

import (
	"net"
	"sync"
	"time"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/dataset"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/logger"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/ml"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/protocol"
)

type RequestCounter struct {
	mu    sync.RWMutex
	count int64
}

func (rc *RequestCounter) Inc() {
	rc.mu.Lock()
	rc.count++
	rc.mu.Unlock()
}

func (rc *RequestCounter) Value() int64 {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return rc.count
}

func HandleConnection(conn net.Conn, model *ml.Model, counter *RequestCounter) {
	defer conn.Close()

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
		recStart := time.Now()
		class, confidence, probs := model.Predict(&rec)
		recLatency := float64(time.Since(recStart).Nanoseconds()) / 1e6

		results[i] = protocol.PredictionResult{
			Class:         class,
			Confidence:    confidence,
			Probabilities: probs,
			LatencyMs:     recLatency,
		}
	}

	latencyMs := float64(time.Since(start).Nanoseconds()) / 1e6

	resp := protocol.InferenceResponse{
		Predictions: results,
		LatencyMs:   latencyMs,
	}

	if err := protocol.WriteMessage(conn, &resp); err != nil {
		logger.Error("failed to write response", "error", err)
		return
	}

	counter.Inc()

	logger.Info("request completed",
		"latency_ms", latencyMs,
		"records", len(req.Records),
	)
}
