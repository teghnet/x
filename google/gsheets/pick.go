package gsheets

import (
	"context"
	"fmt"
	"slices"

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

type Val struct {
	Key string
	Val string
}

func PickSpreadsheets(ctx context.Context, srv *drive.Service, driveID string) ([]Val, error) {
	var opts []huh.Option[string]
	for f, err := range Spreadsheets(ctx, srv, driveID) {
		if err != nil {
			return nil, err
		}
		opts = append(opts, huh.NewOption(f.Name, f.Id))
	}
	if len(opts) == 0 {
		return nil, fmt.Errorf("no spreadsheets found in %s", driveID)
	}

	var fileIDs []string
	s := huh.NewMultiSelect[string]().
		Title(fmt.Sprintf("Select spreadsheet in %s:", driveID)).
		Options(opts...).
		Value(&fileIDs)
	if err := huh.NewForm(
		huh.NewGroup(
			s,
		),
	).Run(); err != nil {
		return nil, err
	}

	var vals []Val
	for _, o := range opts {
		if slices.Contains(fileIDs, o.Value) {
			vals = append(vals, Val{o.Key, o.Value})
		}
	}
	return vals, nil
}
