package mt940

import (
	"bufio"
	"io"
	"iter"
	"regexp"
	"strings"
)

// tagStart matches the start of a line beginning a new SWIFT field, e.g.
// ":20:", ":61:", ":28C:".
var tagStart = regexp.MustCompile(`^:([0-9]{2}[A-Z]?):`)

// record is one SWIFT field: its tag and full value, with continuation lines
// (if any) joined by "\n".
type record struct {
	tag   string
	value string
}

// scan splits raw MT940 text into a sequence of records. A line consisting
// solely of "-" ends the current message; scan emits a record with tag "-"
// so the caller can reset per-message state. Lines that don't start a new
// tag are treated as continuations of the previous field's value.
func scan(r io.Reader) iter.Seq2[record, error] {
	return func(yield func(record, error) bool) {
		sc := bufio.NewScanner(r)
		sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)

		var cur record
		flush := func() bool {
			if cur.tag == "" {
				return true
			}
			ok := yield(cur, nil)
			cur = record{}
			return ok
		}

		for sc.Scan() {
			line := strings.TrimRight(sc.Text(), "\r")
			if line == "" {
				continue
			}
			if line == "-" {
				if !flush() {
					return
				}
				if !yield(record{tag: "-"}, nil) {
					return
				}
				continue
			}
			if m := tagStart.FindStringSubmatchIndex(line); m != nil {
				if !flush() {
					return
				}
				cur = record{tag: line[m[2]:m[3]], value: line[m[1]:]}
				continue
			}
			// Continuation of the previous field's value.
			if cur.tag != "" {
				cur.value += "\n" + line
			}
		}
		if err := sc.Err(); err != nil {
			yield(record{}, err)
			return
		}
		flush()
	}
}
