// Copyright (c) 2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package jsonio

import (
	"io/fs"
	"iter"

	"github.com/teghnet/x"
	"github.com/teghnet/x/parse/json"
)

// ReadFS reads a JSON file and unmarshalls it into type T.
func ReadFS[T any](fsfs fs.FS, name string) (T, error) {
	f, err := fsfs.Open(name)
	if err != nil {
		return *new(T), err
	}
	return Read[T](f)
}

func JSON[T any](fsfs fs.FS, name string) (T, error) {
	f, err := fsfs.Open(name)
	if err != nil {
		return *new(T), err
	}
	return Read[T](f)
}

// JSONList returns an iterator over newline-delimited JSON objects (JSONL).
func JSONList[T any](fsfs fs.FS, name string) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		f, err := fsfs.Open(name)
		if err != nil {
			_ = yield(*new(T), err)
			return
		}
		defer x.Close(f)
		ReadJSONList[T](f)(yield)
	}
}

// JSONArray returns an iterator over elements in a JSON array file.
func JSONArray[T any](fsfs fs.FS, name string) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		f, err := fsfs.Open(name)
		if err != nil {
			yield(*new(T), err)
			return
		}
		defer x.Close(f)
		json.IterArray[T](f)(yield)
	}
}
