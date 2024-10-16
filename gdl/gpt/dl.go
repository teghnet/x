package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

func getClient(config *oauth2.Config) *http.Client {
	// Token file stores the user's access and refresh tokens
	tokenFile := "token.json"

	tok, err := tokenFromFile(tokenFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokenFile, tok)
	}
	return config.Client(context.Background(), tok)
}

func downloadFileFromDrive(service *drive.Service, fileId, fileName string) error {
	// Create the file on disk
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	// Download the file from Google Drive
	res, err := service.Files.Get(fileId).Download()
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// Write the file content to the disk
	_, err = io.Copy(file, res.Body)
	return err
}

func mainDL() {
	// Load the OAuth credentials
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		fmt.Printf("Unable to read client secret file: %v", err)
		return
	}

	// Set up Google OAuth config
	config, err := google.ConfigFromJSON(b, drive.DriveReadonlyScope)
	if err != nil {
		fmt.Printf("Unable to parse client secret file to config: %v", err)
		return
	}

	client := getClient(config)

	// Create a Drive API service
	service, err := drive.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		fmt.Printf("Unable to retrieve Drive client: %v", err)
		return
	}

	// Download the file (replace with actual Google Drive file ID and desired local filename)
	fileId := "your-google-drive-file-id"
	fileName := "video.mp4"

	err = downloadFileFromDrive(service, fileId, fileName)
	if err != nil {
		fmt.Printf("Error downloading file: %v", err)
	} else {
		fmt.Println("Downloaded file successfully!")
	}
}
