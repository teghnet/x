package paths

import (
	"os"
	"path/filepath"
)

// Paths XDG Base Directory paths
// see: https://specifications.freedesktop.org/basedir/latest/
type Paths struct {
	// ConfigHome user-specific settings that you would want to preserve or back up.
	// .<app>/config or $XDG_CONFIG_HOME/<app> or ~/.config/<app>
	ConfigHome string
	// CacheHome non-essential data that can be safely deleted without losing information.
	// Deleting this directory should only result in a slight speed penalty the next time you run the app.
	// .<app>/chache or $XDG_CACHE_HOME/<app> or ~/.cache/<app>
	CacheHome string
	// DataHome for persistent data files that the application needs to function.
	// Examples: Game saves, local mail storage, browser extensions, icon sets, and custom fonts.
	// .<app>/share or $XDG_DATA_HOME/<app> or ~/.local/share/<app>
	DataHome string
	// StateHome temporary application state that should persist between restarts
	// but isn't a configuration or "data" in the traditional sense.
	// .<app>/state or $XDG_STATE_HOME/<app> or ~/.local/state/<app>
	StateHome string
}

// ProfileConfigPath returns the config path for a specific profile.
func (p *Paths) ProfileConfigPath(profileName string) string {
	return filepath.Join(p.ConfigHome, "profiles", profileName)
}

// ProfileDataPath returns the data path for a specific profile.
func (p *Paths) ProfileDataPath(profileName string) string {
	return filepath.Join(p.DataHome, "profiles", profileName)
}

// ProfileSecretsPath returns the path for secrets (credentials) within a profile.
func (p *Paths) ProfileSecretsPath(profileName string) string {
	return filepath.Join(p.DataHome, "profiles", profileName, "secrets")
}

// ProfileTokensPath returns the path for tokens within a profile.
func (p *Paths) ProfileTokensPath(profileName string) string {
	return filepath.Join(p.StateHome, "profiles", profileName, "tokens")
}

func sameDir(a, b string) bool {
	ai, err := os.Stat(a)
	if err != nil {
		return false
	}
	bi, err := os.Stat(b)
	if err != nil {
		return false
	}
	return ai.IsDir() && bi.IsDir() && os.SameFile(ai, bi)
}

// wdIsHome checks if the working directory is in the user's home directory.
func wdIsHome() bool {
	wd, err := os.Getwd()
	if err != nil {
		return false
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	return sameDir(wd, homeDir)
}
