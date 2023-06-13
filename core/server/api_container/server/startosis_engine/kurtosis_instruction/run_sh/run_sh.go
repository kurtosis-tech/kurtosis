package run_sh

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
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
	"reflect"
	"strings"
)

const (
	RunShBuiltinName = "run_sh"

	ImageNameArgName = "image"
	WorkDirArgName   = "workdir"
	RunArgName       = "run"

	DefaultWorkDir   = "task"
	DefaultImageName = "badouralix/curl-jq"
	FilesAttr        = "files"

	runshCodeKey   = "code"
	runshOutputKey = "output"
	newlineChar    = "\n"

	bashCommand = "/bin/sh"

	createAndSwitchDirectoryTemplate = "mkdir -p %v && cd %v"
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
				},
				{
					Name:              WorkDirArgName,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
				},
				{
					Name:              FilesAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.Dict],
				},
			},
		},

		Capabilities: func() kurtosis_plan_instruction.KurtosisPlanInstructionCapabilities {
			return &RunShCapabilities{
				serviceNetwork:    serviceNetwork,
				runtimeValueStore: runtimeValueStore,
				name:              "",
				image:             DefaultImageName, // populated at interpretation time
				run:               "",               // populated at interpretation time
				workdir:           DefaultWorkDir,   // populated at interpretation time
				files:             nil,
				resultUuid:        "", // populated at interpretation time
			}
		},

		DefaultDisplayArguments: map[string]bool{
			RunArgName:       true,
			ImageNameArgName: true,
			WorkDirArgName:   true,
			FilesAttr:        true,
		},
	}
}

type RunShCapabilities struct {
	runtimeValueStore *runtime_value_store.RuntimeValueStore
	serviceNetwork    service_network.ServiceNetwork

	resultUuid string
	name       string
	run        string
	image      string
	workdir    string
	files      map[string]string
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

	if arguments.IsSet(WorkDirArgName) {
		workDirStarlark, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, WorkDirArgName)
		if err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", WorkDirArgName)
		}
		builtin.workdir = workDirStarlark.GoString()
	}

	if arguments.IsSet(FilesAttr) {
		filesStarlark, err := builtin_argument.ExtractArgumentValue[*starlark.Dict](arguments, FilesAttr)
		if err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", FilesAttr)
		}
		if filesStarlark.Len() > 0 {
			filesArtifactMountDirpaths, interpretationErr := kurtosis_types.SafeCastToMapStringString(filesStarlark, FilesAttr)
			if interpretationErr != nil {
				return nil, interpretationErr
			}
			builtin.files = filesArtifactMountDirpaths
		}
	}

	resultUuid, err := builtin.runtimeValueStore.CreateValue()
	if err != nil {
		return nil, startosis_errors.NewInterpretationError("An error occurred while generating UUID for future reference for %v instruction", RunShBuiltinName)
	}
	builtin.resultUuid = resultUuid
	randomUuid := uuid.NewRandom()
	builtin.name = fmt.Sprintf("task-%v", randomUuid.String())

	runShCodeValue := fmt.Sprintf(magic_string_helper.RuntimeValueReplacementPlaceholderFormat, builtin.resultUuid, runshCodeKey)
	dict := &starlark.Dict{}
	if err := dict.SetKey(starlark.String(runshCodeKey), starlark.String(runShCodeValue)); err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "An error happened while creating run_sh return value, setting field '%v'", runshCodeKey)
	}

	runShOutputValue := fmt.Sprintf(magic_string_helper.RuntimeValueReplacementPlaceholderFormat, builtin.resultUuid, runshOutputKey)
	if err := dict.SetKey(starlark.String(runshOutputKey), starlark.String(runShOutputValue)); err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "An error happened while creating run_sh return value, setting field '%v'", runshOutputKey)
	}
	dict.Freeze()
	return dict, nil
}

func (builtin *RunShCapabilities) Validate(_ *builtin_argument.ArgumentValuesSet, validatorEnvironment *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
	return nil
}

// Execute This is just v0 for run_sh task - we can later improve on it.
//	TODO: stop the container as soon as task completed.
//   Create an mechanism for other services to retrieve files from the task container
//   Make task as it's own entity instead of currently shown under services
func (builtin *RunShCapabilities) Execute(ctx context.Context, _ *builtin_argument.ArgumentValuesSet) (string, error) {
	// create work directory and cd into that directory
	createAndSwitchTheDirectoryCmd := fmt.Sprintf(createAndSwitchDirectoryTemplate, builtin.workdir, builtin.workdir)

	// replace future references to actual strings
	maybeSubCommandWithRuntimeValues, err := magic_string_helper.ReplaceRuntimeValueInString(builtin.run, builtin.runtimeValueStore)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while replacing runtime values in run_sh")
	}

	commandWithNoNewLines := strings.ReplaceAll(maybeSubCommandWithRuntimeValues, newlineChar, " ")

	completeRunCommand := fmt.Sprintf("%v && %v", createAndSwitchTheDirectoryCmd, commandWithNoNewLines)
	createDefaultDirectory := []string{bashCommand, "-c", completeRunCommand}
	serviceConfigBuilder := services.NewServiceConfigBuilder(builtin.image)
	serviceConfigBuilder.WithFilesArtifactMountDirpaths(builtin.files)
	// This make sure that the container does not stop as soon as it starts
	// This only is needed for kubernetes at the moment
	// TODO: Instead of creating a service and running exec commands
	//  we could probably run the command as an entrypoint and retrieve the results as soon as the
	//  command is completed
	serviceConfigBuilder.WithCmdArgs(runTailCommandToPreventContainerToStopOnCreating)

	serviceConfig := serviceConfigBuilder.Build()
	_, err = builtin.serviceNetwork.AddService(ctx, service.ServiceName(builtin.name), serviceConfig)

	if err != nil {
		return "", stacktrace.Propagate(err, fmt.Sprintf("error occurred while creating a run_sh task with image: %v", builtin.image))
	}

	// run the command passed in by user in the container
	code, output, err := builtin.serviceNetwork.ExecCommand(ctx, builtin.name, createDefaultDirectory)
	if err != nil {
		return "", stacktrace.Propagate(err, fmt.Sprintf("error occurred while executing one time task command: %v ", builtin.run))
	}

	result := map[string]starlark.Comparable{
		runshOutputKey: starlark.String(output),
		runshCodeKey:   starlark.MakeInt(int(code)),
	}
	builtin.runtimeValueStore.SetValue(builtin.resultUuid, result)

	instructionResult := resultMapToString(result)
	return instructionResult, err
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
