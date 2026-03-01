// Copyright (c) 2024-2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package fsio

import (
	"errors"
	"io/fs"
	"iter"
	"log/slog"
	"os"
	"path"
)

// Glob is a utility function that returns an iterator over files matching
// the given pattern in the provided filesystem.
func Glob(f fs.FS, pattern string) iter.Seq[string] {
	return func(yield func(string) bool) {
		matches, err := fs.Glob(f, pattern)
		if err != nil {
			slog.Debug("fsio.Glob: failed resolve pattern", "err", err)
			return
		}
		for _, match := range matches {
			if !yield(match) {
				return
			}
		}
	}
}

func Remove(dir, pattern string) error {
	var err error
	for name := range Glob(os.DirFS(dir), pattern) {
		err = errors.Join(os.Remove(path.Join(dir, name)))
	}
	return err
}
