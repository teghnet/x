// Package csv parses the CSV export from mBank's "Lista operacji" (list of
// operations) report.
//
// That export isn't a clean transaction table: it starts with a variable
// number of preamble lines (bank letterhead, client name, selected account
// numbers, per-currency inflow/outflow summaries) that share no fixed shape
// with the transaction rows and aren't needed to build a Transaction. This
// parser scans past all of that looking for the fixed table header
// "#Data operacji;#Opis operacji;#Rachunek;#Kategoria;#Kwota;#Saldo po
// operacji;" and treats every row after it as a transaction, ignoring the
// blank lines mBank sometimes trails the file with.
//
// Amounts are Polish-formatted: a space thousands separator, a comma
// decimal separator, and a trailing ISO 4217 currency code (e.g.
// "-1 883,13 PLN").
package mbankcsv

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

const ParserName = "mbank_lista_operacji_csv"

func init() {
	banking.Register(ParserName, func() banking.Parser { return New() })
}

// dateLayout is the Go reference layout for the "Data operacji" column, e.g. "2026-07-10".
const dateLayout = "2006-01-02"

// tableHeader is the fixed header row marking the start of the transaction
// table, past the report's preamble.
const tableHeader = "#Data operacji"

// parser implements banking.Parser for mBank's "Lista operacji" CSV export.
type parser struct {
	cfg banking.Config
}

// New creates a banking.Parser for mBank's "Lista operacji" CSV export. Only
// Encoding is meaningful to override; the report's structure (delimiter,
// header, date format) is fixed by mBank, not configurable, though Delimiter
// can still be overridden if a variant export uses a different one.
func New(opts ...banking.Option) banking.Parser {
	cfg := banking.NewConfig(append([]banking.Option{banking.WithDelimiter(';')}, opts...)...)
	return &parser{cfg: cfg}
}

// Parse implements banking.Parser.
func (p *parser) Parse(ctx context.Context, r io.Reader) iter.Seq2[*banking.Transaction, error] {
	return func(yield func(*banking.Transaction, error) bool) {
		decoded, err := banking.DecodeReader(r, p.cfg.Encoding)
		if err != nil {
			yield(nil, fmt.Errorf("mbank/csv: %w", err))
			return
		}

		reader := csv.NewReader(decoded)
		reader.Comma = p.cfg.Delimiter
		reader.FieldsPerRecord = -1
		reader.LazyQuotes = true

		var inTable bool

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
				if !yield(nil, fmt.Errorf("mbank/csv: read row: %w", err)) {
					return
				}
				continue
			}

			first := ""
			if len(record) > 0 {
				first = strings.TrimSpace(record[0])
			}

			switch {
			case !inTable && first == tableHeader:
				inTable = true

			case inTable && isBlankRecord(record):
				// mBank sometimes trails the file with blank lines; ignore them.

			case inTable:
				tx, err := p.mapRow(record)
				if err != nil {
					if !yield(nil, fmt.Errorf("mbank/csv: map row: %w", err)) {
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
}

// isBlankRecord reports whether every field in record is empty or whitespace.
func isBlankRecord(record []string) bool {
	for _, f := range record {
		if strings.TrimSpace(f) != "" {
			return false
		}
	}
	return true
}

// mapRow builds a Transaction from a single transaction-table row:
// Data operacji, Opis operacji, Rachunek, Kategoria, Kwota, Saldo po operacji.
func (p *parser) mapRow(record []string) (*banking.Transaction, error) {
	if len(record) < 6 {
		return nil, fmt.Errorf("row has %d columns, want at least 6", len(record))
	}

	date, err := time.Parse(dateLayout, strings.TrimSpace(record[0]))
	if err != nil {
		return nil, fmt.Errorf("parse date %q: %w", record[0], err)
	}

	amount, currency, err := parseAmount(record[4])
	if err != nil {
		return nil, fmt.Errorf("parse amount %q: %w", record[4], err)
	}

	tx := &banking.Transaction{
		TransactionDate:  date,
		BookingDate:      date,
		Description:      strings.TrimSpace(record[1]),
		Category:         strings.TrimSpace(record[3]),
		Amount:           amount,
		Currency:         currency,
		StatementAccount: strings.TrimSpace(record[2]),
		RawData:          map[string]string{},
	}
	if balance := strings.TrimSpace(record[5]); balance != "" {
		tx.RawData["balance"] = balance
	}

	return tx, nil
}

// parseAmount parses an mBank money cell, e.g. "-1 883,13 PLN" or
// "85,68 PLN": a Polish-formatted decimal (space thousands separator, comma
// decimal separator) followed by a space and an ISO 4217 currency code.
func parseAmount(s string) (decimal.Decimal, string, error) {
	s = strings.TrimSpace(s)
	idx := strings.LastIndex(s, " ")
	if idx < 0 {
		return decimal.Decimal{}, "", fmt.Errorf("malformed amount %q: no currency", s)
	}
	amountStr, currency := s[:idx], s[idx+1:]
	amountStr = strings.ReplaceAll(amountStr, " ", "")
	amountStr = strings.ReplaceAll(amountStr, ",", ".")

	amount, err := decimal.Parse(amountStr)
	if err != nil {
		return decimal.Decimal{}, "", fmt.Errorf("parse amount %q: %w", amountStr, err)
	}
	return amount, currency, nil
}
