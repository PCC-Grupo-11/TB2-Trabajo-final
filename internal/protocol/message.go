package protocol

import (
	"encoding/json"
	"io"
)

func ReadMessage(r io.Reader, msg any) error {
	return json.NewDecoder(r).Decode(msg)
}

func WriteMessage(w io.Writer, msg any) error {
	return json.NewEncoder(w).Encode(msg)
}
