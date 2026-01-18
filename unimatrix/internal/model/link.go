// Copyright (c) 2024-2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package model

// StrategyType defines how sync should be performed.
type StrategyType int

const (
	// OneWay copies from source to target only.
	OneWay StrategyType = iota
	// BiDirectional merges changes from both sides.
	BiDirectional
	// Mirror makes target an exact copy of source.
	Mirror
)

// Endpoint represents one side of a sync link.
type Endpoint struct {
	Connector string // Connector name (e.g., "local", "notion")
	Path      string // Path within the connector
}

// Link represents a sync relationship between two paths.
type Link struct {
	ID       string
	Name     string
	Source   Endpoint
	Target   Endpoint
	Strategy StrategyType
	Enabled  bool
}

// String implements fmt.Stringer.
func (l Link) String() string {
	return l.Name
}
