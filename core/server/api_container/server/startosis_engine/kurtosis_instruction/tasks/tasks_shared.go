package tasks

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service_directory"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/store_spec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers/magic_string_helper"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	store_spec_starlark_type "github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/store_spec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"reflect"
	"strings"
	"time"
)

// shared constants
const (
	ImageNameArgName  = "image"
	RunArgName        = "run"
	StoreFilesArgName = "store"
	WaitArgName       = "wait"
	FilesArgName      = "files"
	EnvVarsArgName    = "env_vars"

	newlineChar = "\n"

	DefaultWaitTimeoutDurationStr = "180s"
	DisableWaitTimeoutDurationStr = ""

	runResultCodeKey     = "code"
	runResultOutputKey   = "output"
	runFilesArtifactsKey = "files_artifacts"

	shellWrapperCommand = "/bin/sh"
	noNameSet           = ""
	uniqueNameGenErrStr = "error occurred while generating unique name for the file artifact"
)

var runTailCommandToPreventContainerToStopOnCreating = []string{"tail", "-f", "/dev/null"}

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

func createInterpretationResult(resultUuid string, storeSpecList []*store_spec.StoreSpec) *starlarkstruct.Struct {
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
	return result
}

func validateTasksCommon(validatorEnvironment *startosis_validator.ValidatorEnvironment, storeSpecList []*store_spec.StoreSpec, serviceDirpathsToArtifactIdentifiers map[string][]string, imageName string) *startosis_errors.ValidationError {
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

	validatorEnvironment.AppendRequiredImagePull(imageName)
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
	image string,
	filesArtifactExpansion *service_directory.FilesArtifactsExpansion,
	envVars *map[string]string,
) (*service.ServiceConfig, error) {
	serviceConfig, err := service.CreateServiceConfig(
		image,
		nil,
		nil,
		nil,
		nil,
		// This make sure that the container does not stop as soon as it starts
		// This only is needed for kubernetes at the moment
		// TODO: Instead of creating a service and running exec commands
		//  we could probably run the command as an entrypoint and retrieve the results as soon as the
		//  command is completed
		runTailCommandToPreventContainerToStopOnCreating,
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
		nil,
	)
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
		return stacktrace.NewError("error occurred while removing task with name %v", serviceName)
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
