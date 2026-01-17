// Copyright (c) $year Paweł Zaremba
// SPDX-License-Identifier: MIT

package fsio_test

import (
	"testing"
	"testing/fstest"

	"github.com/teghnet/x/fsio"
)

func TestFSLoadJSON(t *testing.T) {
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
			got, err := fsio.FSLoadJSON[testData](tt.fs, tt.fileName)
			if (err != nil) != tt.wantErr {
				t.Errorf("FSLoadJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Name != tt.wantName {
					t.Errorf("FSLoadJSON() Name = %v, want %v", got.Name, tt.wantName)
				}
				if got.Value != tt.wantValue {
					t.Errorf("FSLoadJSON() Value = %v, want %v", got.Value, tt.wantValue)
				}
			}
		})
	}
}

func TestFSJSONList(t *testing.T) {
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
	for item, err := range fsio.FSJSONList[item](fs, "items.jsonl") {
		if err != nil {
			t.Fatalf("FSJSONList() unexpected error: %v", err)
		}
		ids = append(ids, item.ID)
	}

	if len(ids) != 3 {
		t.Errorf("FSIterateJSONs() got %d items, want 3", len(ids))
	}
	for i, want := range []int{1, 2, 3} {
		if ids[i] != want {
			t.Errorf("FSIterateJSONs() ids[%d] = %d, want %d", i, ids[i], want)
		}
	}
}

func TestFSGlob(t *testing.T) {
	fs := fstest.MapFS{
		"data/file1.json":  &fstest.MapFile{Data: []byte(`{}`)},
		"data/file2.json":  &fstest.MapFile{Data: []byte(`{}`)},
		"data/file3.txt":   &fstest.MapFile{Data: []byte(`text`)},
		"other/file4.json": &fstest.MapFile{Data: []byte(`{}`)},
	}

	tests := []struct {
		name    string
		pattern string
		want    []string
	}{
		{
			name:    "match JSON files in data",
			pattern: "data/*.json",
			want:    []string{"data/file1.json", "data/file2.json"},
		},
		{
			name:    "match all JSON files",
			pattern: "*/*.json",
			want:    []string{"data/file1.json", "data/file2.json", "other/file4.json"},
		},
		{
			name:    "no matches",
			pattern: "*.xml",
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []string
			for match := range fsio.FSGlob(fs, tt.pattern) {
				got = append(got, match)
			}
			if len(got) != len(tt.want) {
				t.Errorf("FSGlob() got %d matches, want %d", len(got), len(tt.want))
				return
			}
			for i, want := range tt.want {
				if got[i] != want {
					t.Errorf("FSGlob() got[%d] = %v, want %v", i, got[i], want)
				}
			}
		})
	}
}

func TestFSGlob_InvalidPattern(t *testing.T) {
	fs := fstest.MapFS{}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("FSGlob() expected panic for invalid pattern")
		}
	}()

	// Invalid pattern should cause panic
	for range fsio.FSGlob(fs, "[") {
		// This should not be reached
	}
}
