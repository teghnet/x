// Copyright (c) 2024-2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Preview displays file content or directory info.
type Preview struct {
	title   string
	content string
}

// NewPreview creates a new Preview component.
func NewPreview() Preview {
	return Preview{
		title: "Preview",
		content: "Select a file to preview its contents.\n\n" +
			"Use j/k to navigate the tree.\n" +
			"Press Enter to expand folders.\n" +
			"Press Tab to switch panes.",
	}
}

// View renders the preview pane.
func (p Preview) View(width, height int, focused bool) string {
	var borderColor lipgloss.Color
	if focused {
		borderColor = lipgloss.Color("#00FF00")
	} else {
		borderColor = lipgloss.Color("#444444")
	}

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(width - 2).
		Height(height - 2)

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true).
		MarginBottom(1)

	contentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CCCCCC"))

	title := titleStyle.Render(p.title)
	content := contentStyle.Render(p.content)

	// Truncate content to fit
	maxLines := height - 4
	lines := strings.Split(content, "\n")
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		lines = append(lines, "...")
	}
	content = strings.Join(lines, "\n")

	return style.Render(title + "\n" + content)
}

// SetContent updates the preview content.
func (p *Preview) SetContent(title, content string) {
	p.title = title
	p.content = content
}

// Clear resets the preview to default state.
func (p *Preview) Clear() {
	p.title = "Preview"
	p.content = "Select a file to preview."
}
