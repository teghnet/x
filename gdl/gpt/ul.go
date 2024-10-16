package main

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

func uploadVideoToYouTube(service *youtube.Service, title, description, fileName string) error {
	video := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:       title,
			Description: description,
			CategoryId:  "22", // Category ID for People & Blogs
		},
		Status: &youtube.VideoStatus{PrivacyStatus: "public"},
	}

	// Open the video file for reading
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	// Upload the video
	call := service.Videos.Insert([]string{"snippet", "status"}, video)
	call.Media(file)
	_, err = call.Do()
	return err
}

func mainUL() {
	// Load OAuth credentials
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		fmt.Printf("Unable to read client secret file: %v", err)
		return
	}

	// Set up Google OAuth config
	config, err := google.ConfigFromJSON(b, youtube.YoutubeUploadScope)
	if err != nil {
		fmt.Printf("Unable to parse client secret file to config: %v", err)
		return
	}

	client := getClient(config)

	// Create a YouTube API service
	service, err := youtube.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		fmt.Printf("Unable to create YouTube service: %v", err)
		return
	}

	// Upload the video
	err = uploadVideoToYouTube(service, "My Video Title", "This is the description", "video.mp4")
	if err != nil {
		fmt.Printf("Error uploading video: %v", err)
	} else {
		fmt.Println("Video uploaded successfully!")
	}
}
