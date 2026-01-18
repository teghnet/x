// Copyright (c) 2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package ops

import (
	"crypto/sha256"
	"strings"
)

var zeroByte = string([]byte{0})

func Hash(line []string) []byte {
	h := sha256.Sum256([]byte(strings.Join(line, zeroByte)))
	return h[:16]
}
