package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTokenValid(t *testing.T) {
	now := time.Unix(1000, 0)
	tests := []struct {
		name string
		tok  Token
		want bool
	}{
		{"empty", Token{}, false},
		{"no expiry", Token{Value: "x"}, true},
		{"future expiry", Token{Value: "x", Expiry: now.Add(time.Minute)}, true},
		{"past expiry", Token{Value: "x", Expiry: now.Add(-time.Minute)}, false},
	}
	for _, tt := range tests {
		if got := tt.tok.Valid(now); got != tt.want {
			t.Errorf("%s: Valid = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestTokenHeader(t *testing.T) {
	if got := (Token{Value: "abc"}).Header(); got != "Bearer abc" {
		t.Errorf("default scheme: %q", got)
	}
	if got := (Token{Value: "abc", Type: "Basic"}).Header(); got != "Basic abc" {
		t.Errorf("explicit scheme: %q", got)
	}
}

func TestStaticProvider(t *testing.T) {
	ctx := context.Background()
	p := StaticProvider{
		Default:  StaticToken{Value: "def"},
		Accounts: map[string]Credential{"bob": StaticToken{Value: "bob-tok"}},
	}
	c, err := p.Credential(ctx, "")
	if err != nil {
		t.Fatal(err)
	}
	if tok, _ := c.Token(ctx); tok.Value != "def" {
		t.Errorf("default = %q", tok.Value)
	}
	c, _ = p.Credential(ctx, "bob")
	if tok, _ := c.Token(ctx); tok.Value != "bob-tok" {
		t.Errorf("bob = %q", tok.Value)
	}
	if _, err := p.Credential(ctx, "nobody"); !errors.Is(err, ErrNoCredential) {
		t.Errorf("missing account: %v", err)
	}
}

func TestTransportInjectsHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
	}))
	defer srv.Close()

	rt := &Transport{Credential: StaticToken{Value: "sekret"}}
	hc := &http.Client{Transport: rt}

	req, _ := http.NewRequest(http.MethodGet, srv.URL, nil)
	resp, err := hc.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if gotAuth != "Bearer sekret" {
		t.Fatalf("Authorization = %q", gotAuth)
	}
	// The original request must be untouched.
	if req.Header.Get("Authorization") != "" {
		t.Fatal("transport mutated caller's request")
	}
}

func TestTransportNoCredential(t *testing.T) {
	rt := &Transport{}
	req, _ := http.NewRequest(http.MethodGet, "http://example.invalid", nil)
	if _, err := rt.RoundTrip(req); err == nil {
		t.Fatal("expected error with no credential")
	}
}

func TestMemKeyring(t *testing.T) {
	k := NewMemKeyring()
	if _, err := k.Get("svc", "acct"); !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound, got %v", err)
	}
	if err := k.Set("svc", "acct", "s3cr3t"); err != nil {
		t.Fatal(err)
	}
	v, err := k.Get("svc", "acct")
	if err != nil || v != "s3cr3t" {
		t.Fatalf("got %q err=%v", v, err)
	}
	if err := k.Delete("svc", "acct"); err != nil {
		t.Fatal(err)
	}
	if _, err := k.Get("svc", "acct"); !errors.Is(err, ErrKeyNotFound) {
		t.Fatal("expected deleted")
	}
}

func TestOAuth2RefreshAndCache(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if err := r.ParseForm(); err != nil {
			t.Errorf("parse form: %v", err)
		}
		if r.Form.Get("grant_type") != "refresh_token" {
			t.Errorf("grant_type = %q", r.Form.Get("grant_type"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"at1","token_type":"Bearer","expires_in":3600}`))
	}))
	defer srv.Close()

	now := time.Unix(0, 0)
	c := &OAuth2{
		TokenURL:     srv.URL,
		RefreshToken: "rt",
		now:          func() time.Time { return now },
	}
	tok, err := c.Token(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if tok.Value != "at1" {
		t.Fatalf("token = %q", tok.Value)
	}
	// Second call within validity must be served from cache.
	if _, err := c.Token(context.Background()); err != nil {
		t.Fatal(err)
	}
	if hits != 1 {
		t.Fatalf("token endpoint hit %d times, want 1 (caching)", hits)
	}
	// Advance past expiry (minus early window) and expect a refresh.
	now = now.Add(2 * time.Hour)
	if _, err := c.Token(context.Background()); err != nil {
		t.Fatal(err)
	}
	if hits != 2 {
		t.Fatalf("expected refresh after expiry, hits=%d", hits)
	}
}

func TestOAuth2Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"invalid_grant","error_description":"expired"}`))
	}))
	defer srv.Close()

	c := &OAuth2{TokenURL: srv.URL, RefreshToken: "rt"}
	if _, err := c.Token(context.Background()); err == nil {
		t.Fatal("expected error from token endpoint")
	}
}
