package google

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"charm.land/log/v2"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"

	"github.com/teghnet/x/paths"
)

const clientSecretFilePattern = "client_secret_*.apps.googleusercontent.com.json"
const tokenFilePattern = "token_%x.json"
const tokenFilePattern2 = "%s_" + tokenFilePattern

func TokenSource(ctx context.Context, xdg paths.XDG, scope []string) (oauth2.TokenSource, error) {
	configPath := xdg.ConfigPath(clientSecretFilePattern)

	// 1) Find JWT config if present
	jwtConfig, err := getJWTConfig(configPath, scope...)
	if err == nil {
		return jwtConfig.TokenSource(ctx), nil
	}
	log.Warnf("JWT config: %v", err)

	// 2) Use OAuth2 client secret flow with persisted token
	oauth2Config, err := getOAuth2Config(configPath, scope...)
	if err == nil {
		sum256 := sha256.Sum256([]byte(strings.Join(append(scope, oauth2Config.ClientID), string([]byte{0}))))
		tokenFile := fmt.Sprintf(tokenFilePattern, sum256)
		if splits := strings.Split(oauth2Config.ClientID, "."); len(splits) > 0 {
			tokenFile = fmt.Sprintf(tokenFilePattern2, splits[0], sum256)
		}
		tokenPath := xdg.StatePath(tokenFile)
		t, err := tokenFromFile(tokenPath)
		if err != nil {
			t, err = tokenFromWeb(ctx, oauth2Config)
			if err := saveToken(tokenPath, t); err != nil {
				log.Errorf("could not save token: %v", err)
				// still usable, no error
			}
		}
		if err == nil {
			return &savingTokenSource{
				path: tokenPath,
				src:  oauth2.ReuseTokenSource(t, oauth2Config.TokenSource(ctx, t)),
				last: t,
			}, nil
		}
	}
	log.Warnf("OAuth2 config: %v", err)

	// 3) Try Application Default Credentials (ADC) first
	return google.DefaultTokenSource(ctx, scope...)
}

func getJWTConfig(pat string, scope ...string) (*jwt.Config, error) {
	matches, err := filepath.Glob(pat)
	if err != nil {
		return nil, fmt.Errorf("unable to read app dir: %v", err)
	}
	for _, match := range matches {
		b, err := os.ReadFile(match)
		if err != nil {
			return nil, fmt.Errorf("unable to read client secret file: %v", err)
		}
		config, err := google.JWTConfigFromJSON(b, scope...)
		if err != nil {
			continue
		}
		log.Print("JWT Config", config.PrivateKeyID)
		return config, nil
	}
	return nil, fmt.Errorf("unable to find client secret file: %v", err)
}

func getOAuth2Config(pat string, scopes ...string) (*oauth2.Config, error) {
	matches, err := filepath.Glob(pat)
	if err != nil {
		return nil, fmt.Errorf("unable to read app dir: %v", err)
	}
	for _, match := range matches {
		b, err := os.ReadFile(match)
		if err != nil {
			return nil, fmt.Errorf("unable to read client secret file: %v", err)
		}
		config, err := google.ConfigFromJSON(b, scopes...)
		if err != nil {
			log.Warn("Unable to parse client secret file to config", "err", err)
			continue
		}
		log.Info("OAuth2 Config", "clientID", config.ClientID)
		return config, nil
	}
	return nil, fmt.Errorf("unable to find client secret file: %v", err)
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	return tok, json.NewDecoder(f).Decode(tok)
}

// Request a token from the web, then returns the retrieved token.
func tokenFromWeb(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	config.RedirectURL = "http://localhost:8080"

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	log.Printf("Go to the following link in your browser then type the authorization code: \n\n%v\n\n", authURL)
	if err := openBrowser(authURL); err != nil {
		return nil, err
	}

	tok, err := config.Exchange(ctx, waitForCode(ctx, 8080), oauth2.AccessTypeOffline)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from web: %w", err)
	}
	return tok, nil
}

// Saves a token to a file path (create if not exists and truncate if exists).
func saveToken(path string, token *oauth2.Token) error {
	log.Info("Saving credential file", "path", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to cache oauth token: %w", err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(token)
}

func waitForCode(ctx context.Context, port int) string {
	ctx, cancel := context.WithCancel(ctx)

	var code string
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer cancel()
		code = r.FormValue("code")
		log.Print("code:", code)
		log.Print("state:", r.FormValue("state"))
		log.Print("scope:", r.FormValue("scope"))
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("ok"))
		if err != nil {
			log.Printf("Write error: %v", err)
		}
	})

	server := &http.Server{Addr: fmt.Sprintf(":%d", port)}

	go func() {
		defer cancel()
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Server error: %v", err)
			return
		}
		log.Printf("Server stopped.")
	}()

	<-ctx.Done()

	log.Print("Shutting down server.")
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
	return code
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	}
	return fmt.Errorf("unsupported platform")
}

// savingTokenSource wraps an oauth2.TokenSource and persists tokens to a file whenever a token is retrieved/refreshed.
type savingTokenSource struct {
	src oauth2.TokenSource

	last *oauth2.Token
	path string
	mu   sync.Mutex
}

func (s *savingTokenSource) Token() (*oauth2.Token, error) {
	tok, err := s.src.Token()
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.last == nil || !tokensEqual(s.last, tok) {
		if err := saveToken(s.path, tok); err != nil {
			// Don't fail token retrieval if saving fails; log and proceed.
			log.Printf("unable to save refreshed oauth token: %v", err)
		}
		// store a shallow copy as last
		cpy := *tok
		s.last = &cpy
	}
	return tok, nil
}

// tokensEqual determines if two tokens are equivalent for persistence purposes.
// We consider access token, refresh token, and expiry time.
func tokensEqual(a, b *oauth2.Token) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.AccessToken != b.AccessToken {
		return false
	}
	if a.RefreshToken != b.RefreshToken {
		return false
	}
	// Compare expiry times; zero times should be treated as equal if both zero
	if a.Expiry.IsZero() && b.Expiry.IsZero() {
		return true
	}
	return a.Expiry.Equal(b.Expiry)
}
