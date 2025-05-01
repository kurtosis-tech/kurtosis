package enclave_id_arg

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/stacktrace"
	"sort"
	"strings"
)

const (
	validShortenedUuidOrNameMatches = 1
	uuidDelimiter                   = ", "
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

// This constructor is for allowing 'service logs' command to disable the validation for consuming logs from removed or stopped enclaves
func NewHistoricalEnclaveIdentifiersArgWithValidationDisabled(
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

// Make best-effort attempt to get enclave names
func getCompletions(ctx context.Context, flags *flags.ParsedFlags, previousArgs *args.ParsedArgs) ([]string, error) {
	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred connecting to the Kurtosis engine for retrieving the enclave UUIDs and names for tab completion",
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
	for enclaveName := range enclaves.GetEnclavesByName() {
		enclaveNames = append(enclaveNames, enclaveName)
	}

	// we sort them individually
	sort.Strings(enclaveNames)
	return enclaveNames, nil
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

		enclavesByShortenedUuid := enclaves.GetEnclavesByShortenedUuid()
		enclavesByUuid := enclaves.GetEnclavesByUuid()
		enclavesByName := enclaves.GetEnclavesByName()

		for _, enclaveIdentifier := range enclaveIdentifiersToValidate {
			if _, found := enclavesByUuid[enclaveIdentifier]; found {
				continue
			}
			if matches, found := enclavesByShortenedUuid[enclaveIdentifier]; found {
				if len(matches) > validShortenedUuidOrNameMatches {
					return stacktrace.NewError("Found multiple matching uuids '%v' for shortened uuid '%v' which is ambiguous", enclaveInfosToUuidsStr(matches), enclaveIdentifier)
				}
				continue
			}
			if matches, found := enclavesByName[enclaveIdentifier]; found {
				if len(matches) > validShortenedUuidOrNameMatches {
					return stacktrace.NewError("Found multiple matching uuids '%v' for name '%v' which is ambiguous", enclaveInfosToUuidsStr(matches), enclaveIdentifier)
				}
				continue
			}
			return stacktrace.NewError("No enclave found for identifier '%v'", enclaveIdentifier)
		}
		return nil
	}
}

func enclaveInfosToUuidsStr(infos []*kurtosis_engine_rpc_api_bindings.EnclaveInfo) string {
	var uuids []string
	for _, info := range infos {
		uuids = append(uuids, info.EnclaveUuid)
	}
	return strings.Join(uuids, uuidDelimiter)
}
