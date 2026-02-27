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
	"strings"
)

func TrimXML(r io.Reader, w io.Writer) (err error) {
	prevElemType := ""
	var xpath []string
	for t := range tokens(xml.NewDecoder(r)) {
		switch e := t.(type) {
		case xml.ProcInst:
			_, err = fmt.Fprintf(w, "<?%s %s?>\n", e.Target, e.Inst)
		case xml.Directive:
			_, err = fmt.Fprintf(w, "!%s\n", e)
		case xml.Comment:
			_, err = fmt.Fprintf(w, "<!--%s-->\n", e)
		case xml.StartElement:
			err = fPrint(w, xpath, prevElemType == "xml.StartElement" || prevElemType == "xml.EndElement", e, se)
			xpath = append(xpath, e.Name.Local)
		case xml.EndElement:
			xpath = xpath[:len(xpath)-1]
			err = fPrint(w, xpath, prevElemType == "xml.StartElement" || prevElemType == "xml.EndElement", e, ee)
		case xml.CharData:
			err = xml.EscapeText(w, e)
		default:
			log.Printf("%T", e)
		}
		if err != nil {
			return err
		}
		prevElemType = fmt.Sprintf("%T", t)
	}
	return nil
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

func fPrint[T xml.StartElement | xml.EndElement](w io.Writer, xp []string, prependNewLine bool, e T, f func(T, []string) []any) (err error) {
	if prependNewLine {
		if _, err = fmt.Fprint(w, "\n", strings.Repeat("\t", len(xp))); err != nil {
			return err
		}
	}
	_, err = fmt.Fprint(w, f(e, xp)...)
	return err
}

func se(e xml.StartElement, xp []string) []any {
	if e.Name.Space != "" {
		e.Name.Space += ":"
	}
	ret := []any{"<", e.Name.Space, e.Name.Local}
	for _, a := range e.Attr {
		if a.Name.Space != "" {
			a.Name.Space += ":"
		}
		ret = append(ret, "\n\t", strings.Repeat("\t", len(xp)), a.Name.Space, a.Name.Local, `="`, html.EscapeString(a.Value), `"`)
	}
	return append(ret, ">")
}

func ee(e xml.EndElement, xp []string) []any {
	if e.Name.Space != "" {
		e.Name.Space += ":"
	}
	return []any{"</", e.Name.Space, e.Name.Local, ">"}
}

// tokens
// TODO:
// [ ] ensure proper handling of namespaces
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
