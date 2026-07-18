// Copyright (c) 2024-2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package paths

import (
	"os"
	"path/filepath"
)

const dirProfiles = "profiles"

// ProfileConfig returns the config directory for a specific profile.
// For local dev: .local/<profile> or .<app>/<profile>.
// For system: ~/.config/<app>/profiles/<profile>.
func ProfileConfig(app, profile string) string {
	if app == "" {
		panic("app name must be non-empty")
	}
	if profile == "" {
		panic("profile name must be non-empty")
	}

	if wd, err := os.Getwd(); err == nil && !wdIsHome() {
		dir := filepath.Join(wd, dotLocal, profile, dirConfig)
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			// Return WD/.local/<profile>/config
			return dir
		}
		dir = filepath.Join(wd, "."+app, profile, dirConfig)
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			// Return WD/.<app>/<profile>/config
			return dir
		}
		dir = filepath.Join(wd, dotConfig, profile)
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			// Return WD/.config/<profile>
			return dir
		}
	}

	// System: ~/.config/<app>/profiles/<profile>
	return filepath.Join(AppConfig(app), dirProfiles, profile)
}

// ProfileCache returns the cache directory for a specific profile.
// For local dev: .local/<profile>/cache or .<app>/<profile>/cache.
// For system: ~/.cache/<app>/profiles/<profile>.
func ProfileCache(app, profile string) string {
	if app == "" {
		panic("appName must be non-empty")
	}
	if profile == "" {
		panic("profileName must be non-empty")
	}

	if wd, err := os.Getwd(); err == nil && !wdIsHome() {
		dir := filepath.Join(wd, dotLocal, profile, dirCache)
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			// Return WD/.local/<profile>/cache
			return dir
		}
		dir = filepath.Join(wd, "."+app, profile, dirCache)
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			// Return WD/.<app>/<profile>/cache
			return dir
		}
		dir = filepath.Join(wd, dotCache, profile)
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			// Return WD/.cache/<profile>
			return dir
		}
	}

	return filepath.Join(AppCache(app), dirProfiles, profile)
}

// ProfileData returns the data directory for a specific profile.
// For local dev: .local/<profile>/data or .<app>/<profile>/data.
// For system: ~/.local/share/<app>/profiles/<profile>.
func ProfileData(app, profile string) string {
	if app == "" {
		panic("appName must be non-empty")
	}
	if profile == "" {
		panic("profileName must be non-empty")
	}

	if wd, err := os.Getwd(); err == nil && !wdIsHome() {
		dir := filepath.Join(wd, dotLocal, profile, dirData)
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			// Return WD/.local/<profile>/data
			return dir
		}
		dir = filepath.Join(wd, "."+app, profile, dirData)
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			// Return WD/.<app>/<profile>/data
			return dir
		}
		dir = filepath.Join(wd, dotData, profile)
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			// Return WD/.data/<profile>
			return dir
		}
	}

	return filepath.Join(AppData(app), dirProfiles, profile)
}

// ProfileState returns the state directory for a specific profile.
// For local dev: .local/<profile>/state or .<app>/<profile>/state.
// For system: ~/.local/state/<app>/profiles/<profile>.
func ProfileState(app, profile string) string {
	if app == "" {
		panic("appName must be non-empty")
	}
	if profile == "" {
		panic("profileName must be non-empty")
	}

	if wd, err := os.Getwd(); err == nil && !wdIsHome() {
		dir := filepath.Join(wd, dotLocal, profile, dirState)
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			// Return WD/.local/<profile>/state
			return dir
		}
		dir = filepath.Join(wd, "."+app, profile, dirState)
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			// Return WD/.<app>/<profile>/state
			return dir
		}
		dir = filepath.Join(wd, dotState, profile)
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			// Return WD/.state/<profile>
			return dir
		}
	}

	return filepath.Join(AppState(app), dirProfiles, profile)
}
