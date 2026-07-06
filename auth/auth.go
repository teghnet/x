// Package auth models credentials and injects them into outbound HTTP requests.
//
// The core abstractions are:
//
//   - Token: an access token with an optional expiry.
//   - Credential: something that can produce a (possibly refreshed) Token.
//   - Provider: a source of Credentials, e.g. keyed by account.
//   - Transport: an http.RoundTripper that attaches a credential's token to
//     every request.
//
// The Transport is how auth is injected into the client package: callers build
// a Transport here and hand it to client as the base round tripper. Consumers
// such as api/* therefore never learn how credentials are obtained or stored.
package auth

import (
	"context"
	"errors"
	"time"
)

// ErrNoCredential indicates a Provider has no credential to offer.
var ErrNoCredential = errors.New("auth: no credential")

// Token is an access token. A zero Expiry means the token does not expire.
type Token struct {
	Value  string
	Type   string // e.g. "Bearer"; defaults to "Bearer" when used in a header
	Expiry time.Time
}

// Valid reports whether the token is non-empty and not expired as of now.
func (t Token) Valid(now time.Time) bool {
	if t.Value == "" {
		return false
	}
	return t.Expiry.IsZero() || now.Before(t.Expiry)
}

// Header returns the value for an Authorization header, e.g. "Bearer abc".
func (t Token) Header() string {
	scheme := t.Type
	if scheme == "" {
		scheme = "Bearer"
	}
	return scheme + " " + t.Value
}

// Credential produces access tokens, refreshing as needed. Implementations must
// be safe for concurrent use.
type Credential interface {
	// Token returns a currently-valid access token, refreshing if necessary.
	Token(ctx context.Context) (Token, error)
}

// Provider is a source of credentials. A daemon may hold one Provider and ask
// it for the credential appropriate to a given account or service.
type Provider interface {
	// Credential returns the credential for the given account. An empty account
	// selects the default credential. It returns ErrNoCredential if none exists.
	Credential(ctx context.Context, account string) (Credential, error)
}

// StaticToken is a Credential that always returns the same token.
type StaticToken Token

// Token implements Credential.
func (s StaticToken) Token(context.Context) (Token, error) {
	return Token(s), nil
}

// CredentialFunc adapts a function into a Credential.
type CredentialFunc func(ctx context.Context) (Token, error)

// Token implements Credential.
func (f CredentialFunc) Token(ctx context.Context) (Token, error) { return f(ctx) }

// StaticProvider serves a fixed set of credentials keyed by account, with an
// optional default (empty-account) credential.
type StaticProvider struct {
	Default  Credential
	Accounts map[string]Credential
}

// Credential implements Provider.
func (p StaticProvider) Credential(_ context.Context, account string) (Credential, error) {
	if account == "" {
		if p.Default != nil {
			return p.Default, nil
		}
		return nil, ErrNoCredential
	}
	if c, ok := p.Accounts[account]; ok {
		return c, nil
	}
	return nil, ErrNoCredential
}
