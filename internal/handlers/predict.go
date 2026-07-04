package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/auth"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/logger"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/protocol"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/storage"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/vectorizer"
)

func (h *Handler) Predict(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok {
		protocol.WriteJSON(w, http.StatusUnauthorized, protocol.ErrorResponse{Error: "unauthorized"})
		return
	}

	start := time.Now()

	var req protocol.PredictRequest
	if err := protocol.ReadMessage(r.Body, &req); err != nil {
		protocol.WriteJSON(w, http.StatusBadRequest, protocol.ErrorResponse{Error: "invalid request body"})
		return
	}

	record, err := h.Vec.Vectorize(&req)
	if err != nil {
		h.writeVectorizationError(w, err)
		return
	}

	cacheKey := storage.CacheKey(record)
	if cached := h.checkCache(r.Context(), cacheKey); cached != nil {
		h.writePredictResponse(w, *cached, true)
		return
	}

	inferReq := &protocol.InferenceRequest{Records: []protocol.PredictRecord{record}}
	resp, err := h.LB.Predict(inferReq)
	if err != nil {
		logger.Error("inference failed", "error", err)
		protocol.WriteJSON(w, http.StatusServiceUnavailable, protocol.ErrorResponse{Error: "inference failed"})
		return
	}

	result := resp.Predictions[0]
	result.LatencyMs = float64(time.Since(start).Nanoseconds()) / 1e6

	if h.Repo != nil {
		go func() {
			ctx, cancel := context.WithTimeout(
				context.Background(),
				5*time.Second,
			)
			defer cancel()
			err := h.Repo.SavePrediction(ctx, userID, req, result)
			if err != nil {
				logger.Error("failed to save prediction to database", "error", err)
			}
		}()
	}

	h.writeCache(r.Context(), cacheKey, result)
	if h.Cache != nil {
		err := h.Cache.IncrMetricsPipeline(r.Context(),
			map[string]int64{
				"predictions_count": 1,
				"cache_misses":      1,
			},
			map[string]float64{
				"latency_sum": result.LatencyMs,
			},
		)

		if err != nil {
			logger.Warn("failed to increment metrics pipeline", "error", err)
		}
	}

	h.writePredictResponse(w, result, false)
}

func (h *Handler) writePredictResponse(w http.ResponseWriter, result protocol.PredictionResult, cached bool) {
	latencyMs := result.LatencyMs
	if cached {
		latencyMs = 0
	}
	protocol.WriteJSON(w, http.StatusOK, map[string]any{
		"class":         result.Class,
		"confidence":    result.Confidence,
		"probabilities": result.Probabilities,
		"latency_ms":    latencyMs,
		"cached":        cached,
	})
}

func (h *Handler) checkCache(ctx context.Context, hash string) *protocol.PredictionResult {
	if h.Cache == nil {
		return nil
	}

	data, err := h.Cache.GetPrediction(ctx, hash)
	if err != nil {
		return nil
	}

	var cr protocol.PredictionResult
	if err := json.Unmarshal(data, &cr); err != nil {
		logger.Warn("cache data corrupted", "hash", hash, "error", err)
		return nil
	}

	h.Cache.IncrCounter(ctx, "cache_hits")
	return &cr
}

func (h *Handler) writeVectorizationError(w http.ResponseWriter, err error) {
	if _, ok := err.(vectorizer.ValidationError); ok {
		protocol.WriteJSON(w, http.StatusBadRequest, protocol.ErrorResponse{Error: err.Error()})
		return
	}
	logger.Error("vectorization failed", "error", err)
	protocol.WriteJSON(w, http.StatusInternalServerError, protocol.ErrorResponse{Error: "vectorization failed"})
}

func (h *Handler) writeCache(ctx context.Context, key string, v any) {
	if h.Cache == nil {
		return
	}
	data, err := json.Marshal(v)
	if err != nil {
		logger.Warn("failed to marshal cache entry", "error", err)
		return
	}
	if err := h.Cache.SetPrediction(ctx, key, data); err != nil {
		logger.Error("failed to set cache entry", "error", err)
	}
}
