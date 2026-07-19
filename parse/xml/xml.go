// Package xml provides pure, I/O-free XML decoding helpers.
//
// Like the sibling json package, decoders take bytes or readers and return
// values. No network or filesystem access occurs here.
package xml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"iter"
)

// Decode decodes a single XML document from r into a freshly allocated value of
// type T.
func Decode[T any](r io.Reader) (T, error) {
	var v T
	if err := xml.NewDecoder(r).Decode(&v); err != nil {
		return v, fmt.Errorf("parse/xml: decode: %w", err)
	}
	return v, nil
}

// Unmarshal decodes XML bytes into a freshly allocated value of type T.
func Unmarshal[T any](data []byte) (T, error) {
	return Decode[T](bytes.NewReader(data))
}

// Stream walks r and decodes every element whose local name equals element into
// a value of type T, invoking fn for each. It is suited to large documents with
// many repeated records. Decoding stops at the first error from fn.
func Stream[T any](r io.Reader, element string, fn func(T) error) error {
	for v, err := range Iter[T](r, element) {
		if err != nil {
			return err
		}
		if err := fn(v); err != nil {
			return err
		}
	}
	return nil
}

// Iter returns an iterator over the same sequence of values as Stream, for
// callers that prefer range over a callback. A decode error is delivered as
// the second value of the final pair yielded; ranging over Iter stops there
// unless the loop body already broke out on its own.
func Iter[T any](r io.Reader, element string) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		dec := xml.NewDecoder(r)
		for {
			tok, err := dec.Token()
			if err == io.EOF {
				return
			}
			if err != nil {
				var zero T
				yield(zero, fmt.Errorf("parse/xml: stream token: %w", err))
				return
			}
			start, ok := tok.(xml.StartElement)
			if !ok || start.Name.Local != element {
				continue
			}
			var v T
			if err := dec.DecodeElement(&v, &start); err != nil {
				yield(v, fmt.Errorf("parse/xml: stream decode %q: %w", element, err))
				return
			}
			if !yield(v, nil) {
				return
			}
		}
	}
}
