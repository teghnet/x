// mBank's MT940 export uses the standard SWIFT grammar for Fields 20, 25,
// 28C, 60F/60M, 61, 62F/62M, 64 (handled generically by package mt940), but
// its Tag 86 doesn't follow the "?NN" structured-subfield convention that
// mt940.DefaultDialect understands. Instead it's a free-text line built from
// a leading 3-digit operation code and description, followed by
// semicolon-separated "LABEL: value" pairs whose label vocabulary
// (Z RACH./NA RACH./OD/DLA/TYT./TNR/KURS/...) is mBank-specific — hence this
// dedicated dialect. The export is ISO-8859-2 encoded (not windows-1250: the
// two code pages agree almost everywhere, but diverge for a handful of
// Polish letters including Ą and Ś, which is exactly where a windows-1250
// guess falls apart).
package mbank

import (
	"regexp"
	"strings"

	"github.com/teghnet/x/banking"
	"github.com/teghnet/x/banking/mt940"
)

const MT940ParserName = "mbank_mt940"

func init() {
	banking.Register(MT940ParserName, func() banking.Parser { return NewMT940() })
}

// NewMT940 creates a banking.Parser for mBank's MT940 statement export.
// Encoding defaults to iso-8859-2, as used by mBank's exports; pass
// banking.WithEncoding to override.
func NewMT940(opts ...banking.Option) banking.Parser {
	cfg := append([]banking.Option{banking.WithEncoding("iso-8859-2")}, opts...)
	return mt940.New(mt940Dialect{}, cfg...)
}

// tag86CodeRe matches the leading 3-digit operation code and description at
// the start of a Tag 86 value, e.g. "716 PRZEKSIĘGOWANIE" or
// "944 COMPANYNET PRZELEW KRAJOWY".
var tag86CodeRe = regexp.MustCompile(`^(\d{3})\s+(.*)$`)

// mt940Dialect implements mt940.Dialect for mBank's Tag 86 layout.
type mt940Dialect struct{}

// ParseTag86 implements mt940.Dialect.
func (mt940Dialect) ParseTag86(raw string, tx *banking.Transaction) error {
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
		case "TYT.", "NAZWA ZOBOW.":
			if value != "" {
				desc = append(desc, value)
			}
		case "Z RACH.", "NA RACH.":
			tx.CounterpartyAccount = value
		case "OD", "DLA":
			tx.CounterpartyName = value
		case "TNR":
			// mBank's own transaction reference is more useful than a bare
			// "NONREF" customer reference carried over from Field 61.
			if tx.Reference == "" || tx.Reference == "NONREF" {
				tx.Reference = value
			}
			tx.RawData["tnr"] = value
		case "KURS":
			tx.RawData["exchangeRate"] = value
		case "NR":
			tx.RawData["dealNumber"] = value
		case "DATA STEMPLA":
			tx.RawData["stampDate"] = value
		case "ID UZUP.":
			tx.RawData["taxpayerId"] = value
		case "SYMBOL FORM.":
			tx.RawData["taxForm"] = value
		case "OKRES":
			tx.RawData["taxPeriod"] = value
		case "ID ZOBOW.":
			tx.RawData["taxLiabilityId"] = value
		case "KARTA NR", "NR KARTY":
			tx.RawData["cardNumber"] = value
		default:
			desc = append(desc, seg)
		}
	}

	if len(desc) > 0 {
		tx.Description = strings.Join(desc, "; ")
	}
	return nil
}
