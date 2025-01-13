package exec

import (
	"context"
	"fmt"

	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/enclave_id_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/service_identifier_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/out"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/metrics_client"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	enclaveIdentifierArgKey = "enclave"
	isEnclaveIdArgOptional  = false
	isEnclaveIdArgGreedy    = false

	serviceIdentifierArgKey  = "service"
	isServiceGuidArgOptional = false
	isServiceGuidArgGreedy   = false

	execCommandArgKey        = "command"
	isExecCommandArgOptional = false
	isExecCommandArgGreedy   = false

	containerUserKey      = "user"
	containerUserShortKey = "u"
	containerUserDefault  = "root"

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"

	binShCommand     = "sh"
	binShCommandFlag = "-c"
)

var ServiceShellCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.ServiceExecCmdStr,
	ShortDescription:          "Executes a command in a service",
	LongDescription:           "Execute a command in a service. Note if the command being run is multiple words you should wrap it in quotes",
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:    engineClientCtxKey,
	Flags: []*flags.FlagConfig{{
		Key:       containerUserKey,
		Usage:     "optional service container user for command",
		Shorthand: containerUserShortKey,
		Type:      flags.FlagType_String,
		Default:   containerUserDefault,
	}},
	Args: []*args.ArgConfig{
		enclave_id_arg.NewEnclaveIdentifierArg(
			enclaveIdentifierArgKey,
			engineClientCtxKey,
			isEnclaveIdArgOptional,
			isEnclaveIdArgGreedy,
		),
		service_identifier_arg.NewServiceIdentifierArg(
			serviceIdentifierArgKey,
			enclaveIdentifierArgKey,
			isServiceGuidArgOptional,
			isServiceGuidArgGreedy,
		),
		{
			Key:        execCommandArgKey,
			IsOptional: isExecCommandArgOptional,
			IsGreedy:   isExecCommandArgGreedy,
		},
	},
	RunFunc: run,
}

func run(
	ctx context.Context,
	kurtosisBackend backend_interface.KurtosisBackend,
	_ kurtosis_engine_rpc_api_bindings.EngineServiceClient,
	_ metrics_client.MetricsClient,
	flags *flags.ParsedFlags,
	args *args.ParsedArgs,
) error {
	enclaveIdentifier, err := args.GetNonGreedyArg(enclaveIdentifierArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the enclave identifier using arg key '%v'", enclaveIdentifierArgKey)
	}

	serviceIdentifier, err := args.GetNonGreedyArg(serviceIdentifierArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the service identifier using arg key '%v'", serviceIdentifierArgKey)
	}

	execCommandToRun, err := args.GetNonGreedyArg(execCommandArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the exec command using arg key '%v'", execCommandArgKey)
	}

	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred connecting to the local Kurtosis engine")
	}

	enclaveCtx, err := kurtosisCtx.GetEnclaveContext(ctx, enclaveIdentifier)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while getting enclave context for enclave with identifier '%v' exists", enclaveIdentifier)
	}

	enclaveUuid := enclave.EnclaveUUID(enclaveCtx.GetEnclaveUuid())

	serviceCtx, err := enclaveCtx.GetServiceContext(serviceIdentifier)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while getting service context for service with identifier '%v'", serviceIdentifier)
	}
	serviceUuid := service.ServiceUUID(serviceCtx.GetServiceUUID())

	execContainerUser, err := flags.GetString(containerUserKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the exec container user flag '%v'", containerUserKey)
	}
	// convert "root" to "" as it is implied if empty. This simplifies the backend
	// impl usages that do not support --user
	if execContainerUser == containerUserDefault {
		execContainerUser = ""
	}

	results, resultErrors, err := kurtosisBackend.RunUserServiceExecCommandsAsUser(ctx, enclaveUuid, execContainerUser, map[service.ServiceUUID][]string{
		serviceUuid: {
			binShCommand,
			binShCommandFlag,
			execCommandToRun,
		},
	})

	if err != nil {
		return stacktrace.Propagate(err, "An error occurred executing command '%v' in user service with UUID '%v' in enclave '%v'",
			execCommandToRun, serviceUuid, enclaveIdentifier)
	}

	if err, found := resultErrors[serviceUuid]; found {
		return stacktrace.Propagate(err, "An error occurred executing command '%v' in user service with UUID '%v' in enclave '%v'",
			execCommandToRun, serviceUuid, enclaveIdentifier)
	}

	successResult, found := results[serviceUuid]
	if !found {
		return stacktrace.NewError("The status of the command execution for '%s' in user service with UUID '%s' in enclave '%s' is unknown."+
			" It wasn't returned neither as a success nor a failure. This is a bug in Kurtosis.", execCommandToRun, serviceUuid, enclaveIdentifier)
	}

	if successResult.GetExitCode() != 0 {
		return stacktrace.NewError("The command was successfully executed but returned a non-zero exit code: '%d'. Output was:\n%v", successResult.GetExitCode(), successResult.GetOutput())
	}
	out.PrintOutLn(fmt.Sprintf("The command was successfully executed and returned '%d'. Output was:\n%v", successResult.GetExitCode(), successResult.GetOutput()))
	return nil
}
