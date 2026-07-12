package csv

import (
	"os"
	"strings"
	"testing"

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

// payment mirrors a row of testdata/testdata.csv, a real (anonymized)
// export of a Polish B2B payment ledger: a header naming most but not all
// columns (two are blank in the header, and left unmapped here), rows with
// quoted comma-decimal amounts, a non-breaking-space thousands separator on
// larger amounts, a blank line partway through the file, and a negative
// (correction) amount.
type payment struct {
	Type           string    `csv:"Typ płatności"`
	Payer          string    `csv:"Płatnik"`
	Amount         plnAmount `csv:"Kwota płatności [PLN]"`
	Date           string    `csv:"Data płatności"`
	Status         string    `csv:"Status"`
	PaymentID      string    `csv:"Identyfikator płatności"`
	Counterparty   string    `csv:"Kontrahent"`
	DocumentNumber string    `csv:"Numer dokumentu"`
	DocumentType   string    `csv:"Typ dokumentu"`
	IssueDate      string    `csv:"Data wystawienia"`
	DueDate        string    `csv:"Termin płatności"`
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
		if p.DocumentType != "SL" && p.DocumentType != "SK" {
			t.Errorf("row %d: DocumentType = %q", i, p.DocumentType)
		}
		// "Kwota płatności [PLN]" is blank in every row of this export.
		if !p.Amount.IsZero() {
			t.Errorf("row %d: Amount = %s, want 0", i, p.Amount)
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

	// "1 003,73": non-breaking-space thousands separator must be stripped.
	thousands := got[2]
	if got, want := thousands.GrossValue.String(), "1003.73"; got != want {
		t.Errorf("thousands.GrossValue = %s, want %s (thousands separator must be stripped)", got, want)
	}
	if got, want := thousands.Paid.String(), "1003.73"; got != want {
		t.Errorf("thousands.Paid = %s, want %s", got, want)
	}

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
}
