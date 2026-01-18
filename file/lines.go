// Copyright (c) 2024-2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package file

import (
	"bufio"
	"io"
	"strings"
)

// ReadFirstLine reads the first line of a file and returns it as a string.
func ReadFirstLine(filename string) string {
	lines := mustStrings(ReadFileLines(filename, 1, false))
	if len(lines) == 0 {
		return ""
	}
	return lines[0]
}

// ReadLineNo reads the nth line of a file and returns it as a string.
func ReadLineNo(filename string, no int) string {
	lines := mustStrings(ReadFileLines(filename, no, false))
	if len(lines) == 0 {
		return ""
	}
	return lines[len(lines)-1]
}

// ReadAllLines reads all lines of a file and returns them as a slice of strings.
func ReadAllLines(filename string) []string {
	return mustStrings(ReadFileLines(filename, 0, false))
}

// ReadAllLinesWithOneComment reads all lines of a file and returns them as a slice of strings
// adding one line of comment if it exists.
func ReadAllLinesWithOneComment(filename string) []string {
	return mustStrings(ReadFileLines(filename, 0, true))
}

// ReadFileLines reads lines from a file and returns them as a slice of strings.
func ReadFileLines(filename string, limit int, withOneComment bool) ([]string, error) {
	file, err := openFile(filename)
	if err != nil {
		return nil, err
	}
	defer closeFile(file)
	return ReadLines(file, limit, withOneComment, "#;")
}

// ReadLines reads lines from a reader and returns them as a slice of strings.
func ReadLines(r io.Reader, limit int, withOneComment bool, chars string) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		text := scanner.Text()
		s := strings.Trim(stripFromFirstChar(text, chars), "\t \r\n")
		if len(s) != 0 {
			if withOneComment && len(text) > 0 {
				lines = append(lines, text)
			}
			lines = append(lines, s)
		}
		if 0 < limit && limit <= len(lines) {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}
