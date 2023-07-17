package tasks

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/files_artifacts_expansion"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers/magic_string_helper"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/xtgo/uuid"
	"go.starlark.net/starlark"
	"strings"
)

const (
	RunShBuiltinName = "run_sh"

	defaultRunShImageName = "badouralix/curl-jq"
)

func NewRunShService(serviceNetwork service_network.ServiceNetwork, runtimeValueStore *runtime_value_store.RuntimeValueStore) *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return &kurtosis_plan_instruction.KurtosisPlanInstruction{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: RunShBuiltinName,

			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              RunArgName,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
				},
				{
					Name:              ImageNameArgName,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(argumentValue starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(argumentValue, ImageNameArgName)
					},
				},
				{
					Name:              FilesArgName,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.Dict],
				},
				{
					Name:              StoreFilesArgName,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.List],
				},
				{
					Name:              WaitArgName,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Value],
					// the value can be a string duration, or it can be a Starlark none value (because we are preparing
					// the signature to receive a custom type in the future) when users want to disable it
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.DurationOrNone(value, WaitArgName)
					},
				},
			},
		},

		Capabilities: func() kurtosis_plan_instruction.KurtosisPlanInstructionCapabilities {
			return &RunShCapabilities{
				serviceNetwork:      serviceNetwork,
				runtimeValueStore:   runtimeValueStore,
				name:                "",
				serviceConfig:       nil, // populated at interpretation time
				run:                 "",  // populated at interpretation time
				resultUuid:          "",  // populated at interpretation time
				fileArtifactNames:   nil,
				pathToFileArtifacts: nil,
				wait:                DefaultWaitTimeoutDurationStr,
			}
		},

		DefaultDisplayArguments: map[string]bool{
			RunArgName:        true,
			ImageNameArgName:  true,
			FilesArgName:      true,
			StoreFilesArgName: true,
			WaitArgName:       true,
		},
	}
}

type RunShCapabilities struct {
	runtimeValueStore *runtime_value_store.RuntimeValueStore
	serviceNetwork    service_network.ServiceNetwork

	resultUuid string
	name       string
	run        string

	serviceConfig       *service.ServiceConfig
	fileArtifactNames   []string
	pathToFileArtifacts []string
	wait                string
}

func (builtin *RunShCapabilities) Interpret(arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError) {
	runCommand, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, RunArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", RunArgName)
	}
	builtin.run = runCommand.GoString()

	var image string
	if arguments.IsSet(ImageNameArgName) {
		imageStarlark, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, ImageNameArgName)
		if err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", ImageNameArgName)
		}
		image = imageStarlark.GoString()
	} else {
		image = defaultRunShImageName
	}

	var filesArtifactExpansion *files_artifacts_expansion.FilesArtifactsExpansion
	if arguments.IsSet(FilesArgName) {
		filesStarlark, err := builtin_argument.ExtractArgumentValue[*starlark.Dict](arguments, FilesArgName)
		if err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", FilesArgName)
		}
		if filesStarlark.Len() > 0 {
			filesArtifactMountDirPaths, interpretationErr := kurtosis_types.SafeCastToMapStringString(filesStarlark, FilesArgName)
			if interpretationErr != nil {
				return nil, interpretationErr
			}
			filesArtifactExpansion, interpretationErr = service_config.ConvertFilesArtifactsMounts(filesArtifactMountDirPaths, builtin.serviceNetwork)
			if interpretationErr != nil {
				return nil, interpretationErr
			}
		}
	}

	// build a service config from image and files artifacts expansion.
	builtin.serviceConfig = getServiceConfig(image, filesArtifactExpansion)

	if arguments.IsSet(StoreFilesArgName) {
		storeFilesList, err := builtin_argument.ExtractArgumentValue[*starlark.List](arguments, StoreFilesArgName)
		if err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", StoreFilesArgName)
		}
		if storeFilesList.Len() > 0 {

			storeFilesArray, interpretationErr := kurtosis_types.SafeCastToStringSlice(storeFilesList, StoreFilesArgName)
			if interpretationErr != nil {
				return nil, interpretationErr
			}

			builtin.pathToFileArtifacts = storeFilesArray

			// generate unique names
			var uniqueNames []string
			for range storeFilesArray {
				uniqueNameForArtifact, err := builtin.serviceNetwork.GetUniqueNameForFileArtifact()
				if err != nil {
					return nil, startosis_errors.WrapWithInterpretationError(err, "error occurred while generating unique name for file artifact")
				}
				uniqueNames = append(uniqueNames, uniqueNameForArtifact)
			}

			builtin.fileArtifactNames = uniqueNames
		}
	}

	if arguments.IsSet(WaitArgName) {
		var waitTimeout string
		waitValue, err := builtin_argument.ExtractArgumentValue[starlark.Value](arguments, WaitArgName)
		if err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "error occurred while extracting wait information")
		}
		if waitValueStr, ok := waitValue.(starlark.String); ok {
			waitTimeout = waitValueStr.GoString()
		} else if _, ok := waitValue.(starlark.NoneType); ok {
			waitTimeout = DisableWaitTimeoutDurationStr
		}
		builtin.wait = waitTimeout
	}

	resultUuid, err := builtin.runtimeValueStore.CreateValue()
	if err != nil {
		return nil, startosis_errors.NewInterpretationError("An error occurred while generating UUID for future reference for %v instruction", RunShBuiltinName)
	}
	builtin.resultUuid = resultUuid
	randomUuid := uuid.NewRandom()
	builtin.name = fmt.Sprintf("task-%v", randomUuid.String())

	result := createInterpretationResult(resultUuid, builtin.fileArtifactNames)
	return result, nil
}

func (builtin *RunShCapabilities) Validate(_ *builtin_argument.ArgumentValuesSet, validatorEnvironment *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
	// TODO validate bash
	var serviceDirpathsToArtifactIdentifiers map[string]string
	if builtin.serviceConfig.GetFilesArtifactsExpansion() != nil {
		serviceDirpathsToArtifactIdentifiers = builtin.serviceConfig.GetFilesArtifactsExpansion().ServiceDirpathsToArtifactIdentifiers
	}
	return validateTasksCommon(validatorEnvironment, builtin.fileArtifactNames, builtin.pathToFileArtifacts, serviceDirpathsToArtifactIdentifiers, builtin.serviceConfig.GetContainerImageName())
}

// Execute This is just v0 for run_sh task - we can later improve on it.
//
//		TODO: stop the container as soon as task completed.
//	  Create an mechanism for other services to retrieve files from the task container
//	  Make task as its own entity instead of currently shown under services
func (builtin *RunShCapabilities) Execute(ctx context.Context, _ *builtin_argument.ArgumentValuesSet) (string, error) {
	_, err := builtin.serviceNetwork.AddService(ctx, service.ServiceName(builtin.name), builtin.serviceConfig)
	if err != nil {
		return "", stacktrace.Propagate(err, "error occurred while creating a run_sh task with image: %v", builtin.serviceConfig.GetContainerImageName())
	}

	// create work directory and cd into that directory
	commandToRun, err := getCommandToRun(builtin)
	if err != nil {
		return "", stacktrace.Propagate(err, "error occurred while preparing the sh command to execute on the image")
	}
	fullCommandToRun := []string{shellWrapperCommand, "-c", commandToRun}

	// run the command passed in by user in the container
	createDefaultDirectoryResult, err := executeWithWait(ctx, builtin.serviceNetwork, builtin.name, builtin.wait, fullCommandToRun)
	if err != nil {
		return "", stacktrace.Propagate(err, fmt.Sprintf("error occurred while executing one time task command: %v ", builtin.run))
	}

	result := map[string]starlark.Comparable{
		runResultOutputKey: starlark.String(createDefaultDirectoryResult.GetOutput()),
		runResultCodeKey:   starlark.MakeInt(int(createDefaultDirectoryResult.GetExitCode())),
	}

	builtin.runtimeValueStore.SetValue(builtin.resultUuid, result)
	instructionResult := resultMapToString(result)

	// throw an error as execution of the command failed
	if createDefaultDirectoryResult.GetExitCode() != 0 {
		return "", stacktrace.NewError(
			"error occurred and shell command: %q exited with code %d with output %q",
			commandToRun, createDefaultDirectoryResult.GetExitCode(), createDefaultDirectoryResult.GetOutput())
	}

	if builtin.fileArtifactNames != nil && builtin.pathToFileArtifacts != nil {
		err = copyFilesFromTask(ctx, builtin.serviceNetwork, builtin.name, builtin.fileArtifactNames, builtin.pathToFileArtifacts)
		if err != nil {
			return "", stacktrace.Propagate(err, "error occurred while copying files from  a task")
		}
	}
	return instructionResult, err
}

func getCommandToRun(builtin *RunShCapabilities) (string, error) {
	// replace future references to actual strings
	maybeSubCommandWithRuntimeValues, err := magic_string_helper.ReplaceRuntimeValueInString(builtin.run, builtin.runtimeValueStore)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while replacing runtime values in run_sh")
	}
	commandWithNoNewLines := strings.ReplaceAll(maybeSubCommandWithRuntimeValues, newlineChar, " ")

	return commandWithNoNewLines, nil
}
