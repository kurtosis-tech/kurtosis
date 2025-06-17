package tasks

import (
	"context"
	"fmt"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_build_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_registry_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/nix_build_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service_directory"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/store_spec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/dependency_graph"
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
	RunShBuiltinName = "run_sh"

	defaultRunShImageName  = "badouralix/curl-jq"
	shScriptPrintCharLimit = 80
	runningShScriptPrefix  = "Running sh script"
)

func NewRunShService(
	serviceNetwork service_network.ServiceNetwork,
	runtimeValueStore *runtime_value_store.RuntimeValueStore,
	nonBlockingMode bool,
	packageId string,
	packageContentProvider startosis_packages.PackageContentProvider,
	packageReplaceOptions map[string]string) *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return &kurtosis_plan_instruction.KurtosisPlanInstruction{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: RunShBuiltinName,

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
					Name:              EnvVarsArgName,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.Dict],
					Validator:         nil,
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
			},
		},

		Capabilities: func() kurtosis_plan_instruction.KurtosisPlanInstructionCapabilities {
			return &RunShCapabilities{
				serviceNetwork:         serviceNetwork,
				runtimeValueStore:      runtimeValueStore,
				packageId:              packageId,
				packageReplaceOptions:  packageReplaceOptions,
				packageContentProvider: packageContentProvider,
				name:                   "",
				nonBlockingMode:        nonBlockingMode,
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
			RunArgName:        true,
			ImageNameArgName:  true,
			FilesArgName:      true,
			StoreFilesArgName: true,
			WaitArgName:       true,
			EnvVarsArgName:    true,
		},
	}
}

type RunShCapabilities struct {
	runtimeValueStore *runtime_value_store.RuntimeValueStore
	serviceNetwork    service_network.ServiceNetwork

	resultUuid      string
	name            string
	run             string
	nonBlockingMode bool

	// fields required for image building
	packageId              string
	packageContentProvider startosis_packages.PackageContentProvider
	packageReplaceOptions  map[string]string

	serviceConfig   *service.ServiceConfig
	storeSpecList   []*store_spec.StoreSpec
	returnValue     *starlarkstruct.Struct
	wait            string
	description     string
	acceptableCodes []int64
	skipCodeCheck   bool
}

func (builtin *RunShCapabilities) Interpret(locatorOfModuleInWhichThisBuiltinIsBeingCalled string, arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError) {
	taskName, err := getTaskNameFromArgs(arguments)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to get task name from args.")
	}
	builtin.name = taskName

	runCommand, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, RunArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", RunArgName)
	}
	builtin.run = runCommand.GoString()

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
		maybeImageName = defaultRunShImageName
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
	builtin.serviceConfig, err = getServiceConfig(maybeImageName, maybeImageBuildSpec, maybeImageRegistrySpec, maybeNixBuildSpec, filesArtifactExpansion, envVars)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "An error occurred creating service config using for run sh task.")
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
		return nil, startosis_errors.NewInterpretationError("An error occurred while generating UUID for future reference for %v instruction", RunShBuiltinName)
	}
	builtin.resultUuid = resultUuid

	defaultDescription := runningShScriptPrefix
	if len(builtin.run) < shScriptPrintCharLimit {
		defaultDescription = fmt.Sprintf("%v: `%v`", runningShScriptPrefix, builtin.run)
	}
	builtin.description = builtin_argument.GetDescriptionOrFallBack(arguments, defaultDescription)

	builtin.returnValue = createInterpretationResult(resultUuid, builtin.storeSpecList)
	return builtin.returnValue, nil
}

func (builtin *RunShCapabilities) Validate(_ *builtin_argument.ArgumentValuesSet, validatorEnvironment *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
	// TODO validate bash
	var serviceDirpathsToArtifactIdentifiers map[string][]string
	if builtin.serviceConfig.GetFilesArtifactsExpansion() != nil {
		serviceDirpathsToArtifactIdentifiers = builtin.serviceConfig.GetFilesArtifactsExpansion().ServiceDirpathsToArtifactIdentifiers
	}
	return validateTasksCommon(validatorEnvironment, builtin.storeSpecList, serviceDirpathsToArtifactIdentifiers, builtin.serviceConfig)
}

// Execute This is just v0 for run_sh task - we can later improve on it.
//
//	TODO Create an mechanism for other services to retrieve files from the task container
//	Make task as its own entity instead of currently shown under services
func (builtin *RunShCapabilities) Execute(ctx context.Context, _ *builtin_argument.ArgumentValuesSet) (string, error) {
	// swap env vars with their runtime value
	serviceConfigWithReplacedEnvVars, err := replaceMagicStringsInEnvVars(builtin.runtimeValueStore, builtin.serviceConfig)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred replacing magic strings in env vars.")
	}

	_, err = builtin.serviceNetwork.AddService(ctx, service.ServiceName(builtin.name), serviceConfigWithReplacedEnvVars)
	if err != nil {
		return "", stacktrace.Propagate(err, "error occurred while creating a run_sh task with image: %v", builtin.serviceConfig.GetContainerImageName())
	}

	// create work directory and cd into that directory
	commandToRun, err := getCommandToRun(builtin)
	if err != nil {
		return "", stacktrace.Propagate(err, "error occurred while preparing the sh command to execute on the image")
	}
	fullCommandToRun := getCommandToRunForStreamingLogs(commandToRun)

	// run the command passed in by user in the container
	runShResult, err := executeWithWait(ctx, builtin.serviceNetwork, builtin.name, builtin.wait, fullCommandToRun)
	if err != nil {
		return "", stacktrace.Propagate(err, fmt.Sprintf("error occurred while executing one time task command: %v ", builtin.run))
	}

	result := map[string]starlark.Comparable{
		runResultOutputKey: starlark.String(runShResult.GetOutput()),
		runResultCodeKey:   starlark.MakeInt(int(runShResult.GetExitCode())),
	}

	if err := builtin.runtimeValueStore.SetValue(builtin.resultUuid, result); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred setting value '%+v' using key UUID '%s' in the runtime value store", result, builtin.resultUuid)
	}
	instructionResult := resultMapToString(result, RunShBuiltinName)

	// throw an error as execution if returned code is not an  acceptable exit code
	if !builtin.skipCodeCheck && !isAcceptableCode(builtin.acceptableCodes, result) {
		errorMessage := fmt.Sprintf("Run sh returned exit code '%v' that is not part of the acceptable status codes '%v', with output:", result["code"], builtin.acceptableCodes)
		return "", stacktrace.NewError(formatErrorMessage(errorMessage, result["output"].String()))
	}

	if builtin.storeSpecList != nil {
		err = copyFilesFromTask(ctx, builtin.serviceNetwork, builtin.name, builtin.storeSpecList)
		if err != nil {
			return "", stacktrace.Propagate(err, "error occurred while copying files from  a task")
		}
	}

	// If the user indicated not to block on removing services after tasks, don't remove the service.
	// The user will have to remove the task service themselves, or it will get cleaned up with Kurtosis clean.
	if !builtin.nonBlockingMode {
		if err = removeService(ctx, builtin.serviceNetwork, builtin.name); err != nil {
			return "", stacktrace.Propagate(err, "attempted to remove the temporary task container but failed")
		}
	}

	return instructionResult, err
}

func (builtin *RunShCapabilities) TryResolveWith(instructionsAreEqual bool, _ *enclave_plan_persistence.EnclavePlanInstruction, _ *enclave_structure.EnclaveComponents) enclave_structure.InstructionResolutionStatus {
	if instructionsAreEqual {
		return enclave_structure.InstructionIsEqual
	}
	return enclave_structure.InstructionIsUnknown
}

func (builtin *RunShCapabilities) FillPersistableAttributes(builder *enclave_plan_persistence.EnclavePlanInstructionBuilder) {
	builder.SetType(RunShBuiltinName)
}

func (builtin *RunShCapabilities) UpdatePlan(plan *plan_yaml.PlanYamlGenerator) error {
	err := plan.AddRunSh(builtin.run, builtin.description, builtin.returnValue, builtin.serviceConfig, builtin.storeSpecList)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred adding run sh task to the plan")
	}
	return nil
}

func (builtin *RunShCapabilities) UpdateDependencyGraph(dependencyGraph *dependency_graph.InstructionsDependencyGraph) error {
	// TODO
	return nil
}

func (builtin *RunShCapabilities) Description() string {
	return builtin.description
}

func getCommandToRun(builtin *RunShCapabilities) (string, error) {
	// replace future references to actual strings
	maybeSubCommandWithRuntimeValues, err := magic_string_helper.ReplaceRuntimeValueInString(builtin.run, builtin.runtimeValueStore)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while replacing runtime values in run_sh")
	}

	return maybeSubCommandWithRuntimeValues, nil
}

func replaceMagicStringsInEnvVars(runtimeValueStore *runtime_value_store.RuntimeValueStore, serviceConfig *service.ServiceConfig) (
	*service.ServiceConfig,
	error) {
	var envVars map[string]string
	if serviceConfig.GetEnvVars() != nil {
		envVars = make(map[string]string, len(serviceConfig.GetEnvVars()))
		for envVarName, envVarValue := range serviceConfig.GetEnvVars() {
			envVarValueWithRuntimeValueReplaced, err := magic_string_helper.ReplaceRuntimeValueInString(envVarValue, runtimeValueStore)
			if err != nil {
				return nil, stacktrace.Propagate(err, "Error occurred while replacing runtime value in command args for '%s': '%s'", envVarName, envVarValue)
			}
			envVars[envVarName] = envVarValueWithRuntimeValueReplaced
		}
	}

	renderedServiceConfig, err := service.CreateServiceConfig(serviceConfig.GetContainerImageName(), serviceConfig.GetImageBuildSpec(), serviceConfig.GetImageRegistrySpec(), serviceConfig.GetNixBuildSpec(), serviceConfig.GetPrivatePorts(), serviceConfig.GetPublicPorts(), serviceConfig.GetEntrypointArgs(), serviceConfig.GetCmdArgs(), envVars, serviceConfig.GetFilesArtifactsExpansion(), serviceConfig.GetPersistentDirectories(), serviceConfig.GetCPUAllocationMillicpus(), serviceConfig.GetMemoryAllocationMegabytes(), serviceConfig.GetPrivateIPAddrPlaceholder(), serviceConfig.GetMinCPUAllocationMillicpus(), serviceConfig.GetMinMemoryAllocationMegabytes(), serviceConfig.GetLabels(), serviceConfig.GetUser(), serviceConfig.GetTolerations(), serviceConfig.GetNodeSelectors(), serviceConfig.GetImageDownloadMode(), tiniEnabled)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a service config with env var magric strings replaced.")
	}

	return renderedServiceConfig, nil
}
