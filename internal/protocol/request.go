package protocol

type PredictRequest struct {
	Timestamp     int64   `json:"ts"`
	Latitude      float64 `json:"lat"`
	Longitude     float64 `json:"lon"`
	Agency        string  `json:"agency"`
	ComplaintType string  `json:"complaint_type"`
	Descriptor    string  `json:"descriptor"`
	LocationType  string  `json:"location_type"`
	Borough       string  `json:"borough"`
}

type BulkPredictRequest struct {
	Timestamp     int64    `json:"ts"`
	Agency        string   `json:"agency"`
	ComplaintType string   `json:"complaint_type"`
	Descriptor    string   `json:"descriptor"`
	LocationType  string   `json:"location_type"`
	Borough       string   `json:"borough"`
	H3Hexes       []string `json:"h3_hexes"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
