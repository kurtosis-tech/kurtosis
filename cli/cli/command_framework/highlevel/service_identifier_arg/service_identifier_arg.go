package service_identifier_arg

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
	"sort"
	"strings"
)

const (
	enclaveIdentifierArgKey = "enclave-identifier"
)

func NewServiceIdentifierArg(
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
		ArgCompletionProvider: args.NewManualCompletionsProvider(getOrderedServiceIdentifiersWithoutShortenedUuids),
		ValidationFunc:        validate,
	}
}

// TODO we added this constructor for allowing 'service logs' command to disable the validation for consuming logs from removed or stopped enclaves
// TODO after https://github.com/kurtosis-tech/kurtosis/issues/879 is done
func NewServiceUUIDArgWithValidationDisabled(
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
		ArgCompletionProvider: args.NewManualCompletionsProvider(getOrderedServiceIdentifiersWithoutShortenedUuids),
		ValidationFunc:        noValidationFunc,
	}
}

func getServiceUuidsAndNamesForEnclave(ctx context.Context, enclaveIdentifier string) (map[services.ServiceUUID]bool, map[services.ServiceName]bool, error) {
	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return nil, nil, stacktrace.Propagate(
			err,
			"An error occurred connecting to the Kurtosis engine for retrieving the service UUIDs & names for tab completion",
		)
	}

	enclaveContext, err := kurtosisCtx.GetEnclaveContext(ctx, enclaveIdentifier)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting enclave context")
	}

	serviceNames, err := enclaveContext.GetServices()
	if err != nil {
		return nil, nil, stacktrace.Propagate(
			err,
			"An error occurred getting the services retrieving for enclave ID tab completion",
		)
	}

	serviceUuids := make(map[services.ServiceUUID]bool, len(serviceNames))
	serviceNamesSet := make(map[services.ServiceName]bool, len(serviceNames))
	for serviceName, serviceUuid := range serviceNames {
		if _, ok := serviceUuids[serviceUuid]; !ok {
			serviceUuids[serviceUuid] = true
		}
		if _, ok := serviceNamesSet[serviceName]; !ok {
			serviceNamesSet[serviceName] = true
		}
	}

	return serviceUuids, serviceNamesSet, nil
}

func getOrderedServiceIdentifiersWithoutShortenedUuids(ctx context.Context, flags *flags.ParsedFlags, previousArgs *args.ParsedArgs) ([]string, error) {
	enclaveIdentifier, err := previousArgs.GetNonGreedyArg(enclaveIdentifierArgKey)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the enclave identifier using key '%v'", enclaveIdentifierArgKey)
	}

	serviceUuids, serviceNames, err := getServiceUuidsAndNamesForEnclave(ctx, enclaveIdentifier)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the services retrieving for enclave identifier tab completion")
	}

	serviceNamesList := []string{}
	for serviceName := range serviceNames {
		serviceNamesList = append(serviceNamesList, string(serviceName))
	}
	sort.Strings(serviceNamesList)

	serviceUuidsList := []string{}
	for serviceUuid := range serviceUuids {
		serviceUuidsList = append(serviceUuidsList, string(serviceUuid))
	}
	sort.Strings(serviceUuidsList)

	result := []string{}
	// result is the concatenation of the sorted names list followed by sorted uuids
	// we want names to show up first in completion
	result = append(result, serviceNamesList...)
	result = append(result, serviceUuidsList...)

	return result, nil
}

func getServiceIdentifiersForValidation(ctx context.Context, _ *flags.ParsedFlags, previousArgs *args.ParsedArgs) (map[string]bool, error) {
	enclaveIdentifier, err := previousArgs.GetNonGreedyArg(enclaveIdentifierArgKey)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the enclave identifier using key '%v'", enclaveIdentifierArgKey)
	}

	serviceUuids, serviceNames, err := getServiceUuidsAndNamesForEnclave(ctx, enclaveIdentifier)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the services retrieving for enclave identifier tab completion")
	}

	serviceIdentifiersSet := make(map[string]bool)

	for serviceName := range serviceNames {
		serviceIdentifiersSet[string(serviceName)] = true
	}
	for serviceUuid := range serviceUuids {
		serviceIdentifiersSet[string(serviceUuid)] = true
		serviceIdentifiersSet[uuid_generator.ShortenedUUIDString(string(serviceUuid))] = true
	}
	return serviceIdentifiersSet, nil
}

func getValidationFunc(
	argKey string,
	isGreedy bool,
) func(context.Context, *flags.ParsedFlags, *args.ParsedArgs) error {
	return func(ctx context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error {
		var serviceIdentifiersToValidate []string
		if isGreedy {
			serviceUuid, err := args.GetGreedyArg(argKey)
			if err != nil {
				return stacktrace.Propagate(err, "Expected a value for greedy arg '%v' but didn't find one", argKey)
			}
			serviceIdentifiersToValidate = serviceUuid
		} else {
			serviceUuid, err := args.GetNonGreedyArg(argKey)
			if err != nil {
				return stacktrace.Propagate(err, "Expected a value for non-greedy arg '%v' but didn't find one", argKey)
			}
			serviceIdentifiersToValidate = []string{serviceUuid}
		}

		knownServiceIdentifiers, err := getServiceIdentifiersForValidation(ctx, flags, args)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while getting the services for the enclave")
		}

		var erroredServiceIdentifiers []string
		for _, serviceIdentifierToValidate := range serviceIdentifiersToValidate {
			if _, found := knownServiceIdentifiers[serviceIdentifierToValidate]; !found {
				erroredServiceIdentifiers = append(erroredServiceIdentifiers, serviceIdentifierToValidate)
			}
		}

		if len(erroredServiceIdentifiers) > 0 {
			return stacktrace.NewError("One or more service identifiers do not exist in the enclave: '%s'", strings.Join(erroredServiceIdentifiers, ", "))
		}
		return nil
	}
}
