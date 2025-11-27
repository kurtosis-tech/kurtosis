package instance_id_arg

import (
	"context"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	defaultIsRequired     = true
	defaultValueEmpty     = ""
	validInstanceIdLength = 32
)

// InstanceIdentifierArg pre-builds instance identifier arg which has tab-completion and validation ready out-of-the-box
func InstanceIdentifierArg(
	// The arg key where this context identifier argument will be stored
	argKey string,
	isGreedy bool,
) *args.ArgConfig {

	validate := getValidationFunc(argKey, isGreedy)

	return &args.ArgConfig{
		Key:                   argKey,
		IsOptional:            defaultIsRequired,
		DefaultValue:          defaultValueEmpty,
		IsGreedy:              isGreedy,
		ValidationFunc:        validate,
		ArgCompletionProvider: args.NewManualCompletionsProvider(getCompletionsFunc()),
	}
}

func getCompletionsFunc() func(ctx context.Context, flags *flags.ParsedFlags, previousArgs *args.ParsedArgs) ([]string, error) {
	return func(ctx context.Context, flags *flags.ParsedFlags, previousArgs *args.ParsedArgs) ([]string, error) {
		// TODO: Given the instance id and the API Key, we could potentially query the API for instance ids to
		//  auto complete the typing but those endpoints don't exist (yet).
		return []string{}, nil
	}
}

func getValidationFunc(argKey string, isGreedy bool) func(context.Context, *flags.ParsedFlags, *args.ParsedArgs) error {
	return func(ctx context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error {
		var instanceIdsToValidate []string
		if isGreedy {
			instanceID, err := args.GetGreedyArg(argKey)
			if err != nil {
				return stacktrace.Propagate(err, "Expected a value for greedy arg '%v' but didn't find one", argKey)
			}
			instanceIdsToValidate = instanceID
		} else {
			instanceID, err := args.GetNonGreedyArg(argKey)
			if err != nil {
				return stacktrace.Propagate(err, "Expected a value for non-greedy arg '%v' but didn't find one", argKey)
			}
			instanceIdsToValidate = []string{instanceID}
		}

		for _, instanceIdToValidate := range instanceIdsToValidate {
			if len(instanceIdToValidate) < validInstanceIdLength {
				return stacktrace.NewError("Instance Id is not valid: %s", instanceIdsToValidate)
			}
		}
		return nil
	}
}
