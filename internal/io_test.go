// Copyright (c) $year Paweł Zaremba
// SPDX-License-Identifier: MIT

package internal_test

import (
	"errors"
	"testing"

	"github.com/teghnet/x/internal"
)

// mockCloser is a mock implementation of io.Closer for testing.
type mockCloser struct {
	err error
}

func (m *mockCloser) Close() error {
	return m.err
}

func TestFatalClose_Success(t *testing.T) {
	m := &mockCloser{err: nil}
	// FatalClose should not panic when Close() returns nil
	internal.FatalClose(m)
}

func TestPrintClose_Success(t *testing.T) {
	m := &mockCloser{err: nil}
	// PrintClose should not panic when Close() returns nil
	internal.PrintClose(m)
}

func TestPrintClose_Error(t *testing.T) {
	m := &mockCloser{err: errors.New("close error")}
	// PrintClose should log but not panic when Close() returns an error
	internal.PrintClose(m)
}
