package conf

import (
	"errors"
	"fmt"
	"os"
	"path"
)

var ErrNotFound = errors.New("not found")

const (
	appNameShort = "x"
	appVer       = "0.1.0"

	extYaml = ".yaml"
	dirConf = "config"

	LocationInWD            = "in-wd"
	LocationInConfigUnderWD = "in-conf-under-wd"
	LocationInHome          = "in-home-dir"
	LocationInUserConfig    = "in-user-config"
)

func UA(opts ...string) string {
	app, ver := appNameShort, appVer
	if len(opts) > 0 {
		app = opts[0]
	}
	if len(opts) > 1 {
		ver = opts[1]
	}
	return app + "/" + ver
}

// ConfigFile searches for a config file in the following order:
// 1. `.<app>.<ext>` in working directory
// 2. `.<config>/<app>.<ext>` in working directory
// 3. `.<app>.<ext>` file in home directory
// 4. `<config>.<ext>` file in user default config location
// If none are found, then the one specified by `prefer` is returned.
// Args order:
// - ext:    the extension of the config file
// - app:    the name of the application
// - conf:   the name of the config directory
// - prefer: one of the following: LocationInWD, LocationInConfigUnderWD, LocationInHome, LocationInUserConfig
func ConfigFile(opts ...string) (string, error) {
	ext, app, config, prefer := extYaml, appNameShort, dirConf, LocationInWD
	if len(opts) > 0 {
		ext = opts[0]
	}
	if len(opts) > 1 {
		app = opts[1]
	}
	if len(opts) > 2 {
		config = opts[2]
	}
	if len(opts) > 3 {
		prefer = opts[3]
	}
	return ConfigFilePref(ext, app, config, prefer)
}

func ConfigFilePref(ext, app, config, prefer string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("unable to get working directory: %w", err)
	}
	hd, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to get user home directory: %w", err)
	}
	cd, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("unable to get user config directory: %w", err)
	}
	candidates := map[string]string{
		LocationInWD:            path.Join(wd, "."+app+ext),
		LocationInConfigUnderWD: path.Join(wd, "."+config, app+ext),
		LocationInHome:          path.Join(hd, "."+app+ext),
		LocationInUserConfig:    path.Join(cd, app, config+ext),
	}
	for _, f := range candidates {
		_, err = os.Stat(f)
		if err == nil {
			return f, nil
		}
		// err is not nil, but if it is also not `os.ErrNotExist`, then return it because we can't do anything with it
		if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("unable to determine file (%s) existence: %w", f, err)
		}
	}
	p, ok := candidates[prefer]
	if !ok {
		return "", ErrNotFound
	}
	return p, nil
}

// StateDir searches for a state directory in the following order:
// 1. `.<conf>` in working directory
// 2. `<app>` in user config directory
// If none are found, then the one specified by `prefer` is returned.
// Args order:
// - app:    the name of the application
// - conf:   the name of the config directory
// - prefer: one of the following: LocationInConfigUnderWD, LocationInUserConfig
func StateDir(opts ...string) (string, error) {
	app, config, prefer := appNameShort, dirConf, LocationInConfigUnderWD
	if len(opts) > 0 {
		app = opts[0]
	}
	if len(opts) > 1 {
		config = opts[1]
	}
	if len(opts) > 2 {
		prefer = opts[2]
	}
	return StateDirPref(app, config, prefer)
}

func StateDirPref(app, config, prefer string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("unable to get working directory: %w", err)
	}
	cd, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("unable to get user config directory: %w", err)
	}
	candidates := map[string]string{
		LocationInConfigUnderWD: path.Join(wd, "."+config),
		LocationInUserConfig:    path.Join(cd, app),
	}
	for _, f := range candidates {
		fi, err := os.Stat(f)
		if err == nil && fi.IsDir() {
			return f, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("unable to determine dir (%s) existence: %w", f, err)
		}
	}
	p, ok := candidates[prefer]
	if !ok {
		return "", ErrNotFound
	}
	return p, nil
}
