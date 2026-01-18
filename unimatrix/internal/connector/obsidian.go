// Copyright (c) 2024-2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package connector

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/teghnet/x/unimatrix/internal/model"
)

// Obsidian is a connector for Obsidian vaults.
// It extends Local with awareness of Obsidian-specific files.
type Obsidian struct {
	*Local
	excludePatterns []string
}

// NewObsidian creates a new Obsidian vault connector.
func NewObsidian(name, vaultPath string) *Obsidian {
	return &Obsidian{
		Local: NewLocal(name, vaultPath),
		excludePatterns: []string{
			".obsidian",
			".trash",
			".git",
		},
	}
}

// Connect implements Connector.
func (o *Obsidian) Connect(ctx context.Context) error {
	if err := o.Local.Connect(ctx); err != nil {
		return err
	}

	// Verify it's an Obsidian vault
	obsidianDir := filepath.Join(o.basePath, ".obsidian")
	if _, err := os.Stat(obsidianDir); os.IsNotExist(err) {
		// Not strictly required, just a warning
	}

	return nil
}

// List implements Connector with Obsidian-specific filtering.
func (o *Obsidian) List(ctx context.Context, path string) ([]model.Node, error) {
	nodes, err := o.Local.List(ctx, path)
	if err != nil {
		return nil, err
	}

	// Filter out Obsidian-specific directories
	filtered := make([]model.Node, 0, len(nodes))
	for _, node := range nodes {
		if !o.shouldExclude(node.Name) {
			// Add frontmatter metadata for markdown files
			if strings.HasSuffix(node.Name, ".md") {
				o.enrichWithFrontmatter(ctx, &node)
			}
			filtered = append(filtered, node)
		}
	}

	return filtered, nil
}

func (o *Obsidian) shouldExclude(name string) bool {
	for _, pattern := range o.excludePatterns {
		if name == pattern || strings.HasPrefix(name, pattern) {
			return true
		}
	}
	return false
}

func (o *Obsidian) enrichWithFrontmatter(ctx context.Context, node *model.Node) {
	r, err := o.Local.Read(ctx, *node)
	if err != nil {
		return
	}
	defer r.Close()

	content, err := io.ReadAll(r)
	if err != nil {
		return
	}

	frontmatter := parseFrontmatter(string(content))
	if frontmatter != nil {
		if node.Metadata == nil {
			node.Metadata = make(map[string]any)
		}
		node.Metadata["frontmatter"] = frontmatter
	}
}

// parseFrontmatter extracts YAML frontmatter from markdown content.
func parseFrontmatter(content string) map[string]any {
	if !strings.HasPrefix(content, "---") {
		return nil
	}

	end := strings.Index(content[3:], "---")
	if end == -1 {
		return nil
	}

	yamlContent := content[3 : end+3]
	yamlContent = strings.TrimSpace(yamlContent)

	// Simple key: value parsing (not full YAML)
	result := make(map[string]any)
	for _, line := range strings.Split(yamlContent, "\n") {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			result[key] = value
		}
	}

	if len(result) == 0 {
		return nil
	}
	return result
}

// VaultConfig holds Obsidian vault configuration.
type VaultConfig struct {
	AttachmentsFolder string `json:"attachmentFolderPath"`
	TrashOption       string `json:"trashOption"`
}

// LoadVaultConfig loads the Obsidian vault configuration.
func (o *Obsidian) LoadVaultConfig() (*VaultConfig, error) {
	configPath := filepath.Join(o.basePath, ".obsidian", "app.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config VaultConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// GetBacklinks returns notes that link to the given note.
func (o *Obsidian) GetBacklinks(ctx context.Context, notePath string) ([]model.Node, error) {
	noteName := strings.TrimSuffix(filepath.Base(notePath), ".md")
	linkPattern := "[[" + noteName + "]]"

	var backlinks []model.Node

	err := filepath.Walk(o.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && o.shouldExclude(info.Name()) {
			return filepath.SkipDir
		}
		if !strings.HasSuffix(path, ".md") {
			return nil
		}
		if path == filepath.Join(o.basePath, notePath) {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		if strings.Contains(string(content), linkPattern) {
			relPath, _ := filepath.Rel(o.basePath, path)
			node, _ := o.Stat(ctx, relPath)
			if node != nil {
				backlinks = append(backlinks, *node)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return backlinks, nil
}
