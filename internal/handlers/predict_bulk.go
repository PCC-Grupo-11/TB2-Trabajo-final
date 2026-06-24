package handlers

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/logger"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/protocol"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/storage"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/vectorizer"
)

func parentCacheKey(parentHex string, req *protocol.BulkPredictRequest) string {
	raw := fmt.Sprintf(
		"%s|%d|%s|%s|%s|%s|%s",
		parentHex,
		req.Timestamp,
		req.Agency,
		req.ComplaintType,
		req.Descriptor,
		req.LocationType,
		req.Borough,
	)
	hash := md5.Sum([]byte(raw))
	return hex.EncodeToString(hash[:])
}

func (h *Handler) PredictBulk(w http.ResponseWriter, r *http.Request) {
	var req protocol.BulkPredictRequest
	if err := protocol.ReadMessage(r.Body, &req); err != nil {
		protocol.WriteJSON(w, http.StatusBadRequest, protocol.ErrorResponse{Error: "invalid request body"})
		return
	}

	parentResults := make(map[string]protocol.HexPrediction, len(req.H3Hexes))
	var missingHexes []string

	if h.Cache != nil {
		for _, parentHex := range req.H3Hexes {
			data, err := h.Cache.GetPrediction(r.Context(), parentCacheKey(parentHex, &req))
			if err != nil {
				missingHexes = append(missingHexes, parentHex)
				continue
			}

			var hp protocol.HexPrediction
			if err := json.Unmarshal(data, &hp); err != nil {
				logger.Warn("parent cache data corrupted", "parent_hex", parentHex, "error", err)
				missingHexes = append(missingHexes, parentHex)
				continue
			}
			parentResults[parentHex] = hp
		}
	} else {
		missingHexes = req.H3Hexes
	}

	if len(parentResults) == len(req.H3Hexes) {
		protocol.WriteJSON(w, http.StatusOK, map[string]any{
			"results":    collectHexValues(parentResults),
			"latency_ms": 0,
			"cached":     true,
		})
		return
	}

	filteredReq := req
	filteredReq.H3Hexes = missingHexes
	hexRecords, err := h.Vec.VectorizeBulk(&filteredReq)
	if err != nil {
		h.writeVectorizationError(w, err)
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
				cacheKey := storage.CacheKey(hexRecords[idx].Record)
				h.writeCache(r.Context(), cacheKey, results[idx])
			}
			h.Cache.IncrBy(r.Context(), "predictions_count", int64(len(uncachedIndices)))
			h.Cache.IncrByFloat(r.Context(), "latency_sum", latencyMs)
		}
	}

	for _, hp := range averageByParent(hexRecords, results) {
		h.writeCache(r.Context(), parentCacheKey(hp.Hex, &req), hp)
		parentResults[hp.Hex] = hp
	}

	protocol.WriteJSON(w, http.StatusOK, map[string]any{
		"results":    collectHexValues(parentResults),
		"latency_ms": latencyMs,
		"cached":     false,
	})
}

func collectHexValues(m map[string]protocol.HexPrediction) []protocol.HexPrediction {
	out := make([]protocol.HexPrediction, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
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
