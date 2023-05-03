package binding_constructors

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
)

// The generated bindings don't come with constructors (leaving it up to the user to initialize all the fields), so we
// add them so that our code is safer

// ==============================================================================================
//
//	Shared Objects (Used By Multiple Endpoints)
//
// ==============================================================================================

func NewPort(
	number uint32,
	protocol kurtosis_core_rpc_api_bindings.Port_TransportProtocol,
	maybeApplicationProtocol string,
	maybeWaitTimeout string,
) *kurtosis_core_rpc_api_bindings.Port {
	return &kurtosis_core_rpc_api_bindings.Port{
		Number:                   number,
		TransportProtocol:        protocol,
		MaybeApplicationProtocol: maybeApplicationProtocol,
		MaybeWaitTimeout:         maybeWaitTimeout,
	}
}

func NewServiceConfig(
	containerImageName string,
	privatePorts map[string]*kurtosis_core_rpc_api_bindings.Port,
	publicPorts map[string]*kurtosis_core_rpc_api_bindings.Port, //TODO this is a huge hack to temporarily enable static ports for NEAR until we have a more productized solution
	entrypointArgs []string,
	cmdArgs []string,
	envVars map[string]string,
	filesArtifactMountDirpaths map[string]string,
	cpuAllocationMillicpus uint64,
	memoryAllocationMegabytes uint64,
	privateIPAddrPlaceholder string,
	subnetwork string) *kurtosis_core_rpc_api_bindings.ServiceConfig {
	return &kurtosis_core_rpc_api_bindings.ServiceConfig{
		ContainerImageName:        containerImageName,
		PrivatePorts:              privatePorts,
		PublicPorts:               publicPorts,
		EntrypointArgs:            entrypointArgs,
		CmdArgs:                   cmdArgs,
		EnvVars:                   envVars,
		FilesArtifactMountpoints:  filesArtifactMountDirpaths,
		CpuAllocationMillicpus:    cpuAllocationMillicpus,
		MemoryAllocationMegabytes: memoryAllocationMegabytes,
		PrivateIpAddrPlaceholder:  privateIPAddrPlaceholder,
		Subnetwork:                &subnetwork,
	}
}

func NewUpdateServiceConfig(subnetwork string) *kurtosis_core_rpc_api_bindings.UpdateServiceConfig {
	return &kurtosis_core_rpc_api_bindings.UpdateServiceConfig{
		Subnetwork: &subnetwork,
	}
}

// ==============================================================================================
//
//	Execute Starlark Arguments
//
// ==============================================================================================
func NewRunStarlarkScriptArgs(serializedString string, serializedParams string, dryRun bool, parallelism int32) *kurtosis_core_rpc_api_bindings.RunStarlarkScriptArgs {
	parallelismCopy := new(int32)
	*parallelismCopy = parallelism
	return &kurtosis_core_rpc_api_bindings.RunStarlarkScriptArgs{
		SerializedScript: serializedString,
		SerializedParams: serializedParams,
		DryRun:           &dryRun,
		Parallelism:      parallelismCopy,
	}
}

func NewRunStarlarkPackageArgs(packageId string, serializedParams string, dryRun bool, parallelism int32) *kurtosis_core_rpc_api_bindings.RunStarlarkPackageArgs {
	parallelismCopy := new(int32)
	*parallelismCopy = parallelism
	clonePackage := false
	return &kurtosis_core_rpc_api_bindings.RunStarlarkPackageArgs{
		PackageId:              packageId,
		ClonePackage:           &clonePackage,
		StarlarkPackageContent: nil,
		SerializedParams:       serializedParams,
		DryRun:                 &dryRun,
		Parallelism:            parallelismCopy,
	}
}

func NewRunStarlarkRemotePackageArgs(packageId string, serializedParams string, dryRun bool, parallelism int32) *kurtosis_core_rpc_api_bindings.RunStarlarkPackageArgs {
	parallelismCopy := new(int32)
	*parallelismCopy = parallelism
	clonePackage := true
	return &kurtosis_core_rpc_api_bindings.RunStarlarkPackageArgs{
		PackageId:              packageId,
		ClonePackage:           &clonePackage,
		StarlarkPackageContent: nil,
		SerializedParams:       serializedParams,
		DryRun:                 &dryRun,
		Parallelism:            parallelismCopy,
	}
}

// ==============================================================================================
//
//	Startosis Execution Response
//
// ==============================================================================================
func NewStarlarkRunResponseLineFromInstruction(instruction *kurtosis_core_rpc_api_bindings.StarlarkInstruction) *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine {
	return &kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine{
		RunResponseLine: &kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine_Instruction{
			Instruction: instruction,
		},
	}
}

func NewStarlarkRunResponseLineFromWarning(warningMessage string) *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine {
	return &kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine{
		RunResponseLine: &kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine_Warning{
			Warning: &kurtosis_core_rpc_api_bindings.StarlarkWarning{
				WarningMessage: warningMessage,
			},
		},
	}
}

func NewStarlarkRunResponseLineFromInstructionResult(serializedInstructionResult string) *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine {
	return &kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine{
		RunResponseLine: &kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine_InstructionResult{
			InstructionResult: &kurtosis_core_rpc_api_bindings.StarlarkInstructionResult{
				SerializedInstructionResult: serializedInstructionResult,
			},
		},
	}
}

func NewStarlarkRunResponseLineFromInterpretationError(interpretationError *kurtosis_core_rpc_api_bindings.StarlarkInterpretationError) *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine {
	return &kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine{
		RunResponseLine: &kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine_Error{
			Error: &kurtosis_core_rpc_api_bindings.StarlarkError{
				Error: &kurtosis_core_rpc_api_bindings.StarlarkError_InterpretationError{
					InterpretationError: interpretationError,
				},
			},
		},
	}
}

func NewStarlarkRunResponseLineFromValidationError(validationError *kurtosis_core_rpc_api_bindings.StarlarkValidationError) *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine {
	return &kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine{
		RunResponseLine: &kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine_Error{
			Error: &kurtosis_core_rpc_api_bindings.StarlarkError{
				Error: &kurtosis_core_rpc_api_bindings.StarlarkError_ValidationError{
					ValidationError: validationError,
				},
			},
		},
	}
}

func NewStarlarkRunResponseLineFromExecutionError(executionError *kurtosis_core_rpc_api_bindings.StarlarkExecutionError) *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine {
	return &kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine{
		RunResponseLine: &kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine_Error{
			Error: &kurtosis_core_rpc_api_bindings.StarlarkError{
				Error: &kurtosis_core_rpc_api_bindings.StarlarkError_ExecutionError{
					ExecutionError: executionError,
				},
			},
		},
	}
}

func NewStarlarkRunResponseLineFromSinglelineProgressInfo(currentStepInfo string, currentStepNumber uint32, totalSteps uint32) *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine {
	return &kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine{
		RunResponseLine: &kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine_ProgressInfo{
			ProgressInfo: &kurtosis_core_rpc_api_bindings.StarlarkRunProgress{
				CurrentStepInfo:   []string{currentStepInfo},
				TotalSteps:        totalSteps,
				CurrentStepNumber: currentStepNumber,
			},
		},
	}
}

func NewStarlarkRunResponseLineFromMultilineProgressInfo(currentStepInfoMultiline []string, currentStepNumber uint32, totalSteps uint32) *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine {
	return &kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine{
		RunResponseLine: &kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine_ProgressInfo{
			ProgressInfo: &kurtosis_core_rpc_api_bindings.StarlarkRunProgress{
				CurrentStepInfo:   currentStepInfoMultiline,
				TotalSteps:        totalSteps,
				CurrentStepNumber: currentStepNumber,
			},
		},
	}
}

func NewStarlarkRunResponseLineFromRunFailureEvent() *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine {
	return &kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine{
		RunResponseLine: &kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine_RunFinishedEvent{
			RunFinishedEvent: &kurtosis_core_rpc_api_bindings.StarlarkRunFinishedEvent{
				IsRunSuccessful:  false,
				SerializedOutput: nil,
			},
		},
	}
}

func NewStarlarkRunResponseLineFromRunSuccessEvent(serializedOutputObject string) *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine {
	return &kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine{
		RunResponseLine: &kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine_RunFinishedEvent{
			RunFinishedEvent: &kurtosis_core_rpc_api_bindings.StarlarkRunFinishedEvent{
				IsRunSuccessful:  true,
				SerializedOutput: &serializedOutputObject,
			},
		},
	}
}

func NewStarlarkInstruction(position *kurtosis_core_rpc_api_bindings.StarlarkInstructionPosition, name string, executableInstruction string, arguments []*kurtosis_core_rpc_api_bindings.StarlarkInstructionArg) *kurtosis_core_rpc_api_bindings.StarlarkInstruction {
	return &kurtosis_core_rpc_api_bindings.StarlarkInstruction{
		InstructionName:       name,
		Position:              position,
		ExecutableInstruction: executableInstruction,
		Arguments:             arguments,
	}
}

func NewStarlarkInstructionPosition(filename string, line int32, column int32) *kurtosis_core_rpc_api_bindings.StarlarkInstructionPosition {
	return &kurtosis_core_rpc_api_bindings.StarlarkInstructionPosition{
		Filename: filename,
		Line:     line,
		Column:   column,
	}
}

func NewStarlarkInstructionKwarg(serializedArgValue string, argName string, isRepresentative bool) *kurtosis_core_rpc_api_bindings.StarlarkInstructionArg {
	return &kurtosis_core_rpc_api_bindings.StarlarkInstructionArg{
		SerializedArgValue: serializedArgValue,
		ArgName:            &argName,
		IsRepresentative:   isRepresentative,
	}
}

func NewStarlarkInstructionArg(serializedArgValue string, isRepresentative bool) *kurtosis_core_rpc_api_bindings.StarlarkInstructionArg {
	return &kurtosis_core_rpc_api_bindings.StarlarkInstructionArg{
		SerializedArgValue: serializedArgValue,
		ArgName:            nil,
		IsRepresentative:   isRepresentative,
	}
}

func NewStarlarkInterpretationError(errorMessage string) *kurtosis_core_rpc_api_bindings.StarlarkInterpretationError {
	return &kurtosis_core_rpc_api_bindings.StarlarkInterpretationError{
		ErrorMessage: errorMessage,
	}
}

func NewStarlarkValidationError(errorMessage string) *kurtosis_core_rpc_api_bindings.StarlarkValidationError {
	return &kurtosis_core_rpc_api_bindings.StarlarkValidationError{
		ErrorMessage: errorMessage,
	}
}

func NewStarlarkExecutionError(errorMessage string) *kurtosis_core_rpc_api_bindings.StarlarkExecutionError {
	return &kurtosis_core_rpc_api_bindings.StarlarkExecutionError{
		ErrorMessage: errorMessage,
	}
}

// ==============================================================================================
//
//	Start Service
//
// ==============================================================================================

func NewStartServicesResponse(
	successfulServicesInfo map[string]*kurtosis_core_rpc_api_bindings.ServiceInfo,
	failedServicesErrors map[string]string) *kurtosis_core_rpc_api_bindings.StartServicesResponse {
	return &kurtosis_core_rpc_api_bindings.StartServicesResponse{
		SuccessfulServiceNameToServiceInfo: successfulServicesInfo,
		FailedServiceNameToError:           failedServicesErrors,
	}
}

// ==============================================================================================
//
//	Get Service Info
//
// ==============================================================================================

func NewGetServicesArgs(serviceIdentifiers map[string]bool) *kurtosis_core_rpc_api_bindings.GetServicesArgs {
	return &kurtosis_core_rpc_api_bindings.GetServicesArgs{
		ServiceIdentifiers: serviceIdentifiers,
	}
}

func NewGetServicesResponse(
	serviceInfo map[string]*kurtosis_core_rpc_api_bindings.ServiceInfo,
) *kurtosis_core_rpc_api_bindings.GetServicesResponse {
	return &kurtosis_core_rpc_api_bindings.GetServicesResponse{
		ServiceInfo: serviceInfo,
	}
}

func NewServiceInfo(
	uuid string,
	name string,
	shortenedUuid string,
	privateIpAddr string,
	privatePorts map[string]*kurtosis_core_rpc_api_bindings.Port,
	maybePublicIpAddr string,
	maybePublicPorts map[string]*kurtosis_core_rpc_api_bindings.Port,
) *kurtosis_core_rpc_api_bindings.ServiceInfo {
	return &kurtosis_core_rpc_api_bindings.ServiceInfo{
		ServiceUuid:       uuid,
		Name:              name,
		ShortenedUuid:     shortenedUuid,
		PrivateIpAddr:     privateIpAddr,
		PrivatePorts:      privatePorts,
		MaybePublicIpAddr: maybePublicIpAddr,
		MaybePublicPorts:  maybePublicPorts,
	}
}

// ==============================================================================================
//
//	Remove Service
//
// ==============================================================================================

func NewRemoveServiceResponse(serviceUuid string) *kurtosis_core_rpc_api_bindings.RemoveServiceResponse {
	return &kurtosis_core_rpc_api_bindings.RemoveServiceResponse{
		ServiceUuid: serviceUuid,
	}
}

// ==============================================================================================
//
//	Exec Command
//
// ==============================================================================================

func NewExecCommandArgs(serviceIdentifier string, commandArgs []string) *kurtosis_core_rpc_api_bindings.ExecCommandArgs {
	return &kurtosis_core_rpc_api_bindings.ExecCommandArgs{
		ServiceIdentifier: serviceIdentifier,
		CommandArgs:       commandArgs,
	}
}

func NewExecCommandResponse(exitCode int32, logOutput string) *kurtosis_core_rpc_api_bindings.ExecCommandResponse {
	return &kurtosis_core_rpc_api_bindings.ExecCommandResponse{
		ExitCode:  exitCode,
		LogOutput: logOutput,
	}
}

// ==============================================================================================
//
//	Upload Files Artifact
//
// ==============================================================================================

func NewUploadFilesArtifactArgs(data []byte, name string) *kurtosis_core_rpc_api_bindings.UploadFilesArtifactArgs {
	return &kurtosis_core_rpc_api_bindings.UploadFilesArtifactArgs{Data: data, Name: name}
}

// ==============================================================================================
//
//	Store Web Files Artifact
//
// ==============================================================================================

func NewStoreWebFilesArtifactArgs(url string, name string) *kurtosis_core_rpc_api_bindings.StoreWebFilesArtifactArgs {
	return &kurtosis_core_rpc_api_bindings.StoreWebFilesArtifactArgs{Url: url, Name: name}
}

// ==============================================================================================
//
//	Download Files Artifact
//
// ==============================================================================================

func DownloadFilesArtifactArgs(fileIdentifier string) *kurtosis_core_rpc_api_bindings.DownloadFilesArtifactArgs {
	return &kurtosis_core_rpc_api_bindings.DownloadFilesArtifactArgs{
		Identifier: fileIdentifier,
	}
}

// ==============================================================================================
//
//	Render Templates To Files Artifact
//
// ==============================================================================================

func NewTemplateAndData(template string, dataAsJson string) *kurtosis_core_rpc_api_bindings.RenderTemplatesToFilesArtifactArgs_TemplateAndData {
	return &kurtosis_core_rpc_api_bindings.RenderTemplatesToFilesArtifactArgs_TemplateAndData{
		Template:   template,
		DataAsJson: dataAsJson,
	}
}

func NewRenderTemplatesToFilesArtifactResponse(filesArtifactUuid string) *kurtosis_core_rpc_api_bindings.RenderTemplatesToFilesArtifactResponse {
	return &kurtosis_core_rpc_api_bindings.RenderTemplatesToFilesArtifactResponse{
		Uuid: filesArtifactUuid,
	}
}
