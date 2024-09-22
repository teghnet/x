package file

import (
	"bufio"
	"io"
	"log"
	"os"
	"strings"
	"unicode"
)

func ReadFirstLine(filename string) string {
	lines := must(ReadFileLines(filename, 1, false))
	if len(lines) == 0 {
		return ""
	}
	return lines[0]
}
func ReadLineNo(filename string, no int) string {
	lines := must(ReadFileLines(filename, no, false))
	if len(lines) == 0 {
		return ""
	}
	return lines[len(lines)-1]
}
func ReadAllLines(filename string) []string {
	return must(ReadFileLines(filename, 0, false))
}
func ReadAllLinesCut(filename, cutset string) []string {
	return must(ReadFileLines(filename, 0, false))
}
func ReadAllLinesWithOneComment(filename string) []string {
	return must(ReadFileLines(filename, 0, true))
}
func must(s []string, err error) []string {
	if err != nil {
		log.Fatal(err)
	}
	return s
}

func ReadFileLines(filename string, limit int, withOneComment bool) ([]string, error) {
	file, closeFile, err := openFile(filename)
	if err != nil {
		return nil, err
	}
	defer closeFile()
	return ReadLines(file, limit, withOneComment, "#;")
}
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
func openFile(filename string) (*os.File, func(), error) {
	file, err := os.Open(filename)
	return file, func() {
		if err := file.Close(); err != nil {
			log.Fatal(file.Name(), err)
		}
	}, err
}
func stripFromFirstChar(s, chars string) string {
	if cut := strings.IndexAny(s, chars); cut >= 0 {
		return strings.TrimRightFunc(s[:cut], unicode.IsSpace)
	}
	return s
}
