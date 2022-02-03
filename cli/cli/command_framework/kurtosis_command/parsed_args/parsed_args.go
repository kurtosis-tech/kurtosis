package parsed_args

import "github.com/kurtosis-tech/stacktrace"

type ParsedArgs struct {
	nonGreedyArgs map[string]string
	greedyArgs    map[string][]string
}

func NewParsedArgs(nonGreedyArgs map[string]string, greedyArgs map[string][]string) *ParsedArgs {
	return &ParsedArgs{nonGreedyArgs: nonGreedyArgs, greedyArgs: greedyArgs}
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

