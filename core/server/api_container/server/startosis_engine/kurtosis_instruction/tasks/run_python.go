package tasks

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/shared_utils"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/exec_result"
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
	"go.starlark.net/starlarkstruct"
	"os"
	"path"
	"strings"
)

const (
	RunPythonBuiltinName = "run_python"

	PythonArgumentsArgName = "args"
	PackagesArgName        = "packages"

	defaultRunPythonImageName = "python:3.11-alpine"

	spaceDelimiter = " "

	pythonScriptFileName = "main.py"
	pythonWorkspace      = "/tmp/python"

	defaultTmpDir                  = ""
	pythonScriptReadPermission     = 0644
	enforceMaxSizeLimit            = true
	temporaryPythonDirectoryPrefix = "run-python-*"

	successfulPipRunExitCode = 0
)

func NewRunPythonService(serviceNetwork service_network.ServiceNetwork, runtimeValueStore *runtime_value_store.RuntimeValueStore) *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return &kurtosis_plan_instruction.KurtosisPlanInstruction{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: RunPythonBuiltinName,

			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              RunArgName,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
				},
				{
					Name:              PythonArgumentsArgName,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.List],
				},
				// TODO figure out how this should handle arguments with spaces between them
				{
					Name:              PackagesArgName,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.List],
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
			return &RunPythonCapabilities{
				serviceNetwork:      serviceNetwork,
				runtimeValueStore:   runtimeValueStore,
				pythonArguments:     nil,
				packages:            nil,
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
			RunArgName:             true,
			PythonArgumentsArgName: true,
			PackagesArgName:        true,
			ImageNameArgName:       true,
			FilesArgName:           true,
			StoreFilesArgName:      true,
			WaitArgName:            true,
		},
	}
}

type RunPythonCapabilities struct {
	runtimeValueStore *runtime_value_store.RuntimeValueStore
	serviceNetwork    service_network.ServiceNetwork

	resultUuid string
	name       string
	run        string

	pythonArguments []string
	packages        []string

	serviceConfig       *service.ServiceConfig
	fileArtifactNames   []string
	pathToFileArtifacts []string
	wait                string
}

func (builtin *RunPythonCapabilities) Interpret(arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError) {
	pythonScript, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, RunArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", RunArgName)
	}
	builtin.run = pythonScript.GoString()

	compressedScript, scriptCompressionInterpretationErr := getCompressedPythonScriptForUpload(builtin.run)
	if err != nil {
		return nil, scriptCompressionInterpretationErr
	}
	uniqueFilesArtifactName, err := builtin.serviceNetwork.GetUniqueNameForFileArtifact()
	if err != nil {
		return nil, startosis_errors.NewInterpretationError("an error occurred while generating unique artifact name for python script")
	}
	_, err = builtin.serviceNetwork.UploadFilesArtifact(compressedScript, uniqueFilesArtifactName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "An error occurred while storing the python script to disk")
	}

	var image string
	if arguments.IsSet(ImageNameArgName) {
		imageStarlark, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, ImageNameArgName)
		if err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", ImageNameArgName)
		}
		image = imageStarlark.GoString()
	} else {
		image = defaultRunPythonImageName
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
			filesArtifactMountDirPaths[pythonWorkspace] = uniqueFilesArtifactName
			filesArtifactExpansion, interpretationErr = service_config.ConvertFilesArtifactsMounts(filesArtifactMountDirPaths, builtin.serviceNetwork)
			if interpretationErr != nil {
				return nil, interpretationErr
			}
		}
	} else {
		filesArtifactMountDirPaths := map[string]string{}
		filesArtifactMountDirPaths[pythonWorkspace] = uniqueFilesArtifactName
		var interpretationErr *startosis_errors.InterpretationError
		filesArtifactExpansion, interpretationErr = service_config.ConvertFilesArtifactsMounts(filesArtifactMountDirPaths, builtin.serviceNetwork)
		if interpretationErr != nil {
			return nil, interpretationErr
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

	if arguments.IsSet(PackagesArgName) {
		packagesValue, err := builtin_argument.ExtractArgumentValue[*starlark.List](arguments, PackagesArgName)
		if err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "error occurred while extracting packages information")
		}
		packagesList, sliceParsingErr := kurtosis_types.SafeCastToStringSlice(packagesValue, PackagesArgName)
		if sliceParsingErr != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "error occurred while converting Starlark list of packages to a golang string slice")
		}
		builtin.packages = packagesList
	}

	if arguments.IsSet(PythonArgumentsArgName) {
		argsValue, err := builtin_argument.ExtractArgumentValue[*starlark.List](arguments, PythonArgumentsArgName)
		if err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "error occurred while extracting passed argument information")
		}
		argsList, sliceParsingErr := kurtosis_types.SafeCastToStringSlice(argsValue, PythonArgumentsArgName)
		if sliceParsingErr != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "error occurred while converting Starlark list of passed arguments to a golang string slice")
		}
		builtin.pythonArguments = argsList
	}

	resultUuid, err := builtin.runtimeValueStore.CreateValue()
	if err != nil {
		return nil, startosis_errors.NewInterpretationError("An error occurred while generating UUID for future reference for %v instruction", RunPythonBuiltinName)
	}
	builtin.resultUuid = resultUuid
	randomUuid := uuid.NewRandom()
	builtin.name = fmt.Sprintf("task-%v", randomUuid.String())

	runPythonCodeValue := fmt.Sprintf(magic_string_helper.RuntimeValueReplacementPlaceholderFormat, builtin.resultUuid, runResultCodeKey)
	runPythonOutputValue := fmt.Sprintf(magic_string_helper.RuntimeValueReplacementPlaceholderFormat, builtin.resultUuid, runResultOutputKey)

	dict := map[string]starlark.Value{}
	dict[runResultCodeKey] = starlark.String(runPythonCodeValue)
	dict[runResultOutputKey] = starlark.String(runPythonOutputValue)

	// converting go slice to starlark list
	artifactNamesList := &starlark.List{}
	if len(builtin.fileArtifactNames) > 0 {
		for _, name := range builtin.fileArtifactNames {
			// purposely not checking error for list because it's mutable so should not throw any errors until this point
			_ = artifactNamesList.Append(starlark.String(name))
		}
	}
	dict[runFilesArtifactsKey] = artifactNamesList
	response := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	return response, nil
}

func (builtin *RunPythonCapabilities) Validate(_ *builtin_argument.ArgumentValuesSet, validatorEnvironment *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
	// TODO add validation for python script
	var serviceDirpathsToArtifactIdentifiers map[string]string
	if builtin.serviceConfig.GetFilesArtifactsExpansion() != nil {
		serviceDirpathsToArtifactIdentifiers = builtin.serviceConfig.GetFilesArtifactsExpansion().ServiceDirpathsToArtifactIdentifiers
	}
	return validateTasksCommon(validatorEnvironment, builtin.fileArtifactNames, builtin.pathToFileArtifacts, serviceDirpathsToArtifactIdentifiers, builtin.serviceConfig.GetContainerImageName())
}

// Execute This is just v0 for run_python task - we can later improve on it.
//
//	 These TODOs are copied from run-sh
//		TODO: stop the container as soon as task completed.
//		Create an mechanism for other services to retrieve files from the task container
//		Make task as its own entity instead of currently shown under services
func (builtin *RunPythonCapabilities) Execute(ctx context.Context, _ *builtin_argument.ArgumentValuesSet) (string, error) {
	_, err := builtin.serviceNetwork.AddService(ctx, service.ServiceName(builtin.name), builtin.serviceConfig)
	if err != nil {
		return "", stacktrace.Propagate(err, "error occurred while creating a run_sh task with image: %v", builtin.serviceConfig.GetContainerImageName())
	}

	pipInstallationResult, err := setupRequiredPackages(ctx, builtin)
	if err != nil {
		return "", stacktrace.Propagate(err, "an error occurred while installing dependencies")
	}

	if pipInstallationResult != nil && pipInstallationResult.GetExitCode() != successfulPipRunExitCode {
		return "", stacktrace.NewError("an error occurred while installing dependencies as pip exited with code '%v' instead of '%v'. The error was:\n%v", pipInstallationResult.GetExitCode(), successfulPipRunExitCode, pipInstallationResult.GetOutput())
	}

	commandToRun, err := getPythonCommandToRun(builtin)
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

func setupRequiredPackages(ctx context.Context, builtin *RunPythonCapabilities) (*exec_result.ExecResult, error) {
	var maybePackagesWithRuntimeValuesReplaced []string
	for _, pythonPackage := range builtin.packages {
		maybePackageWithRuntimeValueReplaced, err := magic_string_helper.ReplaceRuntimeValueInString(pythonPackage, builtin.runtimeValueStore)
		if err != nil {
			return nil, stacktrace.Propagate(err, "an error occurred while replacing runtime value in a package passed to run_python")
		}
		maybePackagesWithRuntimeValuesReplaced = append(maybePackagesWithRuntimeValuesReplaced, maybePackageWithRuntimeValueReplaced)
	}

	if len(maybePackagesWithRuntimeValuesReplaced) == 0 {
		return nil, nil
	}

	packageInstallationSubCommand := fmt.Sprintf("pip install %v", strings.Join(maybePackagesWithRuntimeValuesReplaced, spaceDelimiter))
	packageInstallationCommand := []string{shellWrapperCommand, "-c", packageInstallationSubCommand}

	executionResult, err := builtin.serviceNetwork.RunExec(
		ctx,
		builtin.name,
		packageInstallationCommand,
	)

	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while installing required dependencies")
	}

	return executionResult, nil
}

func getPythonCommandToRun(builtin *RunPythonCapabilities) (string, error) {
	var maybePythonArgumentsWithRuntimeValueReplaced []string
	for _, pythonArgument := range builtin.pythonArguments {
		maybePythonArgumentWithRuntimeValueReplaced, err := magic_string_helper.ReplaceRuntimeValueInString(pythonArgument, builtin.runtimeValueStore)
		if err != nil {
			return "", stacktrace.Propagate(err, "an error occurred while replacing runtime value in a python arg to run_python")
		}
		maybePythonArgumentsWithRuntimeValueReplaced = append(maybePythonArgumentsWithRuntimeValueReplaced, maybePythonArgumentWithRuntimeValueReplaced)
	}
	argumentsAsString := strings.Join(maybePythonArgumentsWithRuntimeValueReplaced, spaceDelimiter)

	pythonScriptAbsolutePath := path.Join(pythonWorkspace, pythonScriptFileName)
	if len(argumentsAsString) > 0 {
		return fmt.Sprintf("python %s %s", pythonScriptAbsolutePath, argumentsAsString), nil
	}
	return fmt.Sprintf("python %s", pythonScriptAbsolutePath), nil
}

func getCompressedPythonScriptForUpload(pythonScript string) ([]byte, *startosis_errors.InterpretationError) {
	temporaryPythonScriptDir, err := os.MkdirTemp(defaultTmpDir, temporaryPythonDirectoryPrefix)
	defer os.Remove(temporaryPythonScriptDir)
	if err != nil {
		return nil, startosis_errors.NewInterpretationError("an error occurred while creating a temporary folder to write the python script too")
	}
	pythonScriptFilePath := path.Join(temporaryPythonScriptDir, pythonScriptFileName)
	if err = os.WriteFile(pythonScriptFilePath, []byte(pythonScript), pythonScriptReadPermission); err != nil {
		return nil, startosis_errors.NewInterpretationError("an error occurred while writing python script to disk")
	}
	compressed, err := shared_utils.CompressPath(pythonScriptFilePath, enforceMaxSizeLimit)
	if err != nil {
		return nil, startosis_errors.NewInterpretationError("an error occurred while compressing the python script")
	}
	return compressed, nil
}
