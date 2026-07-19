// Copyright (c) 2024-2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package internal_test

import (
	"errors"
	"testing"

	"github.com/teghnet/x"
)

// mockCloser is a mock implementation of io.Closer for testing.
type mockCloser struct {
	err error
}

func (m *mockCloser) Close() error {
	return m.err
}

func TestCloseFatal_Success(t *testing.T) {
	m := &mockCloser{err: nil}
	// CloseFatal should not panic when Close() returns nil
	x.CloseFatal(m)
}

func TestClosePrint_Success(t *testing.T) {
	m := &mockCloser{err: nil}
	// ClosePrint should not panic when Close() returns nil
	x.Close(m)
}

func TestClosePrint_Error(t *testing.T) {
	m := &mockCloser{err: errors.New("close error")}
	// ClosePrint should log but not panic when Close() returns an error
	x.Close(m)
}
