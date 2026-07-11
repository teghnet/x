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

// carriageReturn, soh, and etx are the raw control bytes stripped from each
// scanned line: soh (SOH, 0x01) and etx (ETX, 0x03) are SWIFT FIN block
// framing characters, not message content.
const (
	carriageReturn = "\r"
	soh            = "\x01"
	etx            = "\x03"
)

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
			// Raw SWIFT FIN deliveries wrap each message in block control
			// characters: a lone SOH (soh) line before the message and a
			// trailing ETX (etx) glued onto the tagMessageBoundary
			// terminator line. Strip them so it still matches the message
			// boundary check and a bare soh line is treated as blank.
			line := strings.Trim(strings.TrimRight(sc.Text(), carriageReturn), soh+etx)
			if line == "" {
				continue
			}
			if line == tagMessageBoundary {
				if !flush() {
					return
				}
				if !yield(record{tag: tagMessageBoundary}, nil) {
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
