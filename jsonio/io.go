package jsonio

import (
	"encoding/json"
	"io"
)

func ReadJSON[T any](r io.Reader) (T, error) {
	var v T
	return v, json.NewDecoder(r).Decode(&v)
}
func WriteJSON[T any](w io.Writer, v T) error {
	return json.NewEncoder(w).Encode(&v)
}
func WritePrettyJSON[T any](w io.Writer, v T) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
