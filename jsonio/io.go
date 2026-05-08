// Copyright (c) 2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package jsonio

import (
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"log"
)

type Result[T any] struct {
	Val T
	Err error
}

func Read[T any](r io.Reader) (T, error) {
	var v T
	return v, json.NewDecoder(r).Decode(&v)
}

// List returns an iterator over newline-delimited JSON objects (JSONL)
// from the provided io.Reader.
func List[T any](f io.Reader) iter.Seq[Result[T]] {
	return func(yield func(Result[T]) bool) {
		dec := json.NewDecoder(f)
		for dec.More() {
			var v T
			if !yield(Result[T]{Val: v, Err: dec.Decode(&v)}) {
				return
			}
		}
	}
}

func Array[T any](f io.Reader) iter.Seq[Result[T]] {
	return func(yield func(Result[T]) bool) {
		dec := json.NewDecoder(f)
		if err := dropToken(dec, '['); err != nil {
			_ = yield(Result[T]{Err: err})
			return
		}
		for dec.More() {
			var v T
			if !yield(Result[T]{Val: v, Err: dec.Decode(&v)}) {
				return
			}
		}
		if err := dropToken(dec, ']'); err != nil {
			// it's debatable if we should return this error
			_ = yield(Result[T]{Err: err})
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
		return fmt.Errorf("expected '%s' at the end, got %v", r, t)
	}
	return nil
}

func Write[T any](w io.Writer, v T) error {
	return json.NewEncoder(w).Encode(&v)
}

func WritePretty[T any](w io.Writer, v T) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// Deprecated functions ============================================================================

// ReadJSON
// Deprecated: use Read.
func ReadJSON[T any](r io.Reader) (T, error) {
	var v T
	return v, json.NewDecoder(r).Decode(&v)
}

// ReadJSONList returns an iterator over newline-delimited JSON objects (JSONL)
// from the provided io.Reader.
// Deprecated: use List.
func ReadJSONList[T any](f io.Reader) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		dec := json.NewDecoder(f)
		for dec.More() {
			var v T
			if !yield(v, dec.Decode(&v)) {
				return
			}
		}
	}
}

// ReadJSONArray returns an iterator over an array
// Deprecated: use Array.
func ReadJSONArray[T any](f io.Reader) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		dec := json.NewDecoder(f)
		if err := dropToken(dec, '['); err != nil {
			log.Printf("failed to drop leading array token: %v", err)
			return
		}
		for dec.More() {
			var v T
			if !yield(v, dec.Decode(&v)) {
				return
			}
		}
		if err := dropToken(dec, ']'); err != nil {
			log.Printf("failed to drop trailing array token: %v", err)
			return
		}
	}
}

// WriteJSON
// Deprecated: use Write.
func WriteJSON[T any](w io.Writer, v T) error {
	return json.NewEncoder(w).Encode(&v)
}

// WritePrettyJSON
// Deprecated: use WritePretty.
func WritePrettyJSON[T any](w io.Writer, v T) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
