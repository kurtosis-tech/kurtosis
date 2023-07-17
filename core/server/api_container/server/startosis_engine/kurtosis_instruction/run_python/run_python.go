package run_python

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
	"github.com/sirupsen/logrus"
	"github.com/xtgo/uuid"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"os"
	"path"
	"reflect"
	"strings"
	"time"
)

const (
	RunPythonBuiltinName = "run_python"

	ImageNameArgName       = "image"
	PythonArgumentsArgName = "args"
	PackagesArgName        = "packages"
	RunArgName             = "run"
	StoreFilesArgName      = "store"
	WaitArgName            = "wait"
	FilesArgName           = "files"

	DefaultImageName = "python:3.11-alpine"

	runPythonCodeKey         = "code"
	runPythonOutputKey       = "output"
	runPythonFileArtifactKey = "files_artifacts"
	newlineChar              = "\n"

	shellWrapperCommand = "/bin/sh"

	DefaultWaitTimeoutDurationStr = "180s"
	DisableWaitTimeoutDurationStr = ""

	pythonScriptFileName = "main.py"
	pythonWorkspace      = "/tmp/python"

	defaultTmpDir                  = ""
	pythonScriptReadPermission     = 0644
	enforceMaxSizeLimit            = true
	temporaryPythonDirectoryPrefix = "run-python-*"

	successfulPipRunExitCode = 0
)

var runTailCommandToPreventContainerToStopOnCreating = []string{"tail", "-f", "/dev/null"}

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
		image = DefaultImageName
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
	builtin.serviceConfig = service.NewServiceConfig(
		image,
		nil,
		nil,
		// This make sure that the container does not stop as soon as it starts
		// This only is needed for kubernetes at the moment
		// TODO: Instead of creating a service and running exec commands
		//  we could probably run the command as an entrypoint and retrieve the results as soon as the
		//  command is completed
		runTailCommandToPreventContainerToStopOnCreating,
		nil,
		nil,
		filesArtifactExpansion,
		0,
		0,
		service_config.DefaultPrivateIPAddrPlaceholder,
		0,
		0,
		// TODO: hardcoding subnetwork to default is what we do now but is incorrect as the run_sh might not be able to
		//  reach some services outside of the default subnetwork. It should be re-worked if users want to use that in
		//  conjunction with subnetworks
		service_config.DefaultSubnetwork,
	)

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

	runPythonCodeValue := fmt.Sprintf(magic_string_helper.RuntimeValueReplacementPlaceholderFormat, builtin.resultUuid, runPythonCodeKey)
	runPythonOutputValue := fmt.Sprintf(magic_string_helper.RuntimeValueReplacementPlaceholderFormat, builtin.resultUuid, runPythonOutputKey)

	dict := map[string]starlark.Value{}
	dict[runPythonCodeKey] = starlark.String(runPythonCodeValue)
	dict[runPythonOutputKey] = starlark.String(runPythonOutputValue)

	// converting go slice to starlark list
	artifactNamesList := &starlark.List{}
	if len(builtin.fileArtifactNames) > 0 {
		for _, name := range builtin.fileArtifactNames {
			// purposely not checking error for list because it's mutable so should not throw any errors until this point
			_ = artifactNamesList.Append(starlark.String(name))
		}
	}
	dict[runPythonFileArtifactKey] = artifactNamesList
	response := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	return response, nil
}

func (builtin *RunPythonCapabilities) Validate(_ *builtin_argument.ArgumentValuesSet, validatorEnvironment *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
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

	if builtin.serviceConfig.GetFilesArtifactsExpansion() != nil {
		for _, artifactName := range builtin.serviceConfig.GetFilesArtifactsExpansion().ServiceDirpathsToArtifactIdentifiers {
			if !validatorEnvironment.DoesArtifactNameExist(artifactName) {
				return startosis_errors.NewValidationError("There was an error validating '%s' as artifact name '%s' does not exist", RunPythonBuiltinName, artifactName)
			}
		}
	}
	// TODO(gm) validate Python Script(run)

	validatorEnvironment.AppendRequiredContainerImage(builtin.serviceConfig.GetContainerImageName())
	return nil
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

	if pipInstallationResult.GetExitCode() != successfulPipRunExitCode {
		return "", stacktrace.NewError("an error occurred while installing dependencies as pip exited with code '%v' instead of '%v'. The error was:\n%v", pipInstallationResult.GetExitCode(), successfulPipRunExitCode, pipInstallationResult.GetOutput())
	}

	commandToRun, err := getCommandToRun(builtin)
	if err != nil {
		return "", stacktrace.Propagate(err, "error occurred while preparing the sh command to execute on the image")
	}
	fullCommandToRun := []string{shellWrapperCommand, "-c", commandToRun}

	// run the command passed in by user in the container
	createDefaultDirectoryResult, err := executeWithWait(ctx, builtin, fullCommandToRun)
	if err != nil {
		return "", stacktrace.Propagate(err, fmt.Sprintf("error occurred while executing one time task command: %v ", builtin.run))
	}

	result := map[string]starlark.Comparable{
		runPythonOutputKey: starlark.String(createDefaultDirectoryResult.GetOutput()),
		runPythonCodeKey:   starlark.MakeInt(int(createDefaultDirectoryResult.GetExitCode())),
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

func copyFilesFromTask(ctx context.Context, builtin *RunPythonCapabilities) error {
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
// TODO: create a utility method that can be used by add_service(s), run_python and run_sh method.
func resultMapToString(resultMap map[string]starlark.Comparable) string {
	exitCode := resultMap[runPythonCodeKey]
	rawOutput := resultMap[runPythonOutputKey]

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

	packageInstallationSubCommand := fmt.Sprintf("pip install %v", strings.Join(maybePackagesWithRuntimeValuesReplaced, " "))
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

func getCommandToRun(builtin *RunPythonCapabilities) (string, error) {
	var maybePythonArgumentsWithRuntimeValueReplaced []string
	for _, pythonArgument := range builtin.pythonArguments {
		maybePythonArgumentWithRuntimeValueReplaced, err := magic_string_helper.ReplaceRuntimeValueInString(pythonArgument, builtin.runtimeValueStore)
		if err != nil {
			return "", stacktrace.Propagate(err, "an error occurred while replacing runtime value in a python arg to run_python")
		}
		maybePythonArgumentsWithRuntimeValueReplaced = append(maybePythonArgumentsWithRuntimeValueReplaced, maybePythonArgumentWithRuntimeValueReplaced)
	}
	argumentsAsString := strings.Join(maybePythonArgumentsWithRuntimeValueReplaced, " ")

	pythonScriptAbsolutePath := path.Join(pythonWorkspace, pythonScriptFileName)
	if len(argumentsAsString) > 0 {
		return fmt.Sprintf("python %s %s", pythonScriptAbsolutePath, argumentsAsString), nil
	}
	return fmt.Sprintf("python %s", pythonScriptAbsolutePath), nil
}

// TODO(gm) put this in utils share with run_sh
func executeWithWait(ctx context.Context, builtin *RunPythonCapabilities, commandToRun []string) (*exec_result.ExecResult, error) {
	// Wait is set to None
	if builtin.wait == DisableWaitTimeoutDurationStr {
		return builtin.serviceNetwork.RunExec(ctx, builtin.name, commandToRun)
	}

	resultChan := make(chan *exec_result.ExecResult, 1)
	errChan := make(chan error, 1)

	timoutStr := builtin.wait

	// we validate timeout string during the validation stage so it cannot be invalid at this stage
	parsedTimeout, _ := time.ParseDuration(timoutStr)

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
					"error occurred while validating field: %v. The file paths in the array must be unique. Found multiple instances of %v", StoreFilesArgName, filePath)
			}
			duplicates[filePath] = 1
		}
	}
	return nil
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
