// Copyright (c) 2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package paths

import (
	"log"
	"path"
)

type XDG interface {
	Config(...string) string
	Data(...string) string
	Cache(...string) string
	State(...string) string
}

func NewXDG(app string, mkLocalUnlessDefaultExist bool) XDG {
	if mkLocalUnlessDefaultExist {
		errLog(MkLocalApp(app))
		errLog(MkLocalAppConfig(app))
		errLog(MkLocalAppData(app))
		errLog(MkLocalAppCache(app))
		errLog(MkLocalAppState(app))
	}
	return xdg{
		app:        App(app),
		configHome: AppConfig(app),
		dataHome:   AppData(app),
		cacheHome:  AppCache(app),
		stateHome:  AppState(app),
	}
}

// XDG Base Directory paths
type xdg struct {
	app        string
	configHome string
	dataHome   string
	cacheHome  string
	stateHome  string
}

func (x xdg) App(elems ...string) string {
	return path.Join(append([]string{x.app}, elems...)...)
}

// Config configHome user-specific settings that you would want to preserve or back up.
// .local/config or $XDG_CONFIG_HOME/<app> or ~/.config/<app>
func (x xdg) Config(elems ...string) string {
	return path.Join(append([]string{x.configHome}, elems...)...)
}

// Data dataHome for persistent data files that the application needs to function.
// Examples: Game saves, local mail storage, browser extensions, icon sets, and custom fonts.
// .local/share or $XDG_DATA_HOME/<app> or ~/.local/share/<app>
func (x xdg) Data(elems ...string) string {
	return path.Join(append([]string{x.dataHome}, elems...)...)
}

// Cache cacheHome non-essential data that can be safely deleted without losing information.
// Deleting this directory should only result in a slight speed penalty the next time you run the app.
// .local/chache or $XDG_CACHE_HOME/<app> or ~/.cache/<app>
func (x xdg) Cache(elems ...string) string {
	return path.Join(append([]string{x.cacheHome}, elems...)...)
}

// State stateHome temporary application state that should persist between restarts
// but isn't a configuration or "data" in the traditional sense.
// .local/state or $XDG_STATE_HOME/<app> or ~/.local/state/<app>
func (x xdg) State(elems ...string) string {
	return path.Join(append([]string{x.stateHome}, elems...)...)
}

func errLog(err error) {
	if err != nil {
		log.Print(err)
	}
}
