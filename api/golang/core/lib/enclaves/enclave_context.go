/*
 *    Copyright 2021 Kurtosis Technologies Inc.
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 *
 */

package enclaves

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/modules"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/mholt/archiver"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"math"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type EnclaveID string

type PartitionID string

const (
	// This will always resolve to the default partition ID (regardless of whether such a partition exists in the enclave,
	//  or it was repartitioned away)
	defaultPartitionId PartitionID = ""

	grpcDataTransferLimit     = 3999000 //3.999 Mb. 1kb wiggle room. 1kb being about the size of a simple 2 paragraph readme.
	tempCompressionDirPattern = "upload-compression-cache-"
	compressionExtension      = ".tgz"
	defaultTmpDir             = ""

	modFilename = "kurtosis.mod"
)

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
type EnclaveContext struct {
	client kurtosis_core_rpc_api_bindings.ApiContainerServiceClient

	enclaveId EnclaveID
}

/*
Creates a new EnclaveContext object with the given parameters.
*/
// TODO Migrate this to take in API container IP & API container GRPC port num, to match the (better) way that
//  Typescript does it, so that the user doesn't have to figure out how to instantiate the ApiContainerServiceClient on their own!
func NewEnclaveContext(
	client kurtosis_core_rpc_api_bindings.ApiContainerServiceClient,
	enclaveId EnclaveID,
) *EnclaveContext {
	return &EnclaveContext{
		client:    client,
		enclaveId: enclaveId,
	}
}

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
func (enclaveCtx *EnclaveContext) GetEnclaveID() EnclaveID {
	return enclaveCtx.enclaveId
}

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
func (enclaveCtx *EnclaveContext) LoadModule(
	moduleId modules.ModuleID,
	image string,
	serializedParams string) (*modules.ModuleContext, error) {
	args := binding_constructors.NewLoadModuleArgs(string(moduleId), image, serializedParams)

	// We proxy calls to execute modules via the API container, so actually no need to use the response here
	_, err := enclaveCtx.client.LoadModule(context.Background(), args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred loading new module '%v' with image '%v' and serialized params '%v'", moduleId, image, serializedParams)
	}
	moduleCtx := modules.NewModuleContext(enclaveCtx.client, moduleId)
	return moduleCtx, nil
}

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
func (enclaveCtx *EnclaveContext) UnloadModule(moduleId modules.ModuleID) error {
	args := binding_constructors.NewUnloadModuleArgs(string(moduleId))

	_, err := enclaveCtx.client.UnloadModule(context.Background(), args)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred unloading module '%v'", moduleId)
	}
	return nil
}

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
func (enclaveCtx *EnclaveContext) GetModuleContext(moduleId modules.ModuleID) (*modules.ModuleContext, error) {
	moduleMapForArgs := map[string]bool{
		string(moduleId): true,
	}
	args := binding_constructors.NewGetModulesArgs(moduleMapForArgs)

	resp, err := enclaveCtx.client.GetModules(context.Background(), args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting info for module '%v'", moduleId)
	}

	if _, found := resp.ModuleInfo[string(moduleId)]; !found {
		return nil, stacktrace.NewError("Module '%v' does not exist", moduleId)
	}

	moduleCtx := modules.NewModuleContext(enclaveCtx.client, moduleId)
	return moduleCtx, nil
}

func (enclaveCtx *EnclaveContext) ExecuteStartosisScript(serializedScript string) (*kurtosis_core_rpc_api_bindings.ExecuteStartosisResponse, error) {
	executeStartosisScriptArgs := binding_constructors.NewExecuteStartosisScriptArgs(serializedScript)
	executeStartosisResponse, err := enclaveCtx.client.ExecuteStartosisScript(context.Background(), executeStartosisScriptArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unexpected error happened executing Startosis script \n%v", serializedScript)
	}
	return executeStartosisResponse, nil
}

func (enclaveCtx *EnclaveContext) ExecuteStartosisModule(moduleRootPath string) (*kurtosis_core_rpc_api_bindings.ExecuteStartosisResponse, error) {
	kurtosisModFilepath := path.Join(moduleRootPath, modFilename)

	kurtosisMod, err := parseKurtosisMod(kurtosisModFilepath)
	if err != nil {
		return nil, stacktrace.Propagate(err, "There was an error parsing the '%v' at '%v'", modFilename, moduleRootPath)
	}

	compressedModule, err := compressPath(moduleRootPath)
	if err != nil {
		return nil, stacktrace.Propagate(err, "There was an error compressing module '%v' before upload", moduleRootPath)
	}
	executeStartosisModuleArgs := binding_constructors.NewExecuteStartosisModuleArgs(kurtosisMod.Module.ModuleName, compressedModule)
	executeStartosisResponse, err := enclaveCtx.client.ExecuteStartosisModule(context.Background(), executeStartosisModuleArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unexpected error happened executing Startosis module \n%v", moduleRootPath)
	}
	return executeStartosisResponse, nil
}

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
func (enclaveCtx *EnclaveContext) AddService(
	serviceID services.ServiceID,
	containerConfig *services.ContainerConfig,
) (*services.ServiceContext, error) {
	containerConfigs := map[services.ServiceID]*services.ContainerConfig{}
	containerConfigs[serviceID] = containerConfig
	serviceContexts, failedServices, err := enclaveCtx.AddServicesToPartition(
		containerConfigs,
		defaultPartitionId,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding service '%v' to the enclave in the default partition", serviceID)
	}
	serviceErr, found := failedServices[serviceID]
	if found {
		return nil, stacktrace.Propagate(serviceErr, "An error occurred adding service '%v' to the enclave in the default partition", serviceID)
	}
	serviceCtx, found := serviceContexts[serviceID]
	if !found {
		return nil, stacktrace.NewError("An error occurred retrieving the service context of service with ID '%v' from result of adding service to partition. This should not happen and is a bug in Kurtosis.", serviceID)
	}
	return serviceCtx, nil
}

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
func (enclaveCtx *EnclaveContext) AddServices(
	containerConfigs map[services.ServiceID]*services.ContainerConfig,
) (
	resultSuccessfulServices map[services.ServiceID]*services.ServiceContext,
	resultFailedServices map[services.ServiceID]error,
	resultErr error,
) {
	successfulServices, failedServices, err := enclaveCtx.AddServicesToPartition(containerConfigs, defaultPartitionId)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred adding services to the enclave in the default partition.")
	}
	return successfulServices, failedServices, err
}

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
func (enclaveCtx *EnclaveContext) AddServiceToPartition(
	serviceID services.ServiceID,
	partitionID PartitionID,
	containerConfig *services.ContainerConfig,
) (*services.ServiceContext, error) {
	containerConfigs := map[services.ServiceID]*services.ContainerConfig{}
	containerConfigs[serviceID] = containerConfig
	serviceContexts, failedServices, err := enclaveCtx.AddServicesToPartition(
		containerConfigs,
		partitionID,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding service '%v' to the enclave in the default partition", serviceID)
	}
	serviceErr, found := failedServices[serviceID]
	if found {
		return nil, stacktrace.Propagate(serviceErr, "An error occurred adding service '%v' to the enclave in the default partition", serviceID)
	}
	serviceCtx, found := serviceContexts[serviceID]
	if !found {
		return nil, stacktrace.NewError("An error occurred retrieving the service context of service with ID '%v' from result of adding service to partition. This should not happen and is a bug in Kurtosis.", serviceID)
	}
	return serviceCtx, nil
}

func (enclaveCtx *EnclaveContext) AddServicesToPartition(
	containerConfigs map[services.ServiceID]*services.ContainerConfig,
	partitionID PartitionID,
) (
	resultSuccessfulServices map[services.ServiceID]*services.ServiceContext,
	resultFailedServices map[services.ServiceID]error,
	resultErr error,
) {
	ctx := context.Background()
	failedServicesPool := map[services.ServiceID]error{}
	partitionIDStr := string(partitionID)

	serviceConfigs := map[string]*kurtosis_core_rpc_api_bindings.ServiceConfig{}
	for serviceID, containerConfig := range containerConfigs {
		logrus.Tracef("Creating files artifact ID str -> mount dirpaths map for service with Id '%v'...", serviceID)
		artifactIdStrToMountDirpath := map[string]string{}
		for filesArtifactID, mountDirpath := range containerConfig.GetFilesArtifactMountpoints() {
			artifactIdStrToMountDirpath[string(filesArtifactID)] = mountDirpath
		}
		logrus.Tracef("Successfully created files artifact ID str -> mount dirpaths map for service with ID '%v'", serviceID)
		privatePorts := containerConfig.GetUsedPorts()
		privatePortsForApi := map[string]*kurtosis_core_rpc_api_bindings.Port{}
		for portId, portSpec := range privatePorts {
			privatePortsForApi[portId] = &kurtosis_core_rpc_api_bindings.Port{
				Number:   uint32(portSpec.GetNumber()),
				Protocol: kurtosis_core_rpc_api_bindings.Port_Protocol(portSpec.GetProtocol()),
			}
		}
		//TODO this is a huge hack to temporarily enable static ports for NEAR until we have a more productized solution
		publicPorts := containerConfig.GetPublicPorts()
		publicPortsForApi := map[string]*kurtosis_core_rpc_api_bindings.Port{}
		for portId, portSpec := range publicPorts {
			publicPortsForApi[portId] = &kurtosis_core_rpc_api_bindings.Port{
				Number:   uint32(portSpec.GetNumber()),
				Protocol: kurtosis_core_rpc_api_bindings.Port_Protocol(portSpec.GetProtocol()),
			}
		}
		//TODO finish the hack

		serviceIDStr := string(serviceID)
		serviceConfigs[serviceIDStr] = binding_constructors.NewServiceConfig(
			containerConfig.GetImage(),
			privatePortsForApi,
			publicPortsForApi,
			containerConfig.GetEntrypointOverrideArgs(),
			containerConfig.GetCmdOverrideArgs(),
			containerConfig.GetEnvironmentVariableOverrides(),
			artifactIdStrToMountDirpath,
			containerConfig.GetCPUAllocationMillicpus(),
			containerConfig.GetMemoryAllocationMegabytes(),
			containerConfig.GetPrivateIPAddrPlaceholder())
	}

	startServicesArgs := binding_constructors.NewStartServicesArgs(serviceConfigs, partitionIDStr)

	logrus.Trace("Starting new services with Kurtosis API...")
	startServicesResp, err := enclaveCtx.client.StartServices(ctx, startServicesArgs)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred starting services with the Kurtosis API")
	}
	// defer-undo removes all successfully started services in case of errors in the future phases
	shouldRemoveServices := map[services.ServiceID]bool{}
	for serviceIDStr := range startServicesResp.GetSuccessfulServiceIdsToServiceInfo() {
		shouldRemoveServices[services.ServiceID(serviceIDStr)] = true
	}
	defer func() {
		for serviceID := range shouldRemoveServices {
			removeServiceArgs := binding_constructors.NewRemoveServiceArgs(string(serviceID))
			_, err = enclaveCtx.client.RemoveService(context.Background(), removeServiceArgs)
			if err != nil {
				logrus.Errorf("Attempted to remove service '%v' to delete its resources after it failed to start, but an error occurred "+
					"while attempting to remove the service:\n'%v'", serviceID, err)
			}
		}
	}()

	for serviceIDStr, errStr := range startServicesResp.GetFailedServiceIdsToError() {
		serviceID := services.ServiceID(serviceIDStr)
		failedServicesPool[serviceID] = stacktrace.Propagate(errors.New(errStr), "The following error occurred when trying to start service '%v'", serviceID)
	}

	successfulServices := map[services.ServiceID]*services.ServiceContext{}
	for serviceIDStr, serviceInfo := range startServicesResp.GetSuccessfulServiceIdsToServiceInfo() {
		serviceID := services.ServiceID(serviceIDStr)

		serviceCtxPrivatePorts, err := convertApiPortsToServiceContextPorts(serviceInfo.GetPrivatePorts())
		if err != nil {
			failedServicesPool[serviceID] = stacktrace.Propagate(err, "An error occurred converting the private ports returned by the API to ports usable by the service context.")
			continue
		}
		serviceCtxPublicPorts, err := convertApiPortsToServiceContextPorts(serviceInfo.GetMaybePublicPorts())
		if err != nil {
			failedServicesPool[serviceID] = stacktrace.Propagate(err, "An error occurred converting the public ports returned by the API to ports usable by the service context.")
			continue
		}

		serviceContext := services.NewServiceContext(
			enclaveCtx.client,
			serviceID,
			serviceInfo.GetPrivateIpAddr(),
			serviceCtxPrivatePorts,
			serviceInfo.GetMaybePublicIpAddr(),
			serviceCtxPublicPorts,
		)
		successfulServices[serviceID] = serviceContext
	}

	// Do not remove resources for successful services
	for serviceID := range successfulServices {
		delete(shouldRemoveServices, serviceID)
	}
	return successfulServices, failedServicesPool, nil
}

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
func (enclaveCtx *EnclaveContext) GetServiceContext(serviceId services.ServiceID) (*services.ServiceContext, error) {
	serviceIdMapForArgs := map[string]bool{string(serviceId): true}
	getServiceInfoArgs := binding_constructors.NewGetServicesArgs(serviceIdMapForArgs)
	response, err := enclaveCtx.client.GetServices(context.Background(), getServiceInfoArgs)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred when trying to get info for service '%v'",
			serviceId)
	}
	serviceInfo, found := response.GetServiceInfo()[string(serviceId)]
	if !found {
		return nil, stacktrace.NewError("Failed to retrieve service information for service '%v'", string(serviceId))
	}
	if serviceInfo.GetPrivateIpAddr() == "" {
		return nil, stacktrace.NewError(
			"Kurtosis API reported an empty private IP address for service '%v' - this should never happen, and is a bug with Kurtosis!",
			serviceId)
	}
	if serviceInfo.GetMaybePublicIpAddr() == "" {
		return nil, stacktrace.NewError(
			"Kurtosis API reported an empty public IP address for service '%v' - this should never happen, and is a bug with Kurtosis!",
			serviceId)
	}

	serviceCtxPrivatePorts, err := convertApiPortsToServiceContextPorts(serviceInfo.GetPrivatePorts())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred converting the private ports returned by the API to ports usable by the service context")
	}
	serviceCtxPublicPorts, err := convertApiPortsToServiceContextPorts(serviceInfo.GetMaybePublicPorts())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred converting the public ports returned by the API to ports usable by the service context")
	}

	serviceContext := services.NewServiceContext(
		enclaveCtx.client,
		serviceId,
		serviceInfo.GetPrivateIpAddr(),
		serviceCtxPrivatePorts,
		serviceInfo.GetMaybePublicIpAddr(),
		serviceCtxPublicPorts,
	)

	return serviceContext, nil
}

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
func (enclaveCtx *EnclaveContext) RemoveService(serviceId services.ServiceID, containerStopTimeoutSeconds uint64) error {

	logrus.Debugf("Removing service '%v'...", serviceId)
	// NOTE: This is kinda weird - when we remove a service we can never get it back so having a container
	//  stop timeout doesn't make much sense. It will make more sense when we can stop/start containers
	// Independent of adding/removing them from the enclave
	args := binding_constructors.NewRemoveServiceArgs(string(serviceId))
	if _, err := enclaveCtx.client.RemoveService(context.Background(), args); err != nil {
		return stacktrace.Propagate(err, "An error occurred removing service '%v' from the enclave", serviceId)
	}

	logrus.Debugf("Successfully removed service ID %v", serviceId)

	return nil
}

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
func (enclaveCtx *EnclaveContext) RepartitionNetwork(
	partitionServices map[PartitionID]map[services.ServiceID]bool,
	partitionConnections map[PartitionID]map[PartitionID]PartitionConnection,
	defaultConnection PartitionConnection) error {

	if partitionServices == nil {
		return stacktrace.NewError("Partition services map cannot be nil")
	}
	if defaultConnection == nil {
		return stacktrace.NewError("Default connection cannot be nil")
	}

	// Cover for lazy/confused users
	if partitionConnections == nil {
		partitionConnections = map[PartitionID]map[PartitionID]PartitionConnection{}
	}

	reqPartitionServices := map[string]*kurtosis_core_rpc_api_bindings.PartitionServices{}
	for partitionId, serviceIdSet := range partitionServices {
		serviceIdStrPseudoSet := map[string]bool{}
		for serviceId := range serviceIdSet {
			serviceIdStr := string(serviceId)
			serviceIdStrPseudoSet[serviceIdStr] = true
		}
		partitionIdStr := string(partitionId)
		reqPartitionServices[partitionIdStr] = binding_constructors.NewPartitionServices(serviceIdStrPseudoSet)
	}

	reqPartitionConns := map[string]*kurtosis_core_rpc_api_bindings.PartitionConnections{}
	for partitionAId, partitionAConnsMap := range partitionConnections {
		partitionAConnsStrMap := map[string]*kurtosis_core_rpc_api_bindings.PartitionConnectionInfo{}
		for partitionBId, conn := range partitionAConnsMap {
			partitionBIdStr := string(partitionBId)
			partitionAConnsStrMap[partitionBIdStr] = conn.getPartitionConnectionInfo()
		}
		partitionAConns := binding_constructors.NewPartitionConnections(partitionAConnsStrMap)
		partitionAIdStr := string(partitionAId)
		reqPartitionConns[partitionAIdStr] = partitionAConns
	}

	reqDefaultConnection := defaultConnection.getPartitionConnectionInfo()

	repartitionArgs := binding_constructors.NewRepartitionArgs(reqPartitionServices, reqPartitionConns, reqDefaultConnection)
	if _, err := enclaveCtx.client.Repartition(context.Background(), repartitionArgs); err != nil {
		return stacktrace.Propagate(err, "An error occurred repartitioning the enclave")
	}
	return nil
}

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
func (enclaveCtx *EnclaveContext) WaitForHttpGetEndpointAvailability(serviceId services.ServiceID, port uint32, path string, initialDelayMilliseconds uint32, retries uint32, retriesDelayMilliseconds uint32, bodyText string) error {

	availabilityArgs := binding_constructors.NewWaitForHttpGetEndpointAvailabilityArgs(
		string(serviceId),
		port,
		path,
		initialDelayMilliseconds,
		retries,
		retriesDelayMilliseconds,
		bodyText,
	)

	if _, err := enclaveCtx.client.WaitForHttpGetEndpointAvailability(context.Background(), availabilityArgs); err != nil {
		return stacktrace.Propagate(
			err,
			"Endpoint '%v' on port '%v' for service '%v' did not become available despite polling %v times with %v between polls",
			path,
			port,
			serviceId,
			retries,
			retriesDelayMilliseconds,
		)
	}
	return nil
}

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
func (enclaveCtx *EnclaveContext) WaitForHttpPostEndpointAvailability(serviceId services.ServiceID, port uint32, path string, requestBody string, initialDelayMilliseconds uint32, retries uint32, retriesDelayMilliseconds uint32, bodyText string) error {

	availabilityArgs := binding_constructors.NewWaitForHttpPostEndpointAvailabilityArgs(
		string(serviceId),
		port,
		path,
		requestBody,
		initialDelayMilliseconds,
		retries,
		retriesDelayMilliseconds,
		bodyText,
	)

	if _, err := enclaveCtx.client.WaitForHttpPostEndpointAvailability(context.Background(), availabilityArgs); err != nil {
		return stacktrace.Propagate(
			err,
			"Endpoint '%v' on port '%v' for service '%v' did not become available despite polling %v times with %v between polls",
			path,
			port,
			serviceId,
			retries,
			retriesDelayMilliseconds,
		)
	}
	return nil
}

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
func (enclaveCtx *EnclaveContext) GetServices() (map[services.ServiceID]bool, error) {
	getAllServicesIdFilter := map[string]bool{}
	getServicesArgs := binding_constructors.NewGetServicesArgs(getAllServicesIdFilter)
	response, err := enclaveCtx.client.GetServices(context.Background(), getServicesArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the service IDs in the enclave")
	}

	serviceIds := make(map[services.ServiceID]bool, len(response.GetServiceInfo()))

	for key := range response.GetServiceInfo() {
		serviceId := services.ServiceID(key)
		if _, ok := serviceIds[serviceId]; !ok {
			serviceIds[serviceId] = true
		}
	}

	return serviceIds, nil
}

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
func (enclaveCtx *EnclaveContext) GetModules() (map[modules.ModuleID]bool, error) {
	getAllModulesIdFilter := map[string]bool{}
	emptyGetModulesArgs := binding_constructors.NewGetModulesArgs(getAllModulesIdFilter)
	response, err := enclaveCtx.client.GetModules(context.Background(), emptyGetModulesArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the IDs of the modules in the enclave")
	}

	moduleIDs := make(map[modules.ModuleID]bool, len(response.GetModuleInfo()))

	for key := range response.GetModuleInfo() {
		moduleID := modules.ModuleID(key)
		if _, ok := moduleIDs[moduleID]; !ok {
			moduleIDs[moduleID] = true
		}
	}

	return moduleIDs, nil
}

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
func (enclaveCtx *EnclaveContext) UploadFiles(pathToUpload string) (services.FilesArtifactUUID, error) {
	content, err := compressPath(pathToUpload)
	if err != nil {
		return "", stacktrace.Propagate(err,
			"There was an error compressing the file '%v' before upload",
			pathToUpload)
	}

	args := binding_constructors.NewUploadFilesArtifactArgs(content)
	response, err := enclaveCtx.client.UploadFilesArtifact(context.Background(), args)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error was encountered while uploading data to the API Container.")
	}
	return services.FilesArtifactUUID(response.Uuid), nil
}

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
func (enclaveCtx *EnclaveContext) StoreWebFiles(ctx context.Context, urlToStoreWeb string) (services.FilesArtifactUUID, error) {
	args := binding_constructors.NewStoreWebFilesArtifactArgs(urlToStoreWeb)
	response, err := enclaveCtx.client.StoreWebFilesArtifact(ctx, args)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred downloading files artifact from URL '%v'", urlToStoreWeb)
	}
	return services.FilesArtifactUUID(response.Uuid), nil
}

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
func (enclaveCtx *EnclaveContext) StoreServiceFiles(ctx context.Context, serviceId services.ServiceID, absoluteFilepathOnServiceContainer string) (services.FilesArtifactUUID, error) {
	serviceIdStr := string(serviceId)
	args := binding_constructors.NewStoreFilesArtifactFromServiceArgs(serviceIdStr, absoluteFilepathOnServiceContainer)
	response, err := enclaveCtx.client.StoreFilesArtifactFromService(ctx, args)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred copying source content from absolute filepath '%v' in service container with ID '%v'", absoluteFilepathOnServiceContainer, serviceIdStr)
	}
	return services.FilesArtifactUUID(response.Uuid), nil
}

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
func (enclaveCtx *EnclaveContext) PauseService(serviceId services.ServiceID) error {
	args := binding_constructors.NewPauseServiceArgs(string(serviceId))
	_, err := enclaveCtx.client.PauseService(context.Background(), args)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to pause service '%v'", serviceId)
	}
	return nil
}

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
func (enclaveCtx *EnclaveContext) UnpauseService(serviceId services.ServiceID) error {
	args := binding_constructors.NewUnpauseServiceArgs(string(serviceId))
	_, err := enclaveCtx.client.UnpauseService(context.Background(), args)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to unpause service '%v'", serviceId)
	}
	return nil
}

func (enclaveCtx *EnclaveContext) RenderTemplates(templateAndDataByDestinationRelFilepaths map[string]*TemplateAndData) (services.FilesArtifactUUID, error) {
	if len(templateAndDataByDestinationRelFilepaths) == 0 {
		return "", stacktrace.NewError("Expected at least one template got 0")
	}

	templateAndDataByRelDestinationFilepathArgs := make(map[string]*kurtosis_core_rpc_api_bindings.RenderTemplatesToFilesArtifactArgs_TemplateAndData)

	for destinationRelFilepath, templateAndData := range templateAndDataByDestinationRelFilepaths {
		template := templateAndData.template
		templateData := templateAndData.templateData

		templateDataAsJson, err := json.Marshal(templateData)
		if err != nil {
			return "", stacktrace.Propagate(err, "Failed to jsonify templateData '%v' for filename '%v'", templateData, destinationRelFilepath)
		}

		templateAndDataAsJsonString := binding_constructors.NewTemplateAndData(
			template,
			string(templateDataAsJson),
		)
		templateAndDataByRelDestinationFilepathArgs[destinationRelFilepath] = templateAndDataAsJsonString
	}

	renderTemplatesToFilesArtifactArgs := binding_constructors.NewRenderTemplatesToFilesArtifactArgs(templateAndDataByRelDestinationFilepathArgs)

	response, err := enclaveCtx.client.RenderTemplatesToFilesArtifact(context.Background(), renderTemplatesToFilesArtifactArgs)
	if err != nil {
		return "", stacktrace.Propagate(err, "Error in rendering templates")
	}

	return services.FilesArtifactUUID(response.Uuid), nil
}

// ====================================================================================================
//
//	Private helper methods
//
// ====================================================================================================
func convertApiPortsToServiceContextPorts(apiPorts map[string]*kurtosis_core_rpc_api_bindings.Port) (map[string]*services.PortSpec, error) {
	result := map[string]*services.PortSpec{}
	for portId, apiPortSpec := range apiPorts {
		apiPortProtocol := apiPortSpec.GetProtocol()
		serviceCtxPortProtocol := services.PortProtocol(apiPortProtocol)
		if !serviceCtxPortProtocol.IsValid() {
			return nil, stacktrace.NewError("Received unrecognized protocol '%v' from the API", apiPortProtocol)
		}
		portNumUint32 := apiPortSpec.GetNumber()
		if portNumUint32 > math.MaxUint16 {
			return nil, stacktrace.NewError(
				"Received port num '%v' from the API which is higher than the max uint16 value '%v'; this is VERY weird because ports should be 16-bit numbers",
				portNumUint32,
				math.MaxUint16,
			)
		}
		portNumUint16 := uint16(portNumUint32)
		result[portId] = services.NewPortSpec(
			portNumUint16,
			serviceCtxPortProtocol,
		)
	}
	return result, nil
}

func compressPath(pathToCompress string) ([]byte, error) {
	pathToCompress = strings.TrimRight(pathToCompress, string(filepath.Separator))
	uploadFileInfo, err := os.Stat(pathToCompress)
	if err != nil {
		return nil, stacktrace.Propagate(err, "There was a path error for '%s' during file compression.", pathToCompress)
	}

	// This allows us to archive contents of dirs in root instead of nesting
	var filepathsToUpload []string
	if uploadFileInfo.IsDir() {
		filesInDirectory, err := ioutil.ReadDir(pathToCompress)
		if err != nil {
			return nil, stacktrace.Propagate(err, "There was an error in getting a list of files in the directory '%s' provided", pathToCompress)
		}
		if len(filesInDirectory) == 0 {
			return nil, stacktrace.NewError("The directory '%s' you are trying to compress is empty", pathToCompress)
		}

		for _, fileInDirectory := range filesInDirectory {
			filepathToUpload := filepath.Join(pathToCompress, fileInDirectory.Name())
			filepathsToUpload = append(filepathsToUpload, filepathToUpload)
		}
	} else {
		filepathsToUpload = append(filepathsToUpload, pathToCompress)
	}

	tempDir, err := ioutil.TempDir(defaultTmpDir, tempCompressionDirPattern)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create temporary directory '%s' for compression.", tempDir)
	}

	compressedFilePath := filepath.Join(tempDir, filepath.Base(pathToCompress)+compressionExtension)
	if err = archiver.Archive(filepathsToUpload, compressedFilePath); err != nil {
		return nil, stacktrace.Propagate(err, "Failed to compress '%s'.", pathToCompress)
	}

	compressedFileInfo, err := os.Stat(compressedFilePath)
	if err != nil {
		return nil, stacktrace.Propagate(err,
			"Failed to create a temporary archive file at '%s' during files upload for '%s'.",
			tempDir, pathToCompress)
	}

	if compressedFileInfo.Size() >= grpcDataTransferLimit {
		return nil, stacktrace.Propagate(err,
			"The files you are trying to upload, which are now compressed, exceed or reach 4mb, a limit imposed by gRPC. "+
				"Please reduce the total file size and ensure it can compress to a size below 4mb.")
	}
	content, err := ioutil.ReadFile(compressedFilePath)
	if err != nil {
		return nil, stacktrace.Propagate(err,
			"There was an error reading from the temporary tar file '%s' recently compressed for upload.",
			compressedFileInfo.Name())
	}

	return content, nil
}
