package gsheets

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"google.golang.org/api/sheets/v4"
)

var ScopesAll = []string{
	sheets.DriveScope,
	sheets.DriveFileScope,
	sheets.DriveReadonlyScope,
	sheets.SpreadsheetsScope,
	sheets.SpreadsheetsReadonlyScope,
}

const mimeSheet = "application/vnd.google-apps.spreadsheet"

func ParseSheetLink(link string) (Sheet, error) {
	u, err := url.Parse(link)
	if err != nil {
		return Sheet{}, err
	}
	if u.Path == link {
		return Sheet{ID: link}, nil
	}

	var s Sheet

	// Extract ID
	if id := u.Query().Get("id"); id != "" {
		s.ID = id
	} else if elems := strings.Split(u.Path, "/"); elems[len(elems)-1] == "edit" {
		s.ID = elems[len(elems)-2]
	}

	// Extract GID
	if gid, err := strconv.ParseInt(u.Query().Get("gid"), 10, 64); err == nil {
		s.GID = gid
	} else if strings.Contains(u.Fragment, "gid=") {
		query, err := url.ParseQuery(u.Fragment)
		if err == nil {
			if gid, err := strconv.ParseInt(query.Get("gid"), 10, 64); err == nil {
				s.GID = gid
			}
		}
	}

	return s, nil
}

// https://docs.google.com/spreadsheets/d/1nT5bZmaCv4X6iJc36BxaJVdoQ3o3QHai/edit?gid=699980524#gid=699980524
func NormalizeID(spreadsheetID string) string {
	u, err := url.Parse(spreadsheetID)
	if err != nil {
		return spreadsheetID
	}
	if u.Path == spreadsheetID {
		return spreadsheetID
	}
	if elems := strings.Split(u.Path, "/"); elems[len(elems)-1] == "edit" {
		return elems[len(elems)-2]
	}
	if id := u.Query().Get("id"); id != "" {
		return id
	}
	return spreadsheetID
}

// FreezeHeaderRow freezes the first row of the default sheet (sheetId 0)
// so the header remains visible when scrolling through large datasets.
func FreezeHeaderRow(ctx context.Context, srv *sheets.Service, spreadsheetID string) error {
	req := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				UpdateSheetProperties: &sheets.UpdateSheetPropertiesRequest{
					Properties: &sheets.SheetProperties{
						SheetId: 0,
						GridProperties: &sheets.GridProperties{
							FrozenRowCount: 1,
						},
					},
					Fields: "gridProperties.frozenRowCount",
				},
			},
		},
	}
	_, err := srv.Spreadsheets.BatchUpdate(spreadsheetID, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("spreadsheets.batchUpdate (freeze row): %w", err)
	}
	return nil
}

// AppendRows appends rows to Sheet1 of the given spreadsheet using the
// Sheets Append API, making data visible incrementally.
func AppendRows(ctx context.Context, srv *sheets.Service, spreadsheetID string, rows [][]any) error {
	vr := &sheets.ValueRange{Values: rows}
	_, err := srv.Spreadsheets.Values.Append(spreadsheetID, "Sheet1", vr).
		ValueInputOption("USER_ENTERED").
		InsertDataOption("INSERT_ROWS").
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("spreadsheets.values.append: %w", err)
	}
	return nil
}
