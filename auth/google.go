package auth

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"slices"
	"strings"
	"sync"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func GoogleClient(credentialsFile string, scopes []string) (*http.Client, error) {
	b, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read credentials: %w", err)
	}
	slices.Sort(scopes)
	scopes = slices.Compact(scopes)
	config, err := google.ConfigFromJSON(b, scopes...)
	if err != nil {
		return nil, fmt.Errorf("unable to create config: %w", err)
	}
	sum256 := sha256.Sum256([]byte(strings.Join(append(scopes, config.ClientID), "|")))
	return getClient(path.Join(path.Dir(credentialsFile), fmt.Sprintf("token_%x.json", sum256)), config)
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(tokFile string, config *oauth2.Config) (*http.Client, error) {
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		log.Println("tokenFromFile error:", err)
		tok, err = tokenFromWeb(config)
		if err != nil {
			return nil, err
		}
		if err := saveToken(tokFile, tok); err != nil {
			return nil, err
		}
	}
	return config.Client(context.Background(), tok), nil
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
func tokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	log.Printf("Go to the following link in your browser then type the authorization code: \n\n%v\n\n", authURL)
	openBrowser(authURL)

	tok, err := config.Exchange(context.TODO(), waitForCode())
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from web: %w", err)
	}
	return tok, nil
}

// Saves a token to a file path (create if not exists and truncate if exists).
func saveToken(path string, token *oauth2.Token) error {
	log.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to cache oauth token: %w", err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(token)
}

func waitForCode() string {
	var code string

	var wg sync.WaitGroup
	wg.Add(1)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer wg.Done()
		code = r.FormValue("code")
		log.Println("code:", code)
		log.Println("state:", r.FormValue("state"))
		log.Println("scope:", r.FormValue("scope"))
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("ok"))
		if err != nil {
			log.Println("write error:", err)
		}
	})

	server := &http.Server{Addr: ":8080"}
	go func() { log.Printf("Server stopped: %v", server.ListenAndServe()) }()

	wg.Wait()
	log.Printf("Shutting down: %v", server.Shutdown(context.Background()))
	return code
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}
