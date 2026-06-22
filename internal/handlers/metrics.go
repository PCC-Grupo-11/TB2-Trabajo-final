package handlers

import (
	"net/http"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/protocol"
)

func (h *Handler) Metrics(w http.ResponseWriter, r *http.Request) {
	var (
		predictionCount int64
		latencySum      float64
		cacheHits       int64
		cacheMisses     int64
	)

	if h.Cache != nil {
		if v, err := h.Cache.GetCounter(r.Context(), "predictions_count"); err == nil {
			predictionCount = v
		}
		if v, err := h.Cache.GetFloat64(r.Context(), "latency_sum"); err == nil {
			latencySum = v
		}
		if v, err := h.Cache.GetCounter(r.Context(), "cache_hits"); err == nil {
			cacheHits = v
		}
		if v, err := h.Cache.GetCounter(r.Context(), "cache_misses"); err == nil {
			cacheMisses = v
		}
	}

	avgLatency := 0.0
	if predictionCount > 0 {
		avgLatency = latencySum / float64(predictionCount)
	}

	cacheHitRate := 0.0
	totalCache := cacheHits + cacheMisses
	if totalCache > 0 {
		cacheHitRate = float64(cacheHits) / float64(totalCache)
	}

	protocol.WriteJSON(w, http.StatusOK, protocol.MetricsResponse{
		AvgLatencyMs:    avgLatency,
		NodeCount:       len(h.Cfg.InferenceAddrs),
		PredictionCount: predictionCount,
		CacheHitRate:    cacheHitRate,
	})
}
