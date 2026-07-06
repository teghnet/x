// Copyright (c) 2024-2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package fsio

import (
	"errors"
	"io/fs"
	"iter"
	"log/slog"
	"os"
)

// Glob is a utility function that returns an iterator over files matching
// the given pattern in the provided filesystem.
func Glob(fsys fs.FS, pattern string) iter.Seq[string] {
	return func(yield func(string) bool) {
		matches, err := fs.Glob(fsys, pattern)
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

func Remove(fsys fs.FS, pattern string) error {
	var errs error
	for name := range Glob(fsys, pattern) {
		s, err := fs.Lstat(fsys, name)
		if err != nil {
			return err
		}
		errs = errors.Join(os.Remove(s.Name()))
	}
	return errs
}
