// Copyright (c) 2024-2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

// Package components provides reusable TUI components for Unimatrix.
package components

import (
	"github.com/charmbracelet/lipgloss"
)

// Header displays the application title and status.
type Header struct {
	profile  string
	status   string
	lastSync string
}

// NewHeader creates a new Header component.
func NewHeader(profile string) Header {
	return Header{
		profile:  profile,
		status:   "Connected",
		lastSync: "Never",
	}
}

// View renders the header.
func (h Header) View(width int) string {
	// Borg-themed header styling
	headerStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#004400")).
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true).
		Width(width).
		Padding(0, 1)

	title := "◼ UNIMATRIX"
	if h.profile != "" {
		title += " " + h.profile
	}

	status := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#44FF44")).
		Render("● " + h.status)

	lastSync := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Render("Last sync: " + h.lastSync)

	// Right-align status info
	leftSide := title
	rightSide := status + "  " + lastSync
	gap := width - lipgloss.Width(leftSide) - lipgloss.Width(rightSide) - 4

	if gap < 1 {
		gap = 1
	}

	content := leftSide + lipgloss.NewStyle().Width(gap).Render("") + rightSide

	return headerStyle.Render(content)
}

// SetStatus updates the connection status.
func (h *Header) SetStatus(status string) {
	h.status = status
}

// SetLastSync updates the last sync time.
func (h *Header) SetLastSync(t string) {
	h.lastSync = t
}
