package html

import (
	"reflect"
	"testing"
)

func TestTokenizeBasic(t *testing.T) {
	toks := Tokenize([]byte(`<p class="x">hi <b>there</b></p>`))
	want := []Token{
		{Type: StartTagToken, Data: "p", Attr: []Attr{{Key: "class", Val: "x"}}},
		{Type: TextToken, Data: "hi "},
		{Type: StartTagToken, Data: "b"},
		{Type: TextToken, Data: "there"},
		{Type: EndTagToken, Data: "b"},
		{Type: EndTagToken, Data: "p"},
	}
	if !reflect.DeepEqual(toks, want) {
		t.Fatalf("got %#v\nwant %#v", toks, want)
	}
}

func TestTokenizeAttributes(t *testing.T) {
	toks := Tokenize([]byte(`<input disabled name=foo value='b a r'>`))
	if len(toks) != 1 {
		t.Fatalf("got %d tokens", len(toks))
	}
	tok := toks[0]
	if v, ok := tok.Get("disabled"); !ok || v != "" {
		t.Errorf("disabled = %q,%v", v, ok)
	}
	if v, _ := tok.Get("NAME"); v != "foo" {
		t.Errorf("name = %q (case-insensitive lookup)", v)
	}
	if v, _ := tok.Get("value"); v != "b a r" {
		t.Errorf("value = %q", v)
	}
}

func TestSelfClosingAndComment(t *testing.T) {
	toks := Tokenize([]byte(`<br/><!-- note --><img src="a.png" />`))
	if toks[0].Type != SelfClosingTagToken || toks[0].Data != "br" {
		t.Errorf("token0 = %+v", toks[0])
	}
	if toks[1].Type != CommentToken || toks[1].Data != " note " {
		t.Errorf("token1 = %+v", toks[1])
	}
	if toks[2].Type != SelfClosingTagToken || toks[2].Data != "img" {
		t.Errorf("token2 = %+v", toks[2])
	}
}

func TestEntitiesUnescaped(t *testing.T) {
	toks := Tokenize([]byte(`<a title="Tom &amp; Jerry">1 &lt; 2</a>`))
	if v, _ := toks[0].Get("title"); v != "Tom & Jerry" {
		t.Errorf("title = %q", v)
	}
	if toks[1].Data != "1 < 2" {
		t.Errorf("text = %q", toks[1].Data)
	}
}

func TestRawTextScriptNotParsed(t *testing.T) {
	toks := Tokenize([]byte(`<script>if (a < b && c > d) {}</script><p>ok</p>`))
	// The script body must be a single text token, not parsed as tags.
	var scriptText string
	for i, tk := range toks {
		if tk.Type == StartTagToken && tk.Data == "script" {
			scriptText = toks[i+1].Data
		}
	}
	if scriptText != "if (a < b && c > d) {}" {
		t.Fatalf("script text = %q", scriptText)
	}
}

func TestLinks(t *testing.T) {
	in := []byte(`<a href="/one">1</a><a>no href</a><a href="https://x/2">2</a>`)
	got := Links(in)
	want := []string{"/one", "https://x/2"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestTitle(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{`<head><title> Hello &amp; Bye </title></head>`, "Hello & Bye"},
		{`<html>no title here</html>`, ""},
		{`<title></title>`, ""},
	}
	for _, tt := range tests {
		if got := Title([]byte(tt.in)); got != tt.want {
			t.Errorf("Title(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestText(t *testing.T) {
	in := []byte(`<html><head><style>p{color:red}</style></head>
		<body><h1>Title</h1><script>var x=1</script><p>Hello   world</p></body></html>`)
	got := Text(in)
	want := "Title Hello world"
	if got != want {
		t.Fatalf("Text = %q, want %q", got, want)
	}
}

func TestTokenizeUnterminated(t *testing.T) {
	// Should not panic on truncated input.
	Tokenize([]byte(`<a href="x`))
	Tokenize([]byte(`<!-- open`))
	Tokenize([]byte(`text < notatag`))
	Tokenize([]byte(`<`))
}
