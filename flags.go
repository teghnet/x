package x

import (
	"encoding"
	"flag"
	"fmt"
	"time"
)

// FlagsParse parses command-line arguments into a flag.FlagSet,
// applying customizations via provided FlagOption functions.
func FlagsParse(args []string, extras ...FlagOption) error {
	return flagSet(extras...).Parse(args)
}

// FlagOption is a function type used to customize a flag.FlagSet during its initialization.
type FlagOption func(*flag.FlagSet)

// flagSet creates a new flag.FlagSet.
func flagSet(extras ...FlagOption) *flag.FlagSet {
	flags := flag.NewFlagSet("", flag.ExitOnError)
	for _, e := range extras {
		e(flags)
	}
	return flags
}

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

// primitives is a type constraint that holds all primitive types supported by [flag]
// and time.Duration.
type primitives interface {
	bool | int | int64 | uint | uint64 | string | float64 |
		time.Duration
}

func Flag[T primitives](p *T, name string, value T, usage string) FlagOption {
	return func(flags *flag.FlagSet) {
		switch v := any(p).(type) {
		case *int:
			flags.IntVar(v, name, (any(value)).(int), usage)
		case *int64:
			flags.Int64Var(v, name, (any(value)).(int64), usage)
		case *uint:
			flags.UintVar(v, name, (any(value)).(uint), usage)
		case *uint64:
			flags.Uint64Var(v, name, (any(value)).(uint64), usage)
		case *string:
			flags.StringVar(v, name, (any(value)).(string), usage)
		case *float64:
			flags.Float64Var(v, name, (any(value)).(float64), usage)
		case *time.Duration:
			flags.DurationVar(v, name, (any(value)).(time.Duration), usage)
		default:
			panic(fmt.Sprintf("unsupported type: %T", v))
		}
	}
}
func FlagInt(p *int, name string, value int, usage string) FlagOption {
	return func(flags *flag.FlagSet) {
		flags.IntVar(p, name, value, usage)
	}
}

// FlagText defines a flag with a custom TextMarshaler and TextUnmarshaler to parse and format its value.
func FlagText(p encoding.TextUnmarshaler, name string, value encoding.TextMarshaler, usage string) FlagOption {
	return func(flags *flag.FlagSet) {
		flags.TextVar(p, name, value, usage)
	}
}
