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
		dec := json.NewDecoder(f)
		for dec.More() {
			var v T
			if !yield(v, dec.Decode(&v)) {
				return
			}
		}
	}
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
		dec := json.NewDecoder(f)
		if err = DropToken(dec, '['); err != nil {
			yield(*new(T), err)
			return
		}
		for dec.More() {
			var v T
			if !yield(v, dec.Decode(&v)) {
				return
			}
		}
		if err = DropToken(dec, ']'); err != nil {
			yield(*new(T), err)
			return
		}
	}
}

func DropToken(dec *json.Decoder, r json.Delim) error {
	t, err := dec.Token()
	if err != nil {
		return fmt.Errorf("failed to read token: %w", err)
	}
	if t.(json.Delim) != r {
		return fmt.Errorf("expected '%s' at the end, got %v", r, t)
	}
	return nil
}

func DecArray[T any](f io.Reader) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		dec := json.NewDecoder(f)
		if err := DropToken(dec, '['); err != nil {
			log.Printf("failed to drop leading array token: %v", err)
			return
		}
		for dec.More() {
			var v T
			if !yield(v, dec.Decode(&v)) {
				return
			}
		}
		if err := DropToken(dec, ']'); err != nil {
			log.Printf("failed to drop trailing array token: %v", err)
			return
		}
	}
}

func DecList[T any](f io.Reader) iter.Seq2[T, error] {
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
