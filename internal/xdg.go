// Copyright (c) 2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package internal

import (
	"log"

	"github.com/teghnet/x/paths"
)

// XDG Base Directory paths
type XDG struct {
	App string
	// ConfigHome user-specific settings that you would want to preserve or back up.
	// .local/config or $XDG_CONFIG_HOME/<app> or ~/.config/<app>
	ConfigHome string
	// DataHome for persistent data files that the application needs to function.
	// Examples: Game saves, local mail storage, browser extensions, icon sets, and custom fonts.
	// .local/share or $XDG_DATA_HOME/<app> or ~/.local/share/<app>
	DataHome string
	// CacheHome non-essential data that can be safely deleted without losing information.
	// Deleting this directory should only result in a slight speed penalty the next time you run the app.
	// .local/chache or $XDG_CACHE_HOME/<app> or ~/.cache/<app>
	CacheHome string
	// StateHome temporary application state that should persist between restarts
	// but isn't a configuration or "data" in the traditional sense.
	// .local/state or $XDG_STATE_HOME/<app> or ~/.local/state/<app>
	StateHome string
}

func NewXDG(app string, mkLocalUnlessDefaultExist bool) XDG {
	if mkLocalUnlessDefaultExist {
		errLog(paths.MkLocalApp(app))
		errLog(paths.MkLocalAppConfig(app))
		errLog(paths.MkLocalAppData(app))
		errLog(paths.MkLocalAppCache(app))
		errLog(paths.MkLocalAppState(app))
	}
	return XDG{
		App:        paths.App(app),
		ConfigHome: paths.AppConfig(app),
		DataHome:   paths.AppData(app),
		CacheHome:  paths.AppCache(app),
		StateHome:  paths.AppState(app),
	}
}

func errLog(err error) {
	if err != nil {
		log.Print(err)
	}
}
