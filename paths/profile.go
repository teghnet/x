// Copyright (c) 2024-2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package paths

import (
	"os"
	"path/filepath"
)

// ProfileConfig returns the config directory for a specific profile.
// For local dev: .local/<profileName> or .<appName>/<profileName>.
// For system: ~/.config/<appName>/profiles/<profileName>.
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

// ProfileCache returns the cache directory for a specific profile.
// For local dev: .local/<profileName>/cache or .<appName>/<profileName>/cache.
// For system: ~/.cache/<appName>/profiles/<profileName>.
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
		dir = filepath.Join(wd, ".cache", profileName)
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}

	return filepath.Join(AppCache(appName), "profiles", profileName)
}

// ProfileData returns the data directory for a specific profile.
// For local dev: .local/<profileName>/data or .<appName>/<profileName>/data.
// For system: ~/.local/share/<appName>/profiles/<profileName>.
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
		dir = filepath.Join(wd, ".data", profileName)
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}

	return filepath.Join(AppData(appName), "profiles", profileName)
}

// ProfileState returns the state directory for a specific profile.
// For local dev: .local/<profileName>/state or .<appName>/<profileName>/state.
// For system: ~/.local/state/<appName>/profiles/<profileName>.
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
		dir = filepath.Join(wd, ".state", profileName)
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}

	return filepath.Join(AppState(appName), "profiles", profileName)
}
