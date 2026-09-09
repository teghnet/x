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

// Middleware returns a [transport.Middleware] that sets the Authorization
// header on each outgoing request from a token minted by ts. Because
// [transport.Retry] re-applies the chain's middleware on every attempt,
// wrapping this in a Transport with retries enabled re-mints the token on
// each retry rather than reusing one that may have expired.
func Middleware(ts oauth2.TokenSource) transport.Middleware {
	return transport.MutateRequest(func(r *http.Request) error {
		token, err := ts.Token()
		if err != nil {
			return err
		}
		token.SetAuthHeader(r)
		return nil
	})
}
