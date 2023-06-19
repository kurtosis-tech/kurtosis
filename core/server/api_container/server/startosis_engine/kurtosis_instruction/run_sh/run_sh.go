package run_sh

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers/magic_string_helper"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/xtgo/uuid"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"reflect"
	"strings"
	"time"
)

const (
	RunShBuiltinName = "run_sh"

	ImageNameArgName = "image"
	RunArgName       = "run"
	StoreFilesName   = "store"
	WaitName         = "wait"
	FilesName        = "files"

	DefaultImageName = "badouralix/curl-jq"

	runshCodeKey         = "code"
	runshOutputKey       = "output"
	runshFileArtifactKey = "files_artifacts"
	newlineChar          = "\n"

	shellCommand = "/bin/sh"

	DefaultWaitTimeoutDurationStr = "180s"
	DisableWaitTimeoutDurationStr = ""
)

var runTailCommandToPreventContainerToStopOnCreating = []string{"tail", "-f", "/dev/null"}

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
					Name:              FilesName,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.Dict],
				},
				{
					Name:              StoreFilesName,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.List],
				},
				{
					Name:              WaitName,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Value],
					// the value can be a string duration, or it can be a Starlark none value (because we are preparing
					// the signature to receive a custom type in the future) when users want to disable it
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.DurationOrNone(value, WaitName)
					},
				},
			},
		},

		Capabilities: func() kurtosis_plan_instruction.KurtosisPlanInstructionCapabilities {
			return &RunShCapabilities{
				serviceNetwork:      serviceNetwork,
				runtimeValueStore:   runtimeValueStore,
				name:                "",
				image:               DefaultImageName, // populated at interpretation time
				run:                 "",               // populated at interpretation time
				files:               nil,
				resultUuid:          "", // populated at interpretation time
				fileArtifactNames:   nil,
				pathToFileArtifacts: nil,
				wait:                DefaultWaitTimeoutDurationStr,
			}
		},

		DefaultDisplayArguments: map[string]bool{
			RunArgName:       true,
			ImageNameArgName: true,
			FilesName:        true,
			StoreFilesName:   true,
			WaitName:         true,
		},
	}
}

type RunShCapabilities struct {
	runtimeValueStore *runtime_value_store.RuntimeValueStore
	serviceNetwork    service_network.ServiceNetwork

	resultUuid          string
	name                string
	run                 string
	image               string
	files               map[string]string
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

	if arguments.IsSet(ImageNameArgName) {
		imageStarlark, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, ImageNameArgName)
		if err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", ImageNameArgName)
		}
		builtin.image = imageStarlark.GoString()
	}

	if arguments.IsSet(FilesName) {
		filesStarlark, err := builtin_argument.ExtractArgumentValue[*starlark.Dict](arguments, FilesName)
		if err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", FilesName)
		}
		if filesStarlark.Len() > 0 {
			filesArtifactMountDirPaths, interpretationErr := kurtosis_types.SafeCastToMapStringString(filesStarlark, FilesName)
			if interpretationErr != nil {
				return nil, interpretationErr
			}
			builtin.files = filesArtifactMountDirPaths
		}
	}

	if arguments.IsSet(StoreFilesName) {
		storeFilesList, err := builtin_argument.ExtractArgumentValue[*starlark.List](arguments, StoreFilesName)
		if err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", StoreFilesName)
		}
		if storeFilesList.Len() > 0 {

			storeFilesArray, interpretationErr := kurtosis_types.SafeCastToStringSlice(storeFilesList, StoreFilesName)
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

	if arguments.IsSet(WaitName) {
		var waitTimeout string
		waitValue, err := builtin_argument.ExtractArgumentValue[starlark.Value](arguments, WaitName)
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

	runShCodeValue := fmt.Sprintf(magic_string_helper.RuntimeValueReplacementPlaceholderFormat, builtin.resultUuid, runshCodeKey)
	runShOutputValue := fmt.Sprintf(magic_string_helper.RuntimeValueReplacementPlaceholderFormat, builtin.resultUuid, runshOutputKey)

	dict := map[string]starlark.Value{}
	dict[runshCodeKey] = starlark.String(runShCodeValue)
	dict[runshOutputKey] = starlark.String(runShOutputValue)

	// converting go slice to starlark list
	artifactNamesList := &starlark.List{}
	if len(builtin.fileArtifactNames) > 0 {
		for _, name := range builtin.fileArtifactNames {
			// purposely not checking error for list because it's mutable so should not throw any errors until this point
			_ = artifactNamesList.Append(starlark.String(name))
		}
	}
	dict[runshFileArtifactKey] = artifactNamesList
	response := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	return response, nil
}

func (builtin *RunShCapabilities) Validate(_ *builtin_argument.ArgumentValuesSet, validatorEnvironment *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
	if builtin.fileArtifactNames != nil {
		if len(builtin.fileArtifactNames) != len(builtin.pathToFileArtifacts) {
			return startosis_errors.NewValidationError("error occurred while validating file artifact name for each file in store array. "+
				"This seems to be a bug, please create a ticket for it. names: %v paths: %v", len(builtin.fileArtifactNames), len(builtin.pathToFileArtifacts))
		}

		err := validatePathIsUniqueWhileCreatingFileArtifact(builtin.pathToFileArtifacts)
		if err != nil {
			return startosis_errors.WrapWithValidationError(err, "error occurred while validating file paths to copy into file artifact")
		}

		for _, name := range builtin.fileArtifactNames {
			validatorEnvironment.AddArtifactName(name)
		}
	}

	if builtin.files != nil {
		for _, artifactName := range builtin.files {
			if !validatorEnvironment.DoesArtifactNameExist(artifactName) {
				return startosis_errors.NewValidationError("There was an error validating '%s' as artifact name '%s' does not exist", RunShBuiltinName, artifactName)
			}
		}
	}

	validatorEnvironment.AppendRequiredContainerImage(builtin.image)
	return nil
}

// Execute This is just v0 for run_sh task - we can later improve on it.
//	TODO: stop the container as soon as task completed.
//   Create an mechanism for other services to retrieve files from the task container
//   Make task as its own entity instead of currently shown under services
func (builtin *RunShCapabilities) Execute(ctx context.Context, _ *builtin_argument.ArgumentValuesSet) (string, error) {
	// create work directory and cd into that directory
	commandToRun, err := getCommandToRun(builtin)
	if err != nil {
		return "", stacktrace.Propagate(err, "error occurred while preparing the sh command to execute on the image")
	}
	fullCommandToRun := []string{shellCommand, "-c", commandToRun}
	serviceConfigBuilder := services.NewServiceConfigBuilder(builtin.image)
	serviceConfigBuilder.WithFilesArtifactMountDirpaths(builtin.files)
	// This make sure that the container does not stop as soon as it starts
	// This only is needed for kubernetes at the moment
	// TODO: Instead of creating a service and running exec commands
	//  we could probably run the command as an entrypoint and retrieve the results as soon as the
	//  command is completed
	serviceConfigBuilder.WithEntryPointArgs(runTailCommandToPreventContainerToStopOnCreating)
	serviceConfig := serviceConfigBuilder.Build()
	_, err = builtin.serviceNetwork.AddService(ctx, service.ServiceName(builtin.name), serviceConfig)

	if err != nil {
		return "", stacktrace.Propagate(err, fmt.Sprintf("error occurred while creating a run_sh task with image: %v", builtin.image))
	}

	// run the command passed in by user in the container
	createDefaultDirectoryResult, err := executeWithWait(ctx, builtin, fullCommandToRun)
	if err != nil {
		return "", stacktrace.Propagate(err, fmt.Sprintf("error occurred while executing one time task command: %v ", builtin.run))
	}

	result := map[string]starlark.Comparable{
		runshOutputKey: starlark.String(createDefaultDirectoryResult.GetOutput()),
		runshCodeKey:   starlark.MakeInt(int(createDefaultDirectoryResult.GetExitCode())),
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
		err = copyFilesFromTask(ctx, builtin)
		if err != nil {
			return "", stacktrace.Propagate(err, "error occurred while copying files from  a task")
		}
	}
	return instructionResult, err
}

func copyFilesFromTask(ctx context.Context, builtin *RunShCapabilities) error {
	if builtin.fileArtifactNames == nil || builtin.pathToFileArtifacts == nil {
		return nil
	}

	for index, fileArtifactPath := range builtin.pathToFileArtifacts {
		fileArtifactName := builtin.fileArtifactNames[index]
		_, err := builtin.serviceNetwork.CopyFilesFromService(ctx, builtin.name, fileArtifactPath, fileArtifactName)
		if err != nil {
			return stacktrace.Propagate(err, fmt.Sprintf("error occurred while copying file or directory at path: %v", fileArtifactPath))
		}
	}
	return nil
}

// Copied some of the command from: exec_recipe.ResultMapToString
// TODO: create a utility method that can be used by add_service(s) and run_sh method.
func resultMapToString(resultMap map[string]starlark.Comparable) string {
	exitCode := resultMap[runshCodeKey]
	rawOutput := resultMap[runshOutputKey]

	outputStarlarkStr, ok := rawOutput.(starlark.String)
	if !ok {
		logrus.Errorf("Result of run_sh was not a string (was: '%v' of type '%s'). This is not fatal but the object might be malformed in CLI output. It is very unexpected and hides a Kurtosis internal bug. This issue should be reported", rawOutput, reflect.TypeOf(rawOutput))
		outputStarlarkStr = starlark.String(outputStarlarkStr.String())
	}
	outputStr := outputStarlarkStr.GoString()
	if outputStr == "" {
		return fmt.Sprintf("Command returned with exit code '%v' with no output", exitCode)
	}
	if strings.Contains(outputStr, newlineChar) {
		return fmt.Sprintf(`Command returned with exit code '%v' and the following output:
--------------------
%v
--------------------`, exitCode, outputStr)
	}
	return fmt.Sprintf("Command returned with exit code '%v' and the following output: %v", exitCode, outputStr)
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

func executeWithWait(ctx context.Context, builtin *RunShCapabilities, commandToRun []string) (*exec_result.ExecResult, error) {
	// Wait is set to None
	if builtin.wait == DisableWaitTimeoutDurationStr {
		return builtin.serviceNetwork.RunExec(ctx, builtin.name, commandToRun)
	}

	resultChan := make(chan *exec_result.ExecResult, 1)
	errChan := make(chan error, 1)

	timoutStr := builtin.wait
	parsedTimeout, parseErr := time.ParseDuration(timoutStr)
	if parseErr != nil {
		return nil, startosis_errors.WrapWithInterpretationError(parseErr, "an error occurred when parsing timeout '%v'", timoutStr)
	}

	timeDuration := time.After(parsedTimeout)
	contextWithDeadline, cancelContext := context.WithTimeout(ctx, parsedTimeout)
	defer cancelContext()

	go func() {
		executionResult, err := builtin.serviceNetwork.RunExec(contextWithDeadline, builtin.name, commandToRun)
		if err != nil {
			errChan <- err
		} else {
			resultChan <- executionResult
		}
	}()

	select {
	case result := <-resultChan:
		return result, nil
	case err := <-errChan:
		return nil, err
	case <-timeDuration: // Timeout duration
		return nil, stacktrace.NewError("The exec request timed out after %v seconds", parsedTimeout.Seconds())
	}
}

func validatePathIsUniqueWhileCreatingFileArtifact(storeFiles []string) *startosis_errors.ValidationError {
	if len(storeFiles) > 0 {
		duplicates := map[string]uint16{}
		for _, filePath := range storeFiles {
			if duplicates[filePath] != 0 {
				return startosis_errors.NewValidationError(
					"error occurred while validating field: %v. The file paths in the array must be unique. Found multiple instances of %v", StoreFilesName, filePath)
			}
			duplicates[filePath] = 1
		}
	}
	return nil
}
