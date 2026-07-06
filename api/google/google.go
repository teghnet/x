// Package google is a thin client for a couple of Google HTTP APIs.
//
// It is an illustration of the api/* contract: it depends on client for
// transport (which carries authentication) and on parse/json for decoding, and
// it knows nothing about how credentials are stored or refreshed. Swapping the
// credential behind the client requires no change here.
package google

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	pjson "github.com/teghnet/x/parse/json"
)

// Doer is the transport contract this package needs. *client.Client satisfies
// it; accepting the interface keeps this package decoupled from client's
// concrete type and trivial to test with a fake.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Service talks to Google APIs over the injected Doer.
type Service struct {
	// HTTP performs requests; typically a *client.Client. Required.
	HTTP Doer
	// BaseURL overrides the API host, mainly for testing. Empty uses the
	// production endpoints.
	BaseURL string
}

// New returns a Service using the given transport.
func New(h Doer) *Service { return &Service{HTTP: h} }

// UserInfo is the subset of the OpenID Connect userinfo response we expose.
type UserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

// UserInfo fetches the profile of the authenticated user from the OIDC userinfo
// endpoint. Authentication is supplied by the underlying transport.
func (s *Service) UserInfo(ctx context.Context) (UserInfo, error) {
	base := s.hostOr("https://openidconnect.googleapis.com")
	return get[UserInfo](ctx, s, base+"/v1/userinfo")
}

// TokenInfo queries Google's tokeninfo endpoint for the given access token. It
// is a convenience for debugging scopes and audiences.
func (s *Service) TokenInfo(ctx context.Context, accessToken string) (map[string]any, error) {
	base := s.hostOr("https://oauth2.googleapis.com")
	u := base + "/tokeninfo?" + url.Values{"access_token": {accessToken}}.Encode()
	return get[map[string]any](ctx, s, u)
}

func (s *Service) hostOr(prod string) string {
	if s.BaseURL != "" {
		return s.BaseURL
	}
	return prod
}

// get performs a GET and decodes a successful JSON response into a value of
// type T. Decoding is routed through parse/json to respect the parse/* boundary.
func get[T any](ctx context.Context, s *Service, url string) (T, error) {
	var zero T
	if s.HTTP == nil {
		return zero, fmt.Errorf("google: no transport configured")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return zero, fmt.Errorf("google: build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := s.HTTP.Do(req)
	if err != nil {
		return zero, fmt.Errorf("google: request %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return zero, fmt.Errorf("google: %s: unexpected status %d: %s", url, resp.StatusCode, body)
	}
	v, err := pjson.Decode[T](resp.Body)
	if err != nil {
		return zero, fmt.Errorf("google: decode %s: %w", url, err)
	}
	return v, nil
}
