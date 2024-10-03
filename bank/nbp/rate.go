package nbp

import (
	"fmt"
	"time"
)

type FXRates struct {
	tabNumIdx     int
	tabNumFullIdx int
	tabNums       map[string]string

	isoIdxMap map[string]int

	nameList []string
	isoList  []string
	unitList []int
	fxRates  map[string][]float64
}

// PrevDayRate returns the exchange rate for the previous working day
// (as is the official way of calculating foreign exchange rates for accounting in Poland).
func (m *FXRates) PrevDayRate(t time.Time, c string) FX {
	t = t.AddDate(0, 0, -1)
	d := t.Format("20060102")
	for _, ok := m.fxRates[d]; !ok; {
		t = t.AddDate(0, 0, -1)
		d = t.Format("20060102")
		_, ok = m.fxRates[d]
	}
	return FX{
		Rate: m.fxRates[d][m.isoIdxMap[c]] / float64(m.unitList[m.isoIdxMap[c]]),
		Name: m.tabNums[d],
		Date: t,
		Calc: fmt.Sprintf("%d %s = %.4f PLN", m.unitList[m.isoIdxMap[c]], c, m.fxRates[d][m.isoIdxMap[c]]),
	}
}

type FX struct {
	Rate float64
	Name string
	Date time.Time
	Calc string
}

func (r FX) String() string {
	return fmt.Sprintf("%s tabela nr %s z dnia %s", r.Calc, r.Name, r.Date.Format("2006-01-02"))
}
