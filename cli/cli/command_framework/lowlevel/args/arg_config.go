package args

import (
	"context"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/lowlevel/flags"
)

type ArgConfig struct {
	Key string

	// TODO Add & render descriptions!!

	// If set to true, this argument can be omitted
	IsOptional bool

	// For non-greedy args, this must be a string
	// For greedy args, this must be a []string
	// Has no effect in non-optional arguments!
	DefaultValue interface{}

	// If set to true, this arg will consume all the remaining arg strings
	IsGreedy bool

	// Define the argument completion provider which can be one of these:
	// 1- A ManualCompletionsProvider: which will be provided by the custom argument type
	// 2- A DefaultFileCompletionProvider: this one enables the default shell file completion functionality which is disabled in the manual type
	// The ArgCompletionProvider contains a function RunCompletionFunction for generating the valid completions that the shell can use for this arg, and can be dynamically
	//modified using previous arg values
	// The previousArgs will only contain values for the args that come before this one (since completion doesn't make
	// sense in the middle of an entry
	ArgCompletionProvider argCompletionProvider

	// Will be run after the user presses ENTER and before we start actually running the command
	ValidationFunc func(ctx context.Context, flags *flags.ParsedFlags, args *ParsedArgs) error
}
