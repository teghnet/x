package gdrive

import (
	"context"
	"fmt"

	"charm.land/log/v2"
	"google.golang.org/api/drive/v3"
)

// ListSharedDrives lists all shared drives the user can access and returns
// a map from drive ID to drive name for use when annotating files.
// Deprecated: use [SharedDrives] instead.
func ListSharedDrives(ctx context.Context, srv *drive.Service) (map[string]string, error) {
	names := make(map[string]string)
	var pageToken string
	for {
		call := srv.Drives.List().PageSize(100).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		result, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf("drives.list: %w", err)
		}

		for _, d := range result.Drives {
			names[d.Id] = d.Name
			log.Print("shared drive",
				"id", d.Id,
				"name", d.Name,
				"link", fmt.Sprintf("https://drive.google.com/drive/folders/%s", d.Id))
		}

		pageToken = result.NextPageToken
		if pageToken == "" {
			break
		}
	}
	return names, nil
}
