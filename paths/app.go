// Copyright (c) 2024-2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package paths

import (
	"fmt"
	"os"
	"path/filepath"
)

// App is similar to AppConfig but does not need a config sub dir.
func App(app string) string {
	if dir, ok := localAppDir(app); ok {
		return dir
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		panic(fmt.Sprintf("unable to determine app config directory: %v", err))
	}
	return filepath.Join(dir, app)
}

// AppConfig returns the configuration directory path for the given app.
// It checks for local dev directories (.local or .<appName>) first,
// then falls back to the system config directory (~/.config/<appName>).
//
// XDG_CONFIG_HOME stores user-specific configuration files.
// These are typically application settings, preferences, and dotfiles.
func AppConfig(app string) string {
	if dir, ok := localAppDir(app, dirConfig); ok {
		return dir
	}

	dir, err := os.UserConfigDir()
	if err != nil {
		panic(fmt.Sprintf("unable to determine app config directory: %v", err))
	}
	return filepath.Join(dir, app)
}

// AppCache returns the cache directory path for the given app.
// It checks for local dev directories first, then falls back to
// the system cache directory (~/.cache/<appName>).
//
// XDG_CACHE_HOME stores user-specific non-essential data files that can be regenerated or deleted without loss.
// This includes cached data, temporary files, and historical information.
func AppCache(app string) string {
	if dir, ok := localAppDir(app, dirCache); ok {
		return dir
	}

	dir, err := os.UserCacheDir()
	if err != nil {
		panic(fmt.Sprintf("unable to determine app cache directory: %v", err))
	}
	return filepath.Join(dir, app)
}

// AppData returns the data directory path for the given app.
// It checks for local dev directories first, then XDG_DATA_HOME,
// then falls back to ~/.local/share/<appName>.
//
// XDG_DATA_HOME stores user-specific data files that are not configuration files
// and are not meant to be shared with other users. This includes application data,
// saved games, and other user-generated content.
func AppData(app string) string {
	if dir, ok := localAppDir(app, dirData); ok {
		return dir
	}

	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		if !filepath.IsAbs(dir) {
			panic("path in XDG_DATA_HOME is relative")
		}
		return filepath.Join(dir, app)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("unable to determine app data directory: %v", err))
	}

	return filepath.Join(homeDir, ".local", dirDataXDGCompliant, app)
}

// AppState returns the state directory path for the given app.
// It checks for local dev directories first, then XDG_STATE_HOME,
// then falls back to ~/.local/state/<appName>.
//
// XDG_STATE_HOME stores data that should persist between application restarts.
// Not important or portable enough to the user to be stored in [AppData]
func AppState(app string) string {
	if dir, ok := localAppDir(app, dirState); ok {
		return dir
	}

	if dir := os.Getenv("XDG_STATE_HOME"); dir != "" {
		if !filepath.IsAbs(dir) {
			panic("path in XDG_STATE_HOME is relative")
		}
		return filepath.Join(dir, app)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("unable to determine app state directory: %v", err))
	}

	return filepath.Join(homeDir, ".local", dirState, app)
}

// localAppDir searches for a local application directory in the current working directory.
// It checks multiple patterns: .local/app/dir, .local/dir, .dir/app, .dir, and .app/dir in priority order.
// Returns the first existing directory path and true, or empty string and false if none exist.
// Skips .local patterns when working directory is the user's home directory.
func localAppDir(app string, dir ...string) (string, bool) {
	if wd, err := os.Getwd(); err == nil {
		if !wdIsHome() {
			d := filepath.Join(wd, ".local", app, filepath.Join(dir...))
			if info, err := os.Stat(d); err == nil && info.IsDir() {
				return d, true
			}
			d = filepath.Join(wd, ".local", filepath.Join(dir...))
			if info, err := os.Stat(d); err == nil && info.IsDir() {
				return d, true
			}
			if len(dir) > 0 {
				d = filepath.Join(wd, "."+filepath.Join(dir...), app)
				if info, err := os.Stat(d); err == nil && info.IsDir() {
					return d, true
				}
				d = filepath.Join(wd, "."+filepath.Join(dir...))
				if info, err := os.Stat(d); err == nil && info.IsDir() {
					return d, true
				}
			}
		}
		d := filepath.Join(wd, "."+app, filepath.Join(dir...))
		if info, err := os.Stat(d); err == nil && info.IsDir() {
			return d, true
		}
	}
	return "", false
}

const (
	dirCache            = "cache"
	dirConfig           = "config"
	dirData             = "data"
	dirDataXDGCompliant = "share"
	dirState            = "state"
)

func mkLocalDir(app string, dir ...string) error {
	_, ok := localAppDir(app, dir...)
	if ok {
		return nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	if len(dir) == 0 {
		if wdIsHome() {
			return fmt.Errorf("mkLocalDir: cannot create `.%s` directory in $HOME", app)
		}
		d := filepath.Join(wd, ".local")
		if info, err := os.Stat(d); err == nil && info.IsDir() {
			return os.MkdirAll(filepath.Join(d, app), 0700)
		}
		return os.MkdirAll(filepath.Join(wd, "."+app), 0700)
	}
	if wdIsHome() {
		return fmt.Errorf("mkLocalDir: cannot create `.%s/%s` directory in $HOME", dir, app)
	}
	d := filepath.Join(wd, ".local", app)
	if info, err := os.Stat(d); err == nil && info.IsDir() {
		return os.MkdirAll(filepath.Join(d, filepath.Join(dir...)), 0700)
	}
	return os.MkdirAll(filepath.Join(wd, "."+filepath.Join(dir...), app), 0700)
}

func mkCurrentDir(dir ...string) error {
	{
		var dd []string
		for _, d := range dir {
			if d != "" {
				dd = append(dd, d)
			}
		}
		dir = dd
	}
	if len(dir) == 0 {
		return fmt.Errorf("mkCurrentDir: empty dir")
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	if wdIsHome() {
		return fmt.Errorf("mkCurrentDir: cannot create local `.%s` directory in $HOME", dir)
	}
	return os.MkdirAll(filepath.Join(wd, "."+filepath.Join(dir...)), 0700)
}
