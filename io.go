// Copyright (c) 2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package x

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
func PrintErr(err error) {
	if err != nil {
		log.Printf("err: %v", err)
	}
}
