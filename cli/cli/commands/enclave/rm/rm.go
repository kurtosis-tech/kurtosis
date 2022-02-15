package rm

import (
	"context"
	"errors"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/kurtosis_command/flags"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/prebuilt_command_components/enclave_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
	"sort"
	"strings"
)

const (
	shouldAcceptMultipleEnclaves = true

	shouldForceRemoveFlagKey = "force"

	defaultShouldForceRemove = "false"
)

var EnclaveRmCmd = &enclave_consuming_kurtosis_command.EnclaveConsumingKurtosisCommand{
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
	ShouldAcceptMultipleEnclaveIDs: shouldAcceptMultipleEnclaves,
	RunFunc:                        run,
}

func run(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
	engineClient kurtosis_engine_rpc_api_bindings.EngineServiceClient,
	enclaveIds []string,
	flags *flags.ParsedFlags,
) error {
	shouldForceRemove, err := flags.GetBool(shouldForceRemoveFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the force-removal flag value using key '%v'; this is a bug in Kurtosis!", shouldForceRemoveFlagKey)
	}

	logrus.Debugf("inputted enclave IDs: %+v", enclaveIds)

	// Condense the enclave IDs down into a unique set, so we don't try to double-destroy an enclave
	enclaveIdsToDestroy := getUniqueSortedEnclaveIDs(enclaveIds)

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
