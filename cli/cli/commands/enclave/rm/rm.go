package rm

import (
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/kurtosis_command"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/kurtosis_command/args"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/kurtosis_command/flags"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/prebuilt_command_components/enclave_id_arg"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/engine_manager"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/object-attributes-schema-lib/schema"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
	"sort"
	"strings"
)

const (
	enclaveIdArgKey          = "enclave-id"
	shouldForceRemoveFlagKey = "force"

	engineClientCtxKey = "engine-client"
	engineClientCloseFuncCtxKey = "engine-client-close-func"

	isEnclaveIdArgOptional = false
	isEnclaveIdArgGreedy = true

	defaultShouldForceRemove = "false"
)

var EnclaveRmCmd = &kurtosis_command.KurtosisCommand{
	CommandStr:       command_str_consts.EnclaveRmCmdStr,
	ShortDescription: "Destroys the specified enclaves",
	LongDescription:  "Destroys the specified enclaves, removing all resources associated with them",
	Flags:            []*flags.FlagConfig{
		{
			Key:       shouldForceRemoveFlagKey,
			Usage:     "Deletes all enclaves, regardless of whether they're already stopped",
			Shorthand: "f",
			Type:      flags.FlagType_Bool,
			Default:   defaultShouldForceRemove,
		},
	},
	Args:             []*args.ArgConfig{
		enclave_id_arg.NewEnclaveIDArg(
			enclaveIdArgKey,
			engineClientCtxKey,
			isEnclaveIdArgOptional,
			isEnclaveIdArgGreedy,
		),
	},
	PreValidationAndRunFunc: setup,
	RunFunc:          run,
	PostValidationAndRunFunc: teardown,
}

// TODO generalize this to something like EngineUsingKurtosisCommand or something
func setup(ctx context.Context) (context.Context, error) {
	result := ctx

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the Docker client")
	}
	dockerManager := docker_manager.NewDockerManager(
		logrus.StandardLogger(),
		dockerClient,
	)

	engineManager := engine_manager.NewEngineManager(dockerManager)
	objAttrsProvider := schema.GetObjectAttributesProvider()
	engineClient, closeClientFunc, err := engineManager.StartEngineIdempotentlyWithDefaultVersion(ctx, objAttrsProvider, defaults.DefaultEngineLogLevel)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a new Kurtosis engine client")
	}
	result = context.WithValue(result, engineClientCtxKey, engineClient)
	result = context.WithValue(result, engineClientCloseFuncCtxKey, closeClientFunc)

	return result, nil
}

func run(ctx context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error {
	uncastedEngineClient := ctx.Value(engineClientCtxKey)
	if uncastedEngineClient == nil {
		return stacktrace.NewError("Expected an engine client to have been stored in the context under key '%v', but none was found; this is a bug in Kurtosis!", engineClientCtxKey)
	}
	engineClient, ok := uncastedEngineClient.(kurtosis_engine_rpc_api_bindings.EngineServiceClient)
	if !ok {
		return stacktrace.NewError("Found an object that should be the engine client stored in the context under key '%v', but this object wasn't of the correct type", engineClientCtxKey)
	}

	inputtedEnclaveIds, err := args.GetGreedyArg(enclaveIdArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the enclave IDs using arg key '%v'; this is a bug in Kurtosis!", enclaveIdArgKey)
	}
	shouldForceRemove, err := flags.GetBool(shouldForceRemoveFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the force-removal flag value using key '%v'; this is a bug in Kurtosis!", shouldForceRemoveFlagKey)
	}

	logrus.Debugf("inputted enclave IDs: %+v", inputtedEnclaveIds)

	// Condense the enclave IDs down into a unique set, so we don't try to double-destroy an enclave
	enclaveIdsToDestroy := getUniqueSortedEnclaveIDs(inputtedEnclaveIds)

	logrus.Debugf("Unique enclave IDs to destroy: %+v", enclaveIdsToDestroy)

	logrus.Info("Destroying enclaves...")
	getEnclavesResp, err := engineClient.GetEnclaves(ctx, &emptypb.Empty{})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting enclaves to check that the ones to destroy are stopped")
	}
	allEnclaveInfo := getEnclavesResp.EnclaveInfo

	enclaveDestructionErrorStrs := []string{}
	for _, enclaveId := range enclaveIdsToDestroy {
		if err := destroyEnclave(ctx, enclaveId, allEnclaveInfo, engineClient, shouldForceRemove); err != nil {
			enclaveDestructionErrorStrs = append(
				enclaveDestructionErrorStrs,
				fmt.Sprintf(
					">>>>>>>>>>>>>>>>> %v <<<<<<<<<<<<<<<<<\n%v",
					enclaveId,
					err.Error(),
				),
			)
		}
	}

	if len(enclaveDestructionErrorStrs) > 0 {
		errorStr := fmt.Sprintf(
			"One or more errors occurred destroying the enclaves:\n%v",
			strings.Join(enclaveDestructionErrorStrs, "\n\n"),
		)
		return errors.New(errorStr)
	}

	logrus.Info("Enclaves successfully destroyed")

	return nil
}

func teardown(ctx context.Context) {
	uncastedEngineClientCloseFunc := ctx.Value(engineClientCloseFuncCtxKey)
	if uncastedEngineClientCloseFunc != nil {
		engineClientCloseFunc, ok := uncastedEngineClientCloseFunc.(func() error)
		if ok {
			if err := engineClientCloseFunc(); err != nil {
				logrus.Warnf("We tried to close the engine client after we're done using it, but doing so threw an error:\n%v", err)
			}
		} else {
			logrus.Errorf("Expected the object at context key '%v' to be an engine client close function, but it wasn't; this is a bug in Kurtosis!", engineClientCloseFuncCtxKey)
		}
	} else {
		logrus.Errorf(
			"Expected to find an engine client close function during teardown at context key '%v', but none was found; this is a bug in Kurtosis!",
			engineClientCloseFuncCtxKey,
		)
	}
}

// ====================================================================================================
// 									   Private helper methods
// ====================================================================================================
func getUniqueSortedEnclaveIDs(rawInput []string) []string {
	uniqueEnclaveIds := map[string]bool{}
	for _, inputId := range rawInput {
		uniqueEnclaveIds[inputId] = true
	}

	result := []string{}
	for inputId := range uniqueEnclaveIds {
		result = append(result, inputId)
	}
	sort.Strings(result)
	return result
}

func destroyEnclave(
	ctx context.Context,
	enclaveId string,
	allEnclaveInfo map[string]*kurtosis_engine_rpc_api_bindings.EnclaveInfo,
	engineClient kurtosis_engine_rpc_api_bindings.EngineServiceClient,
	shouldForceRemove bool,
) error {
	enclaveInfo, found := allEnclaveInfo[enclaveId]
	if !found {
		return stacktrace.NewError("No enclave '%v' exists", enclaveId)
	}

	enclaveStatus := enclaveInfo.ContainersStatus
	var isEnclaveRemovableWithoutForce bool
	switch enclaveStatus {
	case kurtosis_engine_rpc_api_bindings.EnclaveContainersStatus_EnclaveContainersStatus_EMPTY, kurtosis_engine_rpc_api_bindings.EnclaveContainersStatus_EnclaveContainersStatus_STOPPED:
		isEnclaveRemovableWithoutForce = true
	case kurtosis_engine_rpc_api_bindings.EnclaveContainersStatus_EnclaveContainersStatus_RUNNING:
		isEnclaveRemovableWithoutForce = false
	default:
		return stacktrace.NewError("Unrecognized enclave status '%v'; this is a bug in Kurtosis", enclaveStatus)
	}

	if !shouldForceRemove && !isEnclaveRemovableWithoutForce {
		return stacktrace.NewError(
			"Refusing to destroy enclave '%v' because its status is '%v'; to force its removal, rerun this command with the '%v' flag",
			enclaveId,
			enclaveStatus,
			shouldForceRemoveFlagKey,
		)
	}

	destroyEnclaveArgs := &kurtosis_engine_rpc_api_bindings.DestroyEnclaveArgs{EnclaveId: enclaveId}
	if _, err := engineClient.DestroyEnclave(ctx, destroyEnclaveArgs); err != nil {
		return stacktrace.Propagate(err, "An error occurred destroying enclave '%v'", enclaveId)
	}
	return nil
}
