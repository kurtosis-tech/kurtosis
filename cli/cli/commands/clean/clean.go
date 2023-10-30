package clean

import (
	"context"
	"errors"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/out"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/engine"
	metrics_client "github.com/kurtosis-tech/metrics-library/golang/lib/client"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"sort"
	"strings"
)

const (
	shouldCleanRunningEnclavesFlagKey = "all"
	defaultShouldCleanRunningEnclaves = "false"

	// Titles of the cleaning phases
	// Should be lowercased as they'll go into a string like "Cleaning XXXXX...."
	oldEngineCleaningPhaseTitle = "old Kurtosis engine containers"
	enclavesCleaningPhaseTitle  = "enclaves"
	unusedImagesPhaseTitle      = "unused images"

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"
	uuidAndNameDelimiter  = "\t"

	// this prefix converts the returned guid into a docker name that the user is more used to
	kurtosisEngineGuidPrefix = "kurtosis-engine--"
)

var CleanCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:       command_str_consts.CleanCmdStr,
	ShortDescription: "Cleans up Kurtosis leftover artifacts",
	LongDescription: fmt.Sprintf(
		"Removes stopped enclaves (and live ones if the '%v' flag is set), as well as stopped engine containers",
		shouldCleanRunningEnclavesFlagKey,
	),
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:    engineClientCtxKey,
	Flags: []*flags.FlagConfig{
		{
			Key:       shouldCleanRunningEnclavesFlagKey,
			Usage:     "If set, removes running enclaves as well",
			Shorthand: "a",
			Type:      flags.FlagType_Bool,
			Default:   defaultShouldCleanRunningEnclaves,
		},
	},
	Args:    nil,
	RunFunc: run,
}

func run(
	ctx context.Context,
	kurtosisBackend backend_interface.KurtosisBackend,
	engineClient kurtosis_engine_rpc_api_bindings.EngineServiceClient,
	_ metrics_client.MetricsClient,
	flags *flags.ParsedFlags,
	args *args.ParsedArgs,
) error {
	shouldCleanAll, err := flags.GetBool(shouldCleanRunningEnclavesFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a boolean flag with key '%v' but none was found; this is an error in Kurtosis!", shouldCleanAll)
	}

	// Map of cleaning_phase_title -> (successfully_destroyed_object_id, object_destruction_errors, clean_error)
	cleaningPhaseFunctions := map[string]func() ([]string, []error, error){
		oldEngineCleaningPhaseTitle: func() ([]string, []error, error) {
			// Don't use stacktrace b/c the only reason this function exists is to pass in the right args
			return cleanStoppedEngineContainers(ctx, kurtosisBackend)
		},
		enclavesCleaningPhaseTitle: func() ([]string, []error, error) {
			// Don't use stacktrace b/c the only reason this function exists is to pass in the right args
			return cleanEnclaves(ctx, engineClient, shouldCleanAll)
		},
		unusedImagesPhaseTitle: func() ([]string, []error, error) {
			// Don't use stacktrace b/c the only reason this function exists is to pass in the right args
			return cleanUnusedImages(ctx, kurtosisBackend)
		},
	}

	phasesWithErrors := []string{}
	for phaseTitle, cleaningFunc := range cleaningPhaseFunctions {
		logrus.Infof("Cleaning %v...", phaseTitle)
		successfullyRemovedArtifactUuids, removalErrors, err := cleaningFunc()
		if err != nil {
			logrus.Errorf("Errors occurred cleaning %v:\n%v", phaseTitle, err)
			phasesWithErrors = append(phasesWithErrors, phaseTitle)
			continue
		}

		if len(successfullyRemovedArtifactUuids) > 0 {
			logrus.Infof("Successfully removed the following %v:", phaseTitle)
			sort.Strings(successfullyRemovedArtifactUuids)
			for _, successfulArtifactUuid := range successfullyRemovedArtifactUuids {
				out.PrintOutLn(successfulArtifactUuid)
			}
		}

		if len(removalErrors) > 0 {
			logrus.Errorf("Errors occurred removing the following %v:", phaseTitle)
			for _, err := range removalErrors {
				out.PrintErrLn("")
				out.PrintErrLn(err.Error())
			}
			phasesWithErrors = append(phasesWithErrors, phaseTitle)
			continue
		}
		logrus.Infof("Successfully cleaned %v", phaseTitle)
	}

	if len(phasesWithErrors) > 0 {
		errorStr := "Errors occurred cleaning " + strings.Join(phasesWithErrors, ", ")
		return errors.New(errorStr)
	}
	return nil
}

// ====================================================================================================
//
//	Private Helper Functions
//
// ====================================================================================================

func cleanStoppedEngineContainers(ctx context.Context, kurtosisBackend backend_interface.KurtosisBackend) ([]string, []error, error) {

	engineFilters := &engine.EngineFilters{
		GUIDs: nil,
		Statuses: map[container.ContainerStatus]bool{
			container.ContainerStatus_Stopped: true,
		},
	}

	successfulEngineGuids, erroredEngineGuids, err := kurtosisBackend.DestroyEngines(ctx, engineFilters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred destroying engines using filters '%+v'", engineFilters)
	}

	successfulEngineContainerNames := []string{}
	for engineGuid := range successfulEngineGuids {
		successfulEngineContainerNames = append(successfulEngineContainerNames, kurtosisEngineGuidPrefix+string(engineGuid))
	}

	removeEngineErrors := []error{}
	for engineGuid, err := range erroredEngineGuids {
		wrappedErr := stacktrace.Propagate(err, "An error occurred destroying stopped engine '%v'", engineGuid)
		removeEngineErrors = append(removeEngineErrors, wrappedErr)
	}

	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred cleaning stopped Kurtosis engine containers")
	}
	return successfulEngineContainerNames, removeEngineErrors, nil
}

func cleanEnclaves(ctx context.Context, engineClient kurtosis_engine_rpc_api_bindings.EngineServiceClient, shouldCleanAll bool) ([]string, []error, error) {
	cleanArgs := &kurtosis_engine_rpc_api_bindings.CleanArgs{ShouldCleanAll: &shouldCleanAll}
	cleanResp, err := engineClient.Clean(ctx, cleanArgs)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred while calling clean")
	}

	successfullyDestroyedEnclaveUuidsAndNames := []string{}
	for _, enclaveUuidWithName := range cleanResp.RemovedEnclaveNameAndUuids {
		successfullyDestroyedEnclaveUuidsAndNames = append(successfullyDestroyedEnclaveUuidsAndNames, formattedUuidAndName(enclaveUuidWithName))
	}
	return successfullyDestroyedEnclaveUuidsAndNames, nil, nil
}

func formattedUuidAndName(enclaveUuidWithName *kurtosis_engine_rpc_api_bindings.EnclaveNameAndUuid) string {
	return fmt.Sprintf("%v%v%v", enclaveUuidWithName.Uuid, uuidAndNameDelimiter, enclaveUuidWithName.Name)
}

func cleanUnusedImages(ctx context.Context, kurtosisBackend backend_interface.KurtosisBackend) ([]string, []error, error) {
	cleanedImages, cleanErr := kurtosisBackend.PruneUnusedImages(ctx)
	return cleanedImages, nil, cleanErr
}
