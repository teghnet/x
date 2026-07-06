// Package xml provides pure, I/O-free XML decoding helpers.
//
// Like the sibling json package, decoders take bytes or readers and return
// values. No network or filesystem access occurs here.
package xml

import (
	"bytes"
	stdxml "encoding/xml"
	"fmt"
	"io"
)

// Decode decodes a single XML document from r into a freshly allocated value of
// type T.
func Decode[T any](r io.Reader) (T, error) {
	var v T
	if err := stdxml.NewDecoder(r).Decode(&v); err != nil {
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
	dec := stdxml.NewDecoder(r)
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("parse/xml: stream token: %w", err)
		}
		start, ok := tok.(stdxml.StartElement)
		if !ok || start.Name.Local != element {
			continue
		}
		var v T
		if err := dec.DecodeElement(&v, &start); err != nil {
			return fmt.Errorf("parse/xml: stream decode %q: %w", element, err)
		}
		if err := fn(v); err != nil {
			return err
		}
	}
}
