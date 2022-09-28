package enclave_id_arg

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis-sdk/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-sdk/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/stacktrace"
	"google.golang.org/protobuf/types/known/emptypb"
	"sort"
)

// Prebuilt enclave ID arg which has tab-completion and validation ready out-of-the-box
func NewEnclaveIDArg(
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
		Key:             argKey,
		IsOptional:      isOptional,
		DefaultValue:    "",
		IsGreedy:        isGreedy,
		CompletionsFunc: getCompletions,
		ValidationFunc:  validate,
	}
}

// Make best-effort attempt to get enclave IDs
func getCompletions(ctx context.Context, flags *flags.ParsedFlags, previousArgs *args.ParsedArgs) ([]string, error) {
	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred connecting to the Kurtosis engine for retrieving the enclave IDs for tab completion",
		)
	}

	// TODO close the client inside the kurtosisCtx, but requires https://github.com/kurtosis-tech/kurtosis-engine-server/issues/89

	enclaves, err := kurtosisCtx.GetEnclaves(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred getting the enclaves retrieving for enclave ID tab completion",
		)
	}

	result := []string{}
	for enclaveId := range enclaves {
		result = append(result, string(enclaveId))
	}
	sort.Strings(result)

	// NOTE: If this arg is greedy, we could actually examine the enclave IDs already stored for this arg in ParsedArgs
	//  and remove enclave IDs that are already set so that we don't repeat any

	return result, nil
}

// Create a validation function using the previously-created
func getValidationFunc(argKey string, engineClientCtxKey string, isGreedy bool) func(context.Context, *flags.ParsedFlags, *args.ParsedArgs) error {
	return func(ctx context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error {
		uncastedEngineClient := ctx.Value(engineClientCtxKey)
		if uncastedEngineClient == nil {
			return stacktrace.NewError("Expected an engine client to have been stored in the context under key '%v', but none was found; this is a bug in Kurtosis!", engineClientCtxKey)
		}
		engineClient, ok := uncastedEngineClient.(kurtosis_engine_rpc_api_bindings.EngineServiceClient)
		if !ok {
			return stacktrace.NewError("Found an object that should be the engine client stored in the context under key '%v', but this object wasn't of the correct engine client type", engineClientCtxKey)
		}

		var enclaveIdsToValidate []string
		if isGreedy {
			enclaveIds, err := args.GetGreedyArg(argKey)
			if err != nil {
				return stacktrace.Propagate(err, "Expected a value for greedy arg '%v' but didn't find one", argKey)
			}
			enclaveIdsToValidate = enclaveIds
		} else {
			enclaveId, err := args.GetNonGreedyArg(argKey)
			if err != nil {
				return stacktrace.Propagate(err, "Expected a value for non-greedy arg '%v' but didn't find one", argKey)
			}
			enclaveIdsToValidate = []string{enclaveId}
		}

		getEnclavesResp, err := engineClient.GetEnclaves(ctx, &emptypb.Empty{})
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting enclaves, which is necessary to check if the enclaves exist")
		}

		for _, enclaveId := range enclaveIdsToValidate {
			if _, found := getEnclavesResp.EnclaveInfo[enclaveId]; !found {
				return stacktrace.NewError("No enclave found with ID '%v'", enclaveId)
			}
		}
		return nil
	}
}


