// Copyright (c) 2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"log"
)

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
