// Copyright (c) 2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package paths

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func Test_localAppDir(t *testing.T) {
	dirsToCreate := []string{".local/app/dir", ".local/dir", ".dir", ".app/dir"}
	tests := []struct {
		name    string
		dirs    []string // directories to create relative to WD
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
			name:    "falls back to .dir",
			dirs:    dirsToCreate[2:],
			app:     "app",
			dir:     "dir",
			wantSfx: ".dir",
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
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Each subtest gets its own working directory; index-named so
			// subtest names containing "/" or spaces don't shape real paths.
			wd := filepath.Join(t.TempDir(), "case-"+strconv.Itoa(i))
			if err := os.MkdirAll(wd, 0o755); err != nil {
				t.Fatal(err)
			}
			for _, d := range tt.dirs {
				if err := os.MkdirAll(filepath.Join(wd, d), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			t.Chdir(wd)

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

	t.Run("panics on empty app name", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Errorf("expected panic")
			}
		}()
		localAppDir("", "dir")
	})
}

func Test_mkLocalAppDir(t *testing.T) {
	t.Run("valid app name does not panic and creates .<app>/<dir>", func(t *testing.T) {
		wd := t.TempDir()
		t.Chdir(wd)

		if err := mkLocalAppDir("app", "dir"); err != nil {
			t.Fatalf("mkLocalAppDir returned error: %v", err)
		}
		want := filepath.Join(wd, ".app", "dir")
		if !isDir(want) {
			t.Errorf("expected %q to exist and be a directory", want)
		}
	})

	t.Run("panics on empty app name", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Errorf("expected panic")
			}
		}()
		_ = mkLocalAppDir("")
	})

	t.Run("refuses to create in $HOME", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		t.Chdir(home)

		if err := mkLocalAppDir("app", "dir"); err == nil {
			t.Errorf("expected error when working directory is $HOME")
		}
	})
}

func Test_mkDotDir(t *testing.T) {
	t.Run("creates .<dir>", func(t *testing.T) {
		wd := t.TempDir()
		t.Chdir(wd)

		if err := mkDotDir("dir"); err != nil {
			t.Fatalf("mkDotDir returned error: %v", err)
		}
		want := filepath.Join(wd, ".dir")
		if !isDir(want) {
			t.Errorf("expected %q to exist and be a directory", want)
		}
	})

	t.Run("errors on empty dir", func(t *testing.T) {
		if err := mkDotDir(); err == nil {
			t.Errorf("expected error for empty dir")
		}
	})

	t.Run("refuses to create in $HOME", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		t.Chdir(home)

		if err := mkDotDir("dir"); err == nil {
			t.Errorf("expected error when working directory is $HOME")
		}
	})
}
