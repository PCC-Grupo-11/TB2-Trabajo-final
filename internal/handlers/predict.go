package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/auth"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/logger"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/protocol"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/storage"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/vectorizer"
)

func (h *Handler) Predict(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(auth.UserIDKey).(string)

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
		protocol.WriteJSON(w, http.StatusOK, map[string]any{
			"class":         cached.Class,
			"confidence":    cached.Confidence,
			"probabilities": cached.Probabilities,
			"latency_ms":    0,
			"cached":        true,
		})
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

	if h.Repo != nil {
		go func() {
			err := h.Repo.SavePrediction(context.Background(), userID, req, result)
			if err != nil {
				logger.Error("failed to save prediction to database", "error", err)
			}
		}()
	}

	h.writeCache(r.Context(), cacheKey, result)
	if h.Cache != nil {
		h.Cache.IncrCounter(r.Context(), "predictions_count")
		h.Cache.IncrByFloat(r.Context(), "latency_sum", result.LatencyMs)
		h.Cache.IncrCounter(r.Context(), "cache_misses")
	}

	protocol.WriteJSON(w, http.StatusOK, map[string]any{
		"class":         result.Class,
		"confidence":    result.Confidence,
		"probabilities": result.Probabilities,
		"latency_ms":    result.LatencyMs,
		"cached":        false,
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
