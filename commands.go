package x

import (
	"context"
	"encoding/json"
	"iter"
)

func ErrCommand(err error) Command { return func(ctx context.Context) error { return err } }

// Command represents an executable operation that accepts a context and returns an error.
type Command func(context.Context) error

// CommandSelector is a function type that selects and returns a [Command] based on a command name and arguments.
type CommandSelector func(args []string) Command

// BatchOnce returns the Batch if the first argument is "batch",
// otherwise it runs the CommandSelector.
//
// This is to prevent an infinite batch command loop.
func BatchOnce(args []string, cs CommandSelector) Command {
	if len(args) > 0 && args[0] == "batch" {
		return Batch(args[1:], cs)
	}
	return cs(args)
}

// Batch creates a Command that executes multiple commands from a JSON-formatted input file.
// Each line in the input should be a JSON array of strings representing command arguments.
// Accepts `-i` flag to specify input file (default: stdin)
// and `--continue` flag to keep running on errors.
// Uses the provided CommandSelector to resolve and execute each command from the batch input.
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
		defer Close(f)
		for err := range batchExec(ctx, json.NewDecoder(f), cs) {
			if err != nil {
				if !keepRunningWhenError {
					return err
				}
				Warn("command failed", "err", err)
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
			if !yield(cs(args)(ctx)) {
				return
			}
		}
	}
}
