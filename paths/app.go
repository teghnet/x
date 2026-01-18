// Copyright (c) 2024-2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package paths

import (
	"fmt"
	"os"
	"path/filepath"
)

func AppConfig(appName string) string {
	if appName == "" {
		panic("appName must be non-empty")
	}

	if wd, err := os.Getwd(); err == nil {
		if !wdIsHome() {
			dir := filepath.Join(wd, ".local")
			if info, err := os.Stat(dir); err == nil && info.IsDir() {
				return dir
			}
		}
		dir := filepath.Join(wd, "."+appName)
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}

	dir, err := os.UserConfigDir()
	if err != nil {
		panic(fmt.Sprintf("unable to determine app config directory: %v", err))
	}
	return filepath.Join(dir, appName)
}

func AppCache(appName string) string {
	if appName == "" {
		panic("appName must be non-empty")
	}

	if wd, err := os.Getwd(); err == nil {
		if !wdIsHome() {
			dir := filepath.Join(wd, ".local", "cache")
			if info, err := os.Stat(dir); err == nil && info.IsDir() {
				return dir
			}
		}
		dir := filepath.Join(wd, "."+appName, "cache")
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}

	dir, err := os.UserCacheDir()
	if err != nil {
		panic(fmt.Sprintf("unable to determine app cache directory: %v", err))
	}
	return filepath.Join(dir, appName)

}

func AppData(appName string) string {
	if appName == "" {
		panic("appName must be non-empty")
	}

	if wd, err := os.Getwd(); err == nil {
		if !wdIsHome() {
			dir := filepath.Join(wd, ".local", "data")
			if info, err := os.Stat(dir); err == nil && info.IsDir() {
				return dir
			}
		}
		dir := filepath.Join(wd, "."+appName, "data")
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}

	dir := os.Getenv("XDG_DATA_HOME")
	if dir != "" {
		return filepath.Join(dir, appName)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("unable to determine app data directory: %v", err))
	}

	return filepath.Join(homeDir, ".local", "share", appName)
}

func AppState(appName string) string {
	if appName == "" {
		panic("appName must be non-empty")
	}

	if wd, err := os.Getwd(); err == nil {
		if !wdIsHome() {
			dir := filepath.Join(wd, ".local", "state")
			if info, err := os.Stat(dir); err == nil && info.IsDir() {
				return dir
			}
		}
		dir := filepath.Join(wd, "."+appName, "state")
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}

	dir := os.Getenv("XDG_STATE_HOME")
	if dir != "" {
		return filepath.Join(dir, appName)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("unable to determine app state directory: %v", err))
	}

	return filepath.Join(homeDir, ".local", "state", appName)
}

func ProfileConfig(appName, profileName string) string {
	if appName == "" {
		panic("appName must be non-empty")
	}
	if profileName == "" {
		panic("profileName must be non-empty")
	}

	if wd, err := os.Getwd(); err == nil && !wdIsHome() {
		// Local dev: .local/<profileName>
		dir := filepath.Join(wd, ".local", profileName)
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
		// Local dev: .<appName>/<profileName>
		dir = filepath.Join(wd, "."+appName, profileName)
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}

	// System: ~/.config/<appName>/profiles/<profileName>
	return filepath.Join(AppConfig(appName), "profiles", profileName)
}

func ProfileCache(appName, profileName string) string {
	if appName == "" {
		panic("appName must be non-empty")
	}
	if profileName == "" {
		panic("profileName must be non-empty")
	}

	if wd, err := os.Getwd(); err == nil && !wdIsHome() {
		dir := filepath.Join(wd, ".local", profileName, "cache")
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
		dir = filepath.Join(wd, "."+appName, profileName, "cache")
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}

	return filepath.Join(AppCache(appName), "profiles", profileName)
}

func ProfileData(appName, profileName string) string {
	if appName == "" {
		panic("appName must be non-empty")
	}
	if profileName == "" {
		panic("profileName must be non-empty")
	}

	if wd, err := os.Getwd(); err == nil && !wdIsHome() {
		dir := filepath.Join(wd, ".local", profileName, "data")
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
		dir = filepath.Join(wd, "."+appName, profileName, "data")
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}

	return filepath.Join(AppData(appName), "profiles", profileName)
}

func ProfileState(appName, profileName string) string {
	if appName == "" {
		panic("appName must be non-empty")
	}
	if profileName == "" {
		panic("profileName must be non-empty")
	}

	if wd, err := os.Getwd(); err == nil && !wdIsHome() {
		dir := filepath.Join(wd, ".local", profileName, "state")
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
		dir = filepath.Join(wd, "."+appName, profileName, "state")
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}

	return filepath.Join(AppState(appName), "profiles", profileName)
}
