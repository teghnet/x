// Copyright (c) 2024-2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package internal

import (
	"io"
	"log"
)

// CloseFatal closes the given Closer and calls log.Fatalf on error.
func CloseFatal(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Fatalf("could not close: %v", err)
	}
}

// ClosePrint closes the given Closer and logs any error without terminating.
func ClosePrint(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Printf("could not close: %v", err)
	}
}
