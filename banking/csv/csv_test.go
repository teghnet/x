package csv

import (
	"strings"
	"testing"
	"time"

	"github.com/teghnet/x/banking"
)

func collect(t *testing.T, p banking.Parser, data string) []*banking.Transaction {
	t.Helper()
	var txs []*banking.Transaction
	for tx, err := range p.Parse(t.Context(), strings.NewReader(data)) {
		if err != nil {
			t.Fatalf("parse: %v", err)
		}
		txs = append(txs, tx)
	}
	return txs
}

func TestHappyPath(t *testing.T) {
	const data = "2026-01-05,Acme Corp,PL61109010140000071219812874,100.50,Invoice 123\n" +
		"2026-01-06,Bob,PL27114020040000300201355387,-42.00,Coffee\n"

	mapper := ColumnMapper{
		TransactionDateIdx:  0,
		CounterpartyNameIdx: new(1),
		CounterpartyAcctIdx: new(2),
		AmountIdx:           3,
		DescriptionIndices:  []int{4},
	}
	p := New(mapper)
	txs := collect(t, p, data)

	if len(txs) != 2 {
		t.Fatalf("got %d transactions, want 2", len(txs))
	}
	if !txs[0].TransactionDate.Equal(time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("txs[0].TransactionDate = %v", txs[0].TransactionDate)
	}
	if got, want := txs[0].Amount.String(), "100.50"; got != want {
		t.Errorf("txs[0].Amount = %s, want %s", got, want)
	}
	if txs[0].CounterpartyName != "Acme Corp" {
		t.Errorf("txs[0].CounterpartyName = %q", txs[0].CounterpartyName)
	}
	if txs[1].Amount.IsNeg() != true {
		t.Errorf("txs[1].Amount should be negative, got %s", txs[1].Amount.String())
	}
	if txs[1].Description != "Coffee" {
		t.Errorf("txs[1].Description = %q", txs[1].Description)
	}
	if txs[0].BookingDate != txs[0].TransactionDate {
		t.Errorf("BookingDate should default to TransactionDate when unset")
	}
}

func TestDelimiterAndHeaderSkip(t *testing.T) {
	const data = "date;amount\n2026-02-01;10.00\n"
	mapper := ColumnMapper{TransactionDateIdx: 0, AmountIdx: 1}
	p := New(mapper, banking.WithDelimiter(';'), banking.WithSkipHeaderLines(1))
	txs := collect(t, p, data)
	if len(txs) != 1 {
		t.Fatalf("got %d transactions, want 1", len(txs))
	}
	if got, want := txs[0].Amount.String(), "10.00"; got != want {
		t.Errorf("Amount = %s, want %s", got, want)
	}
}

func TestEncoding(t *testing.T) {
	// "z\xb9" is "zą" in windows-1250.
	data := "2026-03-01,10.00,kwota w z\xb9\n"
	mapper := ColumnMapper{TransactionDateIdx: 0, AmountIdx: 1, DescriptionIndices: []int{2}}
	p := New(mapper, banking.WithEncoding("windows-1250"))
	txs := collect(t, p, data)
	if len(txs) != 1 {
		t.Fatalf("got %d transactions, want 1", len(txs))
	}
	if want := "kwota w zą"; txs[0].Description != want {
		t.Errorf("Description = %q, want %q", txs[0].Description, want)
	}
}

func TestCustomFn(t *testing.T) {
	const data = "2026-04-01,10.00,FEE\n"
	mapper := ColumnMapper{
		TransactionDateIdx: 0,
		AmountIdx:          1,
		CustomFn: func(record []string, tx *banking.Transaction) error {
			tx.Category = record[2]
			return nil
		},
	}
	p := New(mapper)
	txs := collect(t, p, data)
	if len(txs) != 1 {
		t.Fatalf("got %d transactions, want 1", len(txs))
	}
	if txs[0].Category != "FEE" {
		t.Errorf("Category = %q, want FEE", txs[0].Category)
	}
}

func TestMalformedRowSurfacesError(t *testing.T) {
	const data = "not-a-date,10.00\n"
	mapper := ColumnMapper{TransactionDateIdx: 0, AmountIdx: 1}
	p := New(mapper)

	var errs int
	for tx, err := range p.Parse(t.Context(), strings.NewReader(data)) {
		if err != nil {
			errs++
			continue
		}
		if tx != nil {
			t.Fatalf("unexpected transaction: %+v", tx)
		}
	}
	if errs != 1 {
		t.Fatalf("got %d errors, want 1", errs)
	}
}

func TestNewPanicsOnMissingRequiredColumns(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("want panic when AmountIdx is unset")
		}
	}()
	New(ColumnMapper{TransactionDateIdx: 0, AmountIdx: -1})
}
