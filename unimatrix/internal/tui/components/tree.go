// Copyright (c) 2024-2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TreeNode represents a node in the file tree.
type TreeNode struct {
	Name       string
	Path       string
	IsDir      bool
	Expanded   bool
	SyncStatus SyncStatus
	Children   []TreeNode
	Depth      int
}

// SyncStatus represents the sync state of a node.
type SyncStatus int

const (
	SyncNone SyncStatus = iota
	SyncPending
	SyncDone
	SyncConflict
)

// Tree is a hierarchical file browser component.
type Tree struct {
	nodes    []TreeNode
	cursor   int
	flatList []flatNode
}

type flatNode struct {
	node  *TreeNode
	depth int
}

// NewTree creates a new Tree component with mock data.
func NewTree() Tree {
	// Mock data for initial development
	nodes := []TreeNode{
		{
			Name:     "Local",
			Path:     "local://",
			IsDir:    true,
			Expanded: true,
			Children: []TreeNode{
				{
					Name:     "Documents",
					Path:     "local://Documents",
					IsDir:    true,
					Expanded: true,
					Children: []TreeNode{
						{Name: "README.md", Path: "local://Documents/README.md", SyncStatus: SyncDone},
						{Name: "notes.md", Path: "local://Documents/notes.md", SyncStatus: SyncPending},
					},
				},
				{
					Name:     "Projects",
					Path:     "local://Projects",
					IsDir:    true,
					Expanded: false,
					Children: []TreeNode{
						{Name: "unimatrix/", Path: "local://Projects/unimatrix", IsDir: true},
					},
				},
			},
		},
		{
			Name:     "Notion",
			Path:     "notion://",
			IsDir:    true,
			Expanded: false,
			Children: []TreeNode{
				{Name: "Workspace", Path: "notion://Workspace", IsDir: true},
			},
		},
	}

	t := Tree{nodes: nodes}
	t.flatten()
	return t
}

// flatten creates a flat list of visible nodes for rendering.
func (t *Tree) flatten() {
	t.flatList = nil
	for i := range t.nodes {
		t.flattenNode(&t.nodes[i], 0)
	}
}

func (t *Tree) flattenNode(node *TreeNode, depth int) {
	t.flatList = append(t.flatList, flatNode{node: node, depth: depth})
	if node.IsDir && node.Expanded {
		for i := range node.Children {
			t.flattenNode(&node.Children[i], depth+1)
		}
	}
}

// Update handles key events.
func (t Tree) Update(msg tea.Msg) (Tree, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if t.cursor < len(t.flatList)-1 {
				t.cursor++
			}
		case "k", "up":
			if t.cursor > 0 {
				t.cursor--
			}
		case "enter", "l", "right":
			if t.cursor < len(t.flatList) {
				node := t.flatList[t.cursor].node
				if node.IsDir {
					node.Expanded = !node.Expanded
					t.flatten()
				}
			}
		case "h", "left":
			if t.cursor < len(t.flatList) {
				node := t.flatList[t.cursor].node
				if node.IsDir && node.Expanded {
					node.Expanded = false
					t.flatten()
				}
			}
		}
	}
	return t, nil
}

// View renders the tree.
func (t Tree) View(width, height int, focused bool) string {
	var style lipgloss.Style
	if focused {
		style = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#00FF00")).
			Width(width - 2).
			Height(height - 2)
	} else {
		style = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#444444")).
			Width(width - 2).
			Height(height - 2)
	}

	var lines []string
	for i, fn := range t.flatList {
		line := t.renderNode(fn, i == t.cursor)
		lines = append(lines, line)
	}

	content := strings.Join(lines, "\n")
	return style.Render(content)
}

func (t Tree) renderNode(fn flatNode, selected bool) string {
	indent := strings.Repeat("  ", fn.depth)

	// Icon
	var icon string
	if fn.node.IsDir {
		if fn.node.Expanded {
			icon = "▼ "
		} else {
			icon = "▶ "
		}
	} else {
		icon = "  "
	}

	// Sync indicator
	var syncIcon string
	switch fn.node.SyncStatus {
	case SyncPending:
		syncIcon = " →"
	case SyncDone:
		syncIcon = " ✓"
	case SyncConflict:
		syncIcon = " ⚠"
	}

	name := fn.node.Name
	if fn.node.IsDir && !strings.HasSuffix(name, "/") {
		name += "/"
	}

	line := indent + icon + name + syncIcon

	if selected {
		return lipgloss.NewStyle().
			Background(lipgloss.Color("#004400")).
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true).
			Render(line)
	}

	if fn.node.IsDir {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4488FF")).
			Render(line)
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CCCCCC")).
		Render(line)
}

// Selected returns the currently selected node.
func (t Tree) Selected() *TreeNode {
	if t.cursor < len(t.flatList) {
		return t.flatList[t.cursor].node
	}
	return nil
}
