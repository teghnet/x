// Copyright (c) 2024-2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	dirCache  = "cache"
	dirConfig = "config"
	dirData   = "share"
	dirState  = "state"

	dotLocal = ".local"
)

// kind describes one XDG base directory: its dir-name segment and how to
// resolve its default (non-local) base directory.
type kind struct {
	dir  string
	home func() (string, error)
}

var (
	kindConfig = kind{dirConfig, os.UserConfigDir}
	kindCache  = kind{dirCache, os.UserCacheDir}
	kindData   = kind{dirData, func() (string, error) { return xdgHome("XDG_DATA_HOME", dotLocal, dirData) }}
	kindState  = kind{dirState, func() (string, error) { return xdgHome("XDG_STATE_HOME", dotLocal, dirState) }}
)

// xdgHome returns the value of env if it is set and absolute, or
// $HOME/rel otherwise. It returns an error for a relative env value,
// matching the contract of os.UserConfigDir/os.UserCacheDir.
func xdgHome(env string, rel ...string) (string, error) {
	if dir := os.Getenv(env); dir != "" {
		if !filepath.IsAbs(dir) {
			return "", fmt.Errorf("path in %s is relative", env)
		}
		return dir, nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, filepath.Join(rel...)), nil
}

// appDir resolves the directory of the given kind for app: a local dev
// directory if one exists in the working directory, otherwise
// <kind's base dir>/<app>.
func appDir(app string, k kind) string {
	if dir, ok := localAppDir(app, k.dir); ok {
		return dir
	}
	base, err := k.home()
	if err != nil {
		panic(fmt.Sprintf("paths: unable to determine app %s directory: %v", k.dir, err))
	}
	return filepath.Join(base, app)
}

// AppConfig returns the configuration directory path for the given app.
// It checks for local dev directories (.local or .<appName>) first,
// then falls back to the system config directory (~/.config/<appName>).
//
// XDG_CONFIG_HOME stores user-specific configuration files.
// These are typically application settings, preferences, and dotfiles.
func AppConfig(app string) string {
	return appDir(app, kindConfig)
}

// AppCache returns the cache directory path for the given app.
// It checks for local dev directories first, then falls back to
// the system cache directory (~/.cache/<appName>).
//
// XDG_CACHE_HOME stores user-specific non-essential data files that can be regenerated or deleted without loss.
// This includes cached data, temporary files, and historical information.
func AppCache(app string) string {
	return appDir(app, kindCache)
}

// AppData returns the data directory path for the given app.
// It checks for local dev directories first, then XDG_DATA_HOME,
// then falls back to ~/.local/share/<appName>.
//
// XDG_DATA_HOME stores user-specific data files that are not configuration files
// and are not meant to be shared with other users. This includes application data,
// saved games, and other user-generated content.
func AppData(app string) string {
	return appDir(app, kindData)
}

// AppState returns the state directory path for the given app.
// It checks for local dev directories first, then XDG_STATE_HOME,
// then falls back to ~/.local/state/<appName>.
//
// XDG_STATE_HOME stores data that should persist between application restarts.
// Not important or portable enough to the user to be stored in [AppData]
func AppState(app string) string {
	return appDir(app, kindState)
}

// localAppDir searches for a local application directory in the current working directory.
// It checks multiple patterns: .local/<app>/<dir>, .local/<dir>, .<dir>, and .<app>/<dir> in priority order.
// Returns the first existing directory path and true, or empty string and false if none exist.
// Skips .local & .<dir> patterns when the working directory is the user's home directory.
func localAppDir(app string, dir ...string) (string, bool) {
	if strings.TrimSpace(app) == "" {
		panic("localAppDir: app name must not be empty")
	}
	if wd, err := os.Getwd(); err == nil {
		if !wdIsHome() {
			d := filepath.Join(wd, dotLocal, app, filepath.Join(dir...))
			if isDir(d) {
				// Return WD/.local/<app>/<dir>
				return d, true
			}
			if len(dir) > 0 {
				d = filepath.Join(wd, dotLocal, filepath.Join(dir...))
				if isDir(d) {
					// Return WD/.local/<dir>
					return d, true
				}
				d = filepath.Join(wd, "."+filepath.Join(dir...))
				if isDir(d) {
					// Return WD/.<dir>
					return d, true
				}
			}
		}
		d := filepath.Join(wd, "."+app, filepath.Join(dir...))
		if isDir(d) {
			// Return WD/.<app>/<dir>
			return d, true
		}
	}
	return "", false
}

func mkLocalAppDir(app string, dir ...string) error {
	if strings.TrimSpace(app) == "" {
		panic("mkLocalAppDir: app name must not be empty")
	}
	dir = nonEmpty(dir)
	_, ok := localAppDir(app, dir...)
	if ok {
		return nil
	}
	if wdIsHome() {
		return fmt.Errorf("mkLocalAppDir: cannot create `.%s` directory in $HOME", app)
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	d := filepath.Join(wd, dotLocal, app)
	if isDir(d) {
		// Make WD/.local/<app>/<dir>
		return os.MkdirAll(filepath.Join(d, filepath.Join(dir...)), 0700)
	}
	// Make WD/.<app>/<dir>
	return os.MkdirAll(filepath.Join(wd, "."+app, filepath.Join(dir...)), 0700)
}

func mkDotDir(dir ...string) error {
	dir = nonEmpty(dir)
	if len(dir) == 0 {
		return fmt.Errorf("mkDotDir: empty dir")
	}
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("mkDotDir: %w", err)
	}
	if wdIsHome() {
		return fmt.Errorf("mkDotDir: cannot create %q directory in $HOME", "."+filepath.Join(dir...))
	}
	// Make WD/.<dir>
	return os.MkdirAll(filepath.Join(wd, "."+filepath.Join(dir...)), 0700)
}
