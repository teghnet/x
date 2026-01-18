// Copyright (c) 2024-2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Help displays a help overlay with key bindings.
type Help struct{}

// NewHelp creates a new Help component.
func NewHelp() Help {
	return Help{}
}

// View renders the help overlay.
func (h Help) View(width, height int) string {
	// Styling
	overlay := lipgloss.NewStyle().
		Background(lipgloss.Color("#1a1a1a")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00FF00")).
		Padding(1, 2).
		Width(50)

	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true).
		Width(46).
		Align(lipgloss.Center).
		MarginBottom(1).
		Render("◼ UNIMATRIX HELP")

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true).
		Align(lipgloss.Left).
		Width(20)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CCCCCC")).
		Align(lipgloss.Right).
		Width(26)

	bindings := []struct {
		key  string
		desc string
	}{
		{"j / ↓", "Move down"},
		{"k / ↑", "Move up"},
		{"Enter", "Expand/collapse folder"},
		{"h / ←", "Collapse folder"},
		{"l / →", "Expand folder"},
		{"Tab", "Switch pane"},
		{"s", "Sync selected"},
		{"S", "Sync all"},
		{"r", "Refresh"},
		{"?", "Toggle help"},
		{"q", "Quit"},
	}

	var keys, descs []string
	for _, b := range bindings {
		keys = append(keys, b.key)
		descs = append(descs, b.desc)
	}

	keyCol := keyStyle.Render(strings.Join(keys, "\n"))
	descCol := descStyle.Render(strings.Join(descs, "\n"))

	content := title + "\n\n" + lipgloss.JoinHorizontal(lipgloss.Top, keyCol, descCol)
	helpBox := overlay.Render(content)

	// Center the help box
	boxWidth := lipgloss.Width(helpBox)
	boxHeight := lipgloss.Height(helpBox)

	padLeft := (width - boxWidth) / 2
	padTop := (height - boxHeight) / 2

	if padLeft < 0 {
		padLeft = 0
	}
	if padTop < 0 {
		padTop = 0
	}

	// Build centered output
	var result strings.Builder
	for i := 0; i < padTop; i++ {
		result.WriteString("\n")
	}

	for _, line := range strings.Split(helpBox, "\n") {
		result.WriteString(strings.Repeat(" ", padLeft))
		result.WriteString(line)
		result.WriteString("\n")
	}

	return result.String()
}
