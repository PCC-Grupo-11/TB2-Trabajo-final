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
		if _, ok := err.(vectorizer.ValidationError); ok {
			protocol.WriteJSON(w, http.StatusBadRequest, protocol.ErrorResponse{Error: err.Error()})
		} else {
			logger.Error("vectorization failed", "error", err)
			protocol.WriteJSON(w, http.StatusInternalServerError, protocol.ErrorResponse{Error: "vectorization failed"})
		}
		return
	}

	cacheKey := storage.CacheKey(record)
	if cached := h.checkCache(r.Context(), cacheKey); cached != nil {
		protocol.WriteJSON(w, http.StatusOK, map[string]any{
			"class":         cached.Class,
			"confidence":    cached.Confidence,
			"probabilities": cached.Probabilities,
			"latency_ms":    cached.LatencyMs,
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

	if h.Cache != nil {
		if data, err := json.Marshal(result); err != nil {
			logger.Warn("failed to marshal cache entry", "error", err)
		} else {
			err := h.Cache.SetPrediction(r.Context(), cacheKey, data)
			if err != nil {
				logger.Error("failed to set cache entry", "error", err)
			}
		}
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

func (h *Handler) PredictBulk(w http.ResponseWriter, r *http.Request) {
	var req protocol.BulkPredictRequest
	if err := protocol.ReadMessage(r.Body, &req); err != nil {
		protocol.WriteJSON(w, http.StatusBadRequest, protocol.ErrorResponse{Error: "invalid request body"})
		return
	}

	hexRecords, err := h.Vec.VectorizeBulk(&req)
	if err != nil {
		if _, ok := err.(vectorizer.ValidationError); ok {
			protocol.WriteJSON(w, http.StatusBadRequest, protocol.ErrorResponse{Error: err.Error()})
		} else {
			logger.Error("bulk vectorization failed", "error", err)
			protocol.WriteJSON(w, http.StatusInternalServerError, protocol.ErrorResponse{Error: "vectorization failed"})
		}
		return
	}

	n := len(hexRecords)
	results := make([]protocol.PredictionResult, n)

	var uncachedRecords []protocol.PredictRecord
	uncachedIndices := make([]int, 0, n)

	for i, hr := range hexRecords {
		if h.Cache != nil {
			cacheKey := storage.CacheKey(hr.Record)
			data, err := h.Cache.GetPrediction(r.Context(), cacheKey)
			if err == nil {
				var entry protocol.PredictionResult
				if err := json.Unmarshal(data, &entry); err != nil {
					logger.Warn("child cache data corrupted", "cache_key", cacheKey, "error", err)
				} else {
					results[i] = entry
					continue
				}
			}
		}
		uncachedRecords = append(uncachedRecords, hr.Record)
		uncachedIndices = append(uncachedIndices, i)
	}

	var latencyMs float64
	if len(uncachedRecords) > 0 {
		inferResp, err := h.LB.Predict(&protocol.InferenceRequest{Records: uncachedRecords})
		if err != nil {
			logger.Error("bulk inference failed", "error", err)
			protocol.WriteJSON(w, http.StatusServiceUnavailable, protocol.ErrorResponse{Error: "inference failed"})
			return
		}

		latencyMs = inferResp.LatencyMs

		for j, idx := range uncachedIndices {
			results[idx] = inferResp.Predictions[j]
		}
		if h.Cache != nil {
			for _, idx := range uncachedIndices {
				if data, err := json.Marshal(results[idx]); err != nil {
					logger.Warn("failed to marshal child cache entry", "error", err)
				} else {
					cacheKey := storage.CacheKey(hexRecords[idx].Record)
					h.Cache.SetPrediction(r.Context(), cacheKey, data)
				}
			}
			h.Cache.IncrBy(r.Context(), "predictions_count", int64(len(uncachedRecords)))
			h.Cache.IncrByFloat(r.Context(), "latency_sum", latencyMs)
		}
	}

	hexPredictions := averageByParent(hexRecords, results)

	resp := protocol.BulkPredictResponse{
		Results:   hexPredictions,
		LatencyMs: latencyMs,
	}

	protocol.WriteJSON(w, http.StatusOK, resp)
}

func averageByParent(
	records []vectorizer.HexRecord,
	results []protocol.PredictionResult,
) []protocol.HexPrediction {
	type accum struct {
		count    int
		probSums []float32
	}

	accums := make(map[string]*accum)
	order := make([]string, 0)

	for i, hr := range records {
		if _, exists := accums[hr.ParentHex]; !exists {
			order = append(order, hr.ParentHex)
			accums[hr.ParentHex] = &accum{probSums: make([]float32, len(results[i].Probabilities))}
		}
		a := accums[hr.ParentHex]
		a.count++
		for k := range a.probSums {
			a.probSums[k] += results[i].Probabilities[k]
		}
	}

	out := make([]protocol.HexPrediction, 0, len(order))
	for _, parentHex := range order {
		a := accums[parentHex]
		n := float32(a.count)

		maxProb := float32(-1)
		bestClass := 0
		for k, sum := range a.probSums {
			avg := sum / n
			if avg > maxProb {
				maxProb = avg
				bestClass = k
			}
		}

		out = append(out, protocol.HexPrediction{
			Hex:        parentHex,
			Class:      bestClass,
			Confidence: maxProb,
		})
	}

	return out
}
