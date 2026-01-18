// Copyright (c) 2024-2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

// Package model defines the domain types for Unimatrix.
package model

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"time"
)

// NodeType represents the type of a filesystem node.
type NodeType int

const (
	FileNode NodeType = iota
	FolderNode
)

// Node represents a file or folder in any connected system.
type Node struct {
	ID        string
	Path      string
	Name      string
	Type      NodeType
	Size      int64
	ModTime   time.Time
	Checksum  string
	Connector string         // Which connector owns this node
	Metadata  map[string]any // Connector-specific metadata
}

// IsDir returns true if the node is a directory.
func (n Node) IsDir() bool {
	return n.Type == FolderNode
}

// String implements fmt.Stringer.
func (n Node) String() string {
	return n.Path
}

// ComputeChecksum calculates SHA256 checksum from a reader.
func ComputeChecksum(r io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
