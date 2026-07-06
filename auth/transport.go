package auth

import (
	"fmt"
	"net/http"
)

// Transport is an http.RoundTripper that attaches a credential's token to every
// outbound request as an Authorization header. It is the injection point
// between auth and client: build a Transport and pass it as client's base
// transport.
type Transport struct {
	// Credential supplies the token. Required.
	Credential Credential
	// Base is the underlying transport; nil uses http.DefaultTransport.
	Base http.RoundTripper
	// Header overrides the target header name; empty means "Authorization".
	Header string
}

// RoundTrip implements http.RoundTripper. It never mutates the caller's request:
// it operates on a shallow clone with a copied header map, per the
// RoundTripper contract.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.Credential == nil {
		return nil, fmt.Errorf("auth: transport has no credential")
	}
	tok, err := t.Credential.Token(req.Context())
	if err != nil {
		return nil, fmt.Errorf("auth: obtain token: %w", err)
	}

	clone := req.Clone(req.Context())
	name := t.Header
	if name == "" {
		name = "Authorization"
	}
	clone.Header.Set(name, tok.Header())

	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(clone)
}
