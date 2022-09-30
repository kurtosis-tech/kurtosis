package set_selection_arg

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/stacktrace"
	"sort"
	"strings"
)

// TODO We could make this greedy if something requires it down the line
// Creates an ArgConfig that allows for selecting one of a set of choices
func NewSetSelectionArg(argKey string, validValues map[string]bool) *args.ArgConfig {
	sortedValidValues := []string{}
	for validValue := range validValues {
		sortedValidValues = append(sortedValidValues, validValue)
	}
	sort.Strings(sortedValidValues)

	completionsFunc := func(ctx context.Context, flags *flags.ParsedFlags, previousArgs *args.ParsedArgs) ([]string, error) {
		return sortedValidValues, nil
	}

	validationFunc := func(ctx context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error {
		value, err := args.GetNonGreedyArg(argKey)
		if err != nil {
			return stacktrace.Propagate(err, "Expected a value for non-greedy arg '%v' but didn't find one", argKey)
		}
		if _, found := validValues[value]; !found {
			return stacktrace.NewError(
				"Value for arg '%v' was '%v', but must be in set {%v}",
				argKey,
				value,
				strings.Join(sortedValidValues, ", "),
			)
		}
		return nil
	}

	return &args.ArgConfig{
		Key:             argKey,
		CompletionsFunc: completionsFunc,
		ValidationFunc: validationFunc,
	}
}
