package gsheets

// TODO: Add more types as needed for specific Google Sheets interactions.
// For example, configuration for specific sheet ranges, or custom request types.

// Sheet represents a single sheet within a Google Spreadsheet.
type Sheet struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"name"`
	GID   int64  `json:"gid,omitempty"`
	Title string `json:"title,omitempty"`
}

// SheetField represents a field in a Google Sheet.
type SheetField struct {
	Header string
	Value  any
}

type SheetInfo struct {
	URL             string `json:"url,omitempty"`
	SpreadsheetID   string `json:"spreadsheet_id,omitempty"`
	SpreadsheetName string `json:"spreadsheet_name,omitempty"`
	SheetID         int64  `json:"sheet_id"`
	SheetName       string `json:"sheet_name"`
}
