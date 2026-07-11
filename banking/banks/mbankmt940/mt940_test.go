package mbankmt940

import (
	"strings"
	"testing"

	"github.com/teghnet/x/banking"
)

// fixture is a synthetic mBank MT940 export, trimmed and with fake account
// numbers/names, but structurally faithful to a real export: Tag 86 rows
// built from a leading 3-digit code plus mBank's free-text "LABEL: value"
// segments, covering an incoming domestic transfer, an outgoing transfer, an
// FX exchange, a card fee, and a tax-office payment with the tax subfields.
const fixture = "" +
	":20:ST260101CYC/1\r\n" +
	":25:PL13114020620000767335001001\r\n" +
	":28C:1/1\r\n" +
	":60F:C260101PLN0,00\r\n" +
	":61:2601010101CN8281,00NTRFNONREF//EL015968-019543\r\n" +
	"770-PRZELEW PRZYCHODZACY - ELIXIR\r\n" +
	":86:770 PRZELEW URZAD SKARBOWY; Z RACH.: 69101000712222525300066500; OD: \r\n" +
	"PIERWSZY URZAD SKARBOWY WARSZAWA; TYT.: ZWROT Z PODATKU VAT 6/2025; \r\n" +
	"TNR: 210860197059371.390001\r\n" +
	":61:2601020102DN7500,00NTRFNONREF//BR25245243002130\r\n" +
	"944-PRZEL.KRAJ.WYCH.MT.ELX\r\n" +
	":86:944 COMPANYNET PRZELEW KRAJOWY; NA RACH.: \r\n" +
	"79105010251000009213587026; DLA: POBUDKA SP. Z O.O; TYT.: \r\n" +
	"FP-OFFICE/2025/08/2; TNR: 210654014653068.000001\r\n" +
	":61:2601030103CN145680,00NFEXNONE//FX2525400124\r\n" +
	"306-TRANSAKCJA WYMIANY WALUT\r\n" +
	":86:306 TR. WYM. WALUT; NR: 12773712; KURS: 3.642; ; TNR: \r\n" +
	"210748035430317.020001\r\n" +
	":61:2601040104DN5,00NTRF5174992036794540//CC51749920367945\r\n" +
	"697-OPLATY ZA PROWADZENIE KARTY\r\n" +
	":86:697 200 MIESIECZNE UZYTKOWANIE KARTY; NR KARTY: \r\n" +
	"5174992036794540; TNR: 210654013506503.000001\r\n" +
	":61:2601050105DN69788,00NTRFNONREF//BR25267243002220\r\n" +
	"946-PRZEL.KRAJ.WYCH.MT.ELX.US\r\n" +
	":86:946 COMPANYNET PRZELEW ORGAN PODATKOWY; NA RACH.: \r\n" +
	"69101000712222525300066500; DLA: URZAD SKARBOWY CENTRUM \r\n" +
	"ROZLICZENIOWE; NAZWA ZOBOW.: ETH POLAND SP Z O.O.; ID UZUP.: N \r\n" +
	"5253000665; SYMBOL FORM.: VAT7; OKRES: 25M08; ID ZOBOW.: VAT; TNR: \r\n" +
	"210870243968165.000001\r\n" +
	":62F:D260105PLN73011,00\r\n" +
	":64:D260105PLN73011,00\r\n" +
	"-\r\n"

func collect(t *testing.T) []*banking.Transaction {
	t.Helper()
	p := New(banking.WithEncoding("utf-8"))
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
	if len(txs) != 5 {
		t.Fatalf("got %d transactions, want 5", len(txs))
	}

	incoming := txs[0]
	if incoming.Type != "770" {
		t.Errorf("txs[0].Type = %q, want 770", incoming.Type)
	}
	if incoming.CounterpartyAccount != "69101000712222525300066500" {
		t.Errorf("txs[0].CounterpartyAccount = %q", incoming.CounterpartyAccount)
	}
	if incoming.CounterpartyName != "PIERWSZY URZAD SKARBOWY WARSZAWA" {
		t.Errorf("txs[0].CounterpartyName = %q", incoming.CounterpartyName)
	}
	if want := "PRZELEW URZAD SKARBOWY; ZWROT Z PODATKU VAT 6/2025"; incoming.Description != want {
		t.Errorf("txs[0].Description = %q, want %q", incoming.Description, want)
	}
	if incoming.Reference != "210860197059371.390001" {
		t.Errorf("txs[0].Reference = %q, want the TNR value (NONREF should be replaced)", incoming.Reference)
	}
	if incoming.RawData["tnr"] != "210860197059371.390001" {
		t.Errorf("txs[0].RawData[tnr] = %q", incoming.RawData["tnr"])
	}

	outgoing := txs[1]
	if outgoing.Type != "944" {
		t.Errorf("txs[1].Type = %q, want 944", outgoing.Type)
	}
	if outgoing.CounterpartyAccount != "79105010251000009213587026" {
		t.Errorf("txs[1].CounterpartyAccount = %q", outgoing.CounterpartyAccount)
	}
	if outgoing.CounterpartyName != "POBUDKA SP. Z O.O" {
		t.Errorf("txs[1].CounterpartyName = %q", outgoing.CounterpartyName)
	}
	if want := "COMPANYNET PRZELEW KRAJOWY; FP-OFFICE/2025/08/2"; outgoing.Description != want {
		t.Errorf("txs[1].Description = %q, want %q", outgoing.Description, want)
	}
	if got, want := outgoing.Amount.String(), "-7500.00"; got != want {
		t.Errorf("txs[1].Amount = %s, want %s", got, want)
	}

	fx := txs[2]
	if fx.Type != "306" {
		t.Errorf("txs[2].Type = %q, want 306", fx.Type)
	}
	if fx.RawData["dealNumber"] != "12773712" {
		t.Errorf("txs[2].RawData[dealNumber] = %q", fx.RawData["dealNumber"])
	}
	if fx.RawData["exchangeRate"] != "3.642" {
		t.Errorf("txs[2].RawData[exchangeRate] = %q", fx.RawData["exchangeRate"])
	}
	if fx.CounterpartyAccount != "" {
		t.Errorf("txs[2].CounterpartyAccount = %q, want empty (FX rows have no counterparty account)", fx.CounterpartyAccount)
	}

	cardFee := txs[3]
	if cardFee.Type != "697" {
		t.Errorf("txs[3].Type = %q, want 697", cardFee.Type)
	}
	if cardFee.RawData["cardNumber"] != "5174992036794540" {
		t.Errorf("txs[3].RawData[cardNumber] = %q", cardFee.RawData["cardNumber"])
	}

	tax := txs[4]
	if tax.Type != "946" {
		t.Errorf("txs[4].Type = %q, want 946", tax.Type)
	}
	if tax.RawData["taxpayerId"] != "N 5253000665" {
		t.Errorf("txs[4].RawData[taxpayerId] = %q", tax.RawData["taxpayerId"])
	}
	if tax.RawData["taxForm"] != "VAT7" {
		t.Errorf("txs[4].RawData[taxForm] = %q", tax.RawData["taxForm"])
	}
	if tax.RawData["taxPeriod"] != "25M08" {
		t.Errorf("txs[4].RawData[taxPeriod] = %q", tax.RawData["taxPeriod"])
	}
	if tax.RawData["taxLiabilityId"] != "VAT" {
		t.Errorf("txs[4].RawData[taxLiabilityId] = %q", tax.RawData["taxLiabilityId"])
	}
	if want := "COMPANYNET PRZELEW ORGAN PODATKOWY; ETH POLAND SP Z O.O."; tax.Description != want {
		t.Errorf("txs[4].Description = %q, want %q", tax.Description, want)
	}
}

func TestDefaultsToISO88592(t *testing.T) {
	// 0xA1 is "Ą" in ISO-8859-2 but "ˇ" (caron) in windows-1250 — one of the
	// handful of code points where the two encodings diverge, which is
	// exactly where mBank's real exports were found to use ISO-8859-2, not
	// windows-1250. Verifies New decodes correctly without an explicit
	// encoding override.
	data := ":20:R\r\n:25:ACC\r\n:28C:1\r\n:60F:C260101PLN0,00\r\n" +
		":61:2601010101DN45,00NCHGNONREF//CH1\r\n" +
		":86:341 OGRANICZON\xa1 ODPOWIEDZIALNO\xa6CI\xa1; TNR: 1\r\n" +
		":62F:D260101PLN45,00\r\n-\r\n"

	p := New()
	var txs []*banking.Transaction
	for tx, err := range p.Parse(t.Context(), strings.NewReader(data)) {
		if err != nil {
			t.Fatalf("parse: %v", err)
		}
		txs = append(txs, tx)
	}
	if len(txs) != 1 {
		t.Fatalf("got %d transactions, want 1", len(txs))
	}
	if want := "OGRANICZONĄ ODPOWIEDZIALNOŚCIĄ"; txs[0].Description != want {
		t.Errorf("Description = %q, want %q (iso-8859-2 default not applied)", txs[0].Description, want)
	}
}

func TestRegistered(t *testing.T) {
	p, ok := banking.GetParser(ParserName)
	if !ok {
		t.Fatal(`GetParser("` + ParserName + `"): ok = false, want true`)
	}
	if p == nil {
		t.Fatal("GetParser(mbank_mt940) returned a nil parser")
	}
}
