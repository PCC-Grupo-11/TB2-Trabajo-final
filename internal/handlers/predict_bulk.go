package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/logger"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/protocol"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/storage"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/vectorizer"
)

func parentCacheKey(parentHex, borough string, req *protocol.BulkPredictRequest) string {
	raw := fmt.Sprintf(
		"%s|%d|%s|%s|%s|%s|%s",
		parentHex,
		req.Timestamp,
		req.Agency,
		req.ComplaintType,
		req.Descriptor,
		req.LocationType,
		borough,
	)
	return storage.HashString(raw)
}

func (h *Handler) PredictBulk(w http.ResponseWriter, r *http.Request) {
	var req protocol.BulkPredictRequest
	if err := protocol.ReadMessage(r.Body, &req); err != nil {
		protocol.WriteJSON(w, http.StatusBadRequest, protocol.ErrorResponse{Error: "invalid request body"})
		return
	}

	parentResults := make(map[string]protocol.HexPrediction, len(req.Hexes))
	var missingHexes []string
	boroughs := make(map[string]string, len(req.Hexes))

	if h.Cache != nil {
		for _, hex := range req.Hexes {
			key := parentCacheKey(hex.Hex, hex.Borough, &req)
			boroughs[hex.Hex] = hex.Borough
			data, err := h.Cache.GetPrediction(r.Context(), key)
			if err != nil {
				missingHexes = append(missingHexes, hex.Hex)
				continue
			}
			var hp protocol.HexPrediction
			if err := json.Unmarshal(data, &hp); err != nil {
				logger.Warn("parent cache data corrupted", "parent_hex", hex.Hex, "error", err)
				missingHexes = append(missingHexes, hex.Hex)
				continue
			}
			parentResults[hex.Hex] = hp
		}
	} else {
		for _, hex := range req.Hexes {
			missingHexes = append(missingHexes, hex.Hex)
			boroughs[hex.Hex] = hex.Borough
		}
	}

	// 7 = avg children per hex at res+1; if VectorizeBulk resolution changes, update this
	parentHits := int64(len(parentResults)) * 7

	if len(parentResults) == len(req.Hexes) {
		if h.Cache != nil {
			h.Cache.IncrBy(r.Context(), "cache_hits", parentHits)
		}
		writeBulkPredictResponse(w, collectHexValues(parentResults), 0, true)
		return
	}

	if h.Cache != nil {
		h.Cache.IncrBy(r.Context(), "cache_hits", parentHits)
	}

	hexRecords, err := h.Vec.VectorizeBulk(missingHexes, boroughs, &protocol.BulkPredictRequest{
		Timestamp:     req.Timestamp,
		Agency:        req.Agency,
		ComplaintType: req.ComplaintType,
		Descriptor:    req.Descriptor,
		LocationType:  req.LocationType,
	})
	if err != nil {
		h.writeVectorizationError(w, err)
		return
	}

	n := len(hexRecords)
	results := make([]protocol.PredictionResult, n)

	var uncachedRecords []protocol.PredictRecord
	uncachedIndices := make([]int, 0, n)

	for i, hr := range hexRecords {
		cacheKey := storage.CacheKey(hr.Record)
		if cached := h.checkCache(r.Context(), cacheKey); cached != nil {
			results[i] = *cached
			continue
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
			h.Cache.IncrBy(r.Context(), "cache_misses", int64(len(uncachedIndices)))
		}
	}

	for _, hp := range averageByParent(hexRecords, results) {
		h.writeCache(r.Context(), parentCacheKey(hp.Hex, boroughs[hp.Hex], &req), hp)
		parentResults[hp.Hex] = hp
	}

	writeBulkPredictResponse(w, collectHexValues(parentResults), latencyMs, false)
}

func writeBulkPredictResponse(w http.ResponseWriter, results []protocol.HexPrediction, latencyMs float64, cached bool) {
	protocol.WriteJSON(w, http.StatusOK, protocol.BulkPredictResponse{
		Results:   results,
		LatencyMs: latencyMs,
		Cached:    cached,
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
