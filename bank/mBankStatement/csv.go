package mBankStatement

import (
	"crypto/sha256"
	"encoding/csv"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/charmap"
)

var (
	// markers
	markerClient           = "#Klient"
	markerPeriod           = "#Za okres:"
	markerType             = "#Rodzaj rachunku"
	markerCurrency         = "#Waluta"
	markerAccountNumber    = "#Numer rachunku"
	markerNextCapDate      = "#Data następnej kapitalizacji"
	markerPercentage       = "#Oprocentowanie rachunku"
	markerCreditLimit      = "#Limit kredytu"
	markerCreditPercentage = "#Oprocentowanie kredytu"
	markerSummary          = "#Podsumowanie obrotów na rachunku"
	markerOpCount          = "#Liczba operacji"
	markerOpAmount         = "#Wartość operacji"
	markerOpenBal          = "#Saldo początkowe"
	markerCloseBal         = "#Saldo końcowe"
	markerStatement        = "Elektroniczne zestawienie operacji"
)

type Meta struct {
	colIdxMap map[string]int

	Desc string

	Client        string
	PeriodBegin   string
	PeriodEnd     string
	AccountType   string
	Currency      string
	AccountNumber string
	NextCapDate   string
	SavingsRate   string
	CreditLimit   string
	CreditRate    string

	OpCount    string
	OpAmount   string
	OpCountDb  string
	OpAmountDb string
	OpCountCd  string
	OpAmountCd string

	OpeningBal string
	ClosingBal string
}

func (m Meta) String() string {
	return fmt.Sprintf("%s [%s] mBank %s", m.Client, m.Currency, m.AccountNumber)
}

type Oper struct {
	AccClient   string
	AccCurrency string
	AccBank     string
	AccNumber   string

	PostingDate string
	Date        string

	Desc       string
	Title      string
	ExtName    string
	ExtAccount string

	Amount  string
	Balance string

	Hash string
}

var zeroByte = string([]byte{0})

func (o Oper) hash() []byte {
	h := sha256.Sum256([]byte(strings.Join([]string{
		o.AccClient,
		o.AccCurrency,
		o.AccBank,
		o.AccNumber,
		o.PostingDate,
		o.Date,
		o.Desc,
		o.Title,
		o.ExtName,
		o.ExtAccount,
		o.Amount,
		o.Balance,
	}, zeroByte)))
	return h[:16]
}

var (
	// headers
	headerAccDate    = "#Data księgowania"
	headerDate       = "#Data operacji"
	headerDesc       = "#Opis operacji"
	headerTitle      = "#Tytuł"
	headerExtName    = "#Nadawca/Odbiorca"
	headerExtAccount = "#Numer konta"
	headerAmount     = "#Kwota"
	headerBalance    = "#Saldo po operacji"
)

func ReadCSV(filePath string) (*Meta, []Oper, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, nil, err
	}
	defer closeFile(f)

	r := csv.NewReader(charmap.Windows1250.NewDecoder().Reader(f))
	r.Comma = ';'
	r.FieldsPerRecord = -1
	r.LazyQuotes = true

	rows, err := r.ReadAll()
	if err != nil {
		return nil, nil, err
	}

	m, o, err := extract(rows)
	if err != nil {
		return nil, nil, err
	}

	return m, o, nil
}

func extract(rows [][]string) (*Meta, []Oper, error) {
	var start, end int
	var m Meta
	for i, row := range rows {
		switch row[0] {
		case markerClient:
			m.Client = cleanQuote(rows[i+1][0])
		case markerPeriod:
			m.PeriodBegin = cleanDate(rows[i+1][0])
			m.PeriodEnd = cleanDate(rows[i+1][1])
		case markerType:
			m.AccountType = cleanQuote(rows[i+1][0])
		case markerCurrency:
			m.Currency = cleanQuote(rows[i+1][0])
		case markerAccountNumber:
			m.AccountNumber = cleanAccount(rows[i+1][0])
		case markerNextCapDate:
			m.NextCapDate = cleanDate(rows[i+1][0])
		case markerPercentage:
			m.SavingsRate = cleanNumber(rows[i+1][0])
		case markerCreditLimit:
			m.CreditLimit = cleanCurrency(rows[i+1][0], m.Currency)
		case markerCreditPercentage:
			m.CreditRate = cleanNumber(rows[i+1][0])
		case markerSummary:
			if row[1] == markerOpCount && row[2] == markerOpAmount {
				m.OpCountCd = cleanCurrency(rows[i+1][1], "")
				m.OpAmountCd = cleanCurrency(rows[i+1][2], m.Currency)
				m.OpCountDb = cleanCurrency(rows[i+2][1], "")
				m.OpAmountDb = cleanCurrency(rows[i+2][2], m.Currency)
				m.OpCount = cleanCurrency(rows[i+3][1], "")
				m.OpAmount = cleanCurrency(rows[i+3][2], m.Currency)
			}
		case markerOpenBal:
			m.OpeningBal = cleanCurrency(row[1], m.Currency)
		case markerStatement:
			m.Desc = row[0]
		case headerAccDate:
			start = i + 1
			m.colIdxMap = make(map[string]int)
			for j, v := range row {
				if v == "" {
					continue
				}
				m.colIdxMap[v] = j
			}
		default:
			if len(row) > 6 && row[6] == markerCloseBal {
				m.ClosingBal = cleanCurrency(row[7], m.Currency)
				end = i
			}
		}
	}

	var rxs []Oper
	for _, row := range rows[start:end] {
		rx := Oper{
			AccClient:   m.Client,
			AccCurrency: m.Currency,
			AccBank:     "mBank",
			AccNumber:   m.AccountNumber,
			PostingDate: cleanDate(row[m.colIdxMap[headerAccDate]]),
			Date:        cleanDate(row[m.colIdxMap[headerDate]]),
			Desc:        cleanQuote(row[m.colIdxMap[headerDesc]]),
			Title:       cleanQuote(row[m.colIdxMap[headerTitle]]),
			ExtName:     cleanQuote(row[m.colIdxMap[headerExtName]]),
			ExtAccount:  cleanAccount(row[m.colIdxMap[headerExtAccount]]),
			Amount:      cleanCurrency(row[m.colIdxMap[headerAmount]], m.Currency),
			Balance:     cleanCurrency(row[m.colIdxMap[headerBalance]], m.Currency),
		}
		rx.Hash = big.NewInt(0).SetBytes(rx.hash()).Text(62)
		rxs = append(rxs, rx)
	}
	return &m, rxs, nil
}

func cleanDate(s string) string {
	t, err := parseDate(cleanNumber(s))
	if err != nil {
		panic(err)
	}
	return t.Format("2006-01-02")
}

func cleanAccount(s string) string {
	return cleanNumber(s)
}

func cleanCurrency(s, cur string) string {
	s = strings.ReplaceAll(s, cur, "")
	if cur == "" {
		return cleanNumber(s)
	}
	return cleanNumber(s) + " " + cur
}

func cleanNumber(s string) string {
	s = cleanQuote(s)
	s = strings.ReplaceAll(s, " ", "")
	if strings.Count(s, ",") == 1 {
		s = strings.ReplaceAll(s, ",", ".")
	}
	if strings.HasSuffix(s, "%") {
		s = strings.TrimSuffix(s, "%")
		return movePointLeft(s)
	}
	return s
}

func cleanQuote(s string) string {
	s = strings.Trim(s, "'")
	s = strings.Join(strings.Fields(s), " ")
	return strings.TrimSpace(s)
}

func movePointLeft(s string) string {
	if s == "" {
		return "0"
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Fatal(err)
	}
	s = fmt.Sprintf("%.8f", v/100)
	return strings.TrimRight(strings.TrimRight(s, "0"), ".")
}

func closeFile(f *os.File) {
	if err := f.Close(); err != nil {
		log.Fatal(f.Name(), err)
	}
}

func parseDate(s string) (time.Time, error) {
	if strings.Count(s, ".") == 2 {
		t, err := time.Parse("02.01.2006", s)
		if err != nil {
			return time.Time{}, err
		}
		return t, nil
	}
	if strings.Count(s, "-") == 2 {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			return time.Time{}, err
		}
		return t, nil
	}
	t, err := time.Parse("20060102", s)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}
