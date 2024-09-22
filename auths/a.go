package auths

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sync"

	"golang.org/x/oauth2"
)

var (
	tokFile  string
	credFile string
	url      string

	spreadsheetId string
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	//TODO: check if on-line access type is enough
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	// fmt.Printf("Go to the following link in your browser then type the "+
	// 	"authorization code: \n%v\n", authURL)
	openBrowser(authURL)

	// var authCode string
	// if _, err := fmt.Scan(&authCode); err != nil {
	// 	log.Fatalf("Unable to read authorization code: %v", err)
	// }

	tok, err := config.Exchange(context.TODO(), waitForCode())
	if err != nil {
		log.Fatalf("unable to retrieve token from web: %s", err)
	}

	// tok, err := config.Exchange(context.TODO(), authCode)
	// if err != nil {
	// 	log.Fatalf("Unable to retrieve token from web: %v", err)
	// }
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	log.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
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
		w.Write([]byte("ok"))
	})

	server := &http.Server{Addr: ":8080"}
	go func() { log.Printf("Server stopped: %v", server.ListenAndServe()) }()
	wg.Wait()

	log.Printf("Shutting down: %v", server.Shutdown(context.Background()))
	return code
}
