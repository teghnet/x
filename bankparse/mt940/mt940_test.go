package mt940

import (
	"strings"
	"testing"
	"time"

	"github.com/teghnet/x/bankparse"
)

// fixture is a synthetic MT940 document constructed to exercise this
// package's parsing rules. It is not taken from any real bank export.
const fixture = "" +
	":20:STMT0001\r\n" +
	":25:PL61109010140000071219812874\r\n" +
	":28C:00001/001\r\n" +
	":60F:C260101EUR1000,00\r\n" +
	":61:2601050105D250,00NMSCNONREF//BANKREF01\r\n" +
	":86:?00PMNT?20SALARY PAYMENT?21FOR JANUARY?30ABCDPLPW?31PL27114020040000300201355387?32JOHN DOE\r\n" +
	":61:2601060106C1500,50NTRFCUSTREF002\r\n" +
	":86:free text without structured subfields\r\n" +
	":61:2601070107RD75,25NMSCNONREF\r\n" +
	":62F:C260107EUR2175,25\r\n" +
	"-\r\n" +
	":20:STMT0002\r\n" +
	":25:PL61109010140000071219812874\r\n" +
	":28C:00002/001\r\n" +
	":60M:C260107EUR2175,25\r\n" +
	":61:2601100110D2175,25RCMSCNONREF\r\n" +
	":62F:C260110EUR0,00\r\n" +
	"-\r\n"

func collect(t *testing.T) []*bankparse.Transaction {
	t.Helper()
	p := New(DefaultDialect{})
	var txs []*bankparse.Transaction
	for tx, err := range p.Parse(t.Context(), strings.NewReader(fixture)) {
		if err != nil {
			t.Fatalf("parse: %v", err)
		}
		txs = append(txs, tx)
	}
	return txs
}

func TestParseFixture(t *testing.T) {
	txs := collect(t)
	if len(txs) != 4 {
		t.Fatalf("got %d transactions, want 4", len(txs))
	}

	tx0 := txs[0]
	if got, want := tx0.Amount.String(), "-250.00"; got != want {
		t.Errorf("txs[0].Amount = %s, want %s", got, want)
	}
	if !tx0.BookingDate.Equal(time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("txs[0].BookingDate = %v", tx0.BookingDate)
	}
	if !tx0.TransactionDate.Equal(time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("txs[0].TransactionDate = %v", tx0.TransactionDate)
	}
	if tx0.Currency != "EUR" {
		t.Errorf("txs[0].Currency = %q, want EUR", tx0.Currency)
	}
	if tx0.StatementAccount != "PL61109010140000071219812874" {
		t.Errorf("txs[0].StatementAccount = %q", tx0.StatementAccount)
	}
	if tx0.StatementNumber != "00001/001" {
		t.Errorf("txs[0].StatementNumber = %q", tx0.StatementNumber)
	}
	if tx0.Description != "SALARY PAYMENT FOR JANUARY" {
		t.Errorf("txs[0].Description = %q", tx0.Description)
	}
	if tx0.Type != "PMNT" {
		t.Errorf("txs[0].Type = %q, want PMNT", tx0.Type)
	}
	if tx0.CounterpartyBankCode != "ABCDPLPW" {
		t.Errorf("txs[0].CounterpartyBankCode = %q", tx0.CounterpartyBankCode)
	}
	if tx0.CounterpartyAccount != "PL27114020040000300201355387" {
		t.Errorf("txs[0].CounterpartyAccount = %q", tx0.CounterpartyAccount)
	}
	if tx0.CounterpartyName != "JOHN DOE" {
		t.Errorf("txs[0].CounterpartyName = %q", tx0.CounterpartyName)
	}
	if tx0.RawData["dcMark"] != "D" {
		t.Errorf("txs[0].RawData[dcMark] = %q, want D", tx0.RawData["dcMark"])
	}
	if tx0.RawData["bankRef"] != "BANKREF01" {
		t.Errorf("txs[0].RawData[bankRef] = %q", tx0.RawData["bankRef"])
	}

	tx1 := txs[1]
	if got, want := tx1.Amount.String(), "1500.50"; got != want {
		t.Errorf("txs[1].Amount = %s, want %s", got, want)
	}
	if tx1.Description != "free text without structured subfields" {
		t.Errorf("txs[1].Description = %q", tx1.Description)
	}

	// RD (reversal of debit) functions as a credit: positive amount.
	tx2 := txs[2]
	if got, want := tx2.Amount.String(), "75.25"; got != want {
		t.Errorf("txs[2].Amount = %s, want %s (RD should be positive)", got, want)
	}
	if tx2.RawData["dcMark"] != "RD" {
		t.Errorf("txs[2].RawData[dcMark] = %q, want RD", tx2.RawData["dcMark"])
	}

	// RC (reversal of credit) functions as a debit: negative amount. Also
	// verifies statement context resets across the "-" message boundary and
	// the second message's own :25:/:28C:/:60M: are picked up.
	tx3 := txs[3]
	if got, want := tx3.Amount.String(), "-2175.25"; got != want {
		t.Errorf("txs[3].Amount = %s, want %s (RC should be negative)", got, want)
	}
	if tx3.StatementNumber != "00002/001" {
		t.Errorf("txs[3].StatementNumber = %q, want 00002/001", tx3.StatementNumber)
	}
	if tx3.Currency != "EUR" {
		t.Errorf("txs[3].Currency = %q, want EUR", tx3.Currency)
	}
}

func TestEntryDateYearRollover(t *testing.T) {
	// Value date late December, entry date in January -> entry date rolls
	// into the following year.
	const data = ":20:R\r\n:25:ACC\r\n:28C:1\r\n:60F:C251230EUR0,00\r\n" +
		":61:2512310102D10,00NMSCNONREF\r\n:86:x\r\n:62F:D260102EUR10,00\r\n-\r\n"
	p := New(DefaultDialect{})
	var txs []*bankparse.Transaction
	for tx, err := range p.Parse(t.Context(), strings.NewReader(data)) {
		if err != nil {
			t.Fatalf("parse: %v", err)
		}
		txs = append(txs, tx)
	}
	if len(txs) != 1 {
		t.Fatalf("got %d transactions, want 1", len(txs))
	}
	if want := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC); !txs[0].TransactionDate.Equal(want) {
		t.Errorf("TransactionDate = %v, want %v", txs[0].TransactionDate, want)
	}
	if want := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC); !txs[0].BookingDate.Equal(want) {
		t.Errorf("BookingDate = %v, want %v", txs[0].BookingDate, want)
	}
}

func TestMalformedTag61SurfacesError(t *testing.T) {
	const data = ":20:R\r\n:25:ACC\r\n:28C:1\r\n:61:not-a-valid-line\r\n-\r\n"
	p := New(DefaultDialect{})
	var errs int
	for tx, err := range p.Parse(t.Context(), strings.NewReader(data)) {
		if err != nil {
			errs++
			continue
		}
		t.Fatalf("unexpected transaction: %+v", tx)
	}
	if errs != 1 {
		t.Fatalf("got %d errors, want 1", errs)
	}
}
