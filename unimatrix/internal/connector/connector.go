// Copyright (c) 2024-2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

// Package connector defines the interface for storage connectors.
package connector

import (
	"context"
	"io"

	"github.com/teghnet/x/unimatrix/internal/model"
)

// EventType represents the type of filesystem event.
type EventType int

const (
	EventCreated EventType = iota
	EventModified
	EventDeleted
)

// Event represents a filesystem change event.
type Event struct {
	Type EventType
	Node model.Node
}

// Connector is the interface for storage system connectors.
type Connector interface {
	// Name returns the connector's unique name.
	Name() string

	// Connect establishes connection to the storage system.
	Connect(ctx context.Context) error

	// Close disconnects from the storage system.
	Close() error

	// List returns all nodes under the given path.
	List(ctx context.Context, path string) ([]model.Node, error)

	// Read returns a reader for the node's content.
	Read(ctx context.Context, node model.Node) (io.ReadCloser, error)

	// Write writes content to a node.
	Write(ctx context.Context, node model.Node, r io.Reader) error

	// Delete removes a node.
	Delete(ctx context.Context, node model.Node) error

	// Stat returns information about a single node.
	Stat(ctx context.Context, path string) (*model.Node, error)
}

// Watcher is an optional interface for connectors that support watching.
type Watcher interface {
	Watch(ctx context.Context, path string) (<-chan Event, error)
}
