package args

import (
	"github.com/kurtosis-tech/stacktrace"
)

type ParsedArgs struct {
	nonGreedyArgs map[string]string
	greedyArgs    map[string][]string
}

// Takes in the currently-entered arg strings, categorizes them according to the arg configs defined, and
//  returns the kurtosis_command.ArgConfig whose completion function should be used
// NOTES:
//  - If the user presses TAB in the middle of several args (e.g. "arg1 arg2  TAB   arg3"), then `input` will only contain
//     the previous args (which is actually good behaviour)
//  - If the input isn't long enough, the resulting ParsedArgs object won't have arg strings for all the args
//  - A nil value for the returned kurtosis_command.ArgConfig indicates that no completion should be used
func ParseArgsForCompletion(argConfigs []*ArgConfig, input []string) (*ParsedArgs, *ArgConfig) {
	nonGreedyArgValues := map[string]string{}
	greedyArgValues := map[string][]string{}

	var nextArg *ArgConfig = nil
	if len(argConfigs) > 0 {
		nextArg = argConfigs[0]
	}

	configIdx := 0
	inputIdx := 0
	for configIdx < len(argConfigs) && inputIdx < len(input) {
		config := argConfigs[configIdx]
		key := config.Key

		// Greedy case (arg must always be last)
		if config.IsGreedy {
			greedyArgValues[key] = input[inputIdx:]
			inputIdx += len(input) - inputIdx
			nextArg = config  // Greedy args must always be at the end, so they'll be infinitely-completable
			break
		}

		// Non-greedy case
		nonGreedyArgValues[key] = input[inputIdx]
		configIdx += 1
		inputIdx += 1
		if configIdx >= len(argConfigs) {
			// If there's not another kurtosis_command.ArgConfig (indicating we've used them all) then we return nil to indicate
			//  that tab completion shouldn't be done (since no more args are needed)
			nextArg = nil
		} else {
			nextArg = argConfigs[configIdx]
		}
	}
	result := &ParsedArgs{
		nonGreedyArgs: nonGreedyArgValues,
		greedyArgs:    greedyArgValues,
	}
	return result, nextArg
}

// Parses all the args, guaranteeing that the required args are filled out and that default values for non-optional arguments get applied
// This means that if no error was returned, the returned ParsedArgs object is guaranteed to have all the args that were
//  passed in
func ParseArgsForValidation(argConfigs []*ArgConfig, input []string) (*ParsedArgs, error) {
	nonGreedyArgValues := map[string]string{}
	greedyArgValues := map[string][]string{}
	inputIdx := 0
	for configIdx, config := range argConfigs {
		key := config.Key
		if config.IsOptional && configIdx < len(argConfigs) - 1 {
			return nil, stacktrace.NewError("Arg '%v' is marked as optional, but isn't the last argument; this is a bug in Kurtosis!", key)
		}
		if inputIdx >= len(input) {
			if !config.IsOptional {
				return nil, stacktrace.NewError("Missing required arg '%v'", config.Key)
			}
			if config.IsGreedy {
				defaultVal, ok := config.DefaultValue.([]string)
				if !ok {
					return nil, stacktrace.NewError("Greedy arg '%v' wasn't provided, but the default value wasn't a string list; this is a bug in Kurtosis!", key)
				}
				greedyArgValues[key] = defaultVal
			} else {
				defaultVal, ok := config.DefaultValue.(string)
				if !ok {
					return nil, stacktrace.NewError("Greedy arg '%v' wasn't provided, but the default value wasn't a string; this is a bug in Kurtosis!", key)
				}
				nonGreedyArgValues[key] = defaultVal
			}
			continue
		}

		// Greedy case (arg must always be last)
		if config.IsGreedy {
			greedyArgValues[key] = input[inputIdx:]
			inputIdx += len(input) - inputIdx
			break
		}

		nonGreedyArgValues[key] = input[inputIdx]
		inputIdx += 1
	}

	numArgTokenGroupings := len(greedyArgValues) + len(nonGreedyArgValues)
	if numArgTokenGroupings != len(argConfigs) {
		return nil, stacktrace.NewError(
			"Expected '%v' arg token groups, but got '%v'",
			len(argConfigs),
			numArgTokenGroupings,
		)
	}


	result := &ParsedArgs{
		nonGreedyArgs: nonGreedyArgValues,
		greedyArgs:    greedyArgValues,
	}
	return result, nil
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

