// Package bankparse provides a unified interface for reading bank statements
// in disparate formats (CSV, MT940, ...) as a single, denormalized stream of
// transactions.
//
// Design philosophy:
//
//   - Streaming first: parsers read incrementally via iter.Seq2, so large
//     statements never need to be loaded into memory in full.
//   - Extensible: new bank-specific CSV mappings or MT940 dialects plug in
//     without changing this package.
//   - Denormalized output: Transaction is a flat struct covering every
//     standard field, plus RawData for whatever doesn't fit.
//   - Precision: monetary amounts use decimal.Decimal, never float64.
package bankparse

import (
	"time"

	"github.com/govalues/decimal"
)

// Transaction represents a single, denormalized bank account ledger entry.
type Transaction struct {
	// -- Identifiers --
	// Unique identifier for the transaction (if provided by the bank)
	TransactionID string
	// Reference provided by the counterparty or the payment system
	Reference string

	// -- Dates --
	// When the transaction was executed
	TransactionDate time.Time
	// When the funds actually settled/cleared
	BookingDate time.Time

	// -- Financials --
	// The monetary value. Positive for credits (income), negative for debits (spend)
	Amount decimal.Decimal
	// ISO 4217 Currency Code (e.g., "USD", "EUR", "PLN")
	Currency string

	// -- Counterparty Info --
	// Name of the sender or receiver
	CounterpartyName string
	// IBAN or local account number of the counterparty
	CounterpartyAccount string
	// SWIFT/BIC of the counterparty's bank
	CounterpartyBankCode string
	// Counterparty address, if provided
	CounterpartyAddress string

	// -- Context & Details --
	// The transaction description or remittance information (e.g., MT940 Tag 86)
	Description string
	// Bank-specific transaction type/code (e.g., "NMSC", "TRF")
	Type string
	// Category of the transaction (e.g., "FEE", "TRANSFER", "CARD_PAYMENT")
	Category string

	// -- Statement Context --
	// The account this transaction belongs to
	StatementAccount string
	// Sequence or statement number (useful for MT940)
	StatementNumber string

	// -- Raw/Unmapped Data --
	// Holds all original fields that don't fit standard categories (e.g., raw CSV row, specific MT940 sub-tags)
	RawData map[string]string
}
