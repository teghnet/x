package revolut

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/teghnet/x/ops"
)

var headers = []string{
	"Date started (UTC)",
	"Date completed (UTC)",
	"Date started (Europe/Warsaw)",
	"Date completed (Europe/Warsaw)",
	"ID",
	"Type",
	"State",
	"Description",
	"Reference",
	"Payer",
	"Card number",
	"Card label",
	"Card state",
	"Orig currency",
	"Orig amount",
	"Payment currency",
	"Amount",
	"Total amount",
	"Exchange rate",
	"Fee",
	"Fee currency",
	"Balance",
	"Account",
	"Beneficiary account number",
	"Beneficiary sort code or routing number",
	"Beneficiary IBAN",
	"Beneficiary BIC",
	"MCC",
	"Related transaction id",
	"Spend program",
}

const (
	idxDateStartedUTC int = iota
	idxDateCompletedUTC
	idxDateStarted
	idxDateCompleted
	idxID
	idxType
	idxState
	idxDesc
	idxRef
	idxPayer
	idxCardNo
	idxCardLab
	idxCardState
	idxOrigCur
	idxOrigAmount
	idxPaymentCur
	idxAmount
	idxTotalAmount
	idxExchangeRate
	idxFee
	idxFeeCur
	idxBalance
	idxAccount
	idxBenAccNum
	idxBenRoute
	idxBenIBAN
	idxBenBIC
	idxMCC
	idxRelTr
	idxSpend
)

type Entry struct {
	DateStartedUTC   string
	DateCompletedUTC string
	DateStarted      string
	DateCompleted    string
	ID               string
	Type             string
	State            string
	Desc             string
	Ref              string
	Payer            string
	CardNo           string
	CardLab          string
	CardState        string
	OrigCur          string
	OrigAmount       string
	PaymentCur       string
	Amount           string
	TotalAmount      string
	ExchangeRate     string
	Fee              string
	FeeCur           string
	Balance          string
	Account          string
	BenAccNum        string
	BenRoute         string
	BenIBAN          string
	BenBIC           string
	MCC              string
	RelTr            string
	Spend            string

	Hash []byte
}

func (e Entry) hash() []byte {
	return ops.Hash([]string{
		e.DateStartedUTC,
		e.DateCompletedUTC,
		e.DateStarted,
		e.DateCompleted,
		e.ID,
		e.Type,
		e.State,
		e.Desc,
		e.Ref,
		e.Payer,
		e.CardNo,
		e.CardLab,
		e.CardState,
		e.OrigCur,
		e.OrigAmount,
		e.PaymentCur,
		e.Amount,
		e.TotalAmount,
		e.ExchangeRate,
		e.Fee,
		e.FeeCur,
		e.Balance,
		e.Account,
		e.BenAccNum,
		e.BenRoute,
		e.BenIBAN,
		e.BenBIC,
		e.MCC,
		e.RelTr,
		e.Spend,
	})
}

func ReadCSV(filePath string) ([]Entry, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	r := csv.NewReader(file)
	r.Comma = ','
	r.Comment = '#'
	r.FieldsPerRecord = 0
	r.LazyQuotes = true
	r.TrimLeadingSpace = true

	var entries []Entry
	for {
		line, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if len(line) == 0 {
			break
		}
		if strings.Join(line, "") == "" {
			continue
		}
		if strings.Count(line[idxDateStartedUTC], "-") != 2 {
			for i, v := range line {
				if i == idxDateStarted || i == idxDateCompleted {
					continue
				}
				if v != headers[i] {
					return nil, fmt.Errorf("invalid header %s at index %d", v, i)
				}
			}
			continue
		}
		entry := Entry{
			DateStartedUTC:   line[idxDateStartedUTC],
			DateCompletedUTC: line[idxDateCompletedUTC],
			DateStarted:      line[idxDateStarted],
			DateCompleted:    line[idxDateCompleted],
			ID:               line[idxID],
			Type:             line[idxType],
			State:            line[idxState],
			Desc:             line[idxDesc],
			Ref:              line[idxRef],
			Payer:            line[idxPayer],
			CardNo:           line[idxCardNo],
			CardLab:          line[idxCardLab],
			CardState:        line[idxCardState],
			OrigCur:          line[idxOrigCur],
			OrigAmount:       line[idxOrigAmount],
			PaymentCur:       line[idxPaymentCur],
			Amount:           line[idxAmount],
			TotalAmount:      line[idxTotalAmount],
			ExchangeRate:     line[idxExchangeRate],
			Fee:              line[idxFee],
			FeeCur:           line[idxFeeCur],
			Balance:          line[idxBalance],
			Account:          line[idxAccount],
			BenAccNum:        line[idxBenAccNum],
			BenRoute:         line[idxBenRoute],
			BenIBAN:          line[idxBenIBAN],
			BenBIC:           line[idxBenBIC],
			MCC:              line[idxMCC],
			RelTr:            line[idxRelTr],
			Spend:            line[idxSpend],
		}
		entry.Hash = entry.hash()

		entries = append(entries, entry)
	}
	return entries, nil
}
