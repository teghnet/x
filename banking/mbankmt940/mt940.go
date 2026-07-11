// Package mt940 parses mBank's MT940 statement export.
//
// mBank's MT940 export uses the standard SWIFT grammar for Fields 20, 25,
// 28C, 60F/60M, 61, 62F/62M, 64 (handled generically by package
// banking/mt940), but its Tag 86 doesn't follow the "?NN" structured-subfield
// convention that mt940.DefaultDialect understands. Instead it's a free-text
// line built from a leading 3-digit operation code and description, followed
// by semicolon-separated "LABEL: value" pairs whose label vocabulary
// (Z RACH./NA RACH./OD/DLA/TYT./TNR/KURS/...) is mBank-specific — hence this
// dedicated dialect. The export is ISO-8859-2 encoded (not windows-1250: the
// two code pages agree almost everywhere, but diverge for a handful of
// Polish letters including Ą and Ś, which is exactly where a windows-1250
// guess falls apart).
package mbankmt940

import (
	"regexp"
	"strings"

	"github.com/teghnet/x/banking"
	"github.com/teghnet/x/banking/mt940"
)

const ParserName = "mbank_mt940"

func init() {
	banking.Register(ParserName, func() banking.Parser { return New() })
}

// New creates a banking.Parser for mBank's MT940 statement export. Encoding
// defaults to banking.EncodingISO88592, as used by mBank's exports; pass
// banking.WithEncoding to override.
func New(opts ...banking.Option) banking.Parser {
	cfg := append([]banking.Option{banking.WithEncoding(banking.EncodingISO88592)}, opts...)
	return mt940.New(dialect{}, cfg...)
}

// tag86CodeRe matches the leading 3-digit operation code and description at
// the start of a Tag 86 value, e.g. "716 PRZEKSIĘGOWANIE" or
// "944 COMPANYNET PRZELEW KRAJOWY".
var tag86CodeRe = regexp.MustCompile(`^(\d{3})\s+(.*)$`)

// Tag 86 field labels in mBank's free-text "LABEL: value" vocabulary.
const (
	labelTitle              = "TYT."
	labelTaxLiabilityName   = "NAZWA ZOBOW."
	labelSourceAccount      = "Z RACH."
	labelDestinationAccount = "NA RACH."
	labelSender             = "OD"
	labelRecipient          = "DLA"
	labelTransactionRef     = "TNR"
	labelExchangeRate       = "KURS"
	labelDealNumber         = "NR"
	labelStampDate          = "DATA STEMPLA"
	labelTaxpayerID         = "ID UZUP."
	labelTaxForm            = "SYMBOL FORM."
	labelTaxPeriod          = "OKRES"
	labelTaxLiabilityID     = "ID ZOBOW."
	labelCardNumber1        = "KARTA NR"
	labelCardNumber2        = "NR KARTY"
)

// noCustomerRef is the placeholder Field 61 customer reference ("NONREF")
// that TNR should override, since mBank's own transaction reference is more
// useful than a bare "no reference" marker.
const noCustomerRef = "NONREF"

// Keys under which ParseTag86 stores label values that don't map to a
// standard Transaction field.
const (
	rawKeyTNR            = "tnr"
	rawKeyExchangeRate   = "exchangeRate"
	rawKeyDealNumber     = "dealNumber"
	rawKeyStampDate      = "stampDate"
	rawKeyTaxpayerID     = "taxpayerId"
	rawKeyTaxForm        = "taxForm"
	rawKeyTaxPeriod      = "taxPeriod"
	rawKeyTaxLiabilityID = "taxLiabilityId"
	rawKeyCardNumber     = "cardNumber"
)

// dialect implements mt940.Dialect for mBank's Tag 86 layout.
type dialect struct{}

// ParseTag86 implements mt940.Dialect.
func (dialect) ParseTag86(raw string, tx *banking.Transaction) error {
	reflowed := strings.Join(strings.Fields(strings.ReplaceAll(raw, "\n", " ")), " ")
	segments := strings.Split(reflowed, ";")

	var desc []string
	for i, seg := range segments {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}

		if i == 0 {
			if m := tag86CodeRe.FindStringSubmatch(seg); m != nil {
				tx.Type = m[1]
				if m[2] != "" {
					desc = append(desc, m[2])
				}
				continue
			}
			desc = append(desc, seg)
			continue
		}

		key, value, ok := strings.Cut(seg, ":")
		if !ok {
			desc = append(desc, seg)
			continue
		}
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)

		switch key {
		case labelTitle, labelTaxLiabilityName:
			if value != "" {
				desc = append(desc, value)
			}
		case labelSourceAccount, labelDestinationAccount:
			tx.CounterpartyAccount = value
		case labelSender, labelRecipient:
			tx.CounterpartyName = value
		case labelTransactionRef:
			if tx.Reference == "" || tx.Reference == noCustomerRef {
				tx.Reference = value
			}
			tx.RawData[rawKeyTNR] = value
		case labelExchangeRate:
			tx.RawData[rawKeyExchangeRate] = value
		case labelDealNumber:
			tx.RawData[rawKeyDealNumber] = value
		case labelStampDate:
			tx.RawData[rawKeyStampDate] = value
		case labelTaxpayerID:
			tx.RawData[rawKeyTaxpayerID] = value
		case labelTaxForm:
			tx.RawData[rawKeyTaxForm] = value
		case labelTaxPeriod:
			tx.RawData[rawKeyTaxPeriod] = value
		case labelTaxLiabilityID:
			tx.RawData[rawKeyTaxLiabilityID] = value
		case labelCardNumber1, labelCardNumber2:
			tx.RawData[rawKeyCardNumber] = value
		default:
			desc = append(desc, seg)
		}
	}

	if len(desc) > 0 {
		tx.Description = strings.Join(desc, "; ")
	}
	return nil
}
