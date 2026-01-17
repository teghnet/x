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

func FSJSONList[T any](db fs.FS, name string) iter.Seq2[T, error] {
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
