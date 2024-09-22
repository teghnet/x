package nbp

import (
	"encoding/csv"
	"fmt"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/charmap"

	"github.com/teghnet/x/file"
)

func archFile(y int) string {
	return fmt.Sprintf("archiwum_tab_a_%d.csv", y)
}
func archPath(y int) string {
	return path.Join(".local", archFile(y))
}
func archLink(y int) string {
	return fmt.Sprintf("https://static.nbp.pl/dane/kursy/Archiwum/%s", archFile(y))
}

func GetArch(y int) error {
	return file.Download(archLink(y), archPath(y))
}

func ReadArch(y int) (*FXRates, error) {
	f, err := os.Open(archPath(y))
	if err != nil {
		return nil, err
	}
	defer file.CloseFile(f)

	r := csv.NewReader(charmap.Windows1250.NewDecoder().Reader(f))
	r.Comma = ';'
	r.FieldsPerRecord = -1
	r.LazyQuotes = true

	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	var rates [][]string
	var maxIdx int
	var meta FXRates
	for _, row := range rows {
		if row[0] == headerDate {
			for i, v := range row[1:] {
				if v == headerTabNum {
					meta.tabNumIdx = i
					continue
				}
				if v == headerTabNumFull {
					meta.tabNumFullIdx = i
					continue
				}
			}
			continue
		}
		if row[0] == "" && len(meta.nameList) == 0 {
			meta.nameList = row[1:]
			continue
		}
		if row[0] == headerISOCode && meta.isoIdxMap == nil {
			meta.isoIdxMap = make(map[string]int)
			for j, v := range row[1:] {
				if v == "" {
					continue
				}
				meta.isoIdxMap[v] = j
				maxIdx = j
			}
			continue
		}
		if row[0] == headerCurrencyName && len(meta.nameList) > 0 {
			if slices.Compare(meta.nameList, row[1:]) != 0 {
				return nil, fmt.Errorf("currency names mismatch")
			}
			continue
		}
		if row[0] == headerUnits && len(meta.unitList) == 0 {
			for _, v := range row[1:] {
				if v == "" {
					continue
				}
				u, err := strconv.Atoi(v)
				if err != nil {
					return nil, err
				}
				meta.unitList = append(meta.unitList, u)
			}
			continue
		}
		rates = append(rates, row)
	}
	meta.fxRates = make(map[string][]float64)
	meta.tabNums = make(map[string]string)
	for _, row := range rates {
		date := row[0]
		var fxs []float64
		for i, v := range row[1:] {
			if i > maxIdx {
				if i == meta.tabNumFullIdx {
					meta.tabNums[date] = v
					break
				}
				continue
			}
			rate, err := strconv.ParseFloat(normalizeFloat(v), 64)
			if err != nil {
				return nil, err
			}
			fxs = append(fxs, rate)
		}
		meta.fxRates[date] = fxs
	}
	return &meta, nil
}

func normalizeFloat(old string) string {
	s := strings.Replace(old, " ", "", -1)
	return strings.Replace(s, ",", ".", 1)
}

var (
	headerDate       = "data"
	headerTabNum     = "nr tabeli"
	headerTabNumFull = "pełny numer tabeli"

	headerISOCode      = "kod ISO"
	headerCurrencyName = "nazwa waluty"
	headerUnits        = "liczba jednostek"
)
