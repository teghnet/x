package nbp

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/charmap"

	"github.com/teghnet/x/file"
)

func Arch(ctx context.Context, stateDir string) (*FXRates, error) {
	if err := GetArch(2024, stateDir); err != nil {
		return nil, fmt.Errorf("unable to GET archive: %w", err)
	}
	fxRates, err := ReadArch(2024, stateDir)
	if err != nil {
		return nil, fmt.Errorf("unable to read archive: %w", err)
	}
	return fxRates, nil
}

// GetArch downloads the NBP FX rates archive for the given year.
// The file is saved in the state directory with the name `archiwum_tab_a_<year>.csv`.
// The year must be in the range 2012 until the current year.
// The file is downloaded only if it is not up to date.
// `opts` is a list of strings that can be used to specify the location of the state directory.
// See: `conf.StateDir` for more information about the `opts` argument.
func GetArch(y int, stateDir string) error {
	if y < 2012 {
		return fmt.Errorf("year out of range (CSV NPB FX rates archive is available since 2012)")
	}
	if y > time.Now().Year() {
		return fmt.Errorf("year out of range (CSV NPB FX rates archive is available until the current year)")
	}
	p, err := archPath(y, stateDir)
	if err != nil {
		return err
	}
	fi, err := os.Stat(p)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if fi != nil && fi.ModTime().After(time.Now().Truncate(24*time.Hour)) {
		log.Printf("Archive %s is up to date", p)
		return nil
	}
	log.Printf("Downloading archive into %s", p)
	return file.Download(fmt.Sprintf("https://static.nbp.pl/dane/kursy/Archiwum/%s", archFile(y)), p)
}

// ReadArch reads the NBP FX rates archive for the given year.
// The file must be in the state directory with the name `archiwum_tab_a_<year>.csv`.
// `opts` is a list of strings that can be used to specify the location of the state directory.
// See: `conf.StateDir` for more information `opts` argument.
func ReadArch(y int, stateDir string) (*FXRates, error) {
	p, err := archPath(y, stateDir)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer closeFile(f)

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
		if len(rates) == 0 {
			parsed, err := time.Parse("20060102", row[0])
			if err != nil {
				return nil, err
			}
			meta.firstDate = parsed
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

func archFile(y int) string {
	return fmt.Sprintf("archiwum_tab_a_%d.csv", y)
}

func archPath(y int, stateDir string) (string, error) {
	return path.Join(stateDir, archFile(y)), nil
}

func normalizeFloat(old string) string {
	s := strings.Replace(old, " ", "", -1)
	return strings.Replace(s, ",", ".", 1)
}

func closeFile(f *os.File) {
	if err := f.Close(); err != nil {
		log.Fatal(f.Name(), err)
	}
}

var (
	headerDate       = "data"
	headerTabNum     = "nr tabeli"
	headerTabNumFull = "pełny numer tabeli"

	headerISOCode      = "kod ISO"
	headerCurrencyName = "nazwa waluty"
	headerUnits        = "liczba jednostek"
)
