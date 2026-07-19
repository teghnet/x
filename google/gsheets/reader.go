package gsheets

import (
	"context"
	"fmt"
	"iter"
	"reflect"
	"strconv"
	"strings"

	"google.golang.org/api/sheets/v4"
)

// Reader provides functionality to read data from Google Sheets.
type Reader struct {
	service       *sheets.Service
	spreadsheetID string
}

// NewReader creates a new Reader for a given spreadsheet.
func NewReader(client *sheets.Service, spreadsheetID string) *Reader {
	return &Reader{
		service:       client,
		spreadsheetID: spreadsheetID,
	}
}

// ReadRows streams rows from sheetRange, decoding each into a T.
// The first row is treated as a header row and maps columns to struct
// fields by `json` tag (falling back to the field name), mirroring the
// header logic in toValueRange.
func ReadRows[T any](ctx context.Context, r *Reader, sheetRange string) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		var zero T
		typ := reflect.TypeOf(zero)
		if typ == nil || typ.Kind() != reflect.Struct {
			yield(zero, fmt.Errorf("gsheets: ReadRows requires a struct type, got %T", zero))
			return
		}

		resp, err := r.service.Spreadsheets.Values.Get(r.spreadsheetID, sheetRange).Context(ctx).
			ValueRenderOption("UNFORMATTED_VALUE").
			Do()
		if err != nil {
			yield(zero, fmt.Errorf("spreadsheets.values.get: %w", err))
			return
		}
		if len(resp.Values) == 0 {
			return
		}

		fieldByCol := headerFieldIndex(typ, resp.Values[0])

		for _, row := range resp.Values[1:] {
			var item T
			v := reflect.ValueOf(&item).Elem()
			var rowErr error
			for col, fieldIdx := range fieldByCol {
				if col >= len(row) {
					continue
				}
				if err := setField(v.Field(fieldIdx), row[col]); err != nil {
					rowErr = fmt.Errorf("field %s: %w", typ.Field(fieldIdx).Name, err)
					break
				}
			}
			if !yield(item, rowErr) {
				return
			}
		}
	}
}

// headerFieldIndex maps a header row's column indexes to struct field
// indexes, matching by `json` tag (or field name) case-insensitively.
func headerFieldIndex(typ reflect.Type, header []any) map[int]int {
	nameToField := make(map[string]int, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}
		name := field.Name
		if jsonTag := field.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
			if idx := strings.Index(jsonTag, ","); idx >= 0 {
				jsonTag = jsonTag[:idx]
			}
			if jsonTag != "" {
				name = jsonTag
			}
		}
		nameToField[strings.ToLower(name)] = i
	}

	colToField := make(map[int]int, len(header))
	for col, h := range header {
		if idx, ok := nameToField[strings.ToLower(fmt.Sprint(h))]; ok {
			colToField[col] = idx
		}
	}
	return colToField
}

// setField assigns a raw cell value (string, float64 or bool, per the
// Sheets API) onto a struct field, converting as needed.
func setField(field reflect.Value, raw any) error {
	if !field.CanSet() {
		return nil
	}
	switch field.Kind() {
	case reflect.String:
		field.SetString(fmt.Sprint(raw))
	case reflect.Bool:
		if b, ok := raw.(bool); ok {
			field.SetBool(b)
			return nil
		}
		b, err := strconv.ParseBool(fmt.Sprint(raw))
		if err != nil {
			return err
		}
		field.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if f, ok := raw.(float64); ok {
			field.SetInt(int64(f))
			return nil
		}
		n, err := strconv.ParseInt(fmt.Sprint(raw), 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if f, ok := raw.(float64); ok {
			field.SetUint(uint64(f))
			return nil
		}
		n, err := strconv.ParseUint(fmt.Sprint(raw), 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(n)
	case reflect.Float32, reflect.Float64:
		if f, ok := raw.(float64); ok {
			field.SetFloat(f)
			return nil
		}
		f, err := strconv.ParseFloat(fmt.Sprint(raw), 64)
		if err != nil {
			return err
		}
		field.SetFloat(f)
	default:
		return fmt.Errorf("unsupported field kind %s", field.Kind())
	}
	return nil
}
