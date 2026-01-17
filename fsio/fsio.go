// Copyright (c) $year Paweł Zaremba
// SPDX-License-Identifier: MIT

package fsio

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"iter"
	"log"

	"github.com/teghnet/x/internal"
)

func FSLoadJSON[T any](db fs.FS, name string) (T, error) {
	var v T
	f, err := fs.ReadFile(db, name)
	if err != nil {
		return v, err
	}
	err = json.Unmarshal(f, &v)
	if err != nil {
		return v, err
	}
	return v, nil
}

func FSIterateJSONs[T any](db fs.FS, name string) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		f, err := db.Open(name)
		if err != nil {
			log.Fatal(err)
		}
		defer internal.FatalClose(f)
		dec := json.NewDecoder(f)
		for dec.More() {
			var v T
			if !yield(v, dec.Decode(&v)) {
				return
			}
		}
	}
}

// FSUnmarshalJSON reads a JSON file from the filesystem
// and unmarshals it into the provided variable.
func FSUnmarshalJSON[T any](f fs.FS, path string, v T) error {
	data, err := fs.ReadFile(f, path)
	if err != nil {
		return fmt.Errorf("fsio.FSUnmarshalJSON: %w", err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("fsio.FSUnmarshalJSON: %w", err)
	}
	return nil
}

// FSGlob is a utility function that returns an iterator over files matching
// the given pattern in the provided filesystem.
func FSGlob(f fs.FS, pattern string) iter.Seq[string] {
	return func(yield func(string) bool) {
		matches, err := fs.Glob(f, pattern)
		if err != nil {
			panic(fmt.Errorf("fsio.FSGlob: %w", err))
		}
		for _, match := range matches {
			if !yield(match) {
				break
			}
		}
	}
}
