package mBankStatementCSV

import (
	"encoding/csv"
	"fmt"
	"log"
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

	desc string

	client        string
	periodBegin   string
	periodEnd     string
	accountType   string
	Currency      string
	accountNumber string
	nextCapDate   string
	savingsRate   string
	creditLimit   string
	creditRate    string

	opCount    string
	opAmount   string
	opCountDb  string
	opAmountDb string
	opCountCd  string
	opAmountCd string

	openBal  string
	CloseBal string
}

type Op struct {
	date      string
	opDate    string
	opDesc    string
	opTitle   string
	party     string
	account   string
	Amount    string
	PostOpBal string
}

var (
	// headers
	headerDate      = "#Data księgowania"
	headerOpDate    = "#Data operacji"
	headerOpDesc    = "#Opis operacji"
	headerOpTitle   = "#Tytuł"
	headerParty     = "#Nadawca/Odbiorca"
	headerAccount   = "#Numer konta"
	headerAmount    = "#Kwota"
	headerPostOpBal = "#Saldo po operacji"
)

type data struct {
	date      string
	opDate    string
	opDesc    string
	opTitle   string
	party     string
	account   string
	amount    string
	postOpBal string
}

func ReadCSV(filename string) (*Meta, []Op, error) {
	f, err := os.Open(filename)
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

func extract(rows [][]string) (*Meta, []Op, error) {
	var start, end int
	var m Meta
	for i, row := range rows {
		switch row[0] {
		case markerClient:
			m.client = cleanQuote(rows[i+1][0])
		case markerPeriod:
			m.periodBegin = cleanDate(rows[i+1][0])
			m.periodEnd = cleanDate(rows[i+1][1])
		case markerType:
			m.accountType = cleanQuote(rows[i+1][0])
		case markerCurrency:
			m.Currency = cleanQuote(rows[i+1][0])
		case markerAccountNumber:
			m.accountNumber = cleanAccount(rows[i+1][0])
		case markerNextCapDate:
			m.nextCapDate = cleanDate(rows[i+1][0])
		case markerPercentage:
			m.savingsRate = cleanNumber(rows[i+1][0])
		case markerCreditLimit:
			m.creditLimit = cleanCurrency(rows[i+1][0], m.Currency)
		case markerCreditPercentage:
			m.creditRate = cleanNumber(rows[i+1][0])
		case markerSummary:
			if row[1] == markerOpCount && row[2] == markerOpAmount {
				m.opCountCd = cleanCurrency(rows[i+1][1], m.Currency)
				m.opAmountCd = cleanCurrency(rows[i+1][2], m.Currency)
				m.opCountDb = cleanCurrency(rows[i+2][1], m.Currency)
				m.opAmountDb = cleanCurrency(rows[i+2][2], m.Currency)
				m.opCount = cleanCurrency(rows[i+3][1], m.Currency)
				m.opAmount = cleanCurrency(rows[i+3][2], m.Currency)
			}
		case markerOpenBal:
			m.openBal = cleanCurrency(row[1], m.Currency)
		case markerStatement:
			m.desc = row[0]
		case headerDate:
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
				m.CloseBal = cleanCurrency(row[7], m.Currency)
				end = i
			}
		}
	}

	var rxs []Op
	for _, row := range rows[start:end] {
		rx := Op{
			date:      cleanDate(row[m.colIdxMap[headerDate]]),
			opDate:    cleanDate(row[m.colIdxMap[headerOpDate]]),
			opDesc:    cleanQuote(row[m.colIdxMap[headerOpDesc]]),
			opTitle:   cleanQuote(row[m.colIdxMap[headerOpTitle]]),
			party:     cleanQuote(row[m.colIdxMap[headerParty]]),
			account:   cleanAccount(row[m.colIdxMap[headerAccount]]),
			Amount:    cleanCurrency(row[m.colIdxMap[headerAmount]], m.Currency),
			PostOpBal: cleanCurrency(row[m.colIdxMap[headerPostOpBal]], m.Currency),
		}
		rxs = append(rxs, rx)
	}
	return &m, rxs, nil
}

func cleanDate(s string) string {
	s = cleanNumber(s)
	if strings.Count(s, ".") == 2 {
		t, err := time.Parse("02.01.2006", s)
		if err != nil {
			panic(err)
		}
		return t.Format("2006-01-02")
	}
	if strings.Count(s, "-") == 2 {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			panic(err)
		}
		return t.Format("2006-01-02")
	}
	return s
}
func cleanAccount(s string) string {
	return cleanNumber(s)
}
func cleanCurrency(s, cur string) string {
	s = strings.ReplaceAll(s, cur, "")
	return cleanNumber(s)
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
