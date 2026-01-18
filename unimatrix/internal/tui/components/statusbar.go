// Copyright (c) 2024-2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package components

import (
	"github.com/charmbracelet/lipgloss"
)

// StatusBar displays key bindings and current action.
type StatusBar struct {
	message string
}

// NewStatusBar creates a new StatusBar component.
func NewStatusBar() StatusBar {
	return StatusBar{}
}

// View renders the status bar.
func (s StatusBar) View(width int) string {
	style := lipgloss.NewStyle().
		Background(lipgloss.Color("#333333")).
		Foreground(lipgloss.Color("#CCCCCC")).
		Width(width).
		Padding(0, 1)

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888"))

	bindings := []struct {
		key  string
		desc string
	}{
		{"j/k", "navigate"},
		{"Enter", "expand"},
		{"Tab", "switch pane"},
		{"s", "sync"},
		{"?", "help"},
		{"q", "quit"},
	}

	var content string
	for i, b := range bindings {
		if i > 0 {
			content += descStyle.Render(" • ")
		}
		content += keyStyle.Render(b.key) + descStyle.Render(":"+b.desc)
	}

	if s.message != "" {
		content = s.message
	}

	return style.Render(" " + content + " ")
}

// SetMessage sets a custom status message.
func (s *StatusBar) SetMessage(msg string) {
	s.message = msg
}

// Clear clears the custom message.
func (s *StatusBar) Clear() {
	s.message = ""
}
