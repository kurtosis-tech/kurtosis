package tasks

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"go.starlark.net/starlark"
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

	newlineChar = "\n"

	DefaultWaitTimeoutDurationStr = "180s"
	DisableWaitTimeoutDurationStr = ""

	runResultCodeKey     = "code"
	runResultOutputKey   = "output"
	runFilesArtifactsKey = "files_artifacts"

	shellWrapperCommand = "/bin/sh"
)

var runTailCommandToPreventContainerToStopOnCreating = []string{"tail", "-f", "/dev/null"}

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

func copyFilesFromTask(ctx context.Context, serviceNetwork service_network.ServiceNetwork, serviceName string, fileArtifactNames []string, pathToFileArtifacts []string) error {
	if fileArtifactNames == nil || pathToFileArtifacts == nil {
		return nil
	}

	for index, fileArtifactPath := range pathToFileArtifacts {
		fileArtifactName := fileArtifactNames[index]
		_, err := serviceNetwork.CopyFilesFromService(ctx, serviceName, fileArtifactPath, fileArtifactName)
		if err != nil {
			return stacktrace.Propagate(err, fmt.Sprintf("error occurred while copying file or directory at path: %v", fileArtifactPath))
		}
	}
	return nil
}

// Copied some of the command from: exec_recipe.ResultMapToString
// TODO: create a utility method that can be used by add_service(s) and run_sh method.
func resultMapToString(resultMap map[string]starlark.Comparable) string {
	exitCode := resultMap[runResultCodeKey]
	rawOutput := resultMap[runResultOutputKey]

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
