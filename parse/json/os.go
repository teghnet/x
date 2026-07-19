package json

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/teghnet/x"
)

func LogStore[T any](path string, v T) {
	x.WarnErr(Store[T](path, v))
}
func Load[T any](path string) (v T, err error) {
	r, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			x.Notice(err)
			return v, nil
		}
		return v, fmt.Errorf("parse/json: load: %w", err)
	}
	defer x.Close(r)
	return Decode[T](r)
}
func Store[T any](path string, v T) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("parse/json: store: %w", err)
	}
	defer x.Close(file)
	return WritePretty(file, v)
}

func Write[T any](w io.Writer, v T) error {
	return json.NewEncoder(w).Encode(&v)
}

func WritePretty[T any](w io.Writer, v T) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
