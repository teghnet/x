// Copyright (c) 2024-2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package connector

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/teghnet/x/unimatrix/internal/model"
)

// Local is a connector for the local filesystem.
type Local struct {
	name     string
	basePath string
}

// NewLocal creates a new local filesystem connector.
func NewLocal(name, basePath string) *Local {
	return &Local{
		name:     name,
		basePath: basePath,
	}
}

// Name implements Connector.
func (l *Local) Name() string {
	return l.name
}

// Connect implements Connector.
func (l *Local) Connect(ctx context.Context) error {
	// Verify base path exists
	info, err := os.Stat(l.basePath)
	if err != nil {
		return fmt.Errorf("base path error: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("base path is not a directory: %s", l.basePath)
	}
	return nil
}

// Close implements Connector.
func (l *Local) Close() error {
	return nil
}

// List implements Connector.
func (l *Local) List(ctx context.Context, path string) ([]model.Node, error) {
	fullPath := filepath.Join(l.basePath, path)

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	nodes := make([]model.Node, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		nodeType := model.FileNode
		if entry.IsDir() {
			nodeType = model.FolderNode
		}

		nodePath := filepath.Join(path, entry.Name())
		nodes = append(nodes, model.Node{
			ID:        nodePath,
			Path:      nodePath,
			Name:      entry.Name(),
			Type:      nodeType,
			Size:      info.Size(),
			ModTime:   info.ModTime(),
			Connector: l.name,
		})
	}

	return nodes, nil
}

// Read implements Connector.
func (l *Local) Read(ctx context.Context, node model.Node) (io.ReadCloser, error) {
	fullPath := filepath.Join(l.basePath, node.Path)
	return os.Open(fullPath)
}

// Write implements Connector.
func (l *Local) Write(ctx context.Context, node model.Node, r io.Reader) error {
	fullPath := filepath.Join(l.basePath, node.Path)

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}

	f, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, r)
	return err
}

// Delete implements Connector.
func (l *Local) Delete(ctx context.Context, node model.Node) error {
	fullPath := filepath.Join(l.basePath, node.Path)
	if node.IsDir() {
		return os.RemoveAll(fullPath)
	}
	return os.Remove(fullPath)
}

// Stat implements Connector.
func (l *Local) Stat(ctx context.Context, path string) (*model.Node, error) {
	fullPath := filepath.Join(l.basePath, path)

	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, err
	}

	nodeType := model.FileNode
	if info.IsDir() {
		nodeType = model.FolderNode
	}

	return &model.Node{
		ID:        path,
		Path:      path,
		Name:      info.Name(),
		Type:      nodeType,
		Size:      info.Size(),
		ModTime:   info.ModTime(),
		Connector: l.name,
	}, nil
}
