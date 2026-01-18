// Copyright (c) 2024-2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package paths

import (
	"os"
	"path/filepath"
)

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
