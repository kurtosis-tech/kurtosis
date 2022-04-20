package clean

import (
	"context"
	"errors"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang/kurtosis_engine_rpc_api_bindings"
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

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey = "engine-client"
)

var CleanCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:              command_str_consts.CleanCmdStr,
	ShortDescription: "Cleans up Kurtosis leftover artifacts",
	LongDescription: fmt.Sprintf(
		"Removes Kurtosis stopped Kurtosis enclaves (and live ones if the '%v' flag is set), as well as stopped engine containers",
		shouldCleanRunningEnclavesFlagKey,
	),
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:  engineClientCtxKey,
	Flags: []*flags.FlagConfig{
		{
			Key:       shouldCleanRunningEnclavesFlagKey,
			Usage:     "If set, removes running enclaves as well",
			Shorthand: "a",
			Type:      flags.FlagType_Bool,
			Default:   defaultShouldCleanRunningEnclaves,
		},
	},
	RunFunc:                 run,
}

func run(
	ctx context.Context,
	kurtosisBackend backend_interface.KurtosisBackend,
	engineClient kurtosis_engine_rpc_api_bindings.EngineServiceClient,
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
	}

	phasesWithErrors := []string{}
	for phaseTitle, cleaningFunc := range cleaningPhaseFunctions {
		logrus.Infof("Cleaning %v...", phaseTitle)
		successfullyRemovedArtifactIds, removalErrors, err := cleaningFunc()
		if err != nil {
			logrus.Errorf("Errors occurred cleaning %v:\n%v", phaseTitle, err)
			phasesWithErrors = append(phasesWithErrors, phaseTitle)
			continue
		}

		if len(successfullyRemovedArtifactIds) > 0 {
			logrus.Infof("Successfully removed the following %v:", phaseTitle)
			sort.Strings(successfullyRemovedArtifactIds)
			for _, successfulArtifactId := range successfullyRemovedArtifactIds {
				fmt.Fprintln(logrus.StandardLogger().Out, successfulArtifactId)
			}
		}

		if len(removalErrors) > 0 {
			logrus.Errorf("Errors occurred removing the following %v:", phaseTitle)
			for _, err := range removalErrors {
				fmt.Fprintln(logrus.StandardLogger().Out, "")
				fmt.Fprintln(logrus.StandardLogger().Out, err.Error())
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
//                                       Private Helper Functions
// ====================================================================================================
func cleanStoppedEngineContainers(ctx context.Context, kurtosisBackend backend_interface.KurtosisBackend) ([]string, []error, error) {

	engineFilters := &engine.EngineFilters{
		Statuses: map[container_status.ContainerStatus]bool{
			container_status.ContainerStatus_Stopped: true,
		},
	}

	successfulEngineIds, erroredEngineIds, err := kurtosisBackend.DestroyEngines(ctx, engineFilters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred destroying engines using filters '%+v'", engineFilters)
	}

	successfullyDestroyedEngineIDs := []string{}
	for engineId := range successfulEngineIds {
		successfullyDestroyedEngineIDs = append(successfullyDestroyedEngineIDs, engineId)
	}

	removeEngineErrors := []error{}
	for engineId, err := range erroredEngineIds {
		wrappedErr := stacktrace.Propagate(err, "An error occurred destroying stopped engine '%v'", engineId)
		removeEngineErrors = append(removeEngineErrors, wrappedErr)
	}

	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred cleaning stopped Kurtosis engine containers")
	}
	return successfullyDestroyedEngineIDs, removeEngineErrors, nil
}

func cleanEnclaves(ctx context.Context, engineClient kurtosis_engine_rpc_api_bindings.EngineServiceClient, shouldCleanAll bool) ([]string, []error, error) {
	cleanArgs := &kurtosis_engine_rpc_api_bindings.CleanArgs{ShouldCleanAll: shouldCleanAll}
	cleanResp, err := engineClient.Clean(ctx, cleanArgs)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred while calling clean")
	}

	successfullyDestroyedEnclaveIds := []string{}
	for enclaveId, _ := range cleanResp.RemovedEnclaveIds {
		successfullyDestroyedEnclaveIds = append(successfullyDestroyedEnclaveIds, enclaveId)
	}
	return successfullyDestroyedEnclaveIds, nil, nil
}
