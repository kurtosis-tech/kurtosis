package rm

import (
	"context"
	"errors"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/v2/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/v2/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/enclave_id_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/metrics_client"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"sort"
	"strings"
)

const (
	enclaveIdentifiersArgKey = "enclaves"
	isEnclaveIdArgOptional   = false
	isEnclaveIdArgGreedy     = true

	shouldForceRemoveFlagKey = "force"
	defaultShouldForceRemove = "false"

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"
)

var EnclaveRmCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.EnclaveRmCmdStr,
	ShortDescription:          "Destroys the specified enclaves",
	LongDescription:           "Destroys the specified enclaves, removing all resources associated with them",
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:    engineClientCtxKey,
	Flags: []*flags.FlagConfig{
		{
			Key:       shouldForceRemoveFlagKey,
			Usage:     "Deletes all enclaves, regardless of whether they're already stopped",
			Shorthand: "f",
			Type:      flags.FlagType_Bool,
			Default:   defaultShouldForceRemove,
		},
	},
	Args: []*args.ArgConfig{
		enclave_id_arg.NewEnclaveIdentifierArg(
			enclaveIdentifiersArgKey,
			engineClientCtxKey,
			isEnclaveIdArgOptional,
			isEnclaveIdArgGreedy,
		),
	},
	RunFunc: run,
}

func run(
	ctx context.Context,
	_ backend_interface.KurtosisBackend,
	_ kurtosis_engine_rpc_api_bindings.EngineServiceClient,
	_ metrics_client.MetricsClient,
	flags *flags.ParsedFlags,
	args *args.ParsedArgs,
) error {
	enclaveIdentifiers, err := args.GetGreedyArg(enclaveIdentifiersArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a value for greedy enclave identifier arg '%v' but none was found; this is a bug in the Kurtosis CLI!", enclaveIdentifiersArgKey)
	}

	shouldForceRemove, err := flags.GetBool(shouldForceRemoveFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the force-removal flag value using key '%v'; this is a bug in Kurtosis!", shouldForceRemoveFlagKey)
	}

	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating Kurtosis Context from local engine")
	}

	logrus.Debugf("inputted enclave UUIDs: %+v", enclaveIdentifiers)

	// Condense the enclave UUIDs down into a unique set, so we don't try to double-destroy an enclave
	enclaveIdsToDestroy := getUniqueSortedEnclaveIdentifiers(enclaveIdentifiers)

	logrus.Debugf("Unique enclave UUIDs to destroy: %+v", enclaveIdsToDestroy)

	logrus.Info("Destroying enclaves...")

	enclaveDestructionErrorStrs := []string{}
	for _, enclaveId := range enclaveIdsToDestroy {
		if err := destroyEnclave(ctx, kurtosisCtx, enclaveId, shouldForceRemove); err != nil {
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
//
//	Private helper methods
//
// ====================================================================================================
func getUniqueSortedEnclaveIdentifiers(rawInput []string) []string {
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
	kurtosisContext *kurtosis_context.KurtosisContext,
	enclaveIdentifier string,
	shouldForceRemove bool,
) error {
	enclaveInfo, err := kurtosisContext.GetEnclave(ctx, enclaveIdentifier)
	if err != nil {
		return stacktrace.NewError("No enclave '%v' exists", enclaveIdentifier)
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
			enclaveIdentifier,
			enclaveStatus,
			shouldForceRemoveFlagKey,
		)
	}

	if err = kurtosisContext.DestroyEnclave(ctx, enclaveIdentifier); err != nil {
		return stacktrace.Propagate(err, "An error occurred destroying enclave '%v'", enclaveIdentifier)
	}
	return nil
}
