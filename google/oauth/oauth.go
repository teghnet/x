package oauth

import (
	"net/http"

	"golang.org/x/oauth2"
	oauth2v2 "google.golang.org/api/oauth2/v2"

	"github.com/teghnet/x/transport"
)

var ScopesAll = []string{
	oauth2v2.OpenIDScope,
	oauth2v2.UserinfoEmailScope,
	oauth2v2.UserinfoProfileScope,
}

func RequestMutator(ts oauth2.TokenSource) transport.RequestMutator {
	return &oAuth2TokenSource{ts: ts}
}

type oAuth2TokenSource struct {
	ts oauth2.TokenSource
}

// ApplyTo sets the Authorization header on the given request.
// Implements [transport.RequestMutator].
func (ts oAuth2TokenSource) ApplyTo(r *http.Request) error {
	token, err := ts.ts.Token()
	if err != nil {
		return err
	}
	token.SetAuthHeader(r)
	return nil
}
