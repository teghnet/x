package gsheets

import (
	"context"
	"fmt"

	"charm.land/huh/v2"
	"google.golang.org/api/drive/v3"
)

func PickSpreadsheet(ctx context.Context, srv *drive.Service, driveID string) (string, string, error) {
	var opts []huh.Option[string]
	for f, err := range Spreadsheets(ctx, srv, driveID) {
		if err != nil {
			return "", "", err
		}
		opts = append(opts, huh.NewOption(f.Name, f.Id))
	}
	if len(opts) == 0 {
		return "", "", fmt.Errorf("no spreadsheets found in %s", driveID)
	}

	var fileID string
	s := huh.NewSelect[string]().
		Title(fmt.Sprintf("Select spreadsheet in %s:", driveID)).
		Options(opts...).
		Value(&fileID)
	if err := huh.NewForm(
		huh.NewGroup(
			s,
		),
	).Run(); err != nil {
		return "", "", err
	}

	for _, o := range opts {
		if o.Value == fileID {
			return fileID, o.Key, nil
		}
	}
	return fileID, "", nil
}
