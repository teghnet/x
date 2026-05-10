// Copyright (c) 2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package x

import (
	"encoding"
	"flag"
	"fmt"
	"strings"
	"time"
)

// FlagsArgs parses command-line arguments and returns the remaining non-flag arguments after parsing.
// It creates a new FlagSet with the provided options, parses the args, and returns unparsed arguments and any error.
func FlagsArgs(args []string, extras ...FlagOption) ([]string, error) {
	extras = append(extras, FlagSetErrorHandling(flag.ContinueOnError))
	fs := flagSet(extras...)
	return fs.Args(), fs.Parse(args)
}

// FlagsParse parses command-line arguments into a flag.FlagSet,
// applying customizations via provided FlagOption functions.
func FlagsParse(args []string, extras ...FlagOption) error {
	extras = append(extras, FlagSetErrorHandling(flag.ContinueOnError))
	return flagSet(extras...).Parse(args)
}

// flagSet creates a new flag.FlagSet.
func flagSet(extras ...FlagOption) *flag.FlagSet {
	flags := flag.NewFlagSet("", flag.PanicOnError)
	for _, e := range extras {
		e(flags)
	}
	return flags
}

// FlagOption is a function type used to customize a flag.FlagSet during its initialization.
type FlagOption func(*flag.FlagSet)

// FlagSetName changes the name in the flag.FlagSet.
func FlagSetName(name string) FlagOption {
	return func(flags *flag.FlagSet) {
		flags.Init(name, flags.ErrorHandling())
	}
}

// FlagSetErrorHandling changes the name in the flag.FlagSet.
func FlagSetErrorHandling(errorHandling flag.ErrorHandling) FlagOption {
	return func(flags *flag.FlagSet) {
		flags.Init(flags.Name(), errorHandling)
	}
}

// flaggables is a type constraint that holds all types supported by [Flag].
type flaggables interface {
	bool | int | int64 | uint | uint64 | string | float64 | []string | time.Duration
}

func Flag[T flaggables](p *T, name, usage string) FlagOption {
	return func(flags *flag.FlagSet) {
		switch v := any(p).(type) {
		case *bool:
			flags.BoolVar(v, name, *v, usage)
		case *int:
			flags.IntVar(v, name, *v, usage)
		case *int64:
			flags.Int64Var(v, name, *v, usage)
		case *uint:
			flags.UintVar(v, name, *v, usage)
		case *uint64:
			flags.Uint64Var(v, name, *v, usage)
		case *string:
			flags.StringVar(v, name, *v, usage)
		case *float64:
			flags.Float64Var(v, name, *v, usage)
		case *[]string:
			flags.Var((*stringSlice)(v), name, usage)
		case *time.Duration:
			flags.DurationVar(v, name, *v, usage)
		default:
			panic(fmt.Sprintf("unsupported type: %T", v))
		}
	}
}

var _ flag.Value = (*stringSlice)(nil)

type stringSlice []string

func (f *stringSlice) String() string {
	return fmt.Sprint([]string(*f))
}

func (f *stringSlice) Set(value string) error {
	*f = append(*f, strings.Split(value, ",")...)
	return nil
}

// FlagText defines a flag with a custom TextMarshaler and TextUnmarshaler to parse and format its value.
func FlagText(p encoding.TextUnmarshaler, name string, value encoding.TextMarshaler, usage string) FlagOption {
	return func(flags *flag.FlagSet) {
		flags.TextVar(p, name, value, usage)
	}
}
