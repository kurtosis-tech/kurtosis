package instance_id_arg

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
)

const (
	defaultIsRequired = true
	defaultValueEmpty = ""
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
		// TODO: Add some basic validation, maybe just checking the string isn't empty
		return nil
	}
}
