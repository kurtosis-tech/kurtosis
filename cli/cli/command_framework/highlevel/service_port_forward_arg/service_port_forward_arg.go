package service_port_forward_arg

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	isGreedy   = false
	isOptional = true
)

func NewServicePortForwardArg(
	servicePortForwardArgKey string,
	enclaveIdentifierArgKey string,
) *args.ArgConfig {

	validate := getValidationFunc(servicePortForwardArgKey, isGreedy)

	return &args.ArgConfig{
		Key:                   servicePortForwardArgKey,
		IsOptional:            isOptional,
		DefaultValue:          "",
		IsGreedy:              isGreedy,
		ArgCompletionProvider: args.NewManualCompletionsProvider(getServicePortCompletions(enclaveIdentifierArgKey)),
		ValidationFunc:        validate,
	}
}

func getValidationFunc(servicePortForwardArgKey string, isGreedy bool) func(context.Context, *flags.ParsedFlags, *args.ParsedArgs) error {
	return func(ctx context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error {
		return nil
	}
}

func getServicePortCompletions(enclaveIdentifierArgKey string) func(ctx context.Context, flags *flags.ParsedFlags, previousArgs *args.ParsedArgs) ([]string, error) {
	return func(ctx context.Context, flags *flags.ParsedFlags, previousArgs *args.ParsedArgs) ([]string, error) {
		enclaveId, err := previousArgs.GetNonGreedyArg(enclaveIdentifierArgKey)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting the enclave identifier using key '%v'", enclaveIdentifierArgKey)
		}

		kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
		if err != nil {
			return nil, stacktrace.Propagate(
				err,
				"An error occurred connecting to the Kurtosis engine for retrieving the enclave UUIDs and names for tab completion",
			)
		}

		// TODO close the client inside the kurtosisCtx, but requires https://github.com/kurtosis-tech/kurtosis-engine-server/issues/89
		// ^ comment brought over from enclave_id_arg.go

		enclaveContext, err := kurtosisCtx.GetEnclaveContext(ctx, enclaveId)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting context for enclave '%v'", enclaveId)
		}

		services, err := enclaveContext.GetServices()
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to lookup services in enclave %v", enclaveId)
		}

		servicePortCompletions := []string{}

		// TODO(omar): runs quite slowly, but other tab-completes do too
		for serviceName := range services {
			serviceId := string(serviceName)
			serviceContext, err := enclaveContext.GetServiceContext(serviceId)
			if err != nil {
				return nil, stacktrace.Propagate(err, "Failed to get service context for service '%v' in enclave '%v", serviceId, enclaveId)
			}

			servicePortCompletions = append(servicePortCompletions, serviceId)
			for portId := range serviceContext.GetPrivatePorts() {
				servicePortCompletions = append(servicePortCompletions, serviceId+"."+portId)
			}
		}

		return servicePortCompletions, nil
	}
}
