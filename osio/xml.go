// Copyright (c) 2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package osio

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"iter"
	"log"
	"log/slog"
	"regexp"
	"strings"
)

func TrimXML(r io.Reader, w io.Writer, asHTML bool) (err error) {
	prevElemType := ""
	var xpath []string
	dec := xml.NewDecoder(r)
	if asHTML {
		dec.Strict = false
		dec.AutoClose = xml.HTMLAutoClose
		dec.Entity = xml.HTMLEntity
	}
	for t := range tokens(dec) {
		switch el := t.(type) {
		case xml.ProcInst:
			_, err = fmt.Fprintf(w, "<?%s %s?>\n", el.Target, el.Inst)
		case xml.Directive:
			_, err = fmt.Fprintf(w, "!%s\n", el)
		case xml.Comment:
			_, err = fmt.Fprintf(w, "<!--%s-->\n", el)
		case xml.StartElement:
			// render start element
			if prevElemType == "xml.StartElement" || prevElemType == "xml.EndElement" {
				if _, err = fmt.Fprint(w, "\n", strings.Repeat("\t", len(xpath))); err != nil {
					break
				}
			}
			_, err = fmt.Fprint(w, startElement(el, len(xpath), strings.TrimSpace, html.EscapeString, NormalizeSpaces)...)

			// increase nesting level
			xpath = append(xpath, el.Name.Local)
		case xml.EndElement:
			// decrease nesting level
			xpath = xpath[:len(xpath)-1]
			// render end element
			if prevElemType == "xml.StartElement" || prevElemType == "xml.EndElement" {
				if _, err = fmt.Fprint(w, "\n", strings.Repeat("\t", len(xpath))); err != nil {
					break
				}
			}
			if el.Name.Space != "" {
				el.Name.Space += ":"
			}
			_, err = fmt.Fprint(w, []any{"</", el.Name.Space, el.Name.Local, ">"}...)
		case xml.CharData:
			err = xml.EscapeText(w, bytes.TrimSpace(reSpaces.ReplaceAll(el, []byte(" "))))
		default:
			log.Printf("%T", el)
		}
		if err != nil {
			return err
		}
		prevElemType = fmt.Sprintf("%T", t)
	}
	return nil
}

var reSpaces = regexp.MustCompile(`\s+`)

func NormalizeSpaces(input string) string {
	return strings.TrimSpace(reSpaces.ReplaceAllString(input, " "))
}

func XMLDicts(r io.Reader) iter.Seq2[string, string] {
	var xpath []string
	return func(yield func(string, string) bool) {
		for t := range tokens(xml.NewDecoder(r)) {
			switch e := t.(type) {
			case xml.StartElement:
				xpath = append(xpath, e.Name.Local)
			case xml.EndElement:
				xpath = xpath[:len(xpath)-1]
			case xml.CharData:
				if !yield(strings.Join(xpath, " > "), string(e)) {
					return
				}
			default:
				log.Printf("%T", e)
			}
		}
	}
}

func startElement(el xml.StartElement, level int, fns ...func(string) string) []any {
	if el.Name.Space != "" {
		el.Name.Space += ":"
	}
	ret := []any{"<", el.Name.Space, el.Name.Local}
	for _, a := range el.Attr {
		if a.Name.Space != "" {
			a.Name.Space += ":"
		}
		for _, fn := range fns {
			a.Value = fn(a.Value)
		}
		ret = append(ret, "\n", strings.Repeat("\t", level), a.Name.Space, a.Name.Local, `="`, a.Value, `"`)
	}
	return append(ret, ">")
}

// tokens
// TODO: ensure proper handling of namespaces
func tokens(dec *xml.Decoder) iter.Seq[xml.Token] {
	eof := new(io.EOF)
	var cd xml.CharData
	return func(yield func(xml.Token) bool) {
		for {
			t, err := dec.Token()
			if err != nil {
				if errors.As(err, eof) { // err=EOF
					if len(cd) != 0 { // there is leftover data
						// yield (no use to check for return value)
						yield(cd)
					}
					return
				}
				slog.Error("errored", "err", err)
				continue
			}

			if c, ok := t.(xml.CharData); ok {
				cd = append(cd, bytes.TrimSpace(c)...)
				continue
			}

			if len(cd) != 0 {
				// yield any captured data
				if !yield(cd) {
					return
				}
				// reset data buffer
				cd = []byte{}
			}

			// yield current token
			if !yield(t) {
				return
			}
		}
	}
}
