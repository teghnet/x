// Copyright (c) 2024-2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package paths

import (
	"os"
	"path/filepath"
)

const dirProfiles = "profiles"

// profileDir resolves the directory of the given kind for a profile of app.
// It checks local dev directories in the working directory first — in
// priority order, WD/.local/<profile>/<dir>, WD/.<app>/<profile>/<dir>, then
// WD/.<dir>/<profile> — then falls back to <appDir(app, k)>/profiles/<profile>.
//
// Note: the local dev patterns are not namespaced by app beyond .<app>/...,
// so two apps sharing a working directory and a profile name can collide on
// WD/.local/<profile>/<dir> or WD/.<dir>/<profile>.
func profileDir(app, profile string, k kind) string {
	if app == "" {
		panic("app name must be non-empty")
	}
	if profile == "" {
		panic("profile name must be non-empty")
	}

	if wd, err := os.Getwd(); err == nil && !wdIsHome() {
		dir := filepath.Join(wd, dotLocal, profile, k.dir)
		if isDir(dir) {
			// Return WD/.local/<profile>/<dir>
			return dir
		}
		dir = filepath.Join(wd, "."+app, profile, k.dir)
		if isDir(dir) {
			// Return WD/.<app>/<profile>/<dir>
			return dir
		}
		dir = filepath.Join(wd, "."+k.dir, profile)
		if isDir(dir) {
			// Return WD/.<dir>/<profile>
			return dir
		}
	}

	return filepath.Join(appDir(app, k), dirProfiles, profile)
}

// ProfileConfig returns the config directory for a specific profile.
// For local dev: .local/<profile>/config or .<app>/<profile>/config.
// For system: ~/.config/<app>/profiles/<profile>.
func ProfileConfig(app, profile string) string {
	return profileDir(app, profile, kindConfig)
}

// ProfileCache returns the cache directory for a specific profile.
// For local dev: .local/<profile>/cache or .<app>/<profile>/cache.
// For system: ~/.cache/<app>/profiles/<profile>.
func ProfileCache(app, profile string) string {
	return profileDir(app, profile, kindCache)
}

// ProfileData returns the data directory for a specific profile.
// For local dev: .local/<profile>/share or .<app>/<profile>/share.
// For system: ~/.local/share/<app>/profiles/<profile>.
func ProfileData(app, profile string) string {
	return profileDir(app, profile, kindData)
}

// ProfileState returns the state directory for a specific profile.
// For local dev: .local/<profile>/state or .<app>/<profile>/state.
// For system: ~/.local/state/<app>/profiles/<profile>.
func ProfileState(app, profile string) string {
	return profileDir(app, profile, kindState)
}
