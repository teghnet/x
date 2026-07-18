package gsheets

import (
	"net/url"
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
