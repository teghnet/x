// Copyright (c) $year Paweł Zaremba
// SPDX-License-Identifier: MIT

package fsio

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"iter"
	"log"
	"os"

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
		defer internal.CloseFatal(f)
		dec := json.NewDecoder(f)
		for dec.More() {
			var v T
			if !yield(v, dec.Decode(&v)) {
				return
			}
		}
	}
}
func FSJSONArray[T any](db fs.FS, name string) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		f, err := db.Open(name)
		if err != nil {
			log.Fatal(err)
		}
		defer internal.CloseFatal(f)
		dec := json.NewDecoder(f)
		if err = dropToken(dec, '['); err != nil {
			log.Printf("failed to drop leading array token: %v", err)
			return
		}
		for dec.More() {
			var v T
			if !yield(v, dec.Decode(&v)) {
				return
			}
		}
		if err = dropToken(dec, ']'); err != nil {
			log.Printf("failed to drop trailing array token: %v", err)
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

func DynamicWriter(name string, append bool) (io.WriteCloser, error) {
	if name == "-" || name == "stdout" {
		return os.Stdout, nil
	}
	if name == "=" || name == "stderr" {
		return os.Stderr, nil
	}
	flag := os.O_WRONLY | os.O_CREATE
	if append {
		flag |= os.O_APPEND
	} else {
		flag |= os.O_TRUNC
	}
	f, err := os.OpenFile(name, flag, 0600)
	if err != nil {
		return nil, err
	}
	return f, nil
}
