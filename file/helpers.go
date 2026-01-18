// Copyright (c) 2024-2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package file

import (
	"io"
	"log"
	"os"
	"strings"
	"unicode"
)

func openFile(filename string) (*os.File, error) {
	file, err := os.Open(filename)
	return file, err
}

func closeFile(f *os.File) {
	if err := f.Close(); err != nil {
		log.Fatal(f.Name(), err)
	}
}

func closeCloser(c io.Closer) {
	if err := c.Close(); err != nil {
		log.Fatal(err)
	}
}

func stripFromFirstChar(s, chars string) string {
	if cut := strings.IndexAny(s, chars); cut >= 0 {
		return strings.TrimRightFunc(s[:cut], unicode.IsSpace)
	}
	return s
}

func mustStrings(s []string, err error) []string {
	if err != nil {
		log.Fatal(err)
	}
	return s
}
