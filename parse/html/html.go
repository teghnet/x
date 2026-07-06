// Package html provides a small, dependency-free HTML tokenizer plus a handful
// of extraction helpers (links, title, text).
//
// It is deliberately not a full HTML5 parser: there is no tree construction and
// no error-recovery state machine. It tokenizes well-formed and typical
// real-world markup well enough for scraping tasks (extracting anchors,
// metadata and visible text). Like the rest of parse/*, it performs no I/O.
package html

import (
	stdhtml "html"
	"strings"
)

// TokenType identifies the kind of a Token.
type TokenType int

const (
	// ErrorToken is never emitted; it marks the zero value.
	ErrorToken TokenType = iota
	// TextToken is a run of character data between tags.
	TextToken
	// StartTagToken is an opening tag such as <a href="x">.
	StartTagToken
	// EndTagToken is a closing tag such as </a>.
	EndTagToken
	// SelfClosingTagToken is a self-closed tag such as <br/>.
	SelfClosingTagToken
	// CommentToken is an HTML comment <!-- ... -->.
	CommentToken
	// DoctypeToken is a <!doctype ...> declaration.
	DoctypeToken
)

// Attr is a single tag attribute. Val has HTML entities unescaped.
type Attr struct {
	Key string
	Val string
}

// Token is a single lexical unit of an HTML document. For tag tokens Data holds
// the lower-cased tag name; for text and comment tokens it holds the (unescaped,
// for text) content.
type Token struct {
	Type TokenType
	Data string
	Attr []Attr
}

// Get returns the value of the named attribute (case-insensitive) and whether
// it was present.
func (t Token) Get(key string) (string, bool) {
	for _, a := range t.Attr {
		if strings.EqualFold(a.Key, key) {
			return a.Val, true
		}
	}
	return "", false
}

// rawTextTags hold content that must not be tokenized as markup.
var rawTextTags = map[string]bool{
	"script":   true,
	"style":    true,
	"textarea": true,
	"title":    true,
}

// Tokenize scans an HTML document and returns its tokens in order.
func Tokenize(data []byte) []Token {
	s := string(data)
	var toks []Token
	i := 0
	for i < len(s) {
		if s[i] != '<' {
			// Text run up to the next tag.
			j := strings.IndexByte(s[i:], '<')
			if j < 0 {
				appendText(&toks, s[i:])
				break
			}
			appendText(&toks, s[i:i+j])
			i += j
			continue
		}

		// s[i] == '<'
		switch {
		case strings.HasPrefix(s[i:], "<!--"):
			end := strings.Index(s[i+4:], "-->")
			if end < 0 {
				toks = append(toks, Token{Type: CommentToken, Data: s[i+4:]})
				i = len(s)
			} else {
				toks = append(toks, Token{Type: CommentToken, Data: s[i+4 : i+4+end]})
				i += 4 + end + 3
			}
		case strings.HasPrefix(strings.ToLower(s[i:min(i+2, len(s))]), "<!"):
			end := strings.IndexByte(s[i:], '>')
			if end < 0 {
				toks = append(toks, Token{Type: DoctypeToken, Data: s[i+2:]})
				i = len(s)
			} else {
				toks = append(toks, Token{Type: DoctypeToken, Data: strings.TrimSpace(s[i+2 : i+end])})
				i += end + 1
			}
		default:
			tok, next, ok := parseTag(s, i)
			if !ok {
				// A lone '<' that does not begin a tag: treat as text.
				appendText(&toks, "<")
				i++
				continue
			}
			toks = append(toks, tok)
			i = next
			// Raw-text elements: swallow content verbatim until the close tag.
			if tok.Type == StartTagToken && rawTextTags[tok.Data] {
				closing := "</" + tok.Data
				rest := strings.ToLower(s[i:])
				end := strings.Index(rest, closing)
				if end < 0 {
					if content := s[i:]; content != "" {
						toks = append(toks, Token{Type: TextToken, Data: content})
					}
					i = len(s)
					continue
				}
				if content := s[i : i+end]; content != "" {
					toks = append(toks, Token{Type: TextToken, Data: content})
				}
				i += end
			}
		}
	}
	return toks
}

func appendText(toks *[]Token, raw string) {
	if raw == "" {
		return
	}
	*toks = append(*toks, Token{Type: TextToken, Data: stdhtml.UnescapeString(raw)})
}

// parseTag parses a start, end or self-closing tag beginning at s[i] (which is
// '<'). It returns the token, the index just past the tag's closing '>', and
// whether parsing succeeded.
func parseTag(s string, i int) (Token, int, bool) {
	j := i + 1
	end := false
	if j < len(s) && s[j] == '/' {
		end = true
		j++
	}
	// Tag name.
	nameStart := j
	for j < len(s) && !isSpace(s[j]) && s[j] != '>' && s[j] != '/' {
		j++
	}
	name := strings.ToLower(s[nameStart:j])
	if name == "" {
		return Token{}, i, false
	}

	tok := Token{Data: name}
	if end {
		tok.Type = EndTagToken
	} else {
		tok.Type = StartTagToken
	}

	// Attributes.
	for j < len(s) {
		for j < len(s) && isSpace(s[j]) {
			j++
		}
		if j >= len(s) {
			return tok, j, true
		}
		if s[j] == '>' {
			return tok, j + 1, true
		}
		if s[j] == '/' {
			// Possible self-close.
			k := j + 1
			for k < len(s) && isSpace(s[k]) {
				k++
			}
			if k < len(s) && s[k] == '>' {
				if tok.Type == StartTagToken {
					tok.Type = SelfClosingTagToken
				}
				return tok, k + 1, true
			}
			j++
			continue
		}
		var a Attr
		a, j = parseAttr(s, j)
		if !end && a.Key != "" {
			tok.Attr = append(tok.Attr, a)
		}
	}
	return tok, j, true
}

func parseAttr(s string, j int) (Attr, int) {
	keyStart := j
	for j < len(s) && !isSpace(s[j]) && s[j] != '=' && s[j] != '>' && s[j] != '/' {
		j++
	}
	key := strings.ToLower(s[keyStart:j])
	for j < len(s) && isSpace(s[j]) {
		j++
	}
	if j >= len(s) || s[j] != '=' {
		return Attr{Key: key}, j // boolean attribute
	}
	j++ // consume '='
	for j < len(s) && isSpace(s[j]) {
		j++
	}
	if j >= len(s) {
		return Attr{Key: key}, j
	}
	var val string
	switch s[j] {
	case '"', '\'':
		quote := s[j]
		j++
		valStart := j
		for j < len(s) && s[j] != quote {
			j++
		}
		val = s[valStart:j]
		if j < len(s) {
			j++ // consume closing quote
		}
	default:
		valStart := j
		for j < len(s) && !isSpace(s[j]) && s[j] != '>' {
			j++
		}
		val = s[valStart:j]
	}
	return Attr{Key: key, Val: stdhtml.UnescapeString(val)}, j
}

func isSpace(b byte) bool {
	switch b {
	case ' ', '\t', '\n', '\r', '\f':
		return true
	}
	return false
}

// Links returns the href value of every <a> element, in document order.
// Relative and absolute URLs are returned verbatim; callers resolve them.
func Links(data []byte) []string {
	var out []string
	for _, t := range Tokenize(data) {
		if (t.Type == StartTagToken || t.Type == SelfClosingTagToken) && t.Data == "a" {
			if href, ok := t.Get("href"); ok && href != "" {
				out = append(out, href)
			}
		}
	}
	return out
}

// Title returns the text content of the first <title> element, or "" if absent.
func Title(data []byte) string {
	toks := Tokenize(data)
	for i, t := range toks {
		if t.Type == StartTagToken && t.Data == "title" {
			if i+1 < len(toks) && toks[i+1].Type == TextToken {
				return strings.TrimSpace(stdhtml.UnescapeString(toks[i+1].Data))
			}
			return ""
		}
	}
	return ""
}

// Text returns the concatenated visible text of the document, with script and
// style contents removed and runs of whitespace collapsed to single spaces.
func Text(data []byte) string {
	var b strings.Builder
	skip := "" // when non-empty, drop text until this tag closes
	for _, t := range Tokenize(data) {
		switch t.Type {
		case StartTagToken:
			if t.Data == "script" || t.Data == "style" || t.Data == "title" {
				skip = t.Data
			}
		case EndTagToken:
			if t.Data == skip {
				skip = ""
			}
		case TextToken:
			if skip == "" {
				b.WriteString(t.Data)
				b.WriteByte(' ')
			}
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}
