package tasks

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service_directory"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/store_spec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_plan_persistence"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_structure"
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
	RunPythonBuiltinName = "run_python"

	PythonArgumentsArgName = "args"
	PackagesArgName        = "packages"

	defaultRunPythonImageName = "python:3.11-alpine"

	spaceDelimiter = " "

	pipInstallCmd = "pip install"

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
				serviceNetwork:    serviceNetwork,
				runtimeValueStore: runtimeValueStore,
				pythonArguments:   nil,
				packages:          nil,
				name:              "",
				serviceConfig:     nil, // populated at interpretation time
				run:               "",  // populated at interpretation time
				resultUuid:        "",  // populated at interpretation time
				storeSpecList:     nil,
				wait:              DefaultWaitTimeoutDurationStr,
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

	serviceConfig *service.ServiceConfig
	storeSpecList []*store_spec.StoreSpec
	wait          string
}

func (builtin *RunPythonCapabilities) Interpret(_ string, arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError) {
	randomUuid := uuid.NewRandom()
	builtin.name = fmt.Sprintf("task-%v", randomUuid.String())

	pythonScript, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, RunArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", RunArgName)
	}
	builtin.run = pythonScript.GoString()

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

	var filesArtifactExpansion *service_directory.FilesArtifactsExpansion
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
			multipleFilesArtifactsMountDirPaths := map[string][]string{}
			for pathToFile, fileArtifactName := range filesArtifactMountDirPaths {
				multipleFilesArtifactsMountDirPaths[pathToFile] = []string{fileArtifactName}
			}
			filesArtifactExpansion, interpretationErr = service_config.ConvertFilesArtifactsMounts(multipleFilesArtifactsMountDirPaths, builtin.serviceNetwork)
			if interpretationErr != nil {
				return nil, interpretationErr
			}
		}
	}

	envVars, interpretationErr := extractEnvVarsIfDefined(arguments)
	if err != nil {
		return nil, interpretationErr
	}

	// build a service config from image and files artifacts expansion.
	builtin.serviceConfig, err = getServiceConfig(image, filesArtifactExpansion, envVars)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "An error occurred creating service config using image '%s'", image)
	}

	if arguments.IsSet(StoreFilesArgName) {
		storeSpecList, interpretationErr := parseStoreFilesArg(builtin.serviceNetwork, arguments)
		if interpretationErr != nil {
			return nil, interpretationErr
		}
		builtin.storeSpecList = storeSpecList
	}

	if arguments.IsSet(WaitArgName) {
		waitTimeout, interpretationErr := parseWaitArg(arguments)
		if interpretationErr != nil {
			return nil, interpretationErr
		}
		builtin.wait = waitTimeout
	}

	resultUuid, err := builtin.runtimeValueStore.CreateValue()
	if err != nil {
		return nil, startosis_errors.NewInterpretationError("An error occurred while generating UUID for future reference for %v instruction", RunPythonBuiltinName)
	}
	builtin.resultUuid = resultUuid

	result := createInterpretationResult(resultUuid, builtin.storeSpecList)
	return result, nil
}

func (builtin *RunPythonCapabilities) Validate(_ *builtin_argument.ArgumentValuesSet, validatorEnvironment *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
	// TODO add validation for python script
	var serviceDirpathsToArtifactIdentifiers map[string][]string
	if builtin.serviceConfig.GetFilesArtifactsExpansion() != nil {
		serviceDirpathsToArtifactIdentifiers = builtin.serviceConfig.GetFilesArtifactsExpansion().ServiceDirpathsToArtifactIdentifiers
	}
	return validateTasksCommon(validatorEnvironment, builtin.storeSpecList, serviceDirpathsToArtifactIdentifiers, builtin.serviceConfig.GetContainerImageName())
}

// Execute This is just v0 for run_python task - we can later improve on it.
//
//	TODO Create an mechanism for other services to retrieve files from the task container
//	TODO Make task as its own entity instead of currently shown under services
func (builtin *RunPythonCapabilities) Execute(ctx context.Context, _ *builtin_argument.ArgumentValuesSet) (string, error) {
	_, err := builtin.serviceNetwork.AddService(ctx, service.ServiceName(builtin.name), builtin.serviceConfig)
	if err != nil {
		return "", stacktrace.Propagate(err, "error occurred while creating a run_python task with image: %v", builtin.serviceConfig.GetContainerImageName())
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
	runPythonExecutionResult, err := executeWithWait(ctx, builtin.serviceNetwork, builtin.name, builtin.wait, fullCommandToRun)
	if err != nil {
		return "", stacktrace.Propagate(err, fmt.Sprintf("error occurred while executing one time task command: %v ", builtin.run))
	}

	result := map[string]starlark.Comparable{
		runResultOutputKey: starlark.String(runPythonExecutionResult.GetOutput()),
		runResultCodeKey:   starlark.MakeInt(int(runPythonExecutionResult.GetExitCode())),
	}

	if err := builtin.runtimeValueStore.SetValue(builtin.resultUuid, result); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred setting value '%+v' using key UUID '%s' in the runtime value store", result, builtin.resultUuid)
	}
	instructionResult := resultMapToString(result, RunPythonBuiltinName)

	// throw an error as execution of the command failed
	if runPythonExecutionResult.GetExitCode() != 0 {
		errorMessage := fmt.Sprintf("Python command: %q exited with code %d and output", commandToRun, runPythonExecutionResult.GetExitCode())
		return "", stacktrace.NewError(formatErrorMessage(errorMessage, runPythonExecutionResult.GetOutput()))
	}

	if builtin.storeSpecList != nil {
		err = copyFilesFromTask(ctx, builtin.serviceNetwork, builtin.name, builtin.storeSpecList)
		if err != nil {
			return "", stacktrace.Propagate(err, "error occurred while copying files from  a task")
		}
	}

	if err = removeService(ctx, builtin.serviceNetwork, builtin.name); err != nil {
		return "", stacktrace.Propagate(err, "attempted to remove the temporary task container but failed")
	}

	return instructionResult, err
}

func (builtin *RunPythonCapabilities) TryResolveWith(instructionsAreEqual bool, _ *enclave_plan_persistence.EnclavePlanInstruction, _ *enclave_structure.EnclaveComponents) enclave_structure.InstructionResolutionStatus {
	if instructionsAreEqual {
		return enclave_structure.InstructionIsEqual
	}
	return enclave_structure.InstructionIsUnknown
}

func (builtin *RunPythonCapabilities) FillPersistableAttributes(builder *enclave_plan_persistence.EnclavePlanInstructionBuilder) {
	builder.SetType(RunPythonBuiltinName)
}

func setupRequiredPackages(ctx context.Context, builtin *RunPythonCapabilities) (*exec_result.ExecResult, error) {
	if len(builtin.packages) == 0 {
		return nil, nil
	}

	packageInstallationSubCommand := fmt.Sprintf("%v %v", pipInstallCmd, strings.Join(builtin.packages, spaceDelimiter))
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

	if len(argumentsAsString) > 0 {
		return fmt.Sprintf("python -c '%s' %s", builtin.run, argumentsAsString), nil
	}
	return fmt.Sprintf("python -c '%s'", builtin.run), nil
}
