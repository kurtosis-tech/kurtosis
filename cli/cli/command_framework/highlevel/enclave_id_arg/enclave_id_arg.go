package enclave_id_arg

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/stacktrace"
	"sort"
)

// Prebuilt enclave identifier arg which has tab-completion and validation ready out-of-the-box
func NewEnclaveIdentifierArg(
	// The arg key where this enclave ID argument will be stored
	argKey string,
	// TODO SWITCH THIS TO A KURTOSISCONTEXT ONCE https://github.com/kurtosis-tech/kurtosis-core/issues/508 IS BUILT!
	// We expect that the engine to be set up via the command's PreValidationAndRunFunc; this is the key where the resulting
	//  EngineServiceClient will be stored in the context.Context object passed to the validation function
	engineClientCtxKey string,
	isOptional bool,
	isGreedy bool,
) *args.ArgConfig {

	validate := getValidationFunc(argKey, engineClientCtxKey, isGreedy)

	return &args.ArgConfig{
		Key:                   argKey,
		IsOptional:            isOptional,
		DefaultValue:          "",
		IsGreedy:              isGreedy,
		ValidationFunc:        validate,
		ArgCompletionProvider: args.NewManualCompletionsProvider(getCompletions),
	}
}

// TODO we added this constructor for allowing 'service logs' command to disable the validation for consuming logs from removed or stopped enclaves
// TODO after https://github.com/kurtosis-tech/kurtosis/issues/879
func NewEnclaveIdentifierArgWithValidationDisabled(
	argKey string,
	isOptional bool,
	isGreedy bool,
) *args.ArgConfig {

	var noValidationFunc func(ctx context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error

	return &args.ArgConfig{
		Key:                   argKey,
		IsOptional:            isOptional,
		DefaultValue:          "",
		IsGreedy:              isGreedy,
		ValidationFunc:        noValidationFunc,
		ArgCompletionProvider: args.NewManualCompletionsProvider(getCompletions),
	}
}

// Make best-effort attempt to get enclave UUIDs
func getCompletions(ctx context.Context, flags *flags.ParsedFlags, previousArgs *args.ParsedArgs) ([]string, error) {
	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred connecting to the Kurtosis engine for retrieving the enclave UUIDs for tab completion",
		)
	}

	// TODO close the client inside the kurtosisCtx, but requires https://github.com/kurtosis-tech/kurtosis-engine-server/issues/89

	enclaves, err := kurtosisCtx.GetEnclaves(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred getting the enclaves retrieving for enclave identifier tab completion",
		)
	}

	enclaveNames := []string{}
	enclaveUuids := []string{}
	for enclaveName := range enclaves.GetEnclavesByName() {
		enclaveNames = append(enclaveNames, enclaveName)
	}
	for enclaveUuid := range enclaves.GetEnclavesByUuid() {
		enclaveUuids = append(enclaveUuids, enclaveUuid)
	}

	// we sort them individually
	sort.Strings(enclaveNames)
	sort.Strings(enclaveUuids)
	// we first add names and then uuids
	result := append(enclaveNames, enclaveUuids...)

	// NOTE: If this arg is greedy, we could actually examine the enclave UUIDs already stored for this arg in ParsedArgs
	//  and remove enclave UUIDs that are already set so that we don't repeat any

	return result, nil
}

// Create a validation function using the previously-created
func getValidationFunc(argKey string, _ string, isGreedy bool) func(context.Context, *flags.ParsedFlags, *args.ParsedArgs) error {
	return func(ctx context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error {
		kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred creating Kurtosis Context from local engine")
		}

		var enclaveIdentifiersToValidate []string
		if isGreedy {
			enclaveIds, err := args.GetGreedyArg(argKey)
			if err != nil {
				return stacktrace.Propagate(err, "Expected a value for greedy arg '%v' but didn't find one", argKey)
			}
			enclaveIdentifiersToValidate = enclaveIds
		} else {
			enclaveIdentifier, err := args.GetNonGreedyArg(argKey)
			if err != nil {
				return stacktrace.Propagate(err, "Expected a value for non-greedy arg '%v' but didn't find one", argKey)
			}
			enclaveIdentifiersToValidate = []string{enclaveIdentifier}
		}

		enclaves, err := kurtosisCtx.GetEnclaves(ctx)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting enclaves, which is necessary to check if the enclaves exist")
		}

		for _, enclaveIdentifier := range enclaveIdentifiersToValidate {
			if _, found := enclaves.GetEnclavesByUuid()[enclaveIdentifier]; found {
				continue
			}
			if _, found := enclaves.GetEnclavesByName()[enclaveIdentifier]; found {
				continue
			}
			if _, found := enclaves.GetEnclavesByShortenedUuid()[enclaveIdentifier]; found {
				continue
			}
			return stacktrace.NewError("No enclave found for identifier '%v'", enclaveIdentifier)
		}
		return nil
	}
}
