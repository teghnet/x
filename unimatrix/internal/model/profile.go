// Copyright (c) 2024-2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package model

import "time"

// ConnectorConfig holds configuration for a connector.
type ConnectorConfig struct {
	Type   string         // Connector type (local, notion, gdrive, obsidian)
	Name   string         // User-defined name
	Config map[string]any // Connector-specific configuration
}

// Profile represents a Unimatrix (sync configuration).
type Profile struct {
	Name       string            // Profile name (e.g., "zero", "one")
	Links      []Link            // Sync links in this profile
	Connectors []ConnectorConfig // Configured connectors
	LastSync   time.Time         // Last successful sync
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NewProfile creates a new profile with the given name.
func NewProfile(name string) *Profile {
	now := time.Now()
	return &Profile{
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// AddLink adds a sync link to the profile.
func (p *Profile) AddLink(link Link) {
	p.Links = append(p.Links, link)
	p.UpdatedAt = time.Now()
}

// AddConnector adds a connector configuration.
func (p *Profile) AddConnector(cfg ConnectorConfig) {
	p.Connectors = append(p.Connectors, cfg)
	p.UpdatedAt = time.Now()
}

// FindLink finds a link by name.
func (p *Profile) FindLink(name string) *Link {
	for i := range p.Links {
		if p.Links[i].Name == name {
			return &p.Links[i]
		}
	}
	return nil
}
