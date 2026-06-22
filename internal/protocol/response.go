package protocol

type BulkPredictResponse struct {
	Results   []HexPrediction `json:"results"`
	LatencyMs float64         `json:"latency_ms"`
}

type HexPrediction struct {
	Hex        string  `json:"hex"`
	Class      int     `json:"class"`
	Confidence float32 `json:"confidence"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type MetricsResponse struct {
	AvgLatencyMs    float64 `json:"avg_latency_ms"`
	NodeCount       int     `json:"node_count"`
	PredictionCount int64   `json:"prediction_count"`
	CacheHitRate    float64 `json:"cache_hit_rate"`
}
