// Copyright (c) 2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package command

import (
	"flag"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"

	"github.com/teghnet/x/internal"
	"github.com/teghnet/x/osio"
)

func XMLStats(args []string) error {
	flags := flag.NewFlagSet("XMLStats", flag.PanicOnError)
	i := flags.String("i", "", "input file name")
	o := flags.String("o", "-", "output file name")
	_ = flags.Parse(args) // Ignore errors; CommandLine is set for PanicOnError.

	r, err := osio.DynamicReader(*i)
	if err != nil {
		return err
	}
	defer internal.ClosePrint(r)

	w, err := osio.DynamicWriter(*o, false)
	if err != nil {
		return err
	}
	defer internal.ClosePrint(w)

	counter := make(map[string]int64)
	dicts := make(map[string][]string)
	for k, v := range osio.XMLDicts(r) {
		if k == "" {
			continue
		}
		if counter[k+":"+v] == 0 {
			dicts[k] = append(dicts[k], v)
		}
		counter[k+":"+v]++
	}
	fPrint(w, "pole\twartość\tlicznik\n")
	keys := slices.Collect(maps.Keys(dicts))
	slices.Sort(keys)
	for _, k := range keys {
		fPrint(w, k, "\n")
	}
	// reDate := regexp.MustCompile(`^\d{4}-\d{1,2}-\d{1,2}$`)
	for k, vals := range dicts {
		if strings.TrimSpace(k) == "" {
			continue
		}
		fPrintf(w, "%s\t\t%d\n", k, len(vals))
		for _, v := range vals {
			// if _, err := strconv.ParseFloat(v, 64); err == nil || reDate.MatchString(v) {
			// 	continue
			// }
			fPrintf(w, "%s\t%s\t%d\n", k, v, counter[k+":"+v])
		}
	}
	return nil
}
func fPrintf(w io.Writer, format string, a ...interface{}) {
	_, err := fmt.Fprintf(w, format, a...)
	if err != nil {
		panic(err)
	}
}
func fPrint(w io.Writer, a ...interface{}) {
	_, err := fmt.Fprint(w, a...)
	if err != nil {
		panic(err)
	}
}
func XMLPassthrough(args []string) error {
	flags := flag.NewFlagSet("XMLPassthrough", flag.PanicOnError)

	i := flags.String("i", "", "input file name")
	o := flags.String("o", "-", "output file name")

	// Ignore errors; CommandLine is set for PanicOnError.
	_ = flags.Parse(args)

	r, err := osio.DynamicReader(*i)
	if err != nil {
		return err
	}
	defer internal.ClosePrint(r)

	w, err := osio.DynamicWriter(*o, false)
	if err != nil {
		return err
	}
	defer internal.ClosePrint(w)

	return osio.TrimXML(r, w)
}
