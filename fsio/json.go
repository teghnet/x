// Copyright (c) 2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package fsio

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"iter"
	"log"

	"github.com/teghnet/x"
)

// JSON reads a JSON file and unmarshalls it into type T.
func JSON[T any](db fs.FS, name string) (T, error) {
	f, err := fs.ReadFile(db, name)
	if err != nil {
		return *new(T), err
	}
	var v T
	return v, json.Unmarshal(f, &v)
}

// JSONList returns an iterator over newline-delimited JSON objects (JSONL).
func JSONList[T any](db fs.FS, name string) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		f, err := db.Open(name)
		if err != nil {
			yield(*new(T), err)
			return
		}
		defer x.ClosePrint(f)
		decodeList(json.NewDecoder(f), yield)
	}
}

func ReadJSONList[T any](f io.Reader) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		decodeList(json.NewDecoder(f), yield)
	}
}

func decodeList[T any](dec *json.Decoder, yield func(T, error) bool) bool {
	for dec.More() {
		var v T
		if !yield(v, dec.Decode(&v)) {
			return false
		}
	}
	return true
}

// JSONArray returns an iterator over elements in a JSON array file.
func JSONArray[T any](db fs.FS, name string) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		f, err := db.Open(name)
		if err != nil {
			yield(*new(T), err)
			return
		}
		defer x.ClosePrint(f)
		decodeArray(json.NewDecoder(f), yield)
	}
}

func ReadJSONArray[T any](f io.Reader) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		decodeArray(json.NewDecoder(f), yield)
	}
}

func decodeArray[T any](dec *json.Decoder, yield func(T, error) bool) {
	if err := dropToken(dec, '['); err != nil {
		log.Printf("failed to drop leading array token: %v", err)
		return
	}
	if !decodeList(dec, yield) {
		return
	}
	if err := dropToken(dec, ']'); err != nil {
		log.Printf("failed to drop trailing array token: %v", err)
		return
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
