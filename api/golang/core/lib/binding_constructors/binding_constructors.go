package binding_constructors

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"google.golang.org/protobuf/types/known/durationpb"
	"time"
)

// The generated bindings don't come with constructors (leaving it up to the user to initialize all the fields), so we
// add them so that our code is safer

// ==============================================================================================
//
//	Shared Objects (Used By Multiple Endpoints)
//
// ==============================================================================================
func NewPort(number uint32, protocol kurtosis_core_rpc_api_bindings.Port_Protocol) *kurtosis_core_rpc_api_bindings.Port {
	return &kurtosis_core_rpc_api_bindings.Port{
		Number:   number,
		Protocol: protocol,
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
	privateIPAddrPlaceholder string) *kurtosis_core_rpc_api_bindings.ServiceConfig {
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
	}
}

// ==============================================================================================
//
//	Load Module
//
// ==============================================================================================
func NewLoadModuleArgs(moduleId string, containerImage string, serializedParams string) *kurtosis_core_rpc_api_bindings.LoadModuleArgs {
	return &kurtosis_core_rpc_api_bindings.LoadModuleArgs{
		ModuleId:         moduleId,
		ContainerImage:   containerImage,
		SerializedParams: serializedParams,
	}
}

func NewLoadModuleResponse(
	guid string,
	privateIpAddr string,
	privatePort *kurtosis_core_rpc_api_bindings.Port,
	publicIpAddr string,
	publicPort *kurtosis_core_rpc_api_bindings.Port,
) *kurtosis_core_rpc_api_bindings.LoadModuleResponse {
	return &kurtosis_core_rpc_api_bindings.LoadModuleResponse{
		Guid:          guid,
		PrivateIpAddr: privateIpAddr,
		PrivatePort:   privatePort,
		PublicIpAddr:  publicIpAddr,
		PublicPort:    publicPort,
	}
}

// ==============================================================================================
//
//	Unload Module
//
// ==============================================================================================
func NewUnloadModuleArgs(moduleId string) *kurtosis_core_rpc_api_bindings.UnloadModuleArgs {
	return &kurtosis_core_rpc_api_bindings.UnloadModuleArgs{
		ModuleId: moduleId,
	}
}

func NewUnloadModuleResponse(moduleGuid string) *kurtosis_core_rpc_api_bindings.UnloadModuleResponse {
	return &kurtosis_core_rpc_api_bindings.UnloadModuleResponse{
		ModuleGuid: moduleGuid,
	}
}

// ==============================================================================================
//
//	Execute Module
//
// ==============================================================================================
func NewExecuteModuleArgs(moduleId string, serializedParams string) *kurtosis_core_rpc_api_bindings.ExecuteModuleArgs {
	return &kurtosis_core_rpc_api_bindings.ExecuteModuleArgs{
		ModuleId:         moduleId,
		SerializedParams: serializedParams,
	}
}

func NewExecuteModuleResponse(serializedResult string) *kurtosis_core_rpc_api_bindings.ExecuteModuleResponse {
	return &kurtosis_core_rpc_api_bindings.ExecuteModuleResponse{
		SerializedResult: serializedResult,
	}
}

// ==============================================================================================
//
//	Get Module Info
//
// ==============================================================================================
func NewGetModulesArgs(moduleIds map[string]bool) *kurtosis_core_rpc_api_bindings.GetModulesArgs {
	return &kurtosis_core_rpc_api_bindings.GetModulesArgs{
		Ids: moduleIds,
	}
}

func NewGetModulesResponse(
	moduleInfoMap map[string]*kurtosis_core_rpc_api_bindings.ModuleInfo,
) *kurtosis_core_rpc_api_bindings.GetModulesResponse {
	return &kurtosis_core_rpc_api_bindings.GetModulesResponse{
		ModuleInfo: moduleInfoMap,
	}
}

func NewModuleInfo(
	guid string,
	privateIpAddr string,
	privateGrpcPort *kurtosis_core_rpc_api_bindings.Port,
	maybePublicIpAddr string,
	maybePublicGrpcPort *kurtosis_core_rpc_api_bindings.Port,
) *kurtosis_core_rpc_api_bindings.ModuleInfo {
	return &kurtosis_core_rpc_api_bindings.ModuleInfo{
		Guid:                guid,
		PrivateIpAddr:       privateIpAddr,
		PrivateGrpcPort:     privateGrpcPort,
		MaybePublicIpAddr:   maybePublicIpAddr,
		MaybePublicGrpcPort: maybePublicGrpcPort,
	}
}

// ==============================================================================================
//
//	Facts Engine
//
// ==============================================================================================

func NewConstantFactRecipe(serviceId string, factName string, constantFactRecipeDefinition *kurtosis_core_rpc_api_bindings.ConstantFactRecipe, refreshInterval time.Duration) *kurtosis_core_rpc_api_bindings.FactRecipe {
	return &kurtosis_core_rpc_api_bindings.FactRecipe{
		ServiceId: serviceId,
		FactName:  factName,
		FactRecipeDefinition: &kurtosis_core_rpc_api_bindings.FactRecipe_ConstantFact{
			ConstantFact: constantFactRecipeDefinition,
		},
		RefreshInterval: durationpb.New(refreshInterval),
	}
}

// ==============================================================================================
//
//	Execute Startosis Script
//
// ==============================================================================================
func NewExecuteStartosisScriptArgs(serializedString string) *kurtosis_core_rpc_api_bindings.ExecuteStartosisScriptArgs {
	return &kurtosis_core_rpc_api_bindings.ExecuteStartosisScriptArgs{
		SerializedScript: serializedString,
	}
}

func NewExecuteStartosisResponse(
	serializedScriptOutput string,
	interpretationError string,
	validationErrors []*kurtosis_core_rpc_api_bindings.StartosisValidationError,
	executionError string,
) *kurtosis_core_rpc_api_bindings.ExecuteStartosisResponse {
	return &kurtosis_core_rpc_api_bindings.ExecuteStartosisResponse{
		SerializedScriptOutput: serializedScriptOutput,
		InterpretationError:    interpretationError,
		ValidationErrors:       validationErrors,
		ExecutionError:         executionError,
	}
}

// ==============================================================================================
//
//	Start Service
//
// ==============================================================================================
func NewStartServicesArgs(serviceConfigs map[string]*kurtosis_core_rpc_api_bindings.ServiceConfig, partitionID string) *kurtosis_core_rpc_api_bindings.StartServicesArgs {
	return &kurtosis_core_rpc_api_bindings.StartServicesArgs{
		ServiceIdsToConfigs: serviceConfigs,
		PartitionId:         partitionID,
	}
}

func NewStartServicesResponse(
	successfulServicesInfo map[string]*kurtosis_core_rpc_api_bindings.ServiceInfo,
	failedServicesErrors map[string]string) *kurtosis_core_rpc_api_bindings.StartServicesResponse {
	return &kurtosis_core_rpc_api_bindings.StartServicesResponse{
		SuccessfulServiceIdsToServiceInfo: successfulServicesInfo,
		FailedServiceIdsToError:           failedServicesErrors,
	}
}

// ==============================================================================================
//
//	Get Service Info
//
// ==============================================================================================
func NewGetServicesArgs(serviceIds map[string]bool) *kurtosis_core_rpc_api_bindings.GetServicesArgs {
	return &kurtosis_core_rpc_api_bindings.GetServicesArgs{
		ServiceIds: serviceIds,
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
	guid string,
	privateIpAddr string,
	privatePorts map[string]*kurtosis_core_rpc_api_bindings.Port,
	maybePublicIpAddr string,
	maybePublicPorts map[string]*kurtosis_core_rpc_api_bindings.Port,
) *kurtosis_core_rpc_api_bindings.ServiceInfo {
	return &kurtosis_core_rpc_api_bindings.ServiceInfo{
		ServiceGuid:       guid,
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
func NewRemoveServiceArgs(serviceId string) *kurtosis_core_rpc_api_bindings.RemoveServiceArgs {
	return &kurtosis_core_rpc_api_bindings.RemoveServiceArgs{
		ServiceId: serviceId,
	}
}

func NewRemoveServiceResponse(serviceGuid string) *kurtosis_core_rpc_api_bindings.RemoveServiceResponse {
	return &kurtosis_core_rpc_api_bindings.RemoveServiceResponse{
		ServiceGuid: serviceGuid,
	}
}

// ==============================================================================================
//
//	Repartition
//
// ==============================================================================================
func NewRepartitionArgs(
	partitionServices map[string]*kurtosis_core_rpc_api_bindings.PartitionServices,
	partitionConnections map[string]*kurtosis_core_rpc_api_bindings.PartitionConnections,
	defaultConnection *kurtosis_core_rpc_api_bindings.PartitionConnectionInfo) *kurtosis_core_rpc_api_bindings.RepartitionArgs {
	return &kurtosis_core_rpc_api_bindings.RepartitionArgs{
		PartitionServices:    partitionServices,
		PartitionConnections: partitionConnections,
		DefaultConnection:    defaultConnection,
	}
}

func NewPartitionServices(serviceIdSet map[string]bool) *kurtosis_core_rpc_api_bindings.PartitionServices {
	return &kurtosis_core_rpc_api_bindings.PartitionServices{
		ServiceIdSet: serviceIdSet,
	}
}

func NewPartitionConnections(connectionInfo map[string]*kurtosis_core_rpc_api_bindings.PartitionConnectionInfo) *kurtosis_core_rpc_api_bindings.PartitionConnections {
	return &kurtosis_core_rpc_api_bindings.PartitionConnections{
		ConnectionInfo: connectionInfo,
	}
}

func NewPartitionConnectionInfo(packetLossPercentage float32) *kurtosis_core_rpc_api_bindings.PartitionConnectionInfo {
	return &kurtosis_core_rpc_api_bindings.PartitionConnectionInfo{
		PacketLossPercentage: packetLossPercentage,
	}
}

// ==============================================================================================
//                                          Pause/Unpause Service
// ==============================================================================================

func NewPauseServiceArgs(serviceId string) *kurtosis_core_rpc_api_bindings.PauseServiceArgs {
	return &kurtosis_core_rpc_api_bindings.PauseServiceArgs{
		ServiceId: serviceId,
	}
}

func NewUnpauseServiceArgs(serviceId string) *kurtosis_core_rpc_api_bindings.UnpauseServiceArgs {
	return &kurtosis_core_rpc_api_bindings.UnpauseServiceArgs{
		ServiceId: serviceId,
	}
}

// ==============================================================================================
//
//	Exec Command
//
// ==============================================================================================
func NewExecCommandArgs(serviceId string, commandArgs []string) *kurtosis_core_rpc_api_bindings.ExecCommandArgs {
	return &kurtosis_core_rpc_api_bindings.ExecCommandArgs{
		ServiceId:   serviceId,
		CommandArgs: commandArgs,
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
//	Wait For Http Get Endpoint Availability
//
// ==============================================================================================
func NewWaitForHttpGetEndpointAvailabilityArgs(
	serviceId string,
	port uint32,
	path string,
	initialDelayMilliseconds uint32,
	retries uint32,
	retriesDelayMilliseconds uint32,
	bodyText string) *kurtosis_core_rpc_api_bindings.WaitForHttpGetEndpointAvailabilityArgs {
	return &kurtosis_core_rpc_api_bindings.WaitForHttpGetEndpointAvailabilityArgs{
		ServiceId:                serviceId,
		Port:                     port,
		Path:                     path,
		InitialDelayMilliseconds: initialDelayMilliseconds,
		Retries:                  retries,
		RetriesDelayMilliseconds: retriesDelayMilliseconds,
		BodyText:                 bodyText,
	}
}

// ==============================================================================================
//
//	Wait For Http Post Endpoint Availability
//
// ==============================================================================================
func NewWaitForHttpPostEndpointAvailabilityArgs(
	serviceId string,
	port uint32,
	path string,
	requestBody string,
	initialDelayMilliseconds uint32,
	retries uint32,
	retriesDelayMilliseconds uint32,
	bodyText string) *kurtosis_core_rpc_api_bindings.WaitForHttpPostEndpointAvailabilityArgs {
	return &kurtosis_core_rpc_api_bindings.WaitForHttpPostEndpointAvailabilityArgs{
		ServiceId:                serviceId,
		Port:                     port,
		Path:                     path,
		RequestBody:              requestBody,
		InitialDelayMilliseconds: initialDelayMilliseconds,
		Retries:                  retries,
		RetriesDelayMilliseconds: retriesDelayMilliseconds,
		BodyText:                 bodyText,
	}
}

// ==============================================================================================
//
//	Upload Files Artifact
//
// ==============================================================================================
func NewUploadFilesArtifactArgs(data []byte) *kurtosis_core_rpc_api_bindings.UploadFilesArtifactArgs {
	return &kurtosis_core_rpc_api_bindings.UploadFilesArtifactArgs{Data: data}
}

// ==============================================================================================
//
//	Store Web Files Artifact
//
// ==============================================================================================
func NewStoreWebFilesArtifactArgs(url string) *kurtosis_core_rpc_api_bindings.StoreWebFilesArtifactArgs {
	return &kurtosis_core_rpc_api_bindings.StoreWebFilesArtifactArgs{Url: url}
}

// ==============================================================================================
//
//	Store Files Artifact From Service
//
// ==============================================================================================
func NewStoreFilesArtifactFromServiceArgs(serviceId string, sourcePath string) *kurtosis_core_rpc_api_bindings.StoreFilesArtifactFromServiceArgs {
	return &kurtosis_core_rpc_api_bindings.StoreFilesArtifactFromServiceArgs{ServiceId: serviceId, SourcePath: sourcePath}
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

func NewRenderTemplatesToFilesArtifactArgs(templatesAndDataByDestinationRelFilepath map[string]*kurtosis_core_rpc_api_bindings.RenderTemplatesToFilesArtifactArgs_TemplateAndData) *kurtosis_core_rpc_api_bindings.RenderTemplatesToFilesArtifactArgs {
	return &kurtosis_core_rpc_api_bindings.RenderTemplatesToFilesArtifactArgs{
		TemplatesAndDataByDestinationRelFilepath: templatesAndDataByDestinationRelFilepath,
	}
}

func NewRenderTemplatesToFilesArtifactResponse(filesArtifactUuid string) *kurtosis_core_rpc_api_bindings.RenderTemplatesToFilesArtifactResponse {
	return &kurtosis_core_rpc_api_bindings.RenderTemplatesToFilesArtifactResponse{
		Uuid: filesArtifactUuid,
	}
}

// ==============================================================================================
//                                 Startosis errors
// ==============================================================================================

func NewStartosisValidationError(error string) *kurtosis_core_rpc_api_bindings.StartosisValidationError {
	return &kurtosis_core_rpc_api_bindings.StartosisValidationError{
		Error: error,
	}
}

// ==============================================================================================
//                                 Startosis Module Exec Args
// ==============================================================================================

func NewExecuteStartosisModuleArgs(moduleId string, compressedModule []byte, serializedParams string) *kurtosis_core_rpc_api_bindings.ExecuteStartosisModuleArgs {
	return &kurtosis_core_rpc_api_bindings.ExecuteStartosisModuleArgs{
		ModuleId:         moduleId,
		Data:             compressedModule,
		SerializedParams: serializedParams,
	}
}
