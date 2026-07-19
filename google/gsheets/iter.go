package gsheets

import (
	"context"
	"fmt"
	"iter"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/sheets/v4"
)

// Spreadsheets returns all spreadsheet files in the given drive.
func Spreadsheets(ctx context.Context, ds *drive.Service, driveID string) iter.Seq2[*drive.File, error] {
	return func(yield func(*drive.File, error) bool) {
		var pageToken string
		for {
			call := ds.Files.List().Context(ctx).
				Q(fmt.Sprintf("mimeType = '%s' and trashed = false", mimeSheet)).
				Fields("nextPageToken, files(id, name)").
				PageSize(100)
			if driveID == "root" {
				call = call.Corpora("user")
			} else {
				call = call.Corpora("drive").
					SupportsAllDrives(true).
					IncludeItemsFromAllDrives(true).
					DriveId(driveID)
			}

			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			result, err := call.Do()
			if err != nil && !yield(nil, fmt.Errorf("files.list: %w", err)) {
				return
			}
			for _, f := range result.Files {
				if !yield(f, nil) {
					return
				}
			}
			if pageToken = result.NextPageToken; pageToken == "" {
				return
			}
		}
	}
}
func Rows(ctx context.Context, ss *sheets.Service, spreadsheetID, sheetRange string) iter.Seq2[[]any, error] {
	return func(yield func([]any, error) bool) {
		resp, err := ss.Spreadsheets.Values.Get(spreadsheetID, sheetRange).Context(ctx).
			ValueRenderOption("UNFORMATTED_VALUE").
			Do()
		if err != nil {
			yield(nil, fmt.Errorf("spreadsheets.values.get: %w", err))
			return
		}
		if len(resp.Values) == 0 {
			return
		}
		for _, row := range resp.Values {
			// var r []string
			// for _, cell := range row {
			// 	r = append(r, cell.(string))
			// }
			if !yield(row, nil) {
				return
			}
		}
	}
}
