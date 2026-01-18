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
