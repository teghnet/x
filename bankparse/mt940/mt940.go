// Package mt940 parses bank statements in the SWIFT MT940 format.
//
// MT940 itself is strict about its tag grammar (:20:, :61:, :86:, ...), but
// banks differ wildly in how they pack information into Tag 86 (Information
// to Account Owner) — that's why parsing it is delegated to a pluggable
// Dialect rather than hard-coded here.
package mt940

import (
	"context"
	"fmt"
	"io"
	"iter"
	"regexp"
	"strings"

	"github.com/teghnet/x/bankparse"
)

// Dialect extracts counterparty and description information from a
// statement's Tag 86 (Information to Account Owner) text, whose structure is
// bank-specific and not standardized by SWIFT.
type Dialect interface {
	// ParseTag86 enriches tx from the unstructured raw Tag 86 value. tx has
	// already been populated from the preceding Field 61 (dates, amount,
	// type code, statement context).
	ParseTag86(raw string, tx *bankparse.Transaction) error
}

// tag86FieldRe matches the "?NN" structured subfield markers used by the
// DefaultDialect.
var tag86FieldRe = regexp.MustCompile(`\?(\d{2})`)

// DefaultDialect implements the "?NN" structured-subfield convention common
// among European banks: ?00 posting text, ?20-?29 free-text description
// lines, ?30 counterparty bank code, ?31 counterparty account, ?32/?33
// counterparty name. This is a common convention, not a SWIFT standard —
// banks that deviate need their own Dialect.
//
// If raw contains no "?NN" markers at all, the entire value is used as
// Description.
type DefaultDialect struct{}

// ParseTag86 implements Dialect.
func (DefaultDialect) ParseTag86(raw string, tx *bankparse.Transaction) error {
	locs := tag86FieldRe.FindAllStringSubmatchIndex(raw, -1)
	if len(locs) == 0 {
		tx.Description = strings.TrimSpace(raw)
		return nil
	}

	fields := make(map[string]string, len(locs))
	for i, loc := range locs {
		code := raw[loc[2]:loc[3]]
		end := len(raw)
		if i+1 < len(locs) {
			end = locs[i+1][0]
		}
		value := strings.TrimSpace(raw[loc[1]:end])
		if existing := fields[code]; existing != "" {
			value = existing + " " + value
		}
		fields[code] = value
	}

	var desc []string
	for n := 20; n <= 29; n++ {
		if v := fields[fmt.Sprintf("%02d", n)]; v != "" {
			desc = append(desc, v)
		}
	}
	if len(desc) > 0 {
		tx.Description = strings.Join(desc, " ")
	}
	if v := fields["00"]; v != "" {
		tx.Type = v
	}
	if v := fields["30"]; v != "" {
		tx.CounterpartyBankCode = v
	}
	if v := fields["31"]; v != "" {
		tx.CounterpartyAccount = v
	}

	var name []string
	for _, code := range [2]string{"32", "33"} {
		if v := fields[code]; v != "" {
			name = append(name, v)
		}
	}
	if len(name) > 0 {
		tx.CounterpartyName = strings.Join(name, " ")
	}

	return nil
}

// parser implements bankparse.Parser for MT940 statements.
type parser struct {
	dialect Dialect
	cfg     bankparse.Config
}

// New creates a bankparse.Parser for MT940 statements, using dialect to
// interpret each transaction's Tag 86. Only the Encoding option is honored;
// the other bankparse.Config fields don't apply to MT940's fixed grammar.
func New(dialect Dialect, opts ...bankparse.Option) bankparse.Parser {
	return &parser{dialect: dialect, cfg: bankparse.NewConfig(opts...)}
}

// Parse implements bankparse.Parser. It streams transactions from every
// statement message in r; MT940 files commonly concatenate several messages
// (one per account or per day), each terminated by a line containing only
// "-".
func (p *parser) Parse(ctx context.Context, r io.Reader) iter.Seq2[*bankparse.Transaction, error] {
	return func(yield func(*bankparse.Transaction, error) bool) {
		decoded, err := bankparse.DecodeReader(r, p.cfg.Encoding)
		if err != nil {
			yield(nil, fmt.Errorf("mt940: %w", err))
			return
		}

		var (
			account  string
			stmtNo   string
			currency string
			pending  *bankparse.Transaction
		)

		flushPending := func() bool {
			if pending == nil {
				return true
			}
			tx := pending
			pending = nil
			return yield(tx, nil)
		}

		for rec, err := range scan(decoded) {
			if err != nil {
				if !yield(nil, fmt.Errorf("mt940: %w", err)) {
					return
				}
				continue
			}
			if err := ctx.Err(); err != nil {
				yield(nil, err)
				return
			}

			switch rec.tag {
			case "-":
				if !flushPending() {
					return
				}
				account, stmtNo, currency = "", "", ""

			case "25":
				account = rec.value

			case "28", "28C":
				stmtNo = rec.value

			case "60F", "60M":
				if !flushPending() {
					return
				}
				bal, err := parseBalance(rec.value)
				if err != nil {
					if !yield(nil, fmt.Errorf("mt940: opening balance: %w", err)) {
						return
					}
					continue
				}
				currency = bal.Currency

			case "62F", "62M", "64", "65":
				if !flushPending() {
					return
				}
				if _, err := parseBalance(rec.value); err != nil {
					if !yield(nil, fmt.Errorf("mt940: balance: %w", err)) {
						return
					}
				}

			case "61":
				if !flushPending() {
					return
				}
				t61, err := parseTag61(rec.value)
				if err != nil {
					if !yield(nil, fmt.Errorf("mt940: statement line: %w", err)) {
						return
					}
					continue
				}
				tx := &bankparse.Transaction{
					TransactionDate:  t61.EntryDate,
					BookingDate:      t61.ValueDate,
					Amount:           t61.Amount,
					Currency:         currency,
					Type:             t61.TypeCode,
					Reference:        t61.CustomerRef,
					StatementAccount: account,
					StatementNumber:  stmtNo,
					RawData:          map[string]string{"dcMark": t61.Mark},
				}
				if t61.FundsCode != "" {
					tx.RawData["fundsCode"] = t61.FundsCode
				}
				if t61.BankRef != "" {
					tx.RawData["bankRef"] = t61.BankRef
				}
				if t61.Supplementary != "" {
					tx.RawData["supplementary"] = t61.Supplementary
				}
				pending = tx

			case "86":
				if pending == nil {
					continue
				}
				pending.RawData["tag86"] = rec.value
				if err := p.dialect.ParseTag86(rec.value, pending); err != nil {
					if !yield(nil, fmt.Errorf("mt940: tag 86: %w", err)) {
						return
					}
					continue
				}
				if !flushPending() {
					return
				}
			}
		}
		flushPending()
	}
}
