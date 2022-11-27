package service_guid_arg

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/stacktrace"
	"sort"
	"strings"
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
		Key:                   argKey,
		IsOptional:            isOptional,
		DefaultValue:          "",
		IsGreedy:              isGreedy,
		ArgCompletionProvider: args.NewManualCompletionsProvider(getCompletions),
		ValidationFunc:        validate,
	}
}

func getServiceGUIDsForEnclave(ctx context.Context, enclaveID enclaves.EnclaveID) (map[services.ServiceGUID]bool, error) {
	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred connecting to the Kurtosis engine for retrieving the enclave IDs for tab completion",
		)
	}
	_, serviceGUIDs, err := kurtosisCtx.GetServices(ctx, enclaveID)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred getting the services retrieving for enclave ID tab completion",
		)
	}
	return serviceGUIDs, nil
}

func getCompletions(ctx context.Context, flags *flags.ParsedFlags, previousArgs *args.ParsedArgs) ([]string, error) {
	enclaveIdStr, err := previousArgs.GetNonGreedyArg(enclaveIdArgKey)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the enclave ID using key '%v'", enclaveIdArgKey)
	}

	enclaveID := enclaves.EnclaveID(enclaveIdStr)
	serviceGUIDs, err := getServiceGUIDsForEnclave(ctx, enclaveID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the services retrieving for enclave ID tab completion")
	}

	result := []string{}
	for serviceGUID := range serviceGUIDs {
		result = append(result, string(serviceGUID))
	}
	sort.Strings(result)

	return result, nil
}

// Create a validation function using the previously-created
func getValidationFunc(
	argKey string,
	engineClientCtxKey string,
	isGreedy bool,
) func(context.Context, *flags.ParsedFlags, *args.ParsedArgs) error {
	return func(ctx context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error {
		var serviceGUIDsToValidate []string
		if isGreedy {
			serviceGUID, err := args.GetGreedyArg(argKey)
			if err != nil {
				return stacktrace.Propagate(err, "Expected a value for greedy arg '%v' but didn't find one", argKey)
			}
			serviceGUIDsToValidate = serviceGUID
		} else {
			serviceGUID, err := args.GetNonGreedyArg(argKey)
			if err != nil {
				return stacktrace.Propagate(err, "Expected a value for non-greedy arg '%v' but didn't find one", argKey)
			}
			serviceGUIDsToValidate = []string{serviceGUID}
		}

		serviceGUIDs, err := getCompletions(ctx, flags, args)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while getting the services for the enclave")
		}

		var errorServiceGUIDs []string
		for _, serviceGUIDtoValidate := range serviceGUIDsToValidate {
			var contains = false
			for _, serviceGUID := range serviceGUIDs {
				if found := serviceGUIDtoValidate == serviceGUID; found {
					contains = true
					break
				}
			}
			if !contains {
				errorServiceGUIDs = append(errorServiceGUIDs, serviceGUIDtoValidate)
			}
		}

		if len(errorServiceGUIDs) > 0 {
			return stacktrace.NewError("One or more service GUIDs do not exist in the enclave: '%s'", strings.Join(errorServiceGUIDs, ", "))
		}
		return nil
	}
}
