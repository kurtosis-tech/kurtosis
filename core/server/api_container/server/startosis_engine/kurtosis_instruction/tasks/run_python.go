package tasks

import (
	"context"
	"fmt"
	"strings"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_build_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_registry_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/nix_build_spec"
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
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/plan_yaml"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const (
	RunPythonBuiltinName = "run_python"

	PythonArgumentsArgName = "args"
	PackagesArgName        = "packages"

	defaultRunPythonImageName = "python:3.11-alpine"

	spaceDelimiter = " "

	pipInstallCmd = "pip install"

	successfulPipRunExitCode = 0

	runPythonDefaultDescription = "Running Python script"
)

var defaultPythonPackages = []string{"bash"}

func NewRunPythonService(
	serviceNetwork service_network.ServiceNetwork,
	runtimeValueStore *runtime_value_store.RuntimeValueStore,
	nonBlockingMode bool,
	packageId string,
	packageContentProvider startosis_packages.PackageContentProvider,
	packageReplaceOptions map[string]string) *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return &kurtosis_plan_instruction.KurtosisPlanInstruction{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: RunPythonBuiltinName,

			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              TaskNameArgName,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
				},
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
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Value],
					Validator:         nil,
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
				{
					Name:              AcceptableCodesArgName,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.List],
					Validator:         nil,
				},
				{
					Name:              SkipCodeCheckArgName,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Bool],
					Validator:         nil,
				},
				{
					Name:              NodeSelectorsArgName,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.Dict],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.StringMappingToString(value, NodeSelectorsArgName)
					},
				},
				{
					Name:              TolerationsArgName,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.List],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.StringMappingToString(value, TolerationsArgName)
					},
				},
			},
		},

		Capabilities: func() kurtosis_plan_instruction.KurtosisPlanInstructionCapabilities {
			return &RunPythonCapabilities{
				serviceNetwork:         serviceNetwork,
				runtimeValueStore:      runtimeValueStore,
				pythonArguments:        nil,
				packages:               nil,
				name:                   "",
				nonBlockingMode:        nonBlockingMode,
				packageId:              packageId,
				packageContentProvider: packageContentProvider,
				packageReplaceOptions:  packageReplaceOptions,
				serviceConfig:          nil, // populated at interpretation time
				run:                    "",  // populated at interpretation time
				resultUuid:             "",  // populated at interpretation time
				storeSpecList:          nil,
				wait:                   DefaultWaitTimeoutDurationStr,
				description:            "",  // populated at interpretation time
				returnValue:            nil, // populated at interpretation time
				skipCodeCheck:          false,
				acceptableCodes:        nil, // populated at interpretation time
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
			NodeSelectorsArgName:   true,
			TolerationsArgName:     true,
		},
	}
}

type RunPythonCapabilities struct {
	runtimeValueStore *runtime_value_store.RuntimeValueStore
	serviceNetwork    service_network.ServiceNetwork

	resultUuid      string
	name            string
	run             string
	nonBlockingMode bool

	pythonArguments []string
	packages        []string

	// fields required for image building
	packageId              string
	packageContentProvider startosis_packages.PackageContentProvider
	packageReplaceOptions  map[string]string

	returnValue     *starlarkstruct.Struct
	serviceConfig   *service.ServiceConfig
	storeSpecList   []*store_spec.StoreSpec
	wait            string
	description     string
	acceptableCodes []int64
	skipCodeCheck   bool
}

func (builtin *RunPythonCapabilities) Interpret(locatorOfModuleInWhichThisBuiltinIsBeingCalled string, arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError) {
	taskName, err := getTaskNameFromArgs(arguments)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to get task name from args.")
	}
	builtin.name = taskName

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

	builtin.packages = defaultPythonPackages
	if arguments.IsSet(PackagesArgName) {
		packagesValue, err := builtin_argument.ExtractArgumentValue[*starlark.List](arguments, PackagesArgName)
		if err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "error occurred while extracting packages information")
		}
		packagesList, sliceParsingErr := kurtosis_types.SafeCastToStringSlice(packagesValue, PackagesArgName)
		if sliceParsingErr != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "error occurred while converting Starlark list of packages to a golang string slice")
		}
		builtin.packages = append(builtin.packages, packagesList...)
	}

	var maybeImageName string
	var maybeImageBuildSpec *image_build_spec.ImageBuildSpec
	var maybeImageRegistrySpec *image_registry_spec.ImageRegistrySpec
	var maybeNixBuildSpec *nix_build_spec.NixBuildSpec
	var interpretationErr *startosis_errors.InterpretationError
	if arguments.IsSet(ImageNameArgName) {
		rawImageVal, err := builtin_argument.ExtractArgumentValue[starlark.Value](arguments, ImageNameArgName)
		if err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract raw image attribute.")
		}

		maybeImageName, maybeImageBuildSpec, maybeImageRegistrySpec, maybeNixBuildSpec, interpretationErr = service_config.ConvertImage(
			rawImageVal,
			locatorOfModuleInWhichThisBuiltinIsBeingCalled,
			builtin.packageId,
			builtin.packageContentProvider,
			builtin.packageReplaceOptions)
		if interpretationErr != nil {
			return nil, startosis_errors.WrapWithInterpretationError(interpretationErr, "An error occurred converting image for run sh.")
		}
	} else {
		maybeImageName = defaultRunPythonImageName
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
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	nodeSelectors, interpretationErr := extractNodeSelectorsIfDefined(arguments)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	tolerations, interpretationErr := extractTolerationsIfDefined(arguments)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	// build a service config from image and files artifacts expansion.
	builtin.serviceConfig, err = getServiceConfig(maybeImageName, maybeImageBuildSpec, maybeImageRegistrySpec, maybeNixBuildSpec, filesArtifactExpansion, envVars, nodeSelectors, tolerations)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "An error occurred creating service config for run python.")
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

	acceptableCodes := defaultAcceptableCodes
	if arguments.IsSet(AcceptableCodesArgName) {
		acceptableCodesValue, err := builtin_argument.ExtractArgumentValue[*starlark.List](arguments, AcceptableCodesArgName)
		if err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%v' argument", acceptableCodes)
		}
		acceptableCodes, err = kurtosis_types.SafeCastToIntegerSlice(acceptableCodesValue)
		if err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to parse '%v' argument", acceptableCodes)
		}
	}
	builtin.acceptableCodes = acceptableCodes

	skipCodeCheck := defaultSkipCodeCheck
	if arguments.IsSet(SkipCodeCheckArgName) {
		skipCodeCheckArgumentValue, err := builtin_argument.ExtractArgumentValue[starlark.Bool](arguments, SkipCodeCheckArgName)
		if err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", SkipCodeCheckArgName)
		}
		skipCodeCheck = bool(skipCodeCheckArgumentValue)
	}
	builtin.skipCodeCheck = skipCodeCheck

	resultUuid, err := builtin.runtimeValueStore.CreateValue()
	if err != nil {
		return nil, startosis_errors.NewInterpretationError("An error occurred while generating UUID for future reference for %v instruction", RunPythonBuiltinName)
	}
	builtin.resultUuid = resultUuid

	builtin.description = builtin_argument.GetDescriptionOrFallBack(arguments, runPythonDefaultDescription)

	builtin.returnValue = createInterpretationResult(resultUuid, builtin.storeSpecList)
	return builtin.returnValue, nil
}

func (builtin *RunPythonCapabilities) Validate(_ *builtin_argument.ArgumentValuesSet, validatorEnvironment *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
	// TODO add validation for python script
	var serviceDirpathsToArtifactIdentifiers map[string][]string
	if builtin.serviceConfig.GetFilesArtifactsExpansion() != nil {
		serviceDirpathsToArtifactIdentifiers = builtin.serviceConfig.GetFilesArtifactsExpansion().ServiceDirpathsToArtifactIdentifiers
	}
	return validateTasksCommon(validatorEnvironment, builtin.storeSpecList, serviceDirpathsToArtifactIdentifiers, builtin.serviceConfig)
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
	fullCommandToRun := getCommandToRunForStreamingLogs(commandToRun)

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

	// throw an error as execution of the command is not a part of acceptable codes
	if !builtin.skipCodeCheck && !isAcceptableCode(builtin.acceptableCodes, result) {
		errorMessage := fmt.Sprintf("Run python returned exit code '%v' that is not part of the acceptable status codes '%v', with output:", result["code"], builtin.acceptableCodes)
		return "", stacktrace.NewError(formatErrorMessage(errorMessage, result["output"].String()))
	}

	if builtin.storeSpecList != nil {
		err = copyFilesFromTask(ctx, builtin.serviceNetwork, builtin.name, builtin.storeSpecList)
		if err != nil {
			return "", stacktrace.Propagate(err, "error occurred while copying files from  a task")
		}
	}

	// If the user indicated not to block on removing services after tasks, don't remove the service.
	// The user will have to remove the task service themselves or it will get cleaned up with Kurtosis clean.
	if !builtin.nonBlockingMode {
		if err = removeService(ctx, builtin.serviceNetwork, builtin.name); err != nil {
			return "", stacktrace.Propagate(err, "attempted to remove the temporary task container but failed")
		}
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

func (builtin *RunPythonCapabilities) UpdatePlan(plan *plan_yaml.PlanYamlGenerator) error {
	err := plan.AddRunPython(builtin.run, builtin.description, builtin.returnValue, builtin.serviceConfig, builtin.storeSpecList, builtin.pythonArguments, builtin.packages)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred updating plan with run python")
	}
	return nil
}

func (builtin *RunPythonCapabilities) Description() string {
	return builtin.description
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
	runEscaped := strings.ReplaceAll(builtin.run, `"`, `\"`)
	if len(argumentsAsString) > 0 {
		return fmt.Sprintf(`python -u -c "%s" %s`, runEscaped, argumentsAsString), nil
	}
	return fmt.Sprintf(`python -u -c "%s"`, runEscaped), nil
}
