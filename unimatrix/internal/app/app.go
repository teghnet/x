// Copyright (c) 2024-2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

// Package app provides application orchestration for Unimatrix.
package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/teghnet/x/paths"
	"github.com/teghnet/x/unimatrix/internal/model"
)

const appName = "unimatrix"

// Config holds application configuration.
type Config struct {
	DefaultProfile string `json:"default_profile"`
	Debug          bool   `json:"debug"`
}

// App is the main application struct.
type App struct {
	config         *Config
	configDir      string
	dataDir        string
	profiles       map[string]*model.Profile
	currentProfile string
}

// New creates a new App instance.
func New() (*App, error) {
	configDir := paths.AppConfig(appName)
	dataDir := paths.AppData(appName)

	app := &App{
		configDir: configDir,
		dataDir:   dataDir,
		profiles:  make(map[string]*model.Profile),
	}

	if err := app.ensureDirs(); err != nil {
		return nil, fmt.Errorf("failed to create directories: %w", err)
	}

	if err := app.loadConfig(); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return app, nil
}

// ensureDirs creates necessary directories.
func (a *App) ensureDirs() error {
	dirs := []string{
		a.configDir,
		a.dataDir,
		filepath.Join(a.dataDir, "profiles"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

// loadConfig loads or creates the configuration.
func (a *App) loadConfig() error {
	configPath := filepath.Join(a.configDir, "config.json")

	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		// Create default config
		a.config = &Config{
			DefaultProfile: "zero",
			Debug:          false,
		}
		return a.saveConfig()
	}
	if err != nil {
		return err
	}

	a.config = &Config{}
	return json.Unmarshal(data, a.config)
}

// saveConfig persists the configuration.
func (a *App) saveConfig() error {
	configPath := filepath.Join(a.configDir, "config.json")
	data, err := json.MarshalIndent(a.config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}

// Profile returns the named profile, loading it if necessary.
func (a *App) Profile(name string) (*model.Profile, error) {
	if p, ok := a.profiles[name]; ok {
		return p, nil
	}

	p, err := a.loadProfile(name)
	if err != nil {
		// Create new profile if not found
		p = model.NewProfile(name)
		a.profiles[name] = p
		return p, nil
	}

	a.profiles[name] = p
	return p, nil
}

// loadProfile loads a profile from disk.
func (a *App) loadProfile(name string) (*model.Profile, error) {
	path := filepath.Join(a.dataDir, "profiles", name+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	p := &model.Profile{}
	if err := json.Unmarshal(data, p); err != nil {
		return nil, err
	}
	return p, nil
}

// SaveProfile persists a profile to disk.
func (a *App) SaveProfile(p *model.Profile) error {
	path := filepath.Join(a.dataDir, "profiles", p.Name+".json")
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// ConfigDir returns the configuration directory path.
func (a *App) ConfigDir() string {
	return a.configDir
}

// DataDir returns the data directory path.
func (a *App) DataDir() string {
	return a.dataDir
}

// DefaultProfile returns the default profile name.
func (a *App) DefaultProfile() string {
	return a.config.DefaultProfile
}
