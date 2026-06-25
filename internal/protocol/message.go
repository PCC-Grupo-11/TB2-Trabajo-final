package protocol

import (
	"encoding/json"
	"io"
	"net/http"
)

func ReadMessage(r io.Reader, msg any) error {
	return json.NewDecoder(r).Decode(msg)
}

func WriteMessage(w io.Writer, msg any) error {
	return json.NewEncoder(w).Encode(msg)
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
