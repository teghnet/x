// Copyright (c) 2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package paths

import (
	"log"
	"path"
)

type XDGProvider func() XDG
type XDG interface {
	// ### User directories

	// CachePath - `XDG_CACHE_HOME`
	//   - Where user-specific non-essential (cached) data should be written (analogous to `/var/cache`).
	//   - Should default to `$HOME/.cache`.
	CachePath(...string) string

	// ConfigPath - `XDG_CONFIG_HOME`
	//   - Where user-specific configurations should be written (analogous to `/etc`).
	//   - Should default to `$HOME/.config`.
	ConfigPath(...string) string

	// DataPath - `XDG_DATA_HOME`
	//   - Where user-specific data files should be written (analogous to `/usr/share`).
	//   - Should default to `$HOME/.local/share`.
	DataPath(...string) string

	// StatePath - `XDG_STATE_HOME`
	//   - Where user-specific state files should be written (analogous to `/var/lib`).
	//   - Should default to `$HOME/.local/state`.
	StatePath(...string) string

	// - `XDG_RUNTIME_DIR`
	//   - Used for non-essential, user-specific data files such as sockets, named pipes, etc.
	//   - Not required to have a default value; warnings should be issued if not set or equivalents provided.
	//   - Must be owned by the user with an access mode of `0700`.
	//   - Filesystem fully featured by standards of OS.
	//   - Must be on the local filesystem.
	//   - May be subject to periodic cleanup.
	//   - Modified every 6 hours or set sticky bit if persistence is desired.
	//   - Can only exist for the duration of the user's login.
	//   - Should not store large files as it may be mounted as a tmpfs.
	//   - pam_systemd sets this to `/run/user/$UID`.
	//
	// ### System directories
	//
	// - `XDG_DATA_DIRS`
	//   - List of directories separated by `:` (analogous to `PATH`).
	//   - Should default to `/usr/local/share:/usr/share`.
	//
	// - `XDG_CONFIG_DIRS`
	//   - List of directories separated by `:` (analogous to `PATH`).
	//   - Should default to `/etc/xdg`.
}

type conf struct {
	mkCurrentDirs             bool
	mkLocalUnlessDefaultExist bool
}
type ConfOpt func(*conf)

func WithPreferWDStore(v bool) ConfOpt {
	return func(c *conf) {
		c.mkCurrentDirs = v
	}
}
func WithPreferDotLocalStore(v bool) ConfOpt {
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
		c.mkCurrentDirs = false
		c.mkLocalUnlessDefaultExist = true
	}
}

func NewXDG(app string, opts ...ConfOpt) XDG {
	c := conf{}
	for _, opt := range opts {
		opt(&c)
	}
	if c.mkCurrentDirs {
		errLog(mkDotDir(dirCache))
		errLog(mkDotDir(dirConfig))
		errLog(mkDotDir(dirData))
		errLog(mkDotDir(dirState))
	} else if c.mkLocalUnlessDefaultExist {
		errLog(mkLocalAppDir(app))
		errLog(mkLocalAppDir(app, dirCache))
		errLog(mkLocalAppDir(app, dirConfig))
		errLog(mkLocalAppDir(app, dirData))
		errLog(mkLocalAppDir(app, dirState))
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

// ConfigPath configHome user-specific settings that you would want to preserve or back up.
// .local/config or $XDG_CONFIG_HOME/<app> or ~/.config/<app>
func (x xdg) ConfigPath(elems ...string) string {
	return path.Join(append([]string{x.configHome}, elems...)...)
}

// DataPath dataHome for persistent data files that the application needs to function.
// Examples: Game saves, local mail storage, browser extensions, icon sets, and custom fonts.
// .local/share or $XDG_DATA_HOME/<app> or ~/.local/share/<app>
func (x xdg) DataPath(elems ...string) string {
	return path.Join(append([]string{x.dataHome}, elems...)...)
}

// CachePath cacheHome non-essential data that can be safely deleted without losing information.
// Deleting this directory should only result in a slight speed penalty the next time you run the app.
// .local/chache or $XDG_CACHE_HOME/<app> or ~/.cache/<app>
func (x xdg) CachePath(elems ...string) string {
	return path.Join(append([]string{x.cacheHome}, elems...)...)
}

// StatePath stateHome temporary application state that should persist between restarts
// but isn't a configuration or "data" in the traditional sense.
// .local/state or $XDG_STATE_HOME/<app> or ~/.local/state/<app>
func (x xdg) StatePath(elems ...string) string {
	return path.Join(append([]string{x.stateHome}, elems...)...)
}

func errLog(err error) {
	if err != nil {
		log.Print(err)
	}
}
