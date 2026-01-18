// Copyright (c) 2024-2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package fsio_test

import (
	"os"
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

func TestFSJSONArray(t *testing.T) {
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
	for item, err := range fsio.FSJSONArray[item](fs, "items.json") {
		if err != nil {
			t.Fatalf("FSJSONArray() unexpected error: %v", err)
		}
		items = append(items, item)
	}

	if len(items) != 3 {
		t.Errorf("FSJSONArray() got %d items, want 3", len(items))
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
			t.Errorf("FSJSONArray() items[%d].ID = %d, want %d", i, items[i].ID, want.id)
		}
		if items[i].Name != want.name {
			t.Errorf("FSJSONArray() items[%d].Name = %s, want %s", i, items[i].Name, want.name)
		}
	}
}

func TestFSJSONArray_Empty(t *testing.T) {
	type item struct {
		ID int `json:"id"`
	}

	fs := fstest.MapFS{
		"empty.json": &fstest.MapFile{
			Data: []byte(`[]`),
		},
	}

	var count int
	for _, err := range fsio.FSJSONArray[item](fs, "empty.json") {
		if err != nil {
			t.Fatalf("FSJSONArray() unexpected error: %v", err)
		}
		count++
	}

	if count != 0 {
		t.Errorf("FSJSONArray() got %d items, want 0", count)
	}
}

func TestDynamicWriter_Stdout(t *testing.T) {
	tests := []string{"-", "stdout"}
	for _, name := range tests {
		t.Run(name, func(t *testing.T) {
			w, err := fsio.DynamicWriter(name, false)
			if err != nil {
				t.Fatalf("DynamicWriter() error = %v", err)
			}
			// Should return os.Stdout, we can't directly compare, but we know it shouldn't be nil
			if w == nil {
				t.Error("DynamicWriter() returned nil for stdout")
			}
			// Don't close stdout
		})
	}
}

func TestDynamicWriter_Stderr(t *testing.T) {
	tests := []string{"=", "stderr"}
	for _, name := range tests {
		t.Run(name, func(t *testing.T) {
			w, err := fsio.DynamicWriter(name, false)
			if err != nil {
				t.Fatalf("DynamicWriter() error = %v", err)
			}
			if w == nil {
				t.Error("DynamicWriter() returned nil for stderr")
			}
			// Don't close stderr
		})
	}
}

func TestDynamicWriter_File(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := tmpDir + "/output.txt"

	w, err := fsio.DynamicWriter(filePath, false)
	if err != nil {
		t.Fatalf("DynamicWriter() error = %v", err)
	}
	defer w.Close()

	_, err = w.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
}

func TestDynamicWriter_FileAppend(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := tmpDir + "/append.txt"

	// Write first content
	w1, err := fsio.DynamicWriter(filePath, false)
	if err != nil {
		t.Fatalf("DynamicWriter() error = %v", err)
	}
	_, _ = w1.Write([]byte("first"))
	w1.Close()

	// Append second content
	w2, err := fsio.DynamicWriter(filePath, true)
	if err != nil {
		t.Fatalf("DynamicWriter() append error = %v", err)
	}
	_, _ = w2.Write([]byte("second"))
	w2.Close()

	// Verify content using os.ReadFile
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(data) != "firstsecond" {
		t.Errorf("DynamicWriter append got %q, want %q", string(data), "firstsecond")
	}
}

func TestDynamicWriter_FileTruncate(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := tmpDir + "/truncate.txt"

	// Write first content
	w1, err := fsio.DynamicWriter(filePath, false)
	if err != nil {
		t.Fatalf("DynamicWriter() error = %v", err)
	}
	_, _ = w1.Write([]byte("first content that is long"))
	w1.Close()

	// Truncate and write second content
	w2, err := fsio.DynamicWriter(filePath, false)
	if err != nil {
		t.Fatalf("DynamicWriter() truncate error = %v", err)
	}
	_, _ = w2.Write([]byte("short"))
	w2.Close()
}

func TestDynamicWriter_InvalidPath(t *testing.T) {
	_, err := fsio.DynamicWriter("/nonexistent/path/to/file.txt", false)
	if err == nil {
		t.Error("DynamicWriter() expected error for invalid path")
	}
}
