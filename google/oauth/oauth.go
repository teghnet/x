package oauth

import (
	oauth2v2 "google.golang.org/api/oauth2/v2"
)

var ScopesAll = []string{
	oauth2v2.OpenIDScope,
	oauth2v2.UserinfoEmailScope,
	oauth2v2.UserinfoProfileScope,
}
