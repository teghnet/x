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
