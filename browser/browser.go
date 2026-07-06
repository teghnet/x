// Package browser provides a lightweight stateful HTTP session for scraping.
//
// A Session keeps cookies across requests (via net/http/cookiejar), performs
// fetches through the shared client transport, and returns pages that expose
// the parse/html extraction helpers. It sits above client and parse/html and
// deliberately holds only session state, not credentials or API knowledge.
package browser

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/teghnet/x/parse/html"
)

// Doer is the transport contract the session needs; *client.Client satisfies it.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Session is a stateful browsing context. It is safe for sequential use; the
// underlying cookie jar is safe for concurrent use, but callers coordinating
// parallel navigations should serialize on their own state.
type Session struct {
	http      Doer
	jar       http.CookieJar
	userAgent string
}

// Option configures a Session.
type Option func(*Session)

// WithUserAgent sets the User-Agent header sent on every request.
func WithUserAgent(ua string) Option {
	return func(s *Session) { s.userAgent = ua }
}

// New creates a Session that issues requests through h. A cookie jar is created
// automatically so that Set-Cookie responses persist across navigations.
func New(h Doer, opts ...Option) (*Session, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("browser: new cookie jar: %w", err)
	}
	s := &Session{http: h, jar: jar, userAgent: "ctld-browser/1.0"}
	for _, opt := range opts {
		opt(s)
	}
	return s, nil
}

// Page is a fetched document plus helpers over its HTML content.
type Page struct {
	URL        *url.URL
	StatusCode int
	Body       []byte
}

// Title returns the document's <title>.
func (p *Page) Title() string { return html.Title(p.Body) }

// Text returns the visible text of the document.
func (p *Page) Text() string { return html.Text(p.Body) }

// Links returns every anchor href resolved against the page URL. Unparseable
// hrefs are skipped.
func (p *Page) Links() []string {
	raw := html.Links(p.Body)
	out := make([]string, 0, len(raw))
	for _, href := range raw {
		if p.URL == nil {
			out = append(out, href)
			continue
		}
		if u, err := p.URL.Parse(href); err == nil {
			out = append(out, u.String())
		}
	}
	return out
}

// Get fetches rawURL and returns the resulting Page. Cookies from the response
// are stored in the session jar and replayed on subsequent requests.
func (s *Session) Get(ctx context.Context, rawURL string) (*Page, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("browser: build request: %w", err)
	}
	return s.do(req)
}

func (s *Session) do(req *http.Request) (*Page, error) {
	if s.userAgent != "" {
		req.Header.Set("User-Agent", s.userAgent)
	}
	// Attach stored cookies for this URL.
	for _, c := range s.jar.Cookies(req.URL) {
		req.AddCookie(c)
	}

	resp, err := s.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("browser: get %s: %w", req.URL, err)
	}
	defer resp.Body.Close()

	// Persist any Set-Cookie headers.
	if cookies := resp.Cookies(); len(cookies) > 0 {
		s.jar.SetCookies(req.URL, cookies)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("browser: read body %s: %w", req.URL, err)
	}
	// Prefer the final request URL (after redirects) when the transport records
	// it; fall back to the request URL otherwise.
	pageURL := req.URL
	if resp.Request != nil && resp.Request.URL != nil {
		pageURL = resp.Request.URL
	}
	return &Page{URL: pageURL, StatusCode: resp.StatusCode, Body: body}, nil
}

// Cookies returns the cookies the session would send to rawURL. It is useful
// for inspection and tests.
func (s *Session) Cookies(rawURL string) ([]*http.Cookie, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("browser: parse url: %w", err)
	}
	return s.jar.Cookies(u), nil
}
