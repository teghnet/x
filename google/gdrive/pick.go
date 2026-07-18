package gdrive

import (
	"context"
	"fmt"

	"charm.land/huh/v2"
	"google.golang.org/api/drive/v3"
)

func PickSharedDrive(ctx context.Context, ds *drive.Service) (string, string, error) {
	var opts []huh.Option[string]
	for f, err := range SharedDrives(ctx, ds) {
		if err != nil {
			return "", "", err
		}
		opts = append(opts, huh.NewOption(f.Name, f.Id))
	}
	if len(opts) == 0 {
		return "", "", fmt.Errorf("no shared drives")
	}

	var id string
	s := huh.NewSelect[string]().
		Title("Select Shared Drive").
		Options(opts...).
		Value(&id)
	err := huh.NewForm(huh.NewGroup(s)).Run()
	if err != nil {
		return "", "", err
	}

	for _, o := range opts {
		if o.Value == id {
			return id, o.Key, nil
		}
	}
	return id, "", nil
}
