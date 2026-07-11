Here is a proposed specification for a Go 1.26+ library designed to parse bank statements in various CSV formats and MT940.

This spec leverages modern Go features (such as `iter.Seq` introduced in 1.23 for efficient stream processing) and focuses on a highly extensible, denormalized data model.

### 1. Package Overview

* **Name Idea:** `bankparse`
* **Purpose:** Provide a unified interface to read disparate bank statement formats and output a standardized, denormalized stream of transactions.
* **Design Philosophy:**
* **Streaming First:** Use Go `iter.Seq2` to process large files without loading the entire statement into memory.
* **Extensible:** Easy to add new bank-specific CSV mappings or MT940 dialect handlers.
* **Denormalized Output:** A single, flat `Transaction` struct containing all possible standard fields, plus a metadata map for format-specific quirks.
* **Precision:** Use exact decimal representation for currency amounts to avoid floating-point errors.



---

### 2. Core Data Model (Denormalized Entry)

The `Transaction` struct is flat and contains all relevant information extracted from a statement.

```go
package bankparse

import (
	"time"
	"github.com/shopspring/decimal" // Industry standard for financial math in Go
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
	BookingDate     time.Time 

	// -- Financials --
	// The monetary value. Positive for credits (income), negative for debits (spend)
	Amount   decimal.Decimal 
	// ISO 4217 Currency Code (e.g., "USD", "EUR", "PLN")
	Currency string          

	// -- Counterparty Info --
	// Name of the sender or receiver
	CounterpartyName          string 
	// IBAN or local account number of the counterparty
	CounterpartyAccount       string 
	// SWIFT/BIC of the counterparty's bank
	CounterpartyBankCode      string 
	// Counterparty address, if provided
	CounterpartyAddress       string 

	// -- Context & Details --
	// The transaction description or remittance information (e.g., MT940 Tag 86)
	Description string 
	// Bank-specific transaction type/code (e.g., "NMSC", "TRF")
	Type        string 
	// Category of the transaction (e.g., "FEE", "TRANSFER", "CARD_PAYMENT")
	Category    string 

	// -- Statement Context --
	// The account this transaction belongs to
	StatementAccount string 
	// Sequence or statement number (useful for MT940)
	StatementNumber  string 

	// -- Raw/Unmapped Data --
	// Holds all original fields that don't fit standard categories (e.g., raw CSV row, specific MT940 sub-tags)
	RawData map[string]string 
}

```

---

### 3. Core Interfaces

Leveraging Go 1.23+ iterators ensures the library is memory-efficient and idiomatic for Go 1.26+.

```go
package bankparse

import (
	"context"
	"io"
	"iter"
)

// Parser defines the contract for all statement format parsers.
type Parser interface {
	// Parse reads from the provided io.Reader and yields a sequence of Transactions.
	// It uses iter.Seq2 to yield either a Transaction or an error, stopping on EOF.
	Parse(ctx context.Context, r io.Reader) iter.Seq2[*Transaction, error]
}

// Configurator allows passing bank-specific options to a parser.
type Option func(*parserOptions)

type parserOptions struct {
	DateFormat      string
	Encoding        string // e.g., "windows-1250", "utf-8"
	SkipHeaderLines int
	Delimiter       rune
}

```

---

### 4. Format-Specific Strategies

#### A. CSV Parsing Strategy

Since every bank's CSV is different, the CSV parser should use a declarative mapping strategy.

```go
package csvparser

import "github.com/teghnet/x/bankparse"

// ColumnMapper defines how a row of CSV translates to a denormalized Transaction.
type ColumnMapper struct {
	TransactionDateIdx    int
	BookingDateIdx        int
	AmountIdx             int
	CurrencyIdx           int
	CounterpartyNameIdx   int
	CounterpartyAcctIdx   int
	DescriptionIndices    []int // Sometimes descriptions are split across columns
	
	// Custom mapping function for complex parsing rules specific to a bank
	CustomFn func(record []string, tx *bankparse.Transaction) error
}

// New creates a new CSV parser with the specific bank's mapping rules.
func New(mapper ColumnMapper, opts ...bankparse.Option) bankparse.Parser {
    // implementation
}

```

#### B. MT940 Parsing Strategy

MT940 relies on strict SWIFT tags (e.g., `:20:`, `:61:`, `:86:`). The complexity here is that different banks format Tag `:86:` (Information to Account Owner) differently.

```go
package mt940parser

import "github.com/teghnet/x/bankparse"

// Dialect allows parsing bank-specific logic inside Tag 86.
type Dialect interface {
	// ParseTag86 extracts counterparty and description from the unstructured Tag 86 text.
	ParseTag86(raw string, tx *bankparse.Transaction) error
}

// DefaultDialect implements the standard SWIFT structure.
type DefaultDialect struct{}

// New creates a new MT940 parser.
func New(dialect Dialect, opts ...bankparse.Option) bankparse.Parser {
	// implementation
}

```

---

### 5. Registry & Usage Implementation

To make the library user-friendly, implement a registry pattern so developers can request a parser by the bank's identifier.

```go
package bankparse

var parsers = make(map[string]func() Parser)

// Register adds a new bank parser to the global registry.
func Register(bankName string, factory func() Parser) {
	parsers[bankName] = factory
}

// GetParser returns a configured parser for the given bank and format.
func GetParser(bankName string) (Parser, bool) {
	factory, exists := parsers[bankName]
	if !exists {
		return nil, false
	}
	return factory(), true
}

```

#### Example Usage

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/teghnet/x/bankparse"
	_ "github.com/teghnet/x/bankparse/banks/ing_mt940" // Registers "ing_mt940"
	_ "github.com/teghnet/x/bankparse/banks/chase_csv" // Registers "chase_csv"
)

func main() {
	file, err := os.Open("statement.sta")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// 1. Get the configured parser for the specific bank/format
	parser, ok := bankparse.GetParser("ing_mt940")
	if !ok {
		panic("parser not found")
	}

	// 2. Iterate over the stream of transactions
	ctx := context.Background()
	for tx, err := range parser.Parse(ctx, file) {
		if err != nil {
			fmt.Printf("Error parsing transaction: %v\n", err)
			continue // Or break, depending on error handling strategy
		}
		
		fmt.Printf("Date: %s | Amount: %s %s | Counterparty: %s\n", 
			tx.TransactionDate.Format("2006-01-02"), 
			tx.Amount.String(), 
			tx.Currency, 
			tx.CounterpartyName,
		)
	}
}

```

### 6. Suggested Dependencies

* `github.com/shopspring/decimal`: For accurate float/currency mathematics.
* `golang.org/x/text/encoding`: Required for handling CSVs generated by banks that use legacy encodings (like ISO-8859-1 or Windows-1250) instead of UTF-8.
* `encoding/csv`: The standard Go library is sufficient for the CSV reader backbone.