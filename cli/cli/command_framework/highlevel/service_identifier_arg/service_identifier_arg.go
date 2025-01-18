package service_identifier_arg

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/v2/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/api/golang/v2/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
	"sort"
	"strings"
)

const (
	validShortenedUuidOrNameMatches = 1
	uuidDelimiter                   = ", "
)

func NewServiceIdentifierArg(
	serviceIdentifierArgKey string,
	enclaveIdentifierArgKey string,
	isOptional bool,
	isGreedy bool,
) *args.ArgConfig {

	validate := getValidationFunc(serviceIdentifierArgKey, isGreedy, enclaveIdentifierArgKey)

	return &args.ArgConfig{
		Key:                   serviceIdentifierArgKey,
		IsOptional:            isOptional,
		DefaultValue:          "",
		IsGreedy:              isGreedy,
		ArgCompletionProvider: args.NewManualCompletionsProvider(getCompletionsOfActiveServices(enclaveIdentifierArgKey)),
		ValidationFunc:        validate,
	}
}

// This constructor is for allowing 'service logs' command to disable the validation for consuming logs from removed or stopped enclaves
func NewHistoricalServiceIdentifierArgWithValidationDisabled(
	serviceIdentifierArgKey string,
	enclaveIdentifierArgKey string,
	isOptional bool,
	isGreedy bool,
) *args.ArgConfig {

	var noValidationFunc func(ctx context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error

	return &args.ArgConfig{
		Key:                   serviceIdentifierArgKey,
		IsOptional:            isOptional,
		DefaultValue:          []string{},
		IsGreedy:              isGreedy,
		ArgCompletionProvider: args.NewManualCompletionsProvider(getCompletionsForExistingAndHistoricalServices(enclaveIdentifierArgKey)),
		ValidationFunc:        noValidationFunc,
	}
}

func getServiceUuidsAndNamesForEnclave(ctx context.Context, enclaveIdentifier string) (map[services.ServiceUUID]bool, map[services.ServiceName]services.ServiceUUID, error) {
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
	serviceNamesToUuid := make(map[services.ServiceName]services.ServiceUUID)
	for serviceName, serviceUuid := range serviceNames {
		serviceUuids[serviceUuid] = true
		serviceNamesToUuid[serviceName] = serviceUuid
	}

	return serviceUuids, serviceNamesToUuid, nil
}

func getCompletionsForExistingAndHistoricalServices(enclaveIdentifierArgKey string) func(ctx context.Context, _ *flags.ParsedFlags, previousArgs *args.ParsedArgs) ([]string, error) {
	return func(ctx context.Context, _ *flags.ParsedFlags, previousArgs *args.ParsedArgs) ([]string, error) {
		enclaveIdentifier, err := previousArgs.GetNonGreedyArg(enclaveIdentifierArgKey)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting the enclave identifier using key '%v'", enclaveIdentifierArgKey)
		}

		kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
		if err != nil {
			return nil, stacktrace.Propagate(
				err,
				"An error occurred connecting to the Kurtosis engine for retrieving the service UUIDs & names for tab completion",
			)
		}

		enclaveContext, err := kurtosisCtx.GetEnclaveContext(ctx, enclaveIdentifier)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred while fetching enclave for identifier '%v'", enclaveIdentifier)
		}

		serviceIdentifiers, err := enclaveContext.GetExistingAndHistoricalServiceIdentifiers(ctx)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred while fetching services for enclave '%v'", enclaveContext.GetEnclaveName())
		}

		return serviceIdentifiers.GetOrderedListOfNames(), nil
	}
}

func getCompletionsOfActiveServices(enclaveIdentifierArgKey string) func(ctx context.Context, flags *flags.ParsedFlags, previousArgs *args.ParsedArgs) ([]string, error) {
	return func(ctx context.Context, _ *flags.ParsedFlags, previousArgs *args.ParsedArgs) ([]string, error) {
		enclaveIdentifier, err := previousArgs.GetNonGreedyArg(enclaveIdentifierArgKey)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting the enclave identifier using key '%v'", enclaveIdentifierArgKey)
		}

		_, serviceNames, err := getServiceUuidsAndNamesForEnclave(ctx, enclaveIdentifier)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting the services retrieving for enclave identifier tab completion")
		}

		serviceNamesList := []string{}
		for serviceName := range serviceNames {
			serviceNamesList = append(serviceNamesList, string(serviceName))
		}
		sort.Strings(serviceNamesList)

		return serviceNamesList, nil
	}
}

func getServiceIdentifiersForValidation(enclaveIdentifierArgKey string) func(ctx context.Context, _ *flags.ParsedFlags, previousArgs *args.ParsedArgs) (map[services.ServiceUUID]bool, map[services.ServiceName]services.ServiceUUID, map[string][]services.ServiceUUID, error) {
	return func(ctx context.Context, _ *flags.ParsedFlags, previousArgs *args.ParsedArgs) (map[services.ServiceUUID]bool, map[services.ServiceName]services.ServiceUUID, map[string][]services.ServiceUUID, error) {
		enclaveIdentifier, err := previousArgs.GetNonGreedyArg(enclaveIdentifierArgKey)
		if err != nil {
			return nil, nil, nil, stacktrace.Propagate(err, "An error occurred getting the enclave identifier using key '%v'", enclaveIdentifierArgKey)
		}

		serviceUuids, serviceNames, err := getServiceUuidsAndNamesForEnclave(ctx, enclaveIdentifier)
		if err != nil {
			return nil, nil, nil, stacktrace.Propagate(err, "An error occurred getting the services retrieving for enclave identifier tab completion")
		}

		shortenedUuidsToUuids := make(map[string][]services.ServiceUUID)
		for serviceUuid := range serviceUuids {
			shortenedUuid := uuid_generator.ShortenedUUIDString(string(serviceUuid))
			shortenedUuidsToUuids[shortenedUuid] = append(shortenedUuidsToUuids[shortenedUuid], serviceUuid)
		}
		return serviceUuids, serviceNames, shortenedUuidsToUuids, nil
	}
}

func getValidationFunc(
	argKey string,
	isGreedy bool,
	enclaveIdentifierArgKey string,
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

		serviceUuids, serviceNames, shortenedUuidsToUuids, err := getServiceIdentifiersForValidation(enclaveIdentifierArgKey)(ctx, flags, args)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while getting the services for the enclave")
		}

		for _, serviceIdentifierToValidate := range serviceIdentifiersToValidate {
			maybeServiceUuid := services.ServiceUUID(serviceIdentifierToValidate)
			maybeServiceName := services.ServiceName(serviceIdentifierToValidate)
			maybeShortenedUuid := serviceIdentifierToValidate

			if _, found := serviceUuids[maybeServiceUuid]; found {
				continue
			}

			if matches, found := shortenedUuidsToUuids[maybeShortenedUuid]; found {
				if len(matches) > validShortenedUuidOrNameMatches {
					return stacktrace.NewError("Found multiple matching uuids '%v' for shortened uuid '%v' which is ambiguous", serviceUuidsToString(matches), maybeShortenedUuid)
				}
				continue
			}
			if _, found := serviceNames[maybeServiceName]; found {
				continue
			}

			return stacktrace.NewError("No service found for identifier '%v'", serviceIdentifierToValidate)
		}
		return nil
	}
}

func serviceUuidsToString(serviceUuids []services.ServiceUUID) string {
	var uuids []string
	for _, uuid := range serviceUuids {
		uuids = append(uuids, string(uuid))
	}
	return strings.Join(uuids, uuidDelimiter)
}
