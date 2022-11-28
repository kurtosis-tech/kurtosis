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
	isOptional bool,
	isGreedy bool,
) *args.ArgConfig {
	validate := getValidationFunc(argKey, isGreedy)

	return &args.ArgConfig{
		Key:                   argKey,
		IsOptional:            isOptional,
		DefaultValue:          "",
		IsGreedy:              isGreedy,
		ArgCompletionProvider: args.NewManualCompletionsProvider(getOrderedEnclaveServiceGuids),
		ValidationFunc:        validate,
	}
}

func getServiceGuidsForEnclave(ctx context.Context, enclaveID enclaves.EnclaveID) (map[services.ServiceGUID]bool, error) {
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

func getOrderedEnclaveServiceGuids(ctx context.Context, flags *flags.ParsedFlags, previousArgs *args.ParsedArgs) ([]string, error) {
	enclaveIdStr, err := previousArgs.GetNonGreedyArg(enclaveIdArgKey)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the enclave ID using key '%v'", enclaveIdArgKey)
	}

	enclaveID := enclaves.EnclaveID(enclaveIdStr)
	serviceGuids, err := getServiceGuidsForEnclave(ctx, enclaveID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the services retrieving for enclave ID tab completion")
	}

	result := []string{}
	for serviceGuid := range serviceGuids {
		result = append(result, string(serviceGuid))
	}
	sort.Strings(result)

	return result, nil
}

func getValidationFunc(
	argKey string,
	isGreedy bool,
) func(context.Context, *flags.ParsedFlags, *args.ParsedArgs) error {
	return func(ctx context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error {
		var serviceGuidsToValidate []string
		if isGreedy {
			serviceGuid, err := args.GetGreedyArg(argKey)
			if err != nil {
				return stacktrace.Propagate(err, "Expected a value for greedy arg '%v' but didn't find one", argKey)
			}
			serviceGuidsToValidate = serviceGuid
		} else {
			serviceGuid, err := args.GetNonGreedyArg(argKey)
			if err != nil {
				return stacktrace.Propagate(err, "Expected a value for non-greedy arg '%v' but didn't find one", argKey)
			}
			serviceGuidsToValidate = []string{serviceGuid}
		}

		serviceGuids, err := getOrderedEnclaveServiceGuids(ctx, flags, args)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while getting the services for the enclave")
		}

		var errorServiceGuids []string
		for _, serviceGuidToValidate := range serviceGuidsToValidate {
			var contains = false
			for _, serviceGuid := range serviceGuids {
				if found := serviceGuidToValidate == serviceGuid; found {
					contains = true
					break
				}
			}
			if !contains {
				errorServiceGuids = append(errorServiceGuids, serviceGuidToValidate)
			}
		}

		if len(errorServiceGuids) > 0 {
			return stacktrace.NewError("One or more service GUIDs do not exist in the enclave: '%s'", strings.Join(errorServiceGuids, ", "))
		}
		return nil
	}
}
