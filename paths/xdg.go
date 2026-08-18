// Copyright (c) 2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package paths

import (
	"log"
	"path/filepath"
)

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

	// RuntimePath - `XDG_RUNTIME_DIR`
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
	RuntimePath(...string) string

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

// WithPreferWDStore controls whether NewXDG creates plain dot-directories
// (.cache, .config, .share, .state, .run) in the current working directory.
func WithPreferWDStore(v bool) ConfOpt {
	return func(c *conf) {
		c.mkCurrentDirs = v
	}
}

// WithPreferDotLocalStore controls whether NewXDG creates a local dev
// directory (.local/<app> or .<app>) in the current working directory.
//
// If both WithPreferWDStore(true) and WithPreferDotLocalStore(true) are
// passed, WithPreferWDStore wins: NewXDG only creates the plain dot-directories.
func WithPreferDotLocalStore(v bool) ConfOpt {
	return func(c *conf) {
		c.mkLocalUnlessDefaultExist = v
	}
}

// NewXDG resolves the XDG base directories for app.
//
// Resolution is CWD-only: it probes the current working directory for a
// matching local dev or dot-directory (see [AppConfig] and friends) and does
// not walk up to parent directories. It falls back to the system XDG
// locations (honoring XDG_CONFIG_HOME, XDG_CACHE_HOME, XDG_DATA_HOME, and
// XDG_STATE_HOME) when no local directory is found.
//
// Depending on opts, NewXDG may create the directories it resolves to (see
// [WithPreferWDStore] and [WithPreferDotLocalStore]); otherwise the returned
// paths are not guaranteed to exist and callers should create them with
// [EnsureDir] before use.
func NewXDG(app string, opts ...ConfOpt) XDG {
	c := conf{}
	for _, opt := range opts {
		opt(&c)
	}
	if c.mkCurrentDirs {
		errLog("NewXDG", mkDotDir(dirCache))
		errLog("NewXDG", mkDotDir(dirConfig))
		errLog("NewXDG", mkDotDir(dirData))
		errLog("NewXDG", mkDotDir(dirState))
		errLog("NewXDG", mkDotDir(dirRuntime))
	} else if c.mkLocalUnlessDefaultExist {
		errLog("NewXDG", mkLocalAppDir(app))
		errLog("NewXDG", mkLocalAppDir(app, dirCache))
		errLog("NewXDG", mkLocalAppDir(app, dirConfig))
		errLog("NewXDG", mkLocalAppDir(app, dirData))
		errLog("NewXDG", mkLocalAppDir(app, dirState))
		errLog("NewXDG", mkLocalAppDir(app, dirRuntime))
	}
	return xdg{
		configHome:  AppConfig(app),
		dataHome:    AppData(app),
		cacheHome:   AppCache(app),
		stateHome:   AppState(app),
		runtimeHome: AppRuntime(app),
	}
}

// XDG Base Directory paths
type xdg struct {
	configHome  string
	dataHome    string
	cacheHome   string
	stateHome   string
	runtimeHome string
}

// ConfigPath configHome user-specific settings that you would want to preserve or back up.
// .local/config or $XDG_CONFIG_HOME/<app> or ~/.config/<app>
func (x xdg) ConfigPath(elems ...string) string {
	return filepath.Join(append([]string{x.configHome}, elems...)...)
}

// DataPath dataHome for persistent data files that the application needs to function.
// Examples: Game saves, local mail storage, browser extensions, icon sets, and custom fonts.
// .local/share or $XDG_DATA_HOME/<app> or ~/.local/share/<app>
func (x xdg) DataPath(elems ...string) string {
	return filepath.Join(append([]string{x.dataHome}, elems...)...)
}

// CachePath cacheHome non-essential data that can be safely deleted without losing information.
// Deleting this directory should only result in a slight speed penalty the next time you run the app.
// .local/chache or $XDG_CACHE_HOME/<app> or ~/.cache/<app>
func (x xdg) CachePath(elems ...string) string {
	return filepath.Join(append([]string{x.cacheHome}, elems...)...)
}

// StatePath stateHome temporary application state that should persist between restarts
// but isn't a configuration or "data" in the traditional sense.
// .local/state or $XDG_STATE_HOME/<app> or ~/.local/state/<app>
func (x xdg) StatePath(elems ...string) string {
	return filepath.Join(append([]string{x.stateHome}, elems...)...)
}

// RuntimePath runtimeHome non-essential runtime files such as sockets and named pipes.
// Not guaranteed to persist beyond the user's login; see [AppRuntime].
// .local/run or $XDG_RUNTIME_DIR/<app> or an equivalent/temporary directory.
func (x xdg) RuntimePath(elems ...string) string {
	return filepath.Join(append([]string{x.runtimeHome}, elems...)...)
}

// errLog logs err, prefixed with context, if err is non-nil.
func errLog(context string, err error) {
	if err != nil {
		log.Printf("%s: %v", context, err)
	}
}
