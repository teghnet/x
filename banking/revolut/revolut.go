// Package revolut parses the CSV export from Revolut's Personal Finance
// report ("Current Accounts Transaction Statements").
//
// That export isn't a flat transaction table: it's a multi-section report
// (account summaries, deposit values, per-currency transaction statements,
// crypto summaries) with repeated "---------" separators and per-currency
// sub-headers. This parser scans the whole document looking for
// "Transaction statement" tables (one per currency wallet the account
// holds) and yields their rows as Transactions, ignoring everything else —
// including the crypto sale/acquisition tables, which describe asset lots
// rather than cash movements and don't fit the Transaction model.
//
// For the base-currency wallet (PLN), the "Money in/out" column holds a
// single value (e.g. "20.00 PLN"). For any other currency, it holds the
// account-native amount followed by the PLN equivalent in parentheses (e.g.
// "$20.00 (77.97 PLN)"); Transaction.Amount/Currency reflect the
// account-native value, and the PLN equivalent is preserved in RawData.
package revolut

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"iter"
	"regexp"
	"strings"
	"time"

	"github.com/teghnet/x/banking"
)

const ParserName = "revolut_consolidated_csv"

func init() {
	banking.Register(ParserName, func() banking.Parser { return New() })
}

// errPrefix identifies errors originating from this package.
const errPrefix = "revolut"

// dateLayout is the Go reference layout for the "Date" column, e.g. "Nov 26, 2018".
const dateLayout = "Jan 2, 2006"

// Markers used to navigate the report's structure: the table marker and
// header row that open a currency wallet's "Transaction statement", and the
// row that closes it.
const (
	transactionStatementMarker = "Transaction statement"
	dateColumnHeader           = "Date"
	totalRowMarker             = "Total"
)

// statementAccountFormat builds Transaction.StatementAccount, e.g.
// "Personal Account (USD)".
const statementAccountFormat = "Personal Account (%s)"

// Keys under which mapRow stores money-cell details that don't map to a
// standard Transaction field.
const (
	rawKeyBaseAmount   = "baseAmount"
	rawKeyBaseCurrency = "baseCurrency"
	rawKeyBalance      = "balance"
	rawKeyTaxWithheld  = "taxWithheld"
	rawKeyOtherTaxes   = "otherTaxes"
	rawKeyFees         = "fees"
)

// sectionHeaderRe matches a currency wallet's section header, e.g.
// "Personal Account (PLN)".
var sectionHeaderRe = regexp.MustCompile(`^Personal Account \(([A-Za-z]+)\)$`)

// parser implements banking.Parser for Revolut's Personal Finance CSV export.
type parser struct {
	cfg banking.Config
}

// New creates a banking.Parser for Revolut's Personal Finance CSV export.
// Only Encoding and Delimiter are honored; the report's structure (headers,
// date format) is fixed by Revolut, not configurable.
func New(opts ...banking.Option) banking.Parser {
	return &parser{cfg: banking.NewConfig(opts...)}
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

		var (
			currency     string
			expectHeader bool
			inTable      bool
		)

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
			first := ""
			if len(record) > 0 {
				first = strings.TrimSpace(record[0])
			}

			switch {
			case sectionHeaderRe.MatchString(first):
				currency = sectionHeaderRe.FindStringSubmatch(first)[1]
				expectHeader, inTable = false, false

			case first == transactionStatementMarker:
				expectHeader, inTable = true, false

			case expectHeader:
				expectHeader = false
				if first != dateColumnHeader {
					if !yield(nil, fmt.Errorf("%s: expected transaction table header, got %q", errPrefix, first)) {
						return
					}
					continue
				}
				inTable = true

			case inTable && (first == "" || first == totalRowMarker):
				inTable = false

			case inTable:
				tx, err := p.mapRow(record, currency)
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
}

// mapRow builds a Transaction from a single transaction-table row:
// Date, Description, Category, Money in/out, Balance, Tax withheld, Other taxes, Fees.
func (p *parser) mapRow(record []string, section string) (*banking.Transaction, error) {
	if len(record) < 4 {
		return nil, fmt.Errorf("row has %d columns, want at least 4", len(record))
	}

	date, err := time.Parse(dateLayout, strings.TrimSpace(record[0]))
	if err != nil {
		return nil, fmt.Errorf("parse date %q: %w", record[0], err)
	}

	m, err := parseMoneyCell(record[3])
	if err != nil {
		return nil, fmt.Errorf("parse money in/out %q: %w", record[3], err)
	}

	tx := &banking.Transaction{
		TransactionDate:  date,
		BookingDate:      date,
		Description:      strings.TrimSpace(record[1]),
		Category:         strings.TrimSpace(record[2]),
		Amount:           m.Amount,
		Currency:         m.Currency,
		StatementAccount: fmt.Sprintf(statementAccountFormat, section),
		RawData:          map[string]string{},
	}
	if m.HasBase {
		tx.RawData[rawKeyBaseAmount] = m.BaseAmount.String()
		tx.RawData[rawKeyBaseCurrency] = m.BaseCurrency
	}
	for i, name := range [...]string{rawKeyBalance, rawKeyTaxWithheld, rawKeyOtherTaxes, rawKeyFees} {
		if idx := 4 + i; idx < len(record) {
			tx.RawData[name] = strings.TrimSpace(record[idx])
		}
	}

	return tx, nil
}
