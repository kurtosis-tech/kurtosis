package rm

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/highlevel/enclave_id_arg"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	enclaveIdArgKey        = "enclave-id"
	isEnclaveIdArgOptional = false
	isEnclaveIdArgGreedy   = false

	serviceGuidArgKey = "service-guid"

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey  = "engine-client"

)

var ServiceRmCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:              command_str_consts.ServiceRmCmdStr,
	ShortDescription:        "Removes a service from an enclave",
	LongDescription:         "Removes the service with the given ID from the given enclave",
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:  engineClientCtxKey,
	Args: []*args.ArgConfig{
		enclave_id_arg.NewEnclaveIDArg(
			enclaveIdArgKey,
			engineClientCtxKey,
			isEnclaveIdArgOptional,
			isEnclaveIdArgGreedy,
		),
		{
			Key: serviceGuidArgKey,
		},
	},
	Flags: nil,
	RunFunc: run,
}

func run(
	ctx context.Context,
	kurtosisBackend backend_interface.KurtosisBackend,
	engineClient kurtosis_engine_rpc_api_bindings.EngineServiceClient,
	flags *flags.ParsedFlags,
	args *args.ParsedArgs,
) error {
	enclaveIdStr, err := args.GetNonGreedyArg(enclaveIdArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the enclave ID value using key '%v'", enclaveIdArgKey)
	}
	enclaveId := enclave.EnclaveID(enclaveIdStr)

	serviceGuidStr, err := args.GetNonGreedyArg(serviceGuidArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the service GUID value using key '%v'", serviceGuidArgKey)
	}
	serviceGuid := service.ServiceGUID(serviceGuidStr)

	userServiceFilters := &service.ServiceFilters{
		GUIDs: map[service.ServiceGUID]bool{
			serviceGuid: true,
		},
		EnclaveIDs: map[enclave.EnclaveID]bool{
			enclaveId: true,
		},
	}

	_, erroredUserServiceGuids, err := kurtosisBackend.DestroyUserServices(ctx, userServiceFilters)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred destroying user services using filters")
	}

	if len(erroredUserServiceGuids) > 0 {
		err, found := erroredUserServiceGuids[serviceGuid]
		if !found {
			return stacktrace.NewError("Expected to find an error for user service with GUID '%v' on user service error map '%+v' but was not found; it should never happens, it is a bug in Kurtosis", serviceGuid, erroredUserServiceGuids)
		}
		return stacktrace.Propagate(err, "An error occurred destroying user service with GUID '%v'", serviceGuid)
	}

	return nil
}
