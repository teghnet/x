package file

import (
	"fmt"
	"path/filepath"

	"charm.land/huh/v2"
)

func SelectFilename(path, title string) (string, error) {
	matches, err := filepath.Glob(path)
	if err != nil {
		return "", fmt.Errorf("unable to read dir: %v", err)
	}
	if len(matches) == 1 {
		return matches[0], nil
	}

	var entry string
	return entry, huh.NewForm(
		huh.NewGroup(huh.NewSelect[string]().
			Value(&entry).
			Title(title).
			Options(huh.NewOptions(matches...)...),
		),
	).Run()
}
