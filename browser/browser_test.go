package browser

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetParsesPage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><head><title>Hi</title></head>
			<body><p>Hello world</p><a href="/next">next</a></body></html>`))
	}))
	defer srv.Close()

	s, err := New(srv.Client())
	if err != nil {
		t.Fatal(err)
	}
	page, err := s.Get(context.Background(), srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if page.StatusCode != 200 {
		t.Fatalf("status = %d", page.StatusCode)
	}
	if page.Title() != "Hi" {
		t.Fatalf("title = %q", page.Title())
	}
	if page.Text() != "Hello world next" {
		t.Fatalf("text = %q", page.Text())
	}
	links := page.Links()
	if len(links) != 1 || links[0] != srv.URL+"/next" {
		t.Fatalf("links = %v (want resolved absolute URL)", links)
	}
}

func TestUserAgentHeader(t *testing.T) {
	var gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
	}))
	defer srv.Close()

	s, _ := New(srv.Client(), WithUserAgent("my-agent/2.0"))
	if _, err := s.Get(context.Background(), srv.URL); err != nil {
		t.Fatal(err)
	}
	if gotUA != "my-agent/2.0" {
		t.Fatalf("user-agent = %q", gotUA)
	}
}

func TestCookiePersistence(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/set" {
			http.SetCookie(w, &http.Cookie{Name: "sid", Value: "abc", Path: "/"})
			return
		}
		// /check echoes whether the cookie came back.
		if c, err := r.Cookie("sid"); err == nil {
			w.Write([]byte(c.Value))
		}
	}))
	defer srv.Close()

	s, _ := New(srv.Client())
	if _, err := s.Get(context.Background(), srv.URL+"/set"); err != nil {
		t.Fatal(err)
	}
	page, err := s.Get(context.Background(), srv.URL+"/check")
	if err != nil {
		t.Fatal(err)
	}
	if string(page.Body) != "abc" {
		t.Fatalf("cookie not replayed, body = %q", page.Body)
	}

	cookies, err := s.Cookies(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if len(cookies) != 1 || cookies[0].Value != "abc" {
		t.Fatalf("jar cookies = %+v", cookies)
	}
}
