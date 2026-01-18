// Copyright (c) 2024-2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

// Package tui provides the terminal user interface for Unimatrix.
package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/teghnet/x/unimatrix/internal/tui/components"
)

// Model is the root Bubble Tea model for the Unimatrix TUI.
type Model struct {
	profile string
	width   int
	height  int

	// Components
	header    components.Header
	tree      components.Tree
	preview   components.Preview
	statusbar components.StatusBar

	// State
	focusedPane Pane
	showHelp    bool
}

// Pane represents which pane is currently focused.
type Pane int

const (
	TreePane Pane = iota
	PreviewPane
)

// New creates a new root TUI model.
func New(profile string) Model {
	return Model{
		profile:     profile,
		focusedPane: TreePane,
		header:      components.NewHeader(profile),
		tree:        components.NewTree(),
		preview:     components.NewPreview(),
		statusbar:   components.NewStatusBar(),
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m = m.updateLayout()

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.focusedPane = (m.focusedPane + 1) % 2
		case "?":
			m.showHelp = !m.showHelp
		}

		// Delegate to focused pane
		if m.focusedPane == TreePane {
			var cmd tea.Cmd
			m.tree, cmd = m.tree.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

// View implements tea.Model.
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	header := m.header.View(m.width)
	statusbar := m.statusbar.View(m.width)

	// Calculate available height for main content
	contentHeight := m.height - lipgloss.Height(header) - lipgloss.Height(statusbar)

	// Calculate pane widths (60/40 split)
	treeWidth := m.width * 6 / 10
	previewWidth := m.width - treeWidth - 1 // -1 for border

	// Render panes
	treeView := m.tree.View(treeWidth, contentHeight, m.focusedPane == TreePane)
	previewView := m.preview.View(previewWidth, contentHeight, m.focusedPane == PreviewPane)

	// Join panes horizontally
	content := lipgloss.JoinHorizontal(lipgloss.Top, treeView, previewView)

	// Stack vertically
	return lipgloss.JoinVertical(lipgloss.Left, header, content, statusbar)
}

func (m Model) updateLayout() Model {
	// Recalculate dimensions when window resizes
	return m
}
