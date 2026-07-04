package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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
	start := time.Now()

	var req protocol.BulkPredictRequest
	if err := protocol.ReadMessage(r.Body, &req); err != nil {
		protocol.WriteJSON(w, http.StatusBadRequest, protocol.ErrorResponse{Error: "invalid request body"})
		return
	}

	totalHexes := 0
	for _, hexes := range req.Hexes {
		totalHexes += len(hexes)
	}

	parentResults := make(map[string]protocol.HexPrediction, totalHexes)
	var missingHexes []string
	boroughs := make(map[string]string, totalHexes)

	if h.Cache != nil {
		allHexes := make([]string, 0, totalHexes)
		for borough, hexes := range req.Hexes {
			for _, hex := range hexes {
				boroughs[hex] = borough
				allHexes = append(allHexes, hex)
			}
		}
		parentKeys := make([]string, len(allHexes))
		keyToHex := make(map[string]string, len(allHexes))
		for i, hex := range allHexes {
			parentKeys[i] = parentCacheKey(hex, boroughs[hex], &req)
			keyToHex[parentKeys[i]] = hex
		}
		cached, err := h.Cache.GetPredictionsPipeline(r.Context(), parentKeys)
		if err != nil {
			logger.Warn("parent cache pipeline failed", "error", err)
		}
		for key, hex := range keyToHex {
			data, ok := cached[key]
			if !ok {
				missingHexes = append(missingHexes, hex)
				continue
			}
			var hp protocol.HexPrediction
			if err := json.Unmarshal(data, &hp); err != nil {
				logger.Warn("parent cache data corrupted", "parent_hex", hex, "error", err)
				missingHexes = append(missingHexes, hex)
				continue
			}
			parentResults[hex] = hp
		}
	} else {
		for borough, hexes := range req.Hexes {
			for _, hex := range hexes {
				missingHexes = append(missingHexes, hex)
				boroughs[hex] = borough
			}
		}
	}

	parentHits := int64(len(parentResults)) * 7

	if len(parentResults) == totalHexes {
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

	if h.Cache != nil {
		childHashes := make([]string, n)
		for i, hr := range hexRecords {
			childHashes[i] = storage.CacheKey(hr.Record)
		}
		cachedMap, err := h.Cache.GetPredictionsPipeline(r.Context(), childHashes)
		if err != nil {
			logger.Warn("child cache pipeline failed", "error", err)
		}
		for i, hash := range childHashes {
			if data, ok := cachedMap[hash]; ok {
				var cr protocol.PredictionResult
				if err := json.Unmarshal(data, &cr); err == nil {
					results[i] = cr
					continue
				}
			}
			uncachedRecords = append(uncachedRecords, hexRecords[i].Record)
			uncachedIndices = append(uncachedIndices, i)
		}
	} else {
		for i, hr := range hexRecords {
			uncachedRecords = append(uncachedRecords, hr.Record)
			uncachedIndices = append(uncachedIndices, i)
		}
	}

	var latencyMs float64
	if len(uncachedRecords) > 0 {
		inferResp, err := h.LB.Predict(&protocol.InferenceRequest{Records: uncachedRecords})
		if err != nil {
			logger.Error("bulk inference failed", "error", err)
			protocol.WriteJSON(w, http.StatusServiceUnavailable, protocol.ErrorResponse{Error: "inference failed"})
			return
		}

		for j, idx := range uncachedIndices {
			results[idx] = inferResp.Predictions[j]
		}
	}

	parentAverages := averageByParent(hexRecords, results)
	for _, hp := range parentAverages {
		parentResults[hp.Hex] = hp
	}

	latencyMs = float64(time.Since(start).Nanoseconds()) / 1e6

	if h.Cache != nil {
		// Batch-SET child predictions
		childEntries := make(map[string][]byte, len(uncachedIndices))
		for _, idx := range uncachedIndices {
			data, err := json.Marshal(results[idx])
			if err == nil {
				childEntries[storage.CacheKey(hexRecords[idx].Record)] = data
			}
		}
		if err := h.Cache.SetPredictionsPipeline(r.Context(), childEntries); err != nil {
			logger.Warn("child cache pipeline write failed", "error", err)
		}

		// Batch-SET parent predictions
		parentEntries := make(map[string][]byte, len(parentAverages))
		for _, hp := range parentAverages {
			data, err := json.Marshal(hp)
			if err == nil {
				parentEntries[parentCacheKey(hp.Hex, boroughs[hp.Hex], &req)] = data
			}
		}
		if err := h.Cache.SetPredictionsPipeline(r.Context(), parentEntries); err != nil {
			logger.Warn("parent cache pipeline write failed", "error", err)
		}

		// Batch metrics
		counters := map[string]int64{
			"predictions_count": int64(len(uncachedIndices)),
			"cache_misses":      int64(len(uncachedIndices)),
		}
		floats := map[string]float64{
			"latency_sum": latencyMs,
		}
		if err := h.Cache.IncrMetricsPipeline(r.Context(), counters, floats); err != nil {
			logger.Warn("metrics pipeline write failed", "error", err)
		}
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
