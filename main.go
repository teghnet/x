package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"time"

	"github.com/teghnet/x/bank/mbank"
	"github.com/teghnet/x/bank/nbp"
)

var (
	filename string
)

func main() {
	log.SetFlags(0)

	if err := nbp.GetArch(2024); err != nil {
		log.Fatal(err)
	}
	rates, err := nbp.ReadArch(2024)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(rates.PrevDayRate(time.Now(), "USD"))
	fmt.Println(rates.PrevDayRate(time.Now(), "EUR"))
	fmt.Println(rates.PrevDayRate(time.Now(), "THB"))

	files, err := filepath.Glob(filename)
	for _, file := range files {
		log.Println(file)
		m, _, err := mbank.ReadCSV(file)
		if err != nil {
			log.Fatal(err)
		}

		if m.Currency != "PLN" {
			bal, err := strconv.ParseFloat(m.CloseBal, 64)
			if err != nil {
				log.Fatal(err)
			}

			m.CloseBal = fmt.Sprintf("%f", bal*rates.PrevDayRate(time.Now(), m.Currency).Rate)
		}

		fmt.Printf("%#v\n", m)
		// for _, op := range ops {
		// 	fmt.Printf("%#v\n", op)
		// }
	}
}