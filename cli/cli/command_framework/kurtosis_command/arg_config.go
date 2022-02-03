package kurtosis_command

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/kurtosis_command/parsed_args"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/kurtosis_command/parsed_flags"
)

type ArgConfig struct {
	Key string

	// If set to true, this argument can be ommitted
	IsOptional bool

	// For non-greedy args, this must be a string
	// For greedy args, this must be a []string
	// Has no effect in non-optional arguments!
	DefaultValue interface{}

	// If set to true, this arg will consume all the remaining arg strings
	IsGreedy bool

	// This function is for generating the valid completions that the shell can use for this arg, and can be dynamically
	//  modified using previous arg values
	// The previousArgs will only contain values for the args that come before this one (since completion doesn't make
	//  sense in the middle of an entry
	CompletionsFunc func(flags *parsed_flags.ParsedFlags, previousArgs *parsed_args.ParsedArgs) ([]string, error)

	// Will be run after the user presses ENTER and before we start actually running the command
	ValidationFunc func(flags *parsed_flags.ParsedFlags, args *parsed_args.ParsedArgs) error
}

