// Copyright (c) 2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package paths

import (
	"os"
	"path/filepath"
	"testing"
)

func Test_localAppDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Save and restore original working directory.
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(origWd) })

	dirsToCreate := []string{".local/app/dir", ".local/dir", ".dir/app", ".app/dir"}
	tests := []struct {
		name    string
		dirs    []string // directories to create relative to tmpDir
		app     string
		dir     string
		wantSfx string // expected suffix of the returned path
		wantOK  bool
	}{
		{
			name:    "prefers .local/app/dir",
			dirs:    dirsToCreate[0:],
			app:     "app",
			dir:     "dir",
			wantSfx: filepath.Join(".local", "app", "dir"),
			wantOK:  true,
		},
		{
			name:    "falls back to .local/dir",
			dirs:    dirsToCreate[1:],
			app:     "app",
			dir:     "dir",
			wantSfx: filepath.Join(".local", "dir"),
			wantOK:  true,
		},
		{
			name:    "falls back to .dir/app",
			dirs:    dirsToCreate[2:],
			app:     "app",
			dir:     "dir",
			wantSfx: filepath.Join(".dir", "app"),
			wantOK:  true,
		},
		{
			name:    ".app/dir works even at home",
			dirs:    dirsToCreate[3:],
			app:     "app",
			dir:     "dir",
			wantSfx: filepath.Join(".app", "dir"),
			wantOK:  true,
		},
		{
			name:   "returns false when nothing exists",
			dirs:   nil,
			app:    "app",
			dir:    "dir",
			wantOK: false,
		},
		// app == "" is a special case.
		{
			name:    "no app; prefers .local/app/dir",
			dirs:    dirsToCreate[0:],
			app:     "",
			dir:     "dir",
			wantSfx: filepath.Join(".local", "dir"),
			wantOK:  true,
		},
		{
			name:    "no app; falls back to .local/dir",
			dirs:    dirsToCreate[1:],
			app:     "",
			dir:     "dir",
			wantSfx: filepath.Join(".local", "dir"),
			wantOK:  true,
		},
		{
			name:    "no app; falls back to .dir",
			dirs:    dirsToCreate[2:],
			app:     "",
			dir:     "dir",
			wantSfx: filepath.Join(".dir"),
			wantOK:  true,
		},
		{
			name:   "no app; .dir NOT works at home",
			dirs:   dirsToCreate[3:],
			app:    "",
			dir:    "dir",
			wantOK: false,
		},
		{
			name:   "no app; returns false when nothing exists",
			dirs:   nil,
			app:    "",
			dir:    "dir",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Each subtest gets its own temp directory.
			wd := filepath.Join(tmpDir, tt.name)
			if err := os.MkdirAll(wd, 0o755); err != nil {
				t.Fatal(err)
			}
			for _, d := range tt.dirs {
				if err := os.MkdirAll(filepath.Join(wd, d), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.Chdir(wd); err != nil {
				t.Fatal(err)
			}

			got, ok := localAppDir(tt.app, tt.dir)
			if ok != tt.wantOK {
				t.Fatalf("localAppDir(%q, %q) ok = %v, want %v", tt.app, tt.dir, ok, tt.wantOK)
			}
			if !tt.wantOK {
				return
			}
			want := filepath.Join(wd, tt.wantSfx)
			if got != want {
				t.Errorf("localAppDir(%q, %q) = %q, want %q", tt.app, tt.dir, got, want)
			}
		})
	}
}
