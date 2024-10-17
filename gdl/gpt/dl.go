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

// ProgressWriter tracks the progress of the download or upload
type ProgressWriter struct {
	Total    int64
	Progress int64
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.Progress += int64(n)
	fmt.Printf("\rDownloading: %.2f%% complete", float64(pw.Progress)/float64(pw.Total)*100)
	return n, nil
}

// getClient handles authentication
func getClient(config *oauth2.Config) *http.Client {
	tokenFile := "token.json"

	tok, err := tokenFromFile(tokenFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokenFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// downloadFileFromDrive downloads a file from Google Drive
func downloadFileFromDrive(service *drive.Service, fileId, fileName string) error {
	// Get the file metadata to obtain the size
	fileInfo, err := service.Files.Get(fileId).Fields("size").Do()
	if err != nil {
		return err
	}

	totalSize := fileInfo.Size

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

	// Create a progress writer
	pw := &ProgressWriter{Total: totalSize}

	// Write the file content to the disk and track progress
	_, err = io.Copy(io.MultiWriter(file, pw), res.Body)
	fmt.Println("\nDownload complete!")
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

	// Download the file
	fileId := "your-google-drive-file-id"
	fileName := "video.mp4"

	err = downloadFileFromDrive(service, fileId, fileName)
	if err != nil {
		fmt.Printf("Error downloading file: %v", err)
	} else {
		fmt.Println("Downloaded file successfully!")
	}
}
