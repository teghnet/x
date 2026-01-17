// Copyright (c) $year Paweł Zaremba
// SPDX-License-Identifier: MIT

package internal

import (
	"io"
	"log"
)

func FatalClose(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Fatalf("could not close: %v", err)
	}
}

func PrintClose(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Printf("could not close: %v", err)
	}
}
