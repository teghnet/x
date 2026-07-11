package revolut

import (
	"strings"
	"testing"
	"time"

	"github.com/teghnet/x/banking"
)

// fixture is a synthetic Revolut Personal Finance export, trimmed and with
// fake account numbers/names, but structurally faithful to a real export:
// account summaries and crypto sections that must be ignored, a base
// currency (PLN) with single-value amounts, a foreign currency (USD) with
// symbol-prefixed dual amounts, and a third currency (AED) with
// code-suffixed dual amounts.
const fixture = `Current Accounts Summaries,,,,,,,
,,,,,,,
Personal Account (PLN),,,,,,,
,,,,,,,
Current account details,,,,,,,
Account Number (LT IBAN),LT000000000000000000000000,Opening date,"Jan 1, 2020",,,,
,,,,,,,
Deposit value,,,,,,,
,,Opening balance,0.00 PLN,,,,
,,Closing balance,"1,062.72 PLN",,,,
,,,,,,,
---------,,,,,,,
,,,,,,,
Personal Account (AED),,,,,,,
,,,,,,,
Deposit value,,,,,,,
,,Opening balance,0.00 AED (0.00 PLN),,,,
,,,,,,,
---------,,,,,,,
,,,,,,,
Crypto Summaries,,,,,,,
,,,,,,,
Capital gains from Sale,,,,,,,
Gross sales value,13.82 PLN,Tax withheld on sales,0.00 PLN,,,,
,,,,,,,
---------,,,,,,,
,,,,,,,
Current Accounts Transaction Statements,,,,,,,
,,,,,,,
Personal Account (PLN),,,,,,,
,,,,,,,
Transaction statement,,,,,,,
Date,Description,Category,Money in/out,Balance,Tax withheld,Other taxes,Fees
"Jan 2, 2020",Test Merchant,Top up,20.00 PLN,20.00 PLN,0.00 PLN,0.00 PLN,0.00 PLN
"Jan 3, 2020",Some Shop,Merchant,-12.00 PLN,8.00 PLN,0.00 PLN,0.00 PLN,0.00 PLN
Total,,,8.00 PLN,,0.00 PLN,0.00 PLN,0.00 PLN
,,,,,,,
---------,,,,,,,
,,,,,,,
Personal Account (USD),,,,,,,
,,,,,,,
Transaction statement,,,,,,,
Date,Description,Category,Money in/out,Balance,Tax withheld,Other taxes,Fees
"Jan 4, 2020",Foreign Merchant,Merchant,-$14.70 (-56.89 PLN),$23.30 (90.17 PLN),$0.00 (0.00 PLN),$0.00 (0.00 PLN),$0.00 (0.00 PLN)
Total,,,-$14.70 (-56.89 PLN),,$0.00 (0.00 PLN),$0.00 (0.00 PLN),$0.00 (0.00 PLN)
,,,,,,,
---------,,,,,,,
,,,,,,,
Personal Account (AED),,,,,,,
,,,,,,,
Transaction statement,,,,,,,
Date,Description,Category,Money in/out,Balance,Tax withheld,Other taxes,Fees
"Jan 5, 2020",AED Merchant,Merchant,"1,098.67 AED (1,170.41 PLN)","1,098.67 AED (1,170.41 PLN)",0.00 AED (0.00 PLN),0.00 AED (0.00 PLN),0.00 AED (0.00 PLN)
Total,,,"1,098.67 AED (1,170.41 PLN)",,0.00 AED (0.00 PLN),0.00 AED (0.00 PLN),0.00 AED (0.00 PLN)
,,,,,,,
---------,,,,,,,
,,,,,,,
Crypto Transaction Statements,,,,,,,
,,,,,,,
Transaction statement (only sales),,,,,,,
"Date (of Sale, of Purchase)",Description & symbol,Age of units,Units sold,"Unit price (on Sale date, on Purchase date)","Value (of Sale, of Purchase)",Capital gains,Fees
"01.01.20, 01.01.20",BTC,0 days,0.001,"+ 1,000.00 PLN, - 900.00 PLN","+ 10.00 PLN, - 9.00 PLN",1.00 PLN,0.00 PLN
Total,,,,,"+ 10.00 PLN, - 9.00 PLN",1.00 PLN,0.00 PLN
,,,,,,,
---------,,,,,,,
`

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
	if len(txs) != 4 {
		t.Fatalf("got %d transactions, want 4 (crypto and summary sections must be ignored)", len(txs))
	}

	pln1 := txs[0]
	if got, want := pln1.Amount.String(), "20.00"; got != want {
		t.Errorf("txs[0].Amount = %s, want %s", got, want)
	}
	if pln1.Currency != "PLN" {
		t.Errorf("txs[0].Currency = %q, want PLN", pln1.Currency)
	}
	if !pln1.TransactionDate.Equal(time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("txs[0].TransactionDate = %v", pln1.TransactionDate)
	}
	if pln1.Description != "Test Merchant" {
		t.Errorf("txs[0].Description = %q", pln1.Description)
	}
	if _, hasBase := pln1.RawData[rawKeyBaseAmount]; hasBase {
		t.Errorf("txs[0].RawData should have no baseAmount for the base-currency account")
	}

	pln2 := txs[1]
	if got, want := pln2.Amount.String(), "-12.00"; got != want {
		t.Errorf("txs[1].Amount = %s, want %s", got, want)
	}

	usd := txs[2]
	if got, want := usd.Amount.String(), "-14.70"; got != want {
		t.Errorf("txs[2].Amount = %s, want %s", got, want)
	}
	if usd.Currency != "USD" {
		t.Errorf("txs[2].Currency = %q, want USD", usd.Currency)
	}
	if got, want := usd.RawData[rawKeyBaseAmount], "-56.89"; got != want {
		t.Errorf("txs[2].RawData[baseAmount] = %q, want %q", got, want)
	}
	if usd.RawData[rawKeyBaseCurrency] != "PLN" {
		t.Errorf("txs[2].RawData[baseCurrency] = %q, want PLN", usd.RawData[rawKeyBaseCurrency])
	}
	if usd.StatementAccount != "Personal Account (USD)" {
		t.Errorf("txs[2].StatementAccount = %q", usd.StatementAccount)
	}

	aed := txs[3]
	if got, want := aed.Amount.String(), "1098.67"; got != want {
		t.Errorf("txs[3].Amount = %s, want %s", got, want)
	}
	if aed.Currency != "AED" {
		t.Errorf("txs[3].Currency = %q, want AED", aed.Currency)
	}
	if got, want := aed.RawData[rawKeyBaseAmount], "1170.41"; got != want {
		t.Errorf("txs[3].RawData[baseAmount] = %q, want %q", got, want)
	}
}

func TestRegistered(t *testing.T) {
	p, ok := banking.GetParser(ParserName)
	if !ok {
		t.Fatal(`GetParser("` + ParserName + `"): ok = false, want true`)
	}
	if p == nil {
		t.Fatal("GetParser(revolut_csv) returned a nil parser")
	}
}
