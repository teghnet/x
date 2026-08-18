// Copyright (c) 2024-2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

// Package paths
//
// ### User directories
//
// - `XDG_CONFIG_HOME`
//   - Where user-specific configurations should be written (analogous to `/etc`).
//   - Should default to `$HOME/.config`.
//
// - `XDG_CACHE_HOME`
//   - Where user-specific non-essential (cached) data should be written (analogous to `/var/cache`).
//   - Should default to `$HOME/.cache`.
//
// - `XDG_DATA_HOME`
//   - Where user-specific data files should be written (analogous to `/usr/share`).
//   - Should default to `$HOME/.local/share`.
//
// - `XDG_STATE_HOME`
//   - Where user-specific state files should be written (analogous to `/var/lib`).
//   - Should default to `$HOME/.local/state`.
//
// - `XDG_RUNTIME_DIR`
//   - Used for non-essential, user-specific data files such as sockets, named pipes, etc.
//   - Not required to have a default value; warnings should be issued if not set or equivalents provided.
//   - Must be owned by the user with an access mode of `0700`.
//   - Filesystem fully featured by standards of OS.
//   - Must be on the local filesystem.
//   - May be subject to periodic cleanup.
//   - Modified every 6 hours or set sticky bit if persistence is desired.
//   - Can only exist for the duration of the user's login.
//   - Should not store large files as it may be mounted as a tmpfs.
//   - pam_systemd sets this to `/run/user/$UID`.
//
// ### System directories
//
// - `XDG_DATA_DIRS`
//   - List of directories separated by `:` (analogous to `PATH`).
//   - Should default to `/usr/local/share:/usr/share`.
//
// - `XDG_CONFIG_DIRS`
//   - List of directories separated by `:` (analogous to `PATH`).
//   - Should default to `/etc/xdg`.
package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func sameDir(a, b string) bool {
	ai, err := os.Stat(a)
	if err != nil {
		return false
	}
	bi, err := os.Stat(b)
	if err != nil {
		return false
	}
	return ai.IsDir() && bi.IsDir() && os.SameFile(ai, bi)
}

// wdIsHome checks if the working directory is in the user's home directory.
// It fails closed: if the working directory or the home directory cannot be
// determined, it reports true so that callers do not create dot-directories
// in a location they cannot prove is safe.
func wdIsHome() bool {
	wd, err := os.Getwd()
	if err != nil {
		return true
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return true
	}
	return sameDir(wd, homeDir)
}

// isDir reports whether p exists and is a directory.
func isDir(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}

// nonEmpty returns dir with blank (whitespace-only) elements removed.
func nonEmpty(dir []string) []string {
	var dd []string
	for _, d := range dir {
		if strings.TrimSpace(d) != "" {
			dd = append(dd, d)
		}
	}
	return dd
}

// EnsureDir creates the directory joined from p (mode 0700, including parents)
// and returns its path. It panics if the directory cannot be created.
func EnsureDir(p ...string) string {
	dir := filepath.Join(p...)
	if err := os.MkdirAll(dir, 0700); err != nil {
		panic(fmt.Errorf("could not create directory: %w", err))
	}
	return dir
}
