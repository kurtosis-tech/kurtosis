package tasks

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/image_build_spec"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/image_download_mode"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/image_registry_spec"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/nix_build_spec"
	"github.com/xtgo/uuid"
	v1 "k8s.io/api/core/v1"

	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/service_directory"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/store_spec"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/service_network"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers/magic_string_helper"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	store_spec_starlark_type "github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/store_spec"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// shared constants
const (
	ImageNameArgName       = "image"
	TaskNameArgName        = "name"
	RunArgName             = "run"
	StoreFilesArgName      = "store"
	WaitArgName            = "wait"
	FilesArgName           = "files"
	EnvVarsArgName         = "env_vars"
	AcceptableCodesArgName = "acceptable_codes"
	SkipCodeCheckArgName   = "skip_code_check"
	NodeSelectorsArgName   = "node_selectors"
	TolerationsArgName     = "tolerations"
	defaultSkipCodeCheck   = false

	newlineChar = "\n"

	DefaultWaitTimeoutDurationStr = "180s"
	DisableWaitTimeoutDurationStr = ""

	runResultCodeKey     = "code"
	runResultOutputKey   = "output"
	runFilesArtifactsKey = "files_artifacts"

	shellWrapperCommand = "/bin/sh"
	taskLogFilePath     = "/tmp/kurtosis-task.log"
	noNameSet           = ""
	uniqueNameGenErrStr = "error occurred while generating unique name for the file artifact"

	//  enables init mode on containers; cleaning up any zombie processes
	tiniEnabled = true
)

var defaultAcceptableCodes = []int64{
	0, // EXIT_SUCCESS
}

// runCommandToStreamTaskLogs sets the entrypoint of a task container with a command that creates and tails a log file for tasks run on the container
// all tasks redirect output to the task log file (see getCommandToRunForStreamingLogs for details) where they will be picked up by the main process
// and streamed to stdout via tail -F
// By sending to stdout, task output get picked up by our logging infrastructure - making task logs available via kurtosis service logs
var runCommandToStreamTaskLogs = []string{shellWrapperCommand, "-c", fmt.Sprintf("touch %s && tail -F %s", taskLogFilePath, taskLogFilePath)}

// Wraps [commandToRun] to enable streaming logs from tasks.
// This command is crafted carefully to allow outputting logs to task log file, outputting to stdout, and retaining the exit code from [commandToRun]
// Solution is adapted from 3rd answer in this stack exchange post: https://unix.stackexchange.com/questions/14270/get-exit-status-of-process-thats-piped-to-another. Read detailed explanation of command.
// compared to solution in post, an extra echo >> is added to add a newline after all logs are processed, this is a hack to make sure the last log line gets processed runCommandToStreamTaskLogs
func getCommandToRunForStreamingLogs(commandToRun string) []string {
	fullCmd := []string{shellWrapperCommand, "-c", fmt.Sprintf("{ { { { %v; echo $? >&3; } | tee %v >&4; echo >> %v; } 3>&1; } | { read xs; exit $xs; } } 4>&1", commandToRun, taskLogFilePath, taskLogFilePath)}
	return fullCmd
}

func parseStoreFilesArg(serviceNetwork service_network.ServiceNetwork, arguments *builtin_argument.ArgumentValuesSet) ([]*store_spec.StoreSpec, *startosis_errors.InterpretationError) {
	var result []*store_spec.StoreSpec

	storeFilesList, err := builtin_argument.ExtractArgumentValue[*starlark.List](arguments, StoreFilesArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", StoreFilesArgName)
	}

	if storeFilesList.Len() == 0 {
		return nil, nil
	}

	for i := 0; i < storeFilesList.Len(); i++ {
		rawStoreSpec := storeFilesList.Index(i)

		storeSpecObjStarlarkType, isStoreSpecObjStarlarkType := rawStoreSpec.(*store_spec_starlark_type.StoreSpec)
		if isStoreSpecObjStarlarkType {
			storeSpecObj, interpretationErr := storeSpecObjStarlarkType.ToKurtosisType()
			if interpretationErr != nil {
				return nil, startosis_errors.WrapWithInterpretationError(interpretationErr, "an error occurred while converting StoreSpec Starlark type to raw type")
			}
			// is a StoreSpecObj but no name was provided
			if storeSpecObj.GetName() == noNameSet {
				uniqueNameForArtifact, artifactCreationErr := serviceNetwork.GetUniqueNameForFileArtifact()
				if artifactCreationErr != nil {
					return nil, startosis_errors.WrapWithInterpretationError(artifactCreationErr, uniqueNameGenErrStr)
				}
				storeSpecObj.SetName(uniqueNameForArtifact)
			}
			result = append(result, storeSpecObj)
			continue
		}

		// this is a pure string
		storeFilesSrcStr, interpretationErr := kurtosis_types.SafeCastToString(rawStoreSpec, StoreFilesArgName)
		if interpretationErr == nil {
			uniqueNameForArtifact, artifactCreationErr := serviceNetwork.GetUniqueNameForFileArtifact()
			if artifactCreationErr != nil {
				return nil, startosis_errors.WrapWithInterpretationError(artifactCreationErr, uniqueNameGenErrStr)
			}
			storeSpecObj := store_spec.NewStoreSpec(storeFilesSrcStr, uniqueNameForArtifact)
			result = append(result, storeSpecObj)
			continue
		}

		return nil, startosis_errors.NewInterpretationError("Couldn't convert '%v' to StoreSpec type", rawStoreSpec)
	}

	return result, nil
}

func parseWaitArg(arguments *builtin_argument.ArgumentValuesSet) (string, *startosis_errors.InterpretationError) {
	var waitTimeout string
	waitValue, err := builtin_argument.ExtractArgumentValue[starlark.Value](arguments, WaitArgName)
	if err != nil {
		return "", startosis_errors.WrapWithInterpretationError(err, "error occurred while extracting wait information")
	}
	if waitValueStr, ok := waitValue.(starlark.String); ok {
		waitTimeout = waitValueStr.GoString()
	} else if _, ok := waitValue.(starlark.NoneType); ok {
		waitTimeout = DisableWaitTimeoutDurationStr
	}
	return waitTimeout, nil
}

func createInterpretationResult(resultUuid string, storeSpecList []*store_spec.StoreSpec) (*starlarkstruct.Struct, string, string) {
	runCodeValue := fmt.Sprintf(magic_string_helper.RuntimeValueReplacementPlaceholderFormat, resultUuid, runResultCodeKey)
	runOutputValue := fmt.Sprintf(magic_string_helper.RuntimeValueReplacementPlaceholderFormat, resultUuid, runResultOutputKey)

	dict := map[string]starlark.Value{}
	dict[runResultCodeKey] = starlark.String(runCodeValue)
	dict[runResultOutputKey] = starlark.String(runOutputValue)

	// converting go slice to starlark list
	artifactNamesList := &starlark.List{}
	if len(storeSpecList) > 0 {
		for _, storeSpec := range storeSpecList {
			// purposely not checking error for list because it's mutable so should not throw any errors until this point
			_ = artifactNamesList.Append(starlark.String(storeSpec.GetName()))
		}
	}
	dict[runFilesArtifactsKey] = artifactNamesList
	result := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	return result, runCodeValue, runOutputValue
}

func validateTasksCommon(validatorEnvironment *startosis_validator.ValidatorEnvironment, storeSpecList []*store_spec.StoreSpec, serviceDirpathsToArtifactIdentifiers map[string][]string, serviceConfig *service.ServiceConfig) *startosis_errors.ValidationError {
	if storeSpecList != nil {
		if err := validatePathIsUniqueWhileCreatingFileArtifact(storeSpecList); err != nil {
			return startosis_errors.WrapWithValidationError(err, "error occurred while validating file paths to copy into file artifact")
		}

		for _, storeSpec := range storeSpecList {
			validatorEnvironment.AddArtifactName(storeSpec.GetName())
		}
	}

	for _, artifactNames := range serviceDirpathsToArtifactIdentifiers {
		for _, artifactName := range artifactNames {
			if validatorEnvironment.DoesArtifactNameExist(artifactName) == startosis_validator.ComponentNotFound {
				return startosis_errors.NewValidationError("There was an error validating '%s' as artifact name '%s' does not exist", RunPythonBuiltinName, artifactName)
			}
		}
	}

	// add the images to be built here
	if serviceConfig.GetImageBuildSpec() != nil {
		validatorEnvironment.AppendRequiredImageBuild(serviceConfig.GetContainerImageName(), serviceConfig.GetImageBuildSpec())
	} else if serviceConfig.GetImageRegistrySpec() != nil {
		validatorEnvironment.AppendImageToPullWithAuth(serviceConfig.GetContainerImageName(), serviceConfig.GetImageRegistrySpec())
	} else if serviceConfig.GetNixBuildSpec() != nil {
		validatorEnvironment.AppendRequiredNixBuild(serviceConfig.GetContainerImageName(), serviceConfig.GetNixBuildSpec())
	} else {
		validatorEnvironment.AppendRequiredImagePull(serviceConfig.GetContainerImageName())
	}
	return nil

}

func executeWithWait(ctx context.Context, serviceNetwork service_network.ServiceNetwork, serviceName string, wait string, commandToRun []string) (*exec_result.ExecResult, error) {
	// Wait is set to None
	if wait == DisableWaitTimeoutDurationStr {
		return serviceNetwork.RunExec(ctx, serviceName, commandToRun)
	}

	resultChan := make(chan *exec_result.ExecResult, 1)
	errChan := make(chan error, 1)

	timoutStr := wait

	// we validate timeout string during the validation stage so it cannot be invalid at this stage
	parsedTimeout, _ := time.ParseDuration(timoutStr)

	timeDuration := time.After(parsedTimeout)
	contextWithDeadline, cancelContext := context.WithTimeout(ctx, parsedTimeout)
	defer cancelContext()

	go func() {
		executionResult, err := serviceNetwork.RunExec(contextWithDeadline, serviceName, commandToRun)
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

func validatePathIsUniqueWhileCreatingFileArtifact(storeSpecList []*store_spec.StoreSpec) *startosis_errors.ValidationError {
	if len(storeSpecList) > 0 {
		duplicates := map[string]uint16{}
		for _, storeSpec := range storeSpecList {
			filePath := storeSpec.GetSrc()
			if duplicates[filePath] != 0 {
				return startosis_errors.NewValidationError(
					"error occurred while validating field: %v. The file paths in the array must be unique. Found multiple instances of %v", StoreFilesArgName, filePath)
			}
			duplicates[filePath] = 1
		}
	}
	return nil
}

func copyFilesFromTask(ctx context.Context, serviceNetwork service_network.ServiceNetwork, serviceName string, storeSpecList []*store_spec.StoreSpec) error {
	if storeSpecList == nil {
		return nil
	}

	for _, storeSpec := range storeSpecList {
		_, err := serviceNetwork.CopyFilesFromService(ctx, serviceName, storeSpec.GetSrc(), storeSpec.GetName())
		if err != nil {
			return stacktrace.Propagate(err, fmt.Sprintf("error occurred while copying file or directory at path: %v", storeSpec.GetSrc()))
		}
	}
	return nil
}

// Copied some of the command from: exec_recipe.ResultMapToString
// TODO: create a utility method that can be used by add_service(s) and run_sh method.
func resultMapToString(resultMap map[string]starlark.Comparable, builtinNameForLogging string) string {
	exitCode := resultMap[runResultCodeKey]
	rawOutput := resultMap[runResultOutputKey]

	outputStarlarkStr, ok := rawOutput.(starlark.String)
	if !ok {
		logrus.Errorf("Result of %s was not a string (was: '%v' of type '%s'). This is not fatal but the object might be malformed in CLI output. It is very unexpected and hides a Kurtosis internal bug. This issue should be reported", builtinNameForLogging, rawOutput, reflect.TypeOf(rawOutput))
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

func getServiceConfig(
	maybeImageName string,
	maybeImageBuildSpec *image_build_spec.ImageBuildSpec,
	maybeImageRegistrySpec *image_registry_spec.ImageRegistrySpec,
	maybeNixBuildSpec *nix_build_spec.NixBuildSpec,
	filesArtifactExpansion *service_directory.FilesArtifactsExpansion,
	envVars *map[string]string,
	nodeSelectors *map[string]string,
	tolerations []v1.Toleration,
) (*service.ServiceConfig, error) {
	serviceConfig, err := service.CreateServiceConfig(
		maybeImageName,
		maybeImageBuildSpec,
		maybeImageRegistrySpec,
		maybeNixBuildSpec,
		nil,
		nil,
		runCommandToStreamTaskLogs,
		nil,
		*envVars,
		filesArtifactExpansion,
		nil,
		0,
		0,
		service_config.DefaultPrivateIPAddrPlaceholder,
		0,
		0,
		map[string]string{},
		nil,
		tolerations,
		*nodeSelectors,
		image_download_mode.ImageDownloadMode_Missing,
		tiniEnabled,
		false,
		[]string{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating service config")
	}
	return serviceConfig, nil
}

func formatErrorMessage(errorMessage string, errorFromExec string) string {
	splitErrorMessageNewLine := strings.Split(errorFromExec, "\n")
	reformattedErrorMessage := strings.Join(splitErrorMessageNewLine, "\n  ")
	return fmt.Sprintf("%v\n  %v", errorMessage, reformattedErrorMessage)
}

func removeService(ctx context.Context, serviceNetwork service_network.ServiceNetwork, serviceName string) error {
	_, err := serviceNetwork.RemoveService(ctx, serviceName)
	if err != nil {
		return stacktrace.Propagate(err, "error occurred while removing task with name %v", serviceName)
	}
	return nil
}

func extractEnvVarsIfDefined(arguments *builtin_argument.ArgumentValuesSet) (*map[string]string, *startosis_errors.InterpretationError) {
	envVars := map[string]string{}
	if arguments.IsSet(EnvVarsArgName) {
		envVarsStarlark, err := builtin_argument.ExtractArgumentValue[*starlark.Dict](arguments, EnvVarsArgName)
		if err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", EnvVarsArgName)
		}
		if envVarsStarlark != nil && envVarsStarlark.Len() > 0 {
			var interpretationErr *startosis_errors.InterpretationError
			envVars, interpretationErr = kurtosis_types.SafeCastToMapStringString(envVarsStarlark, EnvVarsArgName)
			if interpretationErr != nil {
				return nil, interpretationErr
			}
		}
	}
	return &envVars, nil
}

func extractNodeSelectorsIfDefined(arguments *builtin_argument.ArgumentValuesSet) (*map[string]string, *startosis_errors.InterpretationError) {
	nodeSelectors := map[string]string{}
	if arguments.IsSet(NodeSelectorsArgName) {
		nodeSelectorsStarlark, err := builtin_argument.ExtractArgumentValue[*starlark.Dict](arguments, NodeSelectorsArgName)
		if err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", EnvVarsArgName)
		}
		if nodeSelectorsStarlark != nil && nodeSelectorsStarlark.Len() > 0 {
			var interpretationErr *startosis_errors.InterpretationError
			nodeSelectors, interpretationErr = kurtosis_types.SafeCastToMapStringString(nodeSelectorsStarlark, NodeSelectorsArgName)
			if interpretationErr != nil {
				return nil, interpretationErr
			}
		}
	}
	return &nodeSelectors, nil
}

func extractTolerationsIfDefined(arguments *builtin_argument.ArgumentValuesSet) ([]v1.Toleration, *startosis_errors.InterpretationError) {
	tolerations := []v1.Toleration{}
	if arguments.IsSet(TolerationsArgName) {
		tolerationsStarlark, err := builtin_argument.ExtractArgumentValue[*starlark.List](arguments, TolerationsArgName)
		if err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", TolerationsArgName)
		}
		if tolerationsStarlark != nil && tolerationsStarlark.Len() > 0 {
			var interpretationErr *startosis_errors.InterpretationError
			tolerations, interpretationErr = service_config.ConvertTolerations(tolerationsStarlark)
			if interpretationErr != nil {
				return nil, interpretationErr
			}
		}
	}
	return tolerations, nil
}

func getTaskNameFromArgs(arguments *builtin_argument.ArgumentValuesSet) (string, error) {
	if arguments.IsSet(TaskNameArgName) {
		taskName, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, TaskNameArgName)
		if err != nil {
			return "", startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", TaskNameArgName)
		}
		return taskName.GoString(), nil
	} else {
		randomUuid := uuid.NewRandom()
		return fmt.Sprintf("task-%v", randomUuid.String()), nil
	}
}

func isAcceptableCode(acceptableCodes []int64, recipeResult map[string]starlark.Comparable) bool {
	for _, acceptableCode := range acceptableCodes {
		if recipeResult[runResultCodeKey] == starlark.MakeInt64(acceptableCode) {
			return true
		}
	}
	return false
}
