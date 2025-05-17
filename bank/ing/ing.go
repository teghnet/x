package ing

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/charmap"
)

type Transaction struct {
	tx0     string
	tx0Rest string
	tx1     string
	tx1Rest string
	tx2     string

	AccountingMonth string
	AccountName     string
	AccountingDate  string
	Amount          string
	AccountCurrency string
	Fee             string

	Title string
	Ref   string

	Name string
	IBAN string
	BIC  string

	Currency       string
	CurrencyAmount string
	CurrencyDate   string
	ExRate         string
	ExID           string
	exRate2        string

	ID string

	Code  string
	Code1 string
	Code2 string
	code3 string

	NRB   string
	NRozl string
	NRach string

	Payer string
	Card  string

	Hash []byte
}

func (t Transaction) hash() []byte {
	h := sha256.Sum256([]byte(strings.Join([]string{
		t.AccountingMonth,
		t.AccountName,
		t.AccountingDate,
		t.Amount,
		t.AccountCurrency,

		t.ID,
		t.Title,
		t.Ref,

		t.Name,
		t.IBAN,
		t.BIC,

		t.Currency,
		t.CurrencyAmount,
		t.CurrencyDate,
		t.ExRate,
		t.ExID,

		t.Amount,
	}, zeroByte)))
	return h[:16]
}

var zeroByte = string([]byte{0})

type MT940 struct {
	AccountIBAN   string
	StatementNo   string
	StatementDate string

	OpeningBalance  string
	OpeningCurrency string
	OpeningDate     string

	ClosingBalance  string
	ClosingCurrency string
	ClosingDate     string

	AvailableBalance  string
	AvailableCurrency string
	AvailableDate     string

	Transactions []Transaction
}

const (
	TagInit        = ":20:"
	TagAccountIBAN = ":25:/"
	TagStatementNo = ":28C:"

	TagOpeningBalance   = ":60F:"
	TagClosingBalance   = ":62F:"
	TagAvailableBalance = ":64:"

	TagTransaction            = ":61:"
	TagTransactionDescription = ":86:"
	TagTransactionExRate      = "KURS"
)

func (m MT940) Name() string {
	return fmt.Sprintf("%s | %s (%s) | %s #%d | %s | %s %s",
		"", "", "", "", mustAtoi(m.StatementNo), m.OpeningCurrency, m.OpeningDate, m.ClosingDate)
}
func mustAtoi(s string) int {
	if s == "" {
		return 0
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Fatal(err)
	}
	return i
}
func (m MT940) String() string {
	return m.Name() + " | " +
		fmt.Sprintf("%d + %d(%d) - %d(%d) [%d(%d)] = %d",
			mustCents(m.OpeningBalance),
			0, 0,
			0, 0,
			0, 0,
			mustCents(m.OpeningBalance))
}
func mustCents(val string) int64 {
	c, err := cents(val)
	if err != nil {
		log.Fatal(err)
	}
	return c
}
func cents(val string) (int64, error) {
	if val == "" {
		return 0, nil
	}
	val = strings.ReplaceAll(val, " ", "")
	coma := ","
	if c := strings.Count(val, coma); c > 1 {
		return 0, fmt.Errorf("too many commas: %d", c)
	}
	if z := len(val) - strings.Index(val, coma) - 1; z != 2 {
		return 0, fmt.Errorf("expected 2 decimal places, got %d in: %s", z, val)
	}
	val = strings.ReplaceAll(val, coma, "")
	return strconv.ParseInt(val, 10, 64)
}

func ReadMT940(filename string) (MT940, error) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var ac MT940
	var tr Transaction

	scanner := bufio.NewScanner(file)
	scanner.Split(scanLines)
	for scanner.Scan() {
		text := scanner.Text()

		switch {
		case strings.HasPrefix(text, TagInit):
			if text[len(TagInit):] != "MT940" {
				return MT940{}, fmt.Errorf("error: %s", text)
			}

		case strings.HasPrefix(text, TagAccountIBAN):
			ac.AccountIBAN = text[len(TagAccountIBAN):]
		case strings.HasPrefix(text, TagStatementNo):
			ac.StatementNo = text[len(TagStatementNo):]

		case strings.HasPrefix(text, TagOpeningBalance):
			ac.OpeningBalance, ac.OpeningCurrency, ac.OpeningDate = parseBalance(text[len(TagOpeningBalance):])
		case strings.HasPrefix(text, TagClosingBalance):
			ac.ClosingBalance, ac.ClosingCurrency, ac.ClosingDate = parseBalance(text[len(TagClosingBalance):])
		case strings.HasPrefix(text, TagAvailableBalance):
			ac.AvailableBalance, ac.AvailableCurrency, ac.AvailableDate = parseBalance(text[len(TagAvailableBalance):])

		case strings.HasPrefix(text, TagTransaction):
			if tr.tx0 != "" {
				ac.Transactions = append(ac.Transactions, parseRaw(tr))
			}
			tr = Transaction{
				tx0:             text[len(TagTransaction):],
				AccountCurrency: ac.OpeningCurrency,
				AccountName:     "ING " + ac.OpeningCurrency,
			}
		case strings.HasPrefix(text, TagTransactionExRate):
			tr.ExRate = strings.TrimSpace(fixRate(text[len(TagTransactionExRate):]))

		case strings.HasPrefix(text, TagTransactionDescription):
			text = text[len(TagTransactionDescription):]
			if tr.tx1 == "" {
				tr.tx1 = text
			} else {
				tr.tx2 += text
			}
		default:
			tr.tx2 += text
		}
	}
	if tr.tx0 != "" {
		ac.Transactions = append(ac.Transactions, parseRaw(tr))
	}
	if err := scanner.Err(); err != nil {
		return MT940{}, err
	}
	return ac, nil
}

func parseBalance(text string) (bal string, cur string, dat string) {
	if text[0] == 'D' {
		bal = "-"
	}
	dat = text[1 : 1+6]
	cur = text[1+6 : 1+6+3]
	bal += text[1+6+3:]
	return bal, cur, dat
}

var (
	reTx0 = regexp.MustCompile(`^([0-9]{6})([0-9]{4})([CD])([0-9]{1,13},[0-9]{1,13})(S[0-9]{3})([0-9]{1,13})(.*)`)
	reTx1 = regexp.MustCompile(`^([0-9]{3})([0-9a-zA-Z_/]{6})?([a-zA-Z]{3})?([0-9]{1,12},[0-9]{1,12})?(.*)`)
)

func fixDate(d string) string {
	d = strings.TrimSpace(d)
	if len(d) == 4 {
		d = "24" + d
	}
	if len(d) == 6 {
		d = "20" + d
	}
	if len(d) == 8 {
		d = d[0:4] + "-" + d[4:6] + "-" + d[6:8]
	}
	return d
}
func fixDatePx(d string, p string) string {
	d = strings.TrimSpace(d)
	if len(d) == 4 {
		d = p + d
	}
	if len(d) == 6 {
		d = "20" + d
	}
	if len(d) == 8 {
		d = d[0:4] + "-" + d[4:6] + "-" + d[6:8]
	}
	return d
}
func fixRate(r string) string {
	if strings.HasPrefix(r, "KURS") {
		r = strings.TrimSpace(r[len("KURS"):])
	}
	return strings.ReplaceAll(r, ",", ".")
}
func parseRaw(tr Transaction) Transaction {
	if m := reTx0.FindStringSubmatch(tr.tx0); m != nil {
		tr.CurrencyDate = fixDate(m[1])
		tr.AccountingDate = fixDatePx(m[2], tr.CurrencyDate[0:4])
		tr.AccountingMonth = tr.AccountingDate[0:7]
		if m[3] == "D" {
			tr.Amount = "-"
		}
		tr.Amount += fixRate(m[4])
		tr.Code = "'" + m[5]
		tr.ID = "'" + m[6]
		tr.tx0Rest = m[7]
	}
	if m := reTx1.FindStringSubmatch(tr.tx1); m != nil {
		tr.Code1 = "'" + m[1]
		tr.ExID = "'" + m[2]
		tr.Currency = m[3]
		tr.CurrencyAmount = fixRate(m[4])
		tr.tx1Rest = m[5]
	}
	if tr.tx2 != "" {
		cc := make(map[string]string)
		for _, s := range strings.Split(tr.tx2, "~") {
			if len(s) < 2 {
				log.Printf("does not look like a '~' tag: %s", s)
				continue
			}
			cc[s[0:2]] += s[2:]
		}
		tr.Code2 = cc["00"]
		tr.Title = strings.TrimSpace(cc["20"] + " " + cc["21"] + cc["22"] + cc["23"] + cc["24"] + cc["25"])
		tr.Ref = "'" + strings.TrimSpace(cc["26"]+cc["27"]+cc["28"])
		tr.NRB = "'" + strings.TrimSpace(cc["29"])
		tr.NRozl = "'" + strings.TrimSpace(cc["30"])
		tr.NRach = "'" + strings.TrimSpace(cc["31"])
		tr.Name = strings.TrimSpace(cc["32"] + cc["33"] + cc["62"] + cc["63"])
		tr.code3 = "'" + strings.TrimSpace(cc["34"])
		if tr.code3 != "'" && tr.code3 != tr.Code1 {
			log.Printf("code3 != Code1: %s != %s", tr.code3, tr.Code1)
		}
		tr.IBAN = strings.TrimSpace(cc["38"])
		tr.Fee = strings.TrimSpace(cc["60"])
		tr.exRate2 = fixRate(strings.TrimSpace(cc["61"]))
		if tr.exRate2 != "" && tr.exRate2 != tr.ExRate {
			log.Printf("exRate2 != ExRate: %s != %s", tr.exRate2, tr.ExRate)
		}
	}
	tr.Hash = tr.hash()
	return tr
}

func scanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, decode(data[0:i]), nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), decode(data), nil
	}
	// Request more data.
	return 0, nil, nil
}

func decode(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		data = data[0 : len(data)-1]
	}

	decoder := charmap.CodePage852.NewDecoder()
	if l, err := decoder.Bytes(data); err != nil {
		log.Println("error:", err, string(data))
	} else {
		data = l
	}

	return bytes.TrimSpace(data)
}
