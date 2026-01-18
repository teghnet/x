// Copyright (c) 2024-2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

// Unimatrix is a TUI tool for syncing files between APIs and systems.
package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/teghnet/x/unimatrix/internal/tui"
)

var (
	profile = flag.String("profile", "zero", "sync profile (zero, one, two)")
	version = flag.Bool("version", false, "print version and exit")
)

func main() {
	flag.Parse()

	if *version {
		fmt.Println("unimatrix v0.1.0 - Resistance is futile.")
		os.Exit(0)
	}

	p := tea.NewProgram(
		tui.New(*profile),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
