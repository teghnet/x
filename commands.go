package x

import (
	"context"
	"encoding/json"
	"iter"

	"charm.land/log/v2"
)

type Command func(context.Context) error
type CommandSelector func(string, []string) Command

// BatchOnce returns the Batch if the first argument is "batch",
// otherwise it runs the CommandSelector.
//
// This is to prevent an infinite batch command loop.
func BatchOnce(args []string, cs CommandSelector) Command {
	if args[0] == "batch" {
		return Batch(args[1:], cs)
	}
	return cs(args[0], args[2:])
}

func Batch(args []string, cs CommandSelector) Command {
	return func(ctx context.Context) error {
		inputFile := "-"
		keepRunningWhenError := false
		err := FlagsParse(args,
			Flag(&inputFile, "i", "batch input file"),
			Flag(&keepRunningWhenError, "continue", "keep running even if there are errors"),
		)
		if err != nil {
			return err
		}
		f, err := DynamicReader(inputFile)
		if err != nil {
			return err
		}
		defer ClosePrint(f)
		for err := range batchExec(ctx, json.NewDecoder(f), cs) {
			if err != nil {
				if !keepRunningWhenError {
					return err
				}
				log.Warn("command failed", "err", err)
			}
		}
		return nil
	}
}

func batchExec(ctx context.Context, dec *json.Decoder, cs CommandSelector) iter.Seq[error] {
	return func(yield func(error) bool) {
		for dec.More() {
			if err := ctx.Err(); err != nil {
				_ = yield(context.Cause(ctx))
				return
			}
			var args []string
			if err := dec.Decode(&args); err != nil {
				if !yield(err) {
					return
				}
			}
			if len(args) == 0 {
				if !yield(nil) {
					return
				}
			}
			if !yield(cs(args[0], args[1:])(ctx)) {
				return
			}
		}
	}
}
