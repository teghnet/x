// Package json provides pure, I/O-free JSON decoding helpers.
//
// Decoders take bytes or readers and return values; they never touch the
// network or filesystem. This keeps the package trivial to unit-test and
// safe to import from any layer.
package json

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"iter"
)

// Unmarshal decodes JSON bytes into a freshly allocated value of type T.
func Unmarshal[T any](data []byte) (T, error) {
	return Decode[T](bytes.NewReader(data))
}

// UnmarshalStrict decodes JSON bytes into a value of type T, rejecting unknown
// fields.
func UnmarshalStrict[T any](data []byte) (T, error) {
	return DecodeStrict[T](bytes.NewReader(data))
}

// Decode decodes a single JSON value from r into a freshly allocated value of
// type T. Unknown fields are permitted; use DecodeStrict to reject them.
func Decode[T any](r io.Reader) (v T, err error) {
	if err := json.NewDecoder(r).Decode(&v); err != nil {
		return v, fmt.Errorf("parse/json: decode: %w", err)
	}
	return v, nil
}

// DecodeStrict behaves like Decode but returns an error if the input contains
// object keys that do not map to a field of T.
func DecodeStrict[T any](r io.Reader) (v T, err error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&v); err != nil {
		return v, fmt.Errorf("parse/json: decode strict: %w", err)
	}
	return v, nil
}

// Stream decodes a sequence of concatenated or whitespace-separated JSON values
// from r, invoking fn for each. Decoding stops at the first error from the
// decoder or from fn. io.EOF terminates the stream cleanly.
func Stream[T any](r io.Reader, fn func(T) error) error {
	for v, err := range Iter[T](r) {
		if err != nil {
			return err
		}
		if err := fn(v); err != nil {
			return fmt.Errorf("parse/json: stream: %w", err)
		}
	}
	return nil
}

// Iter returns an iterator over the same sequence of values as Stream, for
// callers that prefer range over a callback. A decode error is delivered as
// the second value of the final pair yielded; ranging over Iter stops there
// unless the loop body already broke out on its own.
func Iter[T any](r io.Reader) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		dec := json.NewDecoder(r)
		for {
			var v T
			err := dec.Decode(&v)
			if err == io.EOF {
				return
			}
			if err != nil {
				yield(v, fmt.Errorf("parse/json: iterate: %w", err))
				return
			}
			if !yield(v, nil) {
				return
			}
		}
	}
}

func IterArray[T any](f io.Reader) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		dec := json.NewDecoder(f)
		if err := dropToken(dec, '['); err != nil {
			return
		}
		for dec.More() {
			var v T
			if !yield(v, dec.Decode(&v)) {
				return
			}
		}
		if err := dropToken(dec, ']'); err != nil {
			return
		}
	}
}

func dropToken(dec *json.Decoder, r json.Delim) error {
	t, err := dec.Token()
	if err != nil {
		return fmt.Errorf("failed to read token: %w", err)
	}
	if t.(json.Delim) != r {
		return fmt.Errorf("expected '%s', got %v", r, t)
	}
	return nil
}
