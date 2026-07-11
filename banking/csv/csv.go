// Package csv parses bank statements exported as CSV, driven by a
// declarative ColumnMapper rather than one hand-written parser per bank.
package csv

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"iter"
	"strings"
	"time"

	"github.com/govalues/decimal"

	"github.com/teghnet/x/banking"
)

// errPrefix identifies errors and panics originating from this package.
const errPrefix = "csv"

// rawColumnKeyFormat is the fmt.Sprintf pattern used to key each raw CSV
// column in Transaction.RawData, e.g. "col0", "col1".
const rawColumnKeyFormat = "col%d"

// ColumnMapper defines how a row of CSV translates to a denormalized
// Transaction. Optional columns are *int so a nil pointer unambiguously means
// "not present in this format" — use new(0), new(1), ... to reference a
// column, including column 0.
type ColumnMapper struct {
	// TransactionDateIdx and AmountIdx are required: New panics if either is
	// negative.
	TransactionDateIdx int
	AmountIdx          int

	BookingDateIdx      *int // nil: TransactionDate is used as BookingDate too
	CurrencyIdx         *int // nil: Transaction.Currency is left empty
	CounterpartyNameIdx *int
	CounterpartyAcctIdx *int
	// DescriptionIndices are joined with a space to form Description; nil or
	// empty means no description column.
	DescriptionIndices []int

	// CustomFn runs after the standard mapping above, with the raw record and
	// the Transaction built so far. It can override or augment any field, and
	// is the escape hatch for bank-specific quirks that don't fit ColumnMapper.
	CustomFn func(record []string, tx *banking.Transaction) error
}

// parser implements banking.Parser for a single bank's CSV export.
type parser struct {
	mapper ColumnMapper
	cfg    banking.Config
}

// New creates a banking.Parser for CSV statements shaped as described by
// mapper. It panics if mapper.AmountIdx or mapper.TransactionDateIdx is
// negative, since a parser that can't locate the amount or date column can't
// do anything useful.
func New(mapper ColumnMapper, opts ...banking.Option) banking.Parser {
	if mapper.AmountIdx < 0 {
		panic(errPrefix + ": ColumnMapper.AmountIdx must be set")
	}
	if mapper.TransactionDateIdx < 0 {
		panic(errPrefix + ": ColumnMapper.TransactionDateIdx must be set")
	}
	return &parser{mapper: mapper, cfg: banking.NewConfig(opts...)}
}

// Parse implements banking.Parser.
func (p *parser) Parse(ctx context.Context, r io.Reader) iter.Seq2[*banking.Transaction, error] {
	return func(yield func(*banking.Transaction, error) bool) {
		decoded, err := banking.DecodeReader(r, p.cfg.Encoding)
		if err != nil {
			yield(nil, fmt.Errorf("%s: %w", errPrefix, err))
			return
		}

		reader := csv.NewReader(decoded)
		reader.Comma = p.cfg.Delimiter
		reader.FieldsPerRecord = -1
		reader.LazyQuotes = true

		for range p.cfg.SkipHeaderLines {
			if _, err := reader.Read(); err != nil {
				if err == io.EOF {
					return
				}
				yield(nil, fmt.Errorf("%s: skip header: %w", errPrefix, err))
				return
			}
		}

		for {
			if err := ctx.Err(); err != nil {
				yield(nil, err)
				return
			}
			record, err := reader.Read()
			if err == io.EOF {
				return
			}
			if err != nil {
				if !yield(nil, fmt.Errorf("%s: read row: %w", errPrefix, err)) {
					return
				}
				continue
			}

			tx, err := p.mapRow(record)
			if err != nil {
				if !yield(nil, fmt.Errorf("%s: map row: %w", errPrefix, err)) {
					return
				}
				continue
			}
			if !yield(tx, nil) {
				return
			}
		}
	}
}

// mapRow builds a Transaction from a single CSV record according to p.mapper.
func (p *parser) mapRow(record []string) (*banking.Transaction, error) {
	tx := &banking.Transaction{RawData: make(map[string]string, len(record))}
	for i, v := range record {
		tx.RawData[fmt.Sprintf(rawColumnKeyFormat, i)] = v
	}

	required := func(idx int) (string, error) {
		if idx < 0 || idx >= len(record) {
			return "", fmt.Errorf("column %d out of range (row has %d columns)", idx, len(record))
		}
		return strings.TrimSpace(record[idx]), nil
	}
	optional := func(idx *int) (string, bool) {
		if idx == nil || *idx < 0 || *idx >= len(record) {
			return "", false
		}
		return strings.TrimSpace(record[*idx]), true
	}

	date, err := required(p.mapper.TransactionDateIdx)
	if err != nil {
		return nil, fmt.Errorf("transaction date: %w", err)
	}
	txDate, err := time.Parse(p.cfg.DateFormat, date)
	if err != nil {
		return nil, fmt.Errorf("parse transaction date %q: %w", date, err)
	}
	tx.TransactionDate = txDate
	tx.BookingDate = txDate

	if bookingDate, ok := optional(p.mapper.BookingDateIdx); ok {
		bd, err := time.Parse(p.cfg.DateFormat, bookingDate)
		if err != nil {
			return nil, fmt.Errorf("parse booking date %q: %w", bookingDate, err)
		}
		tx.BookingDate = bd
	}

	amountStr, err := required(p.mapper.AmountIdx)
	if err != nil {
		return nil, fmt.Errorf("amount: %w", err)
	}
	amount, err := decimal.Parse(strings.ReplaceAll(amountStr, ",", "."))
	if err != nil {
		return nil, fmt.Errorf("parse amount %q: %w", amountStr, err)
	}
	tx.Amount = amount

	if currency, ok := optional(p.mapper.CurrencyIdx); ok {
		tx.Currency = currency
	}
	if name, ok := optional(p.mapper.CounterpartyNameIdx); ok {
		tx.CounterpartyName = name
	}
	if acct, ok := optional(p.mapper.CounterpartyAcctIdx); ok {
		tx.CounterpartyAccount = acct
	}

	if len(p.mapper.DescriptionIndices) > 0 {
		parts := make([]string, 0, len(p.mapper.DescriptionIndices))
		for _, idx := range p.mapper.DescriptionIndices {
			if idx < 0 || idx >= len(record) {
				continue
			}
			if v := strings.TrimSpace(record[idx]); v != "" {
				parts = append(parts, v)
			}
		}
		tx.Description = strings.Join(parts, " ")
	}

	if p.mapper.CustomFn != nil {
		if err := p.mapper.CustomFn(record, tx); err != nil {
			return nil, fmt.Errorf("custom mapping: %w", err)
		}
	}

	return tx, nil
}
