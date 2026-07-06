package google

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/teghnet/x/auth"
	"github.com/teghnet/x/client"
)

func TestUserInfo(t *testing.T) {
	var gotAuth, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"sub":"123","email":"a@b.com","email_verified":true,"name":"Ada"}`))
	}))
	defer srv.Close()

	// Wire a real client with auth injection to prove api/* stays credential-agnostic.
	c := client.New(client.WithBaseTransport(&auth.Transport{Credential: auth.StaticToken{Value: "tok"}}))
	svc := &Service{HTTP: c, BaseURL: srv.URL}

	info, err := svc.UserInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if info.Sub != "123" || info.Email != "a@b.com" || !info.EmailVerified || info.Name != "Ada" {
		t.Fatalf("got %+v", info)
	}
	if gotAuth != "Bearer tok" {
		t.Fatalf("auth header = %q", gotAuth)
	}
	if gotPath != "/v1/userinfo" {
		t.Fatalf("path = %q", gotPath)
	}
}

func TestUserInfoErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	defer srv.Close()

	svc := &Service{HTTP: srv.Client(), BaseURL: srv.URL}
	if _, err := svc.UserInfo(context.Background()); err == nil {
		t.Fatal("expected error on 401")
	}
}

func TestNoTransport(t *testing.T) {
	svc := &Service{}
	if _, err := svc.UserInfo(context.Background()); err == nil {
		t.Fatal("expected error with no transport")
	}
}
