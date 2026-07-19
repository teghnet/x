package google

import (
	"context"
	"maps"
	"net/http"
	"slices"

	"github.com/teghnet/x/google/gdrive"
	"github.com/teghnet/x/google/gsheets"
	"github.com/teghnet/x/google/oauth"
	"github.com/teghnet/x/paths"
	"github.com/teghnet/x/transport"
)

func fullScope() []string {
	scopes := make(map[string]struct{})
	for _, s := range gsheets.ScopesAll {
		scopes[s] = struct{}{}
	}
	for _, s := range oauth.ScopesAll {
		scopes[s] = struct{}{}
	}
	for _, s := range gdrive.ScopeFull {
		scopes[s] = struct{}{}
	}
	return slices.Sorted(maps.Keys(scopes))
}

// Client returns an [http.Client] that uses the provided context and credentials to authenticate all requests.
// If no credentials are provided, the default credentials from the environment are used.
func Client(ctx context.Context, xdg paths.XDG, scope ...string) (*http.Client, error) {
	if len(scope) == 0 {
		scope = fullScope()
	}
	ts, err := TokenSource(ctx, xdg, scope)
	if err != nil {
		return nil, err
	}
	return &http.Client{
		Transport: transport.New(
			transport.WithRequestMutator(oauth.RequestMutator(ts)),
		),
	}, nil
}
