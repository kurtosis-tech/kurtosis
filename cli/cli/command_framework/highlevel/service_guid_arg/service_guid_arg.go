package service_guid_arg

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/stacktrace"
	"sort"
)

const (
	enclaveIdArgKey = "enclave-id"
)

func NewServiceGUIDArg(
	argKey string,
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

func getCompletions(ctx context.Context, flags *flags.ParsedFlags, previousArgs *args.ParsedArgs) ([]string, error) {
	enclaveIdStr, err := previousArgs.GetNonGreedyArg(enclaveIdArgKey)
	enclaveId := enclaves.EnclaveID(enclaveIdStr)

	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred connecting to the Kurtosis engine for retrieving the enclave IDs for tab completion",
		)
	}
	_, serviceGUIDs, err := kurtosisCtx.GetServices(ctx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred getting the services retrieving for enclave ID tab completion",
		)
	}

	result := []string{}
	for serviceGUID := range serviceGUIDs {
		result = append(result, string(serviceGUID))
	}
	sort.Strings(result)

	return result, nil
}

// Create a validation function using the previously-created
func getValidationFunc(argKey string, engineClientCtxKey string, isGreedy bool) func(context.Context, *flags.ParsedFlags, *args.ParsedArgs) error {
	return func(ctx context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error {
		//uncastedEngineClient := ctx.Value(engineClientCtxKey)
		//if uncastedEngineClient == nil {
		//	return stacktrace.NewError("Expected an engine client to have been stored in the context under key '%v', but none was found; this is a bug in Kurtosis!", engineClientCtxKey)
		//}
		//engineClient, ok := uncastedEngineClient.(kurtosis_engine_rpc_api_bindings.EngineServiceClient)
		//if !ok {
		//	return stacktrace.NewError("Found an object that should be the engine client stored in the context under key '%v', but this object wasn't of the correct engine client type", engineClientCtxKey)
		//}
		//
		//var enclaveIdsToValidate []string
		//if isGreedy {
		//	enclaveIds, err := args.GetGreedyArg(argKey)
		//	if err != nil {
		//		return stacktrace.Propagate(err, "Expected a value for greedy arg '%v' but didn't find one", argKey)
		//	}
		//	enclaveIdsToValidate = enclaveIds
		//} else {
		//	enclaveId, err := args.GetNonGreedyArg(argKey)
		//	if err != nil {
		//		return stacktrace.Propagate(err, "Expected a value for non-greedy arg '%v' but didn't find one", argKey)
		//	}
		//	enclaveIdsToValidate = []string{enclaveId}
		//}
		//
		//getEnclavesResp, err := engineClient.GetEnclaves(ctx, &emptypb.Empty{})
		//if err != nil {
		//	return stacktrace.Propagate(err, "An error occurred getting enclaves, which is necessary to check if the enclaves exist")
		//}
		//
		//for _, enclaveId := range enclaveIdsToValidate {
		//	if _, found := getEnclavesResp.EnclaveInfo[enclaveId]; !found {
		//		return stacktrace.NewError("No enclave found with ID '%v'", enclaveId)
		//	}
		//}
		return nil
	}
}
