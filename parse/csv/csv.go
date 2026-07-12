// Package csv provides pure, I/O-free CSV decoding helpers.
//
// Like the sibling json and xml packages, decoders take bytes or readers and
// return values; they never touch the network or filesystem. Unlike
// encoding/csv, which only produces [][]string, this package maps the header
// row onto struct fields via a `csv` tag (falling back to a case-insensitive
// field name match), so callers can decode records straight into typed
// values. Columns with a blank header can be targeted by position instead,
// via a tag of the form `csv:",N"`; see columns.
package csv

import (
	"bytes"
	"encoding"
	stdcsv "encoding/csv"
	"fmt"
	"io"
	"iter"
	"reflect"
	"strconv"
	"strings"
)

// Decode reads a CSV document from r, treating the first record as a header
// of column names, and returns one T per remaining record.
func Decode[T any](r io.Reader) ([]T, error) {
	var out []T
	if err := Stream(r, func(v T) error {
		out = append(out, v)
		return nil
	}); err != nil {
		return nil, err
	}
	return out, nil
}

// Unmarshal decodes CSV bytes into a slice of T; see Decode.
func Unmarshal[T any](data []byte) ([]T, error) {
	return Decode[T](bytes.NewReader(data))
}

// Stream reads r as a CSV document, one header row of column names followed
// by records, decoding each record into a value of type T and invoking fn.
// Decoding stops at the first error from the reader or from fn. An empty
// input (no header row) is not an error; fn is simply never called.
func Stream[T any](r io.Reader, fn func(T) error) error {
	for v, err := range All[T](r) {
		if err != nil {
			return err
		}
		if err := fn(v); err != nil {
			return err
		}
	}
	return nil
}

// All returns an iterator over the same sequence of values as Stream, for
// callers that prefer range over a callback. A read or decode error is
// delivered as the second value of the final pair yielded; ranging over All
// stops there unless the loop body already broke out on its own.
func All[T any](r io.Reader) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		var zero T

		cr := stdcsv.NewReader(r)
		header, err := cr.Read()
		if err == io.EOF {
			return
		}
		if err != nil {
			yield(zero, fmt.Errorf("parse/csv: read header: %w", err))
			return
		}

		cols, err := columns[T](header)
		if err != nil {
			yield(zero, fmt.Errorf("parse/csv: %w", err))
			return
		}

		for {
			record, err := cr.Read()
			if err == io.EOF {
				return
			}
			if err != nil {
				yield(zero, fmt.Errorf("parse/csv: read record: %w", err))
				return
			}

			var v T
			rv := reflect.ValueOf(&v).Elem()
			var fieldErr error
			for i, c := range cols {
				if c.fieldIndex < 0 || i >= len(record) {
					continue
				}
				if err := setField(rv.Field(c.fieldIndex), record[i]); err != nil {
					fieldErr = fmt.Errorf("parse/csv: column %q: %w", c.name, err)
					break
				}
			}
			if fieldErr != nil {
				yield(v, fieldErr)
				return
			}
			if !yield(v, nil) {
				return
			}
		}
	}
}

// column pairs a header entry with the struct field of T that receives it.
type column struct {
	name       string
	fieldIndex int // -1 if the header has no matching field
}

// columns maps each header entry to a struct field of T by `csv` tag or,
// failing that, a case-insensitive field name match. A tag of "-" excludes
// the field from matching. Real-world exports sometimes leave a column's
// header blank (e.g. a leading ID column); such columns can't be matched by
// name, so a tag of the form `csv:",N"` (empty name, 0-based column index N)
// targets one by position instead.
func columns[T any](header []string) ([]column, error) {
	t := reflect.TypeFor[T]()
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("type %s is not a struct", t)
	}

	byName := make(map[string]int, t.NumField())
	byPos := make(map[int]int)
	for i := range t.NumField() {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		name := f.Name
		if tag, ok := f.Tag.Lookup("csv"); ok {
			if tag == "-" {
				continue
			}
			n, opt, _ := strings.Cut(tag, ",")
			switch {
			case n != "":
				name = n
			case opt != "":
				pos, err := strconv.Atoi(opt)
				if err != nil {
					return nil, fmt.Errorf("field %s: csv tag %q: position must be an integer", f.Name, tag)
				}
				byPos[pos] = i
				continue
			}
		}
		byName[strings.ToLower(name)] = i
	}

	cols := make([]column, len(header))
	for i, h := range header {
		idx := -1
		if name := strings.TrimSpace(h); name != "" {
			if fi, ok := byName[strings.ToLower(name)]; ok {
				idx = fi
			}
		} else if fi, ok := byPos[i]; ok {
			idx = fi
		}
		cols[i] = column{name: h, fieldIndex: idx}
	}
	return cols, nil
}

// setField parses s into fv. Fields whose type implements
// encoding.TextUnmarshaler (e.g. decimal.Decimal, time.Time, or a
// locale-specific wrapper around either) delegate to it; the type decides
// how to handle an empty cell. Otherwise fv is set according to its kind, and
// an empty string leaves numeric and boolean fields at their zero value
// rather than erroring.
func setField(fv reflect.Value, s string) error {
	if fv.CanAddr() {
		if u, ok := fv.Addr().Interface().(encoding.TextUnmarshaler); ok {
			return u.UnmarshalText([]byte(s))
		}
	}

	switch fv.Kind() {
	case reflect.String:
		fv.SetString(s)
	case reflect.Bool:
		if s == "" {
			return nil
		}
		b, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		fv.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if s == "" {
			return nil
		}
		n, err := strconv.ParseInt(s, 10, fv.Type().Bits())
		if err != nil {
			return err
		}
		fv.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if s == "" {
			return nil
		}
		n, err := strconv.ParseUint(s, 10, fv.Type().Bits())
		if err != nil {
			return err
		}
		fv.SetUint(n)
	case reflect.Float32, reflect.Float64:
		if s == "" {
			return nil
		}
		n, err := strconv.ParseFloat(s, fv.Type().Bits())
		if err != nil {
			return err
		}
		fv.SetFloat(n)
	default:
		return fmt.Errorf("unsupported field kind %s", fv.Kind())
	}
	return nil
}
