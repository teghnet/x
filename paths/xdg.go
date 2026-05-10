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

type conf struct {
	mkCurrentDirs             bool
	mkLocalUnlessDefaultExist bool
}
type ConfOpt func(*conf)

func WithCurrentDirsPreference(v bool) ConfOpt {
	return func(c *conf) {
		c.mkCurrentDirs = v
	}
}
func WithLocalDirsPreference(v bool) ConfOpt {
	return func(c *conf) {
		c.mkLocalUnlessDefaultExist = v
	}
}
func PreferCurrentDirs() ConfOpt {
	return func(c *conf) {
		c.mkCurrentDirs = true
		c.mkLocalUnlessDefaultExist = false
	}
}
func PreferLocalDirs() ConfOpt {
	return func(c *conf) {
		c.mkLocalUnlessDefaultExist = true
		c.mkCurrentDirs = false
	}
}

func NewXDG(app string, opts ...ConfOpt) XDG {
	c := conf{}
	for _, opt := range opts {
		opt(&c)
	}
	if c.mkCurrentDirs {
		errLog(mkCurrentDir(dirConfig))
		errLog(mkCurrentDir(dirData))
		errLog(mkCurrentDir(dirCache))
		errLog(mkCurrentDir(dirState))
	} else if c.mkLocalUnlessDefaultExist {
		errLog(mkLocalDir(app))
		errLog(mkLocalDir(app, dirConfig))
		errLog(mkLocalDir(app, dirData))
		errLog(mkLocalDir(app, dirCache))
		errLog(mkLocalDir(app, dirState))
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
