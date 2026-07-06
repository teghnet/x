package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// OAuth2 is a Credential backed by the OAuth 2.0 refresh-token flow. It caches
// the current access token and transparently refreshes it shortly before
// expiry. It is safe for concurrent use.
//
// It talks to the token endpoint with a plain *http.Client and does not depend
// on the client package, avoiding an import cycle (client depends on auth).
type OAuth2 struct {
	// TokenURL is the OAuth2 token endpoint. Required.
	TokenURL string
	// ClientID and ClientSecret identify the application.
	ClientID     string
	ClientSecret string
	// RefreshToken is exchanged for access tokens.
	RefreshToken string
	// Scopes, if set, are requested on refresh.
	Scopes []string

	// HTTPClient performs the token request; nil uses http.DefaultClient.
	HTTPClient *http.Client
	// EarlyExpiry refreshes this long before the reported expiry. Default 30s.
	EarlyExpiry time.Duration
	// now is an injectable clock for testing; nil uses time.Now.
	now func() time.Time

	mu  sync.Mutex
	tok Token
}

func (c *OAuth2) clock() time.Time {
	if c.now != nil {
		return c.now()
	}
	return time.Now()
}

// Token implements Credential, returning a cached token when still valid and
// refreshing otherwise.
func (c *OAuth2) Token(ctx context.Context) (Token, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	early := c.EarlyExpiry
	if early <= 0 {
		early = 30 * time.Second
	}
	// Treat the token as expired `early` ahead of its real expiry.
	if c.tok.Valid(c.clock().Add(early)) {
		return c.tok, nil
	}
	tok, err := c.refresh(ctx)
	if err != nil {
		return Token{}, err
	}
	c.tok = tok
	return tok, nil
}

// tokenResponse is the subset of the RFC 6749 token endpoint response we use.
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Error        string `json:"error"`
	ErrorDesc    string `json:"error_description"`
}

func (c *OAuth2) refresh(ctx context.Context) (Token, error) {
	if c.TokenURL == "" {
		return Token{}, fmt.Errorf("auth: oauth2: TokenURL is empty")
	}
	form := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {c.RefreshToken},
	}
	if c.ClientID != "" {
		form.Set("client_id", c.ClientID)
	}
	if c.ClientSecret != "" {
		form.Set("client_secret", c.ClientSecret)
	}
	if len(c.Scopes) > 0 {
		form.Set("scope", strings.Join(c.Scopes, " "))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return Token{}, fmt.Errorf("auth: oauth2: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	hc := c.HTTPClient
	if hc == nil {
		hc = http.DefaultClient
	}
	resp, err := hc.Do(req)
	if err != nil {
		return Token{}, fmt.Errorf("auth: oauth2: token request: %w", err)
	}
	defer resp.Body.Close()

	var tr tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return Token{}, fmt.Errorf("auth: oauth2: decode token response (status %d): %w", resp.StatusCode, err)
	}
	if resp.StatusCode/100 != 2 || tr.Error != "" {
		msg := tr.Error
		if tr.ErrorDesc != "" {
			msg += ": " + tr.ErrorDesc
		}
		if msg == "" {
			msg = resp.Status
		}
		return Token{}, fmt.Errorf("auth: oauth2: token endpoint error: %s", msg)
	}
	if tr.AccessToken == "" {
		return Token{}, fmt.Errorf("auth: oauth2: empty access token in response")
	}

	tok := Token{Value: tr.AccessToken, Type: tr.TokenType}
	if tok.Type == "" {
		tok.Type = "Bearer"
	}
	if tr.ExpiresIn > 0 {
		tok.Expiry = c.clock().Add(time.Duration(tr.ExpiresIn) * time.Second)
	}
	// A rotated refresh token replaces the old one for subsequent refreshes.
	if tr.RefreshToken != "" {
		c.RefreshToken = tr.RefreshToken
	}
	return tok, nil
}
