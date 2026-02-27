// Copyright (c) 2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package fsio

import (
	"encoding/json"
	"io/fs"
	"iter"

	internal2 "github.com/teghnet/x/internal"
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
		defer internal2.ClosePrint(f)
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
		defer internal2.ClosePrint(f)
		dec := json.NewDecoder(f)
		if err = internal2.DropToken(dec, '['); err != nil {
			yield(*new(T), err)
			return
		}
		for dec.More() {
			var v T
			if !yield(v, dec.Decode(&v)) {
				return
			}
		}
		if err = internal2.DropToken(dec, ']'); err != nil {
			yield(*new(T), err)
			return
		}
	}
}
