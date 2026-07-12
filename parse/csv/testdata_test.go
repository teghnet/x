package csv

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/govalues/decimal"
)

// plnAmount wraps decimal.Decimal to parse Polish-formatted money cells such
// as `"1 003,73"` or `"-312,38"`: a comma decimal separator and an
// optional non-breaking-space thousands separator. It implements
// encoding.TextUnmarshaler, which Stream/Decode use for any field type that
// needs more than the built-in kinds.
type plnAmount struct {
	decimal.Decimal
}

func (a *plnAmount) UnmarshalText(text []byte) error {
	s := strings.Map(func(r rune) rune {
		if r == ' ' || r == '\u00a0' {
			return -1
		}
		return r
	}, string(text))
	if s == "" {
		return nil
	}
	d, err := decimal.Parse(strings.ReplaceAll(s, ",", "."))
	if err != nil {
		return err
	}
	a.Decimal = d
	return nil
}

// dateOnlyLayout is the plain calendar-date format used by Data płatności,
// Data wystawienia and Termin płatności (e.g. "2026-03-11"), as opposed to
// the RFC 3339 format time.Time.UnmarshalText expects.
const dateOnlyLayout = "2006-01-02"

// dateOnly wraps time.Time to parse dateOnlyLayout dates. It implements
// encoding.TextUnmarshaler, which Stream/Decode use for any field type that
// needs more than the built-in kinds.
type dateOnly struct {
	time.Time
}

func (d *dateOnly) UnmarshalText(text []byte) error {
	s := string(text)
	if s == "" {
		return nil
	}
	t, err := time.Parse(dateOnlyLayout, s)
	if err != nil {
		return err
	}
	d.Time = t
	return nil
}

// payment mirrors a row of testdata/testdata.csv, a real (anonymized)
// export of a Polish B2B payment ledger: a header leaving two columns
// blank (a leading account ID, targeted here by position, and an always-
// empty trailing column, left unmapped), rows with quoted comma-decimal
// amounts, a non-breaking-space thousands separator on larger amounts,
// plain calendar dates, a blank line partway through the file, and a
// negative (correction) amount.
type payment struct {
	AccountID      string    `csv:",0"`
	Type           string    `csv:"Typ płatności"`
	Payer          string    `csv:"Płatnik"`
	Amount         plnAmount `csv:"Kwota płatności [PLN]"`
	Date           dateOnly  `csv:"Data płatności"`
	Status         string    `csv:"Status"`
	PaymentID      string    `csv:"Identyfikator płatności"`
	Counterparty   string    `csv:"Kontrahent"`
	DocumentNumber string    `csv:"Numer dokumentu"`
	DocumentType   string    `csv:"Typ dokumentu"`
	IssueDate      dateOnly  `csv:"Data wystawienia"`
	DueDate        dateOnly  `csv:"Termin płatności"`
	GrossValue     plnAmount `csv:"Wartość brutto [PLN]"`
	Paid           plnAmount `csv:"Zapłacono [PLN]"`
}

func TestDecodeTestdata(t *testing.T) {
	f, err := os.Open("testdata/testdata.csv")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer f.Close()

	got, err := Decode[payment](f)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	// 5 data rows; the blank line partway through the file must be skipped,
	// not decoded into a zero-value row.
	if len(got) != 5 {
		t.Fatalf("got %d rows, want 5", len(got))
	}

	for i, p := range got {
		if p.AccountID != "123456789" {
			t.Errorf("row %d: AccountID = %q, want 123456789", i, p.AccountID)
		}
		if p.DocumentType != "SL" && p.DocumentType != "SK" {
			t.Errorf("row %d: DocumentType = %q", i, p.DocumentType)
		}
		// "Kwota płatności [PLN]" is blank in every row of this export.
		if !p.Amount.IsZero() {
			t.Errorf("row %d: Amount = %s, want 0", i, p.Amount)
		}
		// "Data płatności" is blank in every row of this export.
		if !p.Date.IsZero() {
			t.Errorf("row %d: Date = %s, want zero", i, p.Date)
		}
	}

	wantDate := func(t *testing.T, label string, got dateOnly, y int, m time.Month, d int) {
		t.Helper()
		want := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
		if !got.Equal(want) {
			t.Errorf("%s = %s, want %s", label, got, want)
		}
	}

	first := got[0]
	if first.Counterparty != "125736" {
		t.Errorf("first.Counterparty = %q, want 125736", first.Counterparty)
	}
	if first.DocumentNumber != "987654321" {
		t.Errorf("first.DocumentNumber = %q, want 987654321", first.DocumentNumber)
	}
	if got, want := first.GrossValue.String(), "80.13"; got != want {
		t.Errorf("first.GrossValue = %s, want %s", got, want)
	}
	if got, want := first.Paid.String(), "80.13"; got != want {
		t.Errorf("first.Paid = %s, want %s", got, want)
	}
	wantDate(t, "first.IssueDate", first.IssueDate, 2026, time.March, 11)
	wantDate(t, "first.DueDate", first.DueDate, 2026, time.May, 25)

	// "1 003,73": non-breaking-space thousands separator must be stripped.
	thousands := got[2]
	if got, want := thousands.GrossValue.String(), "1003.73"; got != want {
		t.Errorf("thousands.GrossValue = %s, want %s (thousands separator must be stripped)", got, want)
	}
	if got, want := thousands.Paid.String(), "1003.73"; got != want {
		t.Errorf("thousands.Paid = %s, want %s", got, want)
	}
	wantDate(t, "thousands.IssueDate", thousands.IssueDate, 2026, time.April, 2)
	wantDate(t, "thousands.DueDate", thousands.DueDate, 2026, time.May, 17)

	// "-312,38": a correction row with a negative amount.
	last := got[len(got)-1]
	if last.DocumentType != "SK" {
		t.Errorf("last.DocumentType = %q, want SK", last.DocumentType)
	}
	if got, want := last.GrossValue.String(), "-312.38"; got != want {
		t.Errorf("last.GrossValue = %s, want %s", got, want)
	}
	if !last.GrossValue.IsNeg() {
		t.Errorf("last.GrossValue = %s, want negative", last.GrossValue)
	}
	wantDate(t, "last.IssueDate", last.IssueDate, 2026, time.May, 25)
	wantDate(t, "last.DueDate", last.DueDate, 2026, time.May, 25)
}
