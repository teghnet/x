// Copyright (c) 2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package internal

import (
	"crypto/sha256"
	"strings"
)

var zeroByte = string([]byte{0})

func Hash(lines ...string) []byte {
	h := sha256.Sum256([]byte(strings.Join(lines, zeroByte)))
	return h[:16]
}
