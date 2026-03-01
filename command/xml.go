// Copyright (c) 2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package command

import (
	"flag"
	"fmt"
	"io"
)

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

func defaultIO(name string, args []string, extras ...func(*flag.FlagSet)) (string, string) {
	flags := flag.NewFlagSet(name, flag.ExitOnError)
	i := flags.String("i", "", "input file name")
	o := flags.String("o", "-", "output file name")
	for _, e := range extras {
		e(flags)
	}
	_ = flags.Parse(args) // Ignore errors; CommandLine is set for ExitOnError.
	return *i, *o
}
