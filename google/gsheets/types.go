package gsheets

import "google.golang.org/api/sheets/v4"

// SheetField represents a field in a Google Sheet.
type SheetField struct {
	Header string
	Value  any
}

// TODO: Add more types as needed for specific Google Sheets interactions.
// For example, configuration for specific sheet ranges, or custom request types.

// Spreadsheet represents a Google Spreadsheet with its ID and potentially its properties.
type Spreadsheet struct {
	ID    string
	Title string
	// Add other properties as needed
}

// Sheet represents a single sheet within a Google Spreadsheet.
type Sheet struct {
	ID    int64
	Title string
	// Add other properties as needed
}

// ValueRange is a wrapper around sheets.ValueRange for convenience.
type ValueRange struct {
	*sheets.ValueRange
}
