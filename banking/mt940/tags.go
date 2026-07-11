package mt940

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/govalues/decimal"
)

// Debit/credit marks used in SWIFT Field 61 and the balance fields
// (60F/60M/62F/62M/64/65).
const (
	markCredit         = "C"
	markDebit          = "D"
	markReversalCredit = "RC"
	markReversalDebit  = "RD"
)

// parseSwiftDate parses a 6-digit YYMMDD date as used throughout MT940,
// pivoting the 2-digit year: 00-79 -> 20YY, 80-99 -> 19YY. Statements are
// always recent, so this is unambiguous in practice.
func parseSwiftDate(yyMMdd string) (time.Time, error) {
	if len(yyMMdd) != 6 {
		return time.Time{}, fmt.Errorf("invalid date %q: want 6 digits (YYMMDD)", yyMMdd)
	}
	yy, err := strconv.Atoi(yyMMdd[0:2])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date %q: %w", yyMMdd, err)
	}
	month, err := strconv.Atoi(yyMMdd[2:4])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date %q: %w", yyMMdd, err)
	}
	day, err := strconv.Atoi(yyMMdd[4:6])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date %q: %w", yyMMdd, err)
	}
	year := 2000 + yy
	if yy >= 80 {
		year = 1900 + yy
	}
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC), nil
}

// entryDate combines a 4-digit MMDD entry date with the year of valueDate,
// per the SWIFT convention that Field 61's entry date shares the value
// date's year (it can roll into the next year, e.g. value date late
// December with an entry date in January).
func entryDate(valueDate time.Time, mmdd string) (time.Time, error) {
	if len(mmdd) != 4 {
		return time.Time{}, fmt.Errorf("invalid entry date %q: want 4 digits (MMDD)", mmdd)
	}
	month, err := strconv.Atoi(mmdd[0:2])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid entry date %q: %w", mmdd, err)
	}
	day, err := strconv.Atoi(mmdd[2:4])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid entry date %q: %w", mmdd, err)
	}
	year := valueDate.Year()
	if time.Month(month) < valueDate.Month() {
		year++
	}
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC), nil
}

// tag61Re matches SWIFT Field 61 (Statement Line):
//
//	6!n     value date (YYMMDD)
//	[4!n]   entry date (MMDD)
//	2a      debit/credit mark: C, D, RC (reversal of credit), RD (reversal of debit)
//	[1!a]   funds code (3rd character of the currency, if it differs from the statement currency)
//	15d     amount, comma as decimal separator
//	[1!a3!c] transaction type identification code, e.g. NMSC, NTRF
//	16x     customer reference, up to "//" or end of field
//	[//16x] bank reference
//	[CrLf34x] supplementary details, on a continuation line
var tag61Re = regexp.MustCompile(`^(\d{6})(\d{4})?(RC|RD|C|D)([A-Z])?(\d+,\d*)([A-Za-z]\w{3})?(.*)$`)

// tag61 is the decoded form of a SWIFT Field 61 statement line.
type tag61 struct {
	ValueDate     time.Time
	EntryDate     time.Time // zero if not present; equals ValueDate otherwise
	Mark          string    // one of C, D, RC, RD
	FundsCode     string
	Amount        decimal.Decimal // signed: negative for debits, per Mark
	TypeCode      string
	CustomerRef   string
	BankRef       string
	Supplementary string
}

// parseTag61 decodes a Field 61 value (the record's value with the leading
// ":61:" already stripped).
func parseTag61(raw string) (tag61, error) {
	first, supplementary, _ := strings.Cut(raw, "\n")

	m := tag61Re.FindStringSubmatch(first)
	if m == nil {
		return tag61{}, fmt.Errorf("malformed field 61 %q", raw)
	}

	valueDate, err := parseSwiftDate(m[1])
	if err != nil {
		return tag61{}, err
	}

	out := tag61{
		ValueDate:     valueDate,
		EntryDate:     valueDate,
		Mark:          m[3],
		FundsCode:     m[4],
		TypeCode:      m[6],
		Supplementary: supplementary,
	}
	if m[2] != "" {
		out.EntryDate, err = entryDate(valueDate, m[2])
		if err != nil {
			return tag61{}, err
		}
	}

	amount, err := decimal.Parse(strings.Replace(m[5], ",", ".", 1))
	if err != nil {
		return tag61{}, fmt.Errorf("parse amount %q: %w", m[5], err)
	}
	switch out.Mark {
	case markDebit, markReversalCredit: // debit, or reversal of a credit -> functions as a debit
		amount = amount.Neg()
	case markCredit, markReversalDebit: // credit, or reversal of a debit -> functions as a credit
	default:
		return tag61{}, fmt.Errorf("unknown debit/credit mark %q", out.Mark)
	}
	out.Amount = amount

	refs := m[7]
	out.CustomerRef, out.BankRef, _ = strings.Cut(refs, "//")

	return out, nil
}

// balanceRe matches SWIFT Fields 60F/60M/62F/62M/64/65 (opening/closing/
// available balance): 1!a (D/C mark) 6!n (date, YYMMDD) 3!a (currency) 15d (amount).
var balanceRe = regexp.MustCompile(`^([CD])(\d{6})([A-Z]{3})(\d+,\d*)$`)

// balance is the decoded form of a SWIFT balance field (60F/60M/62F/62M/64/65).
type balance struct {
	Date     time.Time
	Currency string
	Amount   decimal.Decimal
}

// parseBalance decodes a balance field value (the record's value with the
// leading tag, e.g. ":60F:", already stripped).
func parseBalance(raw string) (balance, error) {
	m := balanceRe.FindStringSubmatch(raw)
	if m == nil {
		return balance{}, fmt.Errorf("malformed balance field %q", raw)
	}
	date, err := parseSwiftDate(m[2])
	if err != nil {
		return balance{}, err
	}
	amount, err := decimal.Parse(strings.Replace(m[4], ",", ".", 1))
	if err != nil {
		return balance{}, fmt.Errorf("parse amount %q: %w", m[4], err)
	}
	if m[1] == markDebit {
		amount = amount.Neg()
	}
	return balance{Date: date, Currency: m[3], Amount: amount}, nil
}
