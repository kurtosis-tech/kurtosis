package rm

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/highlevel/enclave_id_arg"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/enclave_liveness_validator"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/services"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/stacktrace"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	enclaveIdArgKey        = "enclave-id"
	isEnclaveIdArgOptional = false
	isEnclaveIdArgGreedy   = false

	serviceIdArgKey = "service-id"

	timeoutFlagKey = "timeout"

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey  = "engine-client"

	defaultContainerStopTimeoutSeconds = uint32(0)
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
			Key: serviceIdArgKey,
		},
	},
	Flags: []*flags.FlagConfig{
		{
			Key:       timeoutFlagKey,
			Usage:     "Number of seconds to wait for the service to gracefully stop before sending SIGKILL",
			Type:      flags.FlagType_Uint32,
			Default:   fmt.Sprintf("%v", defaultContainerStopTimeoutSeconds),
		},
	},
	RunFunc: run,
}

func run(
	ctx context.Context,
	kurtosisBackend backend_interface.KurtosisBackend,
	engineClient kurtosis_engine_rpc_api_bindings.EngineServiceClient,
	flags *flags.ParsedFlags,
	args *args.ParsedArgs,
) error {
	enclaveId, err := args.GetNonGreedyArg(enclaveIdArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the enclave ID value using key '%v'", enclaveIdArgKey)
	}

	serviceId, err := args.GetNonGreedyArg(serviceIdArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the service ID value using key '%v'", serviceIdArgKey)
	}

	timeoutSeconds, err := flags.GetUint32(timeoutFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the timeout seconds value using key '%v'", timeoutFlagKey)
	}

	// TODO SWITCH TO RETURNING A KURTOSIS_CTX
	getEnclavesResp, err := engineClient.GetEnclaves(ctx, &emptypb.Empty{})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting existing enclaves")
	}

	infoForEnclave, found := getEnclavesResp.EnclaveInfo[enclaveId]
	if !found {
		return stacktrace.Propagate(err, "No enclave with ID '%v' exists", enclaveId)
	}

	enclaveCtx, err := getEnclaveContextFromEnclaveInfo(infoForEnclave)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting an enclave context from enclave info for enclave '%v'", enclaveId)
	}

	if err := enclaveCtx.RemoveService(services.ServiceID(serviceId), uint64(timeoutSeconds)); err != nil {
		return stacktrace.Propagate(err, "An error occurred removing service '%v' from enclave '%v'", serviceId, enclaveId)
	}
	return nil
}

// TODO TODO REMOVE ALL THIS WHEN NewEnclaveContext CAN JUST TAKE IN IP ADDR & PORT NUM!!!
func getEnclaveContextFromEnclaveInfo(infoForEnclave *kurtosis_engine_rpc_api_bindings.EnclaveInfo) (*enclaves.EnclaveContext, error) {
	enclaveId := infoForEnclave.EnclaveId

	apiContainerHostMachineIpAddr, apiContainerHostMachineGrpcPortNum, err := enclave_liveness_validator.ValidateEnclaveLiveness(infoForEnclave)
	if err != nil {
		return nil, stacktrace.NewError("Cannot add service because the API container in enclave '%v' is not running", enclaveId)
	}

	apiContainerHostMachineUrl := fmt.Sprintf(
		"%v:%v",
		apiContainerHostMachineIpAddr,
		apiContainerHostMachineGrpcPortNum,
	)
	conn, err := grpc.Dial(apiContainerHostMachineUrl, grpc.WithInsecure())
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred connecting to the API container grpc port at '%v' in enclave '%v'",
			apiContainerHostMachineUrl,
			enclaveId,
		)
	}
	apiContainerClient := kurtosis_core_rpc_api_bindings.NewApiContainerServiceClient(conn)
	enclaveCtx := enclaves.NewEnclaveContext(
		apiContainerClient,
		enclaves.EnclaveID(enclaveId),
	)

	return enclaveCtx, nil
}
