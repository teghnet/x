package revolut

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/govalues/decimal"
)

// currencySymbols maps the currency symbols Revolut prefixes amounts with to
// their ISO 4217 codes. Currencies without a widely recognized symbol are
// instead suffixed as a 3-letter code (e.g. "1,098.67 AED"), handled by the
// non-symbol branch of parseMoneyPart.
var currencySymbols = map[rune]string{
	'$': "USD",
	'€': "EUR",
	'£': "GBP",
}

// parseMoneyPart parses a single signed monetary value in either of
// Revolut's two forms: symbol-prefixed ("$20.00", "-€7.24") or
// code-suffixed ("20.00 PLN", "-1,098.67 AED"). Thousands separators (",")
// are stripped before decimal parsing.
func parseMoneyPart(s string) (decimal.Decimal, string, error) {
	s = strings.TrimSpace(s)
	neg := strings.HasPrefix(s, "-")
	if neg {
		s = s[1:]
	}
	if s == "" {
		return decimal.Decimal{}, "", fmt.Errorf("empty amount")
	}

	var amountStr, currency string
	if r, size := utf8.DecodeRuneInString(s); currencySymbols[r] != "" {
		currency = currencySymbols[r]
		amountStr = s[size:]
	} else {
		var ok bool
		amountStr, currency, ok = strings.Cut(s, " ")
		if !ok {
			return decimal.Decimal{}, "", fmt.Errorf("malformed amount %q: no currency", s)
		}
	}

	amount, err := decimal.Parse(strings.ReplaceAll(amountStr, ",", ""))
	if err != nil {
		return decimal.Decimal{}, "", fmt.Errorf("parse amount %q: %w", amountStr, err)
	}
	if neg {
		amount = amount.Neg()
	}
	return amount, currency, nil
}

// money is a fully decoded Revolut money cell: the account-native amount and
// currency, plus an optional base-currency (PLN) equivalent shown in
// parentheses for non-PLN accounts.
type money struct {
	Amount       decimal.Decimal
	Currency     string
	BaseAmount   decimal.Decimal
	BaseCurrency string
	HasBase      bool
}

// parseMoneyCell parses a Revolut amount cell. For the base-currency account
// it's a single value, e.g. "20.00 PLN". For any other currency it's the
// account-native amount followed by the PLN equivalent in parentheses, e.g.
// "$20.00 (77.97 PLN)" or "1,098.67 AED (1,170.41 PLN)".
func parseMoneyCell(cell string) (money, error) {
	cell = strings.TrimSpace(cell)

	primary, paren := cell, ""
	if idx := strings.Index(cell, " ("); idx >= 0 && strings.HasSuffix(cell, ")") {
		primary, paren = cell[:idx], cell[idx+2:len(cell)-1]
	}

	amount, currency, err := parseMoneyPart(primary)
	if err != nil {
		return money{}, err
	}
	m := money{Amount: amount, Currency: currency}

	if paren != "" {
		baseAmount, baseCurrency, err := parseMoneyPart(paren)
		if err != nil {
			return money{}, err
		}
		m.BaseAmount, m.BaseCurrency, m.HasBase = baseAmount, baseCurrency, true
	}

	return m, nil
}
