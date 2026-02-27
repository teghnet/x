// Copyright (c) 2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package fsio_test

import (
	"testing"
	"testing/fstest"

	"github.com/teghnet/x/fsio"
)

func TestLoadJSON(t *testing.T) {
	type testData struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	tests := []struct {
		name      string
		fs        fstest.MapFS
		fileName  string
		wantName  string
		wantValue int
		wantErr   bool
	}{
		{
			name: "valid JSON",
			fs: fstest.MapFS{
				"data.json": &fstest.MapFile{
					Data: []byte(`{"name":"test","value":42}`),
				},
			},
			fileName:  "data.json",
			wantName:  "test",
			wantValue: 42,
			wantErr:   false,
		},
		{
			name:     "file not found",
			fs:       fstest.MapFS{},
			fileName: "missing.json",
			wantErr:  true,
		},
		{
			name: "invalid JSON",
			fs: fstest.MapFS{
				"bad.json": &fstest.MapFile{
					Data: []byte(`{invalid}`),
				},
			},
			fileName: "bad.json",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fsio.JSON[testData](tt.fs, tt.fileName)
			if (err != nil) != tt.wantErr {
				t.Errorf("JSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Name != tt.wantName {
					t.Errorf("JSON() Name = %v, want %v", got.Name, tt.wantName)
				}
				if got.Value != tt.wantValue {
					t.Errorf("JSON() Value = %v, want %v", got.Value, tt.wantValue)
				}
			}
		})
	}
}

func TestJSONList(t *testing.T) {
	type item struct {
		ID int `json:"id"`
	}

	fs := fstest.MapFS{
		"items.jsonl": &fstest.MapFile{
			Data: []byte(`{"id":1}
{"id":2}
{"id":3}
`),
		},
	}

	var ids []int
	for item, err := range fsio.JSONList[item](fs, "items.jsonl") {
		if err != nil {
			t.Fatalf("JSONList() unexpected error: %v", err)
		}
		ids = append(ids, item.ID)
	}

	if len(ids) != 3 {
		t.Errorf("JSONList() got %d items, want 3", len(ids))
	}
	for i, want := range []int{1, 2, 3} {
		if ids[i] != want {
			t.Errorf("JSONList() ids[%d] = %d, want %d", i, ids[i], want)
		}
	}
}

func TestJSONArray(t *testing.T) {
	type item struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	fs := fstest.MapFS{
		"items.json": &fstest.MapFile{
			Data: []byte(`[{"id":1,"name":"first"},{"id":2,"name":"second"},{"id":3,"name":"third"}]`),
		},
	}

	var items []item
	for item, err := range fsio.JSONArray[item](fs, "items.json") {
		if err != nil {
			t.Fatalf("JSONArray() unexpected error: %v", err)
		}
		items = append(items, item)
	}

	if len(items) != 3 {
		t.Errorf("JSONArray() got %d items, want 3", len(items))
	}

	expected := []struct {
		id   int
		name string
	}{
		{1, "first"},
		{2, "second"},
		{3, "third"},
	}
	for i, want := range expected {
		if items[i].ID != want.id {
			t.Errorf("JSONArray() items[%d].ID = %d, want %d", i, items[i].ID, want.id)
		}
		if items[i].Name != want.name {
			t.Errorf("JSONArray() items[%d].Name = %s, want %s", i, items[i].Name, want.name)
		}
	}
}

func TestJSONArray_Empty(t *testing.T) {
	type item struct {
		ID int `json:"id"`
	}

	fs := fstest.MapFS{
		"empty.json": &fstest.MapFile{
			Data: []byte(`[]`),
		},
	}

	var count int
	for _, err := range fsio.JSONArray[item](fs, "empty.json") {
		if err != nil {
			t.Fatalf("JSONArray() unexpected error: %v", err)
		}
		count++
	}

	if count != 0 {
		t.Errorf("JSONArray() got %d items, want 0", count)
	}
}
