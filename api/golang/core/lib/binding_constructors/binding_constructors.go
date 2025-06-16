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
		Locked:                   nil,
		Alias:                    nil,
	}
}

// ==============================================================================================
//
//	Execute Starlark Arguments
//
// ==============================================================================================
func NewRunStarlarkScriptArgs(
	mainFunctionName string,
	serializedString string,
	serializedParams string,
	dryRun bool,
	parallelism int32,
	experimentalFeatures []kurtosis_core_rpc_api_bindings.KurtosisFeatureFlag,
	cloudInstanceId string,
	cloudUserId string,
	imageDownloadMode kurtosis_core_rpc_api_bindings.ImageDownloadMode,
	nonBlockingMode bool,
) *kurtosis_core_rpc_api_bindings.RunStarlarkScriptArgs {
	cloudInstanceIdCopy := new(string)
	*cloudInstanceIdCopy = cloudInstanceId
	cloudUserIdCopy := new(string)
	*cloudUserIdCopy = cloudUserId
	imageDownloadModeCopy := new(kurtosis_core_rpc_api_bindings.ImageDownloadMode)
	*imageDownloadModeCopy = imageDownloadMode
	parallelismCopy := new(int32)
	*parallelismCopy = parallelism
	return &kurtosis_core_rpc_api_bindings.RunStarlarkScriptArgs{
		SerializedScript:     serializedString,
		SerializedParams:     &serializedParams,
		DryRun:               &dryRun,
		Parallelism:          parallelismCopy,
		MainFunctionName:     &mainFunctionName,
		ExperimentalFeatures: experimentalFeatures,
		CloudInstanceId:      cloudInstanceIdCopy,
		CloudUserId:          cloudUserIdCopy,
		ImageDownloadMode:    imageDownloadModeCopy,
		NonBlockingMode:      &nonBlockingMode,
	}
}

func NewRunStarlarkPackageArgs(
	packageId string,
	relativePathToMainFile string,
	mainFunctionName string,
	serializedParams string,
	dryRun bool,
	parallelism int32,
	experimentalFeatures []kurtosis_core_rpc_api_bindings.KurtosisFeatureFlag,
	cloudInstanceId string,
	cloudUserId string,
	imageDownloadMode kurtosis_core_rpc_api_bindings.ImageDownloadMode,
	nonBlockingMode bool,
	githubAuthToken string,
) *kurtosis_core_rpc_api_bindings.RunStarlarkPackageArgs {
	parallelismCopy := new(int32)
	*parallelismCopy = parallelism
	cloudInstanceIdCopy := new(string)
	*cloudInstanceIdCopy = cloudInstanceId
	cloudUserIdCopy := new(string)
	*cloudUserIdCopy = cloudUserId
	imageDownloadModeCopy := new(kurtosis_core_rpc_api_bindings.ImageDownloadMode)
	*imageDownloadModeCopy = imageDownloadMode
	clonePackage := false
	githubAuthTokenCopy := new(string)
	*githubAuthTokenCopy = githubAuthToken
	return &kurtosis_core_rpc_api_bindings.RunStarlarkPackageArgs{
		PackageId:              packageId,
		ClonePackage:           &clonePackage,
		StarlarkPackageContent: nil,
		SerializedParams:       &serializedParams,
		DryRun:                 &dryRun,
		Parallelism:            parallelismCopy,
		RelativePathToMainFile: &relativePathToMainFile,
		MainFunctionName:       &mainFunctionName,
		ExperimentalFeatures:   experimentalFeatures,
		CloudInstanceId:        cloudInstanceIdCopy,
		CloudUserId:            cloudUserIdCopy,
		ImageDownloadMode:      imageDownloadModeCopy,
		NonBlockingMode:        &nonBlockingMode,
		GithubAuthToken:        githubAuthTokenCopy,
	}
}

func NewRunStarlarkRemotePackageArgs(
	packageId string,
	relativePathToMainFile string,
	mainFunctionName string,
	serializedParams string,
	dryRun bool,
	parallelism int32,
	experimentalFeatures []kurtosis_core_rpc_api_bindings.KurtosisFeatureFlag,
	cloudInstanceId string,
	cloudUserId string,
	imageDownloadMode kurtosis_core_rpc_api_bindings.ImageDownloadMode,
	nonBlockingMode bool,
	githubAuthToken string,
) *kurtosis_core_rpc_api_bindings.RunStarlarkPackageArgs {
	parallelismCopy := new(int32)
	*parallelismCopy = parallelism
	cloudInstanceIdCopy := new(string)
	*cloudInstanceIdCopy = cloudInstanceId
	cloudUserIdCopy := new(string)
	*cloudUserIdCopy = cloudUserId
	imageDownloadModeCopy := new(kurtosis_core_rpc_api_bindings.ImageDownloadMode)
	*imageDownloadModeCopy = imageDownloadMode
	clonePackage := true
	githubAuthTokenCopy := new(string)
	*githubAuthTokenCopy = githubAuthToken
	return &kurtosis_core_rpc_api_bindings.RunStarlarkPackageArgs{
		PackageId:              packageId,
		ClonePackage:           &clonePackage,
		StarlarkPackageContent: nil,
		SerializedParams:       &serializedParams,
		DryRun:                 &dryRun,
		Parallelism:            parallelismCopy,
		RelativePathToMainFile: &relativePathToMainFile,
		MainFunctionName:       &mainFunctionName,
		ExperimentalFeatures:   experimentalFeatures,
		CloudInstanceId:        cloudInstanceIdCopy,
		CloudUserId:            cloudUserIdCopy,
		ImageDownloadMode:      imageDownloadModeCopy,
		NonBlockingMode:        &nonBlockingMode,
		GithubAuthToken:        githubAuthTokenCopy,
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

func NewStarlarkRunResponseLineFromInfoMsg(infoMessage string) *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine {
	return &kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine{
		RunResponseLine: &kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine_Info{
			Info: &kurtosis_core_rpc_api_bindings.StarlarkInfo{
				InfoMessage: infoMessage,
			},
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

func NewStarlarkInstruction(position *kurtosis_core_rpc_api_bindings.StarlarkInstructionPosition, name string, executableInstruction string, arguments []*kurtosis_core_rpc_api_bindings.StarlarkInstructionArg, isSkipped bool, description string) *kurtosis_core_rpc_api_bindings.StarlarkInstruction {
	return &kurtosis_core_rpc_api_bindings.StarlarkInstruction{
		InstructionName:       name,
		Position:              position,
		ExecutableInstruction: executableInstruction,
		Arguments:             arguments,
		IsSkipped:             isSkipped,
		Description:           description,
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
	serviceStatus kurtosis_core_rpc_api_bindings.ServiceStatus,
	container *kurtosis_core_rpc_api_bindings.Container,
	serviceDirPathsToFilesArtifactsList map[string]*kurtosis_core_rpc_api_bindings.FilesArtifactsList,
	minMillicpus uint32,
	maxMillicpus uint32,
	minMemoryMegabytes uint32,
	maxMemoryMegabytes uint32,
	user *kurtosis_core_rpc_api_bindings.User,
	tolerations []*kurtosis_core_rpc_api_bindings.Toleration,
	nodeSelectors map[string]string,
	labels map[string]string,
	tiniEnabled bool,
	ttyEnabled bool,

) *kurtosis_core_rpc_api_bindings.ServiceInfo {
	return &kurtosis_core_rpc_api_bindings.ServiceInfo{
		ServiceUuid:                         uuid,
		Name:                                name,
		ShortenedUuid:                       shortenedUuid,
		PrivateIpAddr:                       privateIpAddr,
		PrivatePorts:                        privatePorts,
		MaybePublicIpAddr:                   maybePublicIpAddr,
		MaybePublicPorts:                    maybePublicPorts,
		ServiceStatus:                       serviceStatus,
		Container:                           container,
		ServiceDirPathsToFilesArtifactsList: serviceDirPathsToFilesArtifactsList,
		MaxMillicpus:                        maxMillicpus,
		MinMillicpus:                        minMillicpus,
		MaxMemoryMegabytes:                  maxMemoryMegabytes,
		MinMemoryMegabytes:                  minMemoryMegabytes,
		User:                                user,
		Tolerations:                         tolerations,
		NodeSelectors:                       nodeSelectors,
		Labels:                              labels,
		TiniEnabled:                         &tiniEnabled,
		TtyEnabled:                          &ttyEnabled,
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

func NewUploadFilesArtifactResponse(artifactUuid string, artifactName string) *kurtosis_core_rpc_api_bindings.UploadFilesArtifactResponse {
	return &kurtosis_core_rpc_api_bindings.UploadFilesArtifactResponse{
		Uuid: artifactUuid,
		Name: artifactName,
	}
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
//	Connect Services arguments and response to configure user services port forwarding
//
// ==============================================================================================

func NewConnectServicesArgs(connect kurtosis_core_rpc_api_bindings.Connect) *kurtosis_core_rpc_api_bindings.ConnectServicesArgs {
	return &kurtosis_core_rpc_api_bindings.ConnectServicesArgs{
		Connect: connect,
	}
}

func NewConnectServicesResponse() *kurtosis_core_rpc_api_bindings.ConnectServicesResponse {
	return &kurtosis_core_rpc_api_bindings.ConnectServicesResponse{}
}
