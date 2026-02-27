// Copyright (c) 2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package osio

import (
	"io"
	"os"
)

// DynamicReader returns a reader based on the name.
// Use "-" or "stdin" for os.Stdin.
func DynamicReader(name string) (io.ReadCloser, error) {
	if name == "" {
		if hasStdin() {
			return os.Stdin, nil
		}
		panic("name must be non-empty")
	}
	if name == "-" || name == "stdin" {
		return os.Stdin, nil
	}
	return os.Open(name)
}

// hasStdin returns true if data was piped to stdin (i.e., stdin is not a terminal).
// This is useful for determining if the command should read from stdin.
func hasStdin() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	// Check if stdin is a character device (terminal).
	// If it's not a character device, data is being piped in.
	return (stat.Mode() & os.ModeCharDevice) == 0
}

// DynamicWriter returns a writer based on the name.
// Use "-" or "stdout" for os.Stdout, "=" or "stderr" for os.Stderr.
// Any other name opens a file in append mode if enabled.
func DynamicWriter(name string, append bool) (io.WriteCloser, error) {
	if name == "" {
		panic("name must be non-empty")
	}
	if name == "-" || name == "stdout" {
		return os.Stdout, nil
	}
	if name == "=" || name == "stderr" {
		return os.Stderr, nil
	}
	flag := os.O_WRONLY | os.O_CREATE
	if append {
		flag |= os.O_APPEND
	} else {
		flag |= os.O_TRUNC
	}
	return os.OpenFile(name, flag, 0600)
}
