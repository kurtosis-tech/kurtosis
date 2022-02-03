package framework

import "github.com/kurtosis-tech/stacktrace"

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
	CompletionsFunc func(flags *ParsedFlags, previousArgs *ParsedArgs) ([]string, error)

	// Will be run after the user presses ENTER and before we start actually running the command
	ValidationFunc func(flags *ParsedFlags, args *ParsedArgs) error
}

type ParsedArgs struct {
	nonGreedyArgs map[string]string
	greedyArgs    map[string][]string
}
// Non-greedy args only have a single value
func (args *ParsedArgs) GetNonGreedyArg(key string) (string, error) {
	result, found := args.nonGreedyArgs[key]
	if !found {
		return "", stacktrace.NewError("No non-greedy arg with key '%v'", key)
	}
	return result, nil
}
// Greedy args can have multiple values
func (args *ParsedArgs) GetGreedyArg(key string) ([]string, error) {
	elems, found := args.greedyArgs[key]
	if !found {
		return nil, stacktrace.NewError("No greedy arg with key '%v'", key)
	}
	return elems, nil
}

