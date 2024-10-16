package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"path"
	"strings"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"

	"github.com/teghnet/x/auth"
	"github.com/teghnet/x/conf"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

var scopes = []string{
	sheets.SpreadsheetsReadonlyScope,
	drive.DriveMetadataReadonlyScope,
}
var clientSecretFile = "client_secret.json"

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
	ctx := context.Background()
	driveService, err := drive.NewService(ctx, option.WithHTTPClient(client), option.WithUserAgent(conf.UA()))
	if err != nil {
		return fmt.Errorf("unable to create `drive.Service`: %w", err)
	}
	sheetsService, err := sheets.NewService(ctx, option.WithHTTPClient(client), option.WithUserAgent(conf.UA()))
	if err != nil {
		return fmt.Errorf("unable to create `sheets.Service`: %w", err)
	}

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
		files, err := driveService.Files.
			List().
			// Q("mimeType='application/vnd.google-apps.spreadsheet' and trashed=false").
			// Q("trashed=false and sharedWithMe = true and (mimeType='application/vnd.google-apps.folder' or mimeType='application/vnd.google-apps.shortcut')").
			// Q("trashed=false and (mimeType = 'application/vnd.google-apps.folder' and parents contains '1pbEUiCesq_mQbcAeIh4WrLHTG-GIJp3o')").
			// Q("trashed=false and mimeType contains 'video/'").
			Q("trashed=false").
			// Q("trashed=false and '1qqO5TpjfUXCQ_OjmK5SDnFipq7sLXkfh' in parents").
			// Fields("nextPageToken, files(id,name,kind,mimeType,trashed,driveId,sharingUser,webContentLink,webViewLink,owners)").
			Fields("nextPageToken, files").
			SupportsAllDrives(true).
			IncludeItemsFromAllDrives(true).
			OrderBy("modifiedTime desc,folder,name").
			Do()
		if err != nil {
			return err
		}
		for _, f := range files.Files {
			var b strings.Builder
			encoder := json.NewEncoder(&b)
			encoder.SetIndent("", "   ")
			if err := encoder.Encode(f); err != nil {
				return err
			}
			fmt.Println(b.String())
		}
	}
	return nil
}
