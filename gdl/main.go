package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"google.golang.org/api/youtube/v3"

	"github.com/teghnet/x/auth"
	"github.com/teghnet/x/conf"
	"github.com/teghnet/x/file"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

var scopes = []string{
	sheets.SpreadsheetsReadonlyScope,
	drive.DriveMetadataReadonlyScope,
	drive.DriveReadonlyScope,
}
var youtubeScopes = []string{
	youtube.YoutubeUploadScope,
	youtube.YoutubeReadonlyScope,
}
var clientSecretFile = "client_secret.json"

var stages = map[string]string{}

func run() error {
	stateDir, err := conf.StateDir("gdl", "local")
	if err != nil {
		return err
	}
	credentialsFile := path.Join(stateDir, clientSecretFile)
	client, err := auth.GoogleClient(credentialsFile, scopes)
	if err != nil {
		return fmt.Errorf("unable to create `auth.GoogleClient`: %w", err)
	}
	client2, err := auth.GoogleClient(credentialsFile, youtubeScopes)
	if err != nil {
		return fmt.Errorf("unable to create `auth.GoogleClient`: %w", err)
	}
	ctx := context.Background()
	// Create a Drive API service
	driveService, err := drive.NewService(ctx, option.WithHTTPClient(client), option.WithUserAgent(conf.UA()))
	if err != nil {
		return fmt.Errorf("unable to create `drive.Service`: %w", err)
	}
	// Create a Google Sheets service
	sheetsService, err := sheets.NewService(ctx, option.WithHTTPClient(client), option.WithUserAgent(conf.UA()))
	if err != nil {
		return fmt.Errorf("unable to create `sheets.Service`: %w", err)
	}
	// Create a YouTube API service
	youtubeService, err := youtube.NewService(context.Background(), option.WithHTTPClient(client2))
	if err != nil {
		return fmt.Errorf("unable to create YouTube service: %v", err)
	}
	// if _, err := listUserChannels(youtubeService); err != nil {
	// 	return err
	// }
	// if err := listUploadedVideos(youtubeService); err != nil {
	// 	return err
	// }

	if false {
		about, err := driveService.About.Get().Fields(
			"kind",
			"storageQuota",
			"driveThemes",
			"canCreateDrives",
			"importFormats",
			"exportFormats",
			"appInstalled",
			"user",
			"folderColorPalette",
			"maxImportSizes",
			"maxUploadSize",
		).Do()
		if err != nil {
			return err
		}
		var b strings.Builder
		encoder := json.NewEncoder(&b)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(about); err != nil {
			return err
		}
		fmt.Println(b.String())
		fmt.Println(sheetsService.BasePath, sheetsService.UserAgent)
	}

	if true {
		stage := "baltic"
		files, err := driveService.Files.
			List().
			// Q("mimeType='application/vnd.google-apps.spreadsheet' and trashed=false").
			// Q("trashed=false and sharedWithMe = true and (mimeType='application/vnd.google-apps.folder' or mimeType='application/vnd.google-apps.shortcut')").
			// Q("trashed=false and mimeType contains 'video/'").
			// Q("trashed=false").
			Q("trashed=false and '" + stages[stage] + "' in parents").
			Fields("nextPageToken, files(id,name,kind,mimeType,trashed,driveId,sharingUser,webContentLink,webViewLink,owners,size,sha256Checksum)").
			// Fields("nextPageToken, files").
			SupportsAllDrives(true).
			IncludeItemsFromAllDrives(true).
			OrderBy("modifiedTime desc,folder,name").
			Do()
		if err != nil {
			return err
		}

		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "   ")
		for _, f := range files.Files {
			if err := encoder.Encode(f); err != nil {
				return err
			}
			dlfile := path.Join(stateDir, "vid", f.Name)
			doDownload := true
			if fileExists(dlfile) {
				size, err := fileSize(dlfile)
				if err != nil {
					fmt.Println("error (size):", f.Name)
					continue
				}
				if size == 0 {
					fmt.Println("skip:", f.Name)
					continue
				}
				if size == f.Size {
					doDownload = false
				}
			}
			if doDownload {
				if err := download(dlfile, f.Id, f.Size, driveService); err != nil {
					return err
				}
			} else {
				fmt.Println("skip download:", f.Name)
			}
			title := strings.TrimSuffix(f.Name, "."+f.FullFileExtension) + " (" + stage + ")"
			if err := uploadVideoToYouTube(youtubeService, title, f.Description, dlfile); err != nil {
				return err
			}

			if err := os.Truncate(dlfile, 0); err != nil {
				return err
			}
		}
	}
	return nil
}

func fileSize(filePath string) (int64, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	return fileInfo.Size(), nil
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

func listUserChannels(service *youtube.Service) ([]*youtube.Channel, error) {
	call := service.Channels.List([]string{"snippet", "contentDetails", "statistics"}).
		Mine(true)
	response, err := call.Do()
	if err != nil {
		return nil, err
	}
	for i, channel := range response.Items {
		fmt.Printf("[%d] %s (ID: %s)\n", i+1, channel.Snippet.Title, channel.Id)
	}
	return response.Items, nil
}

func listVideos(service *youtube.Service) error {
	// Call the API to list uploads
	call := service.Videos.List([]string{
		"contentDetails",
		// "fileDetails",
		"id",
		"liveStreamingDetails",
		"localizations",
		"paidProductPlacementDetails",
		"player",
		// "processingDetails",
		"recordingDetails",
		"snippet",
		"statistics",
		"status",
		// "suggestions",
		"topicDetails",
	}).
		Id("")

	// Execute the API call
	response, err := call.Do()
	if err != nil {
		return err
	}

	// Display the video details
	fmt.Println("Uploaded Videos:")
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "   ")

	for _, item := range response.Items {
		if err := encoder.Encode(item); err != nil {
			return err
		}
		// fmt.Printf("%#v\n", item)
		// fmt.Printf("Title: %s\nVideo ID: %s\nDescription: %s\n\n", item.Snippet.Title, item.Id.VideoId, item.Snippet.Description)
	}
	return nil
}
func listUploadedVideos(service *youtube.Service) error {
	// Call the API to list uploads
	call := service.Search.List([]string{"snippet"}).
		ForMine(true).
		Type("video").
		Order("date").
		MaxResults(10)

	// Execute the API call
	response, err := call.Do()
	if err != nil {
		return err
	}

	// Display the video details
	fmt.Println("Uploaded Videos:")
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "   ")

	for _, item := range response.Items {
		if err := encoder.Encode(item); err != nil {
			return err
		}
		// fmt.Printf("%#v\n", item)
		// fmt.Printf("Title: %s\nVideo ID: %s\nDescription: %s\n\n", item.Snippet.Title, item.Id.VideoId, item.Snippet.Description)
	}
	return nil
}

func download(name, id string, size int64, driveService *drive.Service) error {
	// Create the file on disk
	nf, err := os.Create(name)
	if err != nil {
		return err
	}
	defer nf.Close()

	// Download the file from Google Drive
	res, err := driveService.Files.Get(id).Download()
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// Create a progress writer
	pw := &file.ProgressWriter{Total: size}

	// Write the file content to the disk and track progress
	_, err = io.Copy(io.MultiWriter(nf, pw), res.Body)
	if err != nil {
		fmt.Println("\nDownload failed!")
		return err
	}

	fmt.Println("\nDownload complete!")
	return nil
}

func uploadVideoToYouTube(service *youtube.Service, title, description, fileName string) error {
	video := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:       title,
			Description: description,
			CategoryId:  "28", // Category ID for Science & Technology
		},
		Status: &youtube.VideoStatus{PrivacyStatus: "private"},
	}

	// Open the video file
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	// Get file size for progress tracking
	fileInfo, err := f.Stat()
	if err != nil {
		return err
	}
	totalSize := fileInfo.Size()

	// Create a progress reader
	pr := &file.ProgressReader{Reader: f, Total: totalSize}

	// Upload the video
	call := service.Videos.Insert([]string{"snippet", "status"}, video)
	call.Media(pr)
	_, err = call.Do()
	fmt.Println("\nUpload complete!")
	return err
}
