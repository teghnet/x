package bankparse

import (
	"context"
	"io"
	"iter"
)

// Parser defines the contract for all statement format parsers.
type Parser interface {
	// Parse reads from r and yields a sequence of Transactions. Iteration
	// stops after the first error, which is yielded alongside a nil
	// Transaction, or after ctx is done.
	Parse(ctx context.Context, r io.Reader) iter.Seq2[*Transaction, error]
}
