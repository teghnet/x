// Copyright (c) 2024-2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package tui

import "github.com/charmbracelet/lipgloss"

// Borg-inspired color palette.
var (
	// Primary colors
	BorgGreen     = lipgloss.Color("#00FF00")
	BorgDarkGreen = lipgloss.Color("#004400")
	BorgBlack     = lipgloss.Color("#0a0a0a")
	BorgGray      = lipgloss.Color("#444444")
	BorgLightGray = lipgloss.Color("#888888")

	// Accent colors
	WarningRed   = lipgloss.Color("#FF4444")
	InfoBlue     = lipgloss.Color("#4488FF")
	SuccessGreen = lipgloss.Color("#44FF44")
)

// Styles defines the application styles.
var Styles = struct {
	// Layout
	Header    lipgloss.Style
	StatusBar lipgloss.Style
	Pane      lipgloss.Style
	PaneFocus lipgloss.Style

	// Text
	Title     lipgloss.Style
	Subtitle  lipgloss.Style
	Muted     lipgloss.Style
	Highlight lipgloss.Style

	// Tree
	TreeItem     lipgloss.Style
	TreeSelected lipgloss.Style
	TreeFolder   lipgloss.Style

	// Sync indicators
	SyncPending  lipgloss.Style
	SyncDone     lipgloss.Style
	SyncConflict lipgloss.Style
}{
	Header: lipgloss.NewStyle().
		Background(BorgDarkGreen).
		Foreground(BorgGreen).
		Bold(true).
		Padding(0, 1),

	StatusBar: lipgloss.NewStyle().
		Background(BorgGray).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 1),

	Pane: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(BorgGray),

	PaneFocus: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(BorgGreen),

	Title: lipgloss.NewStyle().
		Foreground(BorgGreen).
		Bold(true),

	Subtitle: lipgloss.NewStyle().
		Foreground(BorgLightGray),

	Muted: lipgloss.NewStyle().
		Foreground(BorgGray),

	Highlight: lipgloss.NewStyle().
		Foreground(BorgGreen).
		Bold(true),

	TreeItem: lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CCCCCC")),

	TreeSelected: lipgloss.NewStyle().
		Background(BorgDarkGreen).
		Foreground(BorgGreen).
		Bold(true),

	TreeFolder: lipgloss.NewStyle().
		Foreground(InfoBlue),

	SyncPending: lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFAA00")),

	SyncDone: lipgloss.NewStyle().
		Foreground(SuccessGreen),

	SyncConflict: lipgloss.NewStyle().
		Foreground(WarningRed).
		Bold(true),
}
