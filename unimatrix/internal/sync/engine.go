// Copyright (c) 2024-2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

// Package sync provides the sync engine for Unimatrix.
package sync

import (
	"context"
	"fmt"

	"github.com/teghnet/x/unimatrix/internal/connector"
	"github.com/teghnet/x/unimatrix/internal/model"
)

// ChangeType represents the type of change detected.
type ChangeType int

const (
	ChangeAdded ChangeType = iota
	ChangeModified
	ChangeDeleted
)

// Change represents a single change to be synced.
type Change struct {
	Type   ChangeType
	Node   model.Node
	Source string // Source connector name
	Target string // Target connector name
}

// Result holds the result of a sync operation.
type Result struct {
	Added     []model.Node
	Modified  []model.Node
	Deleted   []model.Node
	Conflicts []Conflict
	Errors    []error
}

// Conflict represents a sync conflict.
type Conflict struct {
	SourceNode model.Node
	TargetNode model.Node
	Reason     string
}

// Engine orchestrates sync operations.
type Engine struct {
	connectors map[string]connector.Connector
}

// NewEngine creates a new sync engine.
func NewEngine() *Engine {
	return &Engine{
		connectors: make(map[string]connector.Connector),
	}
}

// RegisterConnector registers a connector with the engine.
func (e *Engine) RegisterConnector(c connector.Connector) {
	e.connectors[c.Name()] = c
}

// Connector returns a registered connector by name.
func (e *Engine) Connector(name string) (connector.Connector, bool) {
	c, ok := e.connectors[name]
	return c, ok
}

// Preview analyzes a link and returns what would change without making changes.
func (e *Engine) Preview(ctx context.Context, link model.Link) (*Result, error) {
	source, ok := e.connectors[link.Source.Connector]
	if !ok {
		return nil, fmt.Errorf("source connector not found: %s", link.Source.Connector)
	}

	target, ok := e.connectors[link.Target.Connector]
	if !ok {
		return nil, fmt.Errorf("target connector not found: %s", link.Target.Connector)
	}

	// Connect to both
	if err := source.Connect(ctx); err != nil {
		return nil, fmt.Errorf("source connect failed: %w", err)
	}
	if err := target.Connect(ctx); err != nil {
		return nil, fmt.Errorf("target connect failed: %w", err)
	}

	// List both sides
	sourceNodes, err := source.List(ctx, link.Source.Path)
	if err != nil {
		return nil, fmt.Errorf("source list failed: %w", err)
	}

	targetNodes, err := target.List(ctx, link.Target.Path)
	if err != nil {
		return nil, fmt.Errorf("target list failed: %w", err)
	}

	// Build lookup map for target
	targetMap := make(map[string]model.Node)
	for _, n := range targetNodes {
		targetMap[n.Name] = n
	}

	result := &Result{}

	// Find additions and modifications
	for _, sn := range sourceNodes {
		if tn, exists := targetMap[sn.Name]; exists {
			// Check if modified (by time or size)
			if sn.ModTime.After(tn.ModTime) || sn.Size != tn.Size {
				result.Modified = append(result.Modified, sn)
			}
			delete(targetMap, sn.Name)
		} else {
			result.Added = append(result.Added, sn)
		}
	}

	// Remaining in targetMap are deletions (for mirror strategy)
	if link.Strategy == model.Mirror {
		for _, tn := range targetMap {
			result.Deleted = append(result.Deleted, tn)
		}
	}

	return result, nil
}

// Sync executes the sync for a link.
func (e *Engine) Sync(ctx context.Context, link model.Link) (*Result, error) {
	preview, err := e.Preview(ctx, link)
	if err != nil {
		return nil, err
	}

	source, _ := e.connectors[link.Source.Connector]
	target, _ := e.connectors[link.Target.Connector]

	// Apply changes
	for _, node := range preview.Added {
		if err := e.copyNode(ctx, source, target, node, link.Target.Path); err != nil {
			preview.Errors = append(preview.Errors, err)
		}
	}

	for _, node := range preview.Modified {
		if err := e.copyNode(ctx, source, target, node, link.Target.Path); err != nil {
			preview.Errors = append(preview.Errors, err)
		}
	}

	for _, node := range preview.Deleted {
		node.Path = link.Target.Path + "/" + node.Name
		if err := target.Delete(ctx, node); err != nil {
			preview.Errors = append(preview.Errors, err)
		}
	}

	return preview, nil
}

func (e *Engine) copyNode(ctx context.Context, source, target connector.Connector, node model.Node, targetPath string) error {
	if node.IsDir() {
		// Create directory in target
		dirNode := model.Node{
			Path: targetPath + "/" + node.Name,
			Type: model.FolderNode,
		}
		return target.Write(ctx, dirNode, nil)
	}

	// Copy file
	r, err := source.Read(ctx, node)
	if err != nil {
		return err
	}
	defer r.Close()

	targetNode := node
	targetNode.Path = targetPath + "/" + node.Name

	return target.Write(ctx, targetNode, r)
}
