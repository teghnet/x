package bankparse

import (
	"fmt"
	"io"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

// Config holds the options shared across statement format parsers. Not every
// field is meaningful to every format; each Parser implementation documents
// which ones it honors.
type Config struct {
	// DateFormat is the time.Parse layout used for dates in formats that
	// don't have a fixed date grammar (e.g. CSV).
	DateFormat string
	// Encoding is the source character encoding, e.g. "utf-8" (default),
	// "windows-1250", "iso-8859-2". See DecodeReader for supported names.
	Encoding string
	// SkipHeaderLines is the number of leading lines to discard before
	// parsing begins (e.g. a CSV header row).
	SkipHeaderLines int
	// Delimiter is the field delimiter for delimited formats. Defaults to ','.
	Delimiter rune
}

// Option configures a Config.
type Option func(*Config)

// WithDateFormat sets the time.Parse layout used to parse dates.
func WithDateFormat(layout string) Option {
	return func(c *Config) { c.DateFormat = layout }
}

// WithEncoding sets the source character encoding. See DecodeReader for the
// set of recognized names.
func WithEncoding(name string) Option {
	return func(c *Config) { c.Encoding = name }
}

// WithSkipHeaderLines sets the number of leading lines to discard.
func WithSkipHeaderLines(n int) Option {
	return func(c *Config) { c.SkipHeaderLines = n }
}

// WithDelimiter sets the field delimiter for delimited formats.
func WithDelimiter(r rune) Option {
	return func(c *Config) { c.Delimiter = r }
}

// NewConfig builds a Config from opts, applying defaults for anything left
// unset.
func NewConfig(opts ...Option) Config {
	cfg := Config{
		DateFormat: "2006-01-02",
		Encoding:   "utf-8",
		Delimiter:  ',',
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}

// encodings maps the names accepted by WithEncoding to their charmap.Charmap.
// Names are matched case-insensitively; "utf-8" and "" both mean no
// conversion.
var encodings = map[string]*charmap.Charmap{
	"windows-1250": charmap.Windows1250,
	"windows-1252": charmap.Windows1252,
	"iso-8859-1":   charmap.ISO8859_1,
	"iso-8859-2":   charmap.ISO8859_2,
	"iso-8859-15":  charmap.ISO8859_15,
}

// DecodeReader wraps r so that bytes read from it are transcoded from
// encodingName to UTF-8. An empty name or "utf-8" returns r unchanged.
func DecodeReader(r io.Reader, encodingName string) (io.Reader, error) {
	if encodingName == "" || encodingName == "utf-8" {
		return r, nil
	}
	cm, ok := encodings[encodingName]
	if !ok {
		return nil, fmt.Errorf("bankparse: unsupported encoding %q", encodingName)
	}
	var enc encoding.Encoding = cm
	return transform.NewReader(r, enc.NewDecoder()), nil
}
