// Copyright (c) 2024-2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package paths

import (
	"fmt"
	"os"
	"path"
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
func wdIsHome() bool {
	wd, err := os.Getwd()
	if err != nil {
		return false
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	return sameDir(wd, homeDir)
}

func EnsureDir(p ...string) string {
	err := os.MkdirAll(path.Join(p...), 0700)
	if err != nil {
		panic(fmt.Errorf("could not create directory: %w", err))
	}
	return path.Join(p...)
}
