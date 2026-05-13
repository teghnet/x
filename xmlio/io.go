// Copyright (c) 2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package xmlio

import (
	"encoding/xml"
	"io"
	"iter"
)

func ReadXML[T any](r io.Reader) (T, error) {
	var v T
	return v, xml.NewDecoder(r).Decode(&v)
}

// Deprecated: use List.
func ReadXMLs[T any](r io.Reader, elementName string) iter.Seq2[T, error] {
	// TODO: improve path handling (so that we can make sure the right element is read)
	return func(yield func(T, error) bool) {
		dec := xml.NewDecoder(r)
		for t := range Tokens(dec, false) {
			switch el := t.(type) {
			case xml.StartElement:
				var v T
				if el.Name.Local != elementName {
					continue
				}
				if !yield(v, dec.DecodeElement(&v, &el)) {
					return
				}
			}
		}
	}
}

type Result[T any] struct {
	Val T
	Err error
}

func List[T any](r io.Reader, elementName string) iter.Seq[Result[T]] {
	return func(yield func(Result[T]) bool) {
		dec := xml.NewDecoder(r)
		for t := range Tokens(dec, false) {
			switch el := t.(type) {
			case xml.StartElement:
				var v T
				if el.Name.Local != elementName {
					continue
				}
				r2 := Result[T]{v, dec.DecodeElement(&v, &el)}
				if !yield(r2) {
					return
				}
			}
		}
	}
}
