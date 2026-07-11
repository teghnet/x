package mbankcsv

import (
	"strings"
	"testing"
	"time"

	"github.com/teghnet/x/banking"
)

// fixture is a synthetic mBank "Lista operacji" export, trimmed and with
// fake account numbers/names, but structurally faithful to a real export:
// a ragged preamble (letterhead, client name, selected accounts, per-currency
// summaries) that must be ignored, followed by the fixed transaction table
// header and rows in the base currency (PLN) and a foreign currency (USD),
// plus trailing blank lines.
const fixture = "mBank S.A. Bankowość Detaliczna;\n" +
	"\t\tSkrytka Pocztowa 2108;\n" +
	"\n" +
	"#Klient;\n" +
	"TEST SP. Z O.O.;\n" +
	"\n" +
	"Lista operacji;\n" +
	"\n" +
	"#Za okres:;\n" +
	"01.01.2026;31.01.2026;\n" +
	"\n" +
	"      #Waluta;#Wpływy;#Wydatki;\n" +
	"PLN;1 000,00;-500,00;\n" +
	"\n" +
	"#Data operacji;#Opis operacji;#Rachunek;#Kategoria;#Kwota;#Saldo po operacji;\n" +
	"2026-01-10;\"Test Merchant  \";\"T-PLN 6711 ... 2246\";\"Wpływy - inne\";85,68 PLN;13 407,53 PLN;\n" +
	"2026-01-08;\"OBSIDIAN  ZAKUP PRZY UŻYCIU KARTY\";\"T-USD 4011 ... 3525\";\"Bez kategorii\";-5,00 USD;78 463,57 USD;\n" +
	"2026-01-01;\"BIG INVOICE PAYMENT\";\"T-PLN 6711 ... 2246\";\"Materiały i usługi - inne\";-1 883,13 PLN;13 923,67 PLN;\n" +
	"\n" +
	"\n"

func collect(t *testing.T) []*banking.Transaction {
	t.Helper()
	p := New()
	var txs []*banking.Transaction
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
	if len(txs) != 3 {
		t.Fatalf("got %d transactions, want 3 (preamble and blank lines must be ignored)", len(txs))
	}

	tx0 := txs[0]
	if got, want := tx0.Amount.String(), "85.68"; got != want {
		t.Errorf("txs[0].Amount = %s, want %s", got, want)
	}
	if tx0.Currency != "PLN" {
		t.Errorf("txs[0].Currency = %q, want PLN", tx0.Currency)
	}
	if !tx0.TransactionDate.Equal(time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("txs[0].TransactionDate = %v", tx0.TransactionDate)
	}
	if tx0.Description != "Test Merchant" {
		t.Errorf("txs[0].Description = %q", tx0.Description)
	}
	if tx0.Category != "Wpływy - inne" {
		t.Errorf("txs[0].Category = %q", tx0.Category)
	}
	if tx0.StatementAccount != "T-PLN 6711 ... 2246" {
		t.Errorf("txs[0].StatementAccount = %q", tx0.StatementAccount)
	}
	if tx0.RawData["balance"] != "13 407,53 PLN" {
		t.Errorf("txs[0].RawData[balance] = %q", tx0.RawData["balance"])
	}

	usd := txs[1]
	if got, want := usd.Amount.String(), "-5.00"; got != want {
		t.Errorf("txs[1].Amount = %s, want %s", got, want)
	}
	if usd.Currency != "USD" {
		t.Errorf("txs[1].Currency = %q, want USD", usd.Currency)
	}

	thousands := txs[2]
	if got, want := thousands.Amount.String(), "-1883.13"; got != want {
		t.Errorf("txs[2].Amount = %s, want %s (thousands separator must be stripped)", got, want)
	}
}

func TestRegistered(t *testing.T) {
	p, ok := banking.GetParser(ParserName)
	if !ok {
		t.Fatal(`GetParser("` + ParserName + `"): ok = false, want true`)
	}
	if p == nil {
		t.Fatal("GetParser(mbank_lista_operacji_csv) returned a nil parser")
	}
}
