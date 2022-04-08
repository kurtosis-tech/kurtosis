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
	"github.com/kurtosis-tech/kurtosis-core/api/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-core/api/golang/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis-core/api/golang/lib/modules"
	"github.com/kurtosis-tech/kurtosis-core/api/golang/lib/services"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
	"math"
	"path/filepath"
)

type EnclaveID string

type PartitionID string

const (
	// This will always resolve to the default partition ID (regardless of whether such a partition exists in the enclave,
	//  or it was repartitioned away)
	defaultPartitionId PartitionID = ""

	// The path on the user service container where the enclave data dir will be bind-mounted
	serviceEnclaveDataDirMountpoint = "/kurtosis-enclave-data"
)

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
type EnclaveContext struct {
	client kurtosis_core_rpc_api_bindings.ApiContainerServiceClient

	enclaveId EnclaveID

	// The location on the filesystem where this code is running where the enclave data dir exists
	enclaveDataDirpath string
}

/*
Creates a new EnclaveContext object with the given parameters.
*/
// TODO Migrate this to take in API container IP & API container GRPC port num, to match the (better) way that
//  Typescript does it, so that the user doesn't have to figure out how to instantiate the ApiContainerServiceClient on their own!
func NewEnclaveContext(
	client kurtosis_core_rpc_api_bindings.ApiContainerServiceClient,
	enclaveId EnclaveID,
	enclaveDataDirpath string,
) *EnclaveContext {
	return &EnclaveContext{
		client:             client,
		enclaveId:          enclaveId,
		enclaveDataDirpath: enclaveDataDirpath,
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
	args := binding_constructors.NewGetModuleInfoArgs(string(moduleId))

	// NOTE: As of 2021-07-18, we actually don't use any of the info that comes back because the ModuleContext doesn't require it!
	_, err := enclaveCtx.client.GetModuleInfo(context.Background(), args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting info for module '%v'", moduleId)
	}
	moduleCtx := modules.NewModuleContext(enclaveCtx.client, moduleId)
	return moduleCtx, nil
}

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
func (enclaveCtx *EnclaveContext) RegisterFilesArtifacts(filesArtifactUrls map[services.FilesArtifactID]string) error {
	filesArtifactIdStrsToUrls := map[string]string{}
	for artifactId, url := range filesArtifactUrls {
		filesArtifactIdStrsToUrls[string(artifactId)] = url
	}
	args := binding_constructors.NewRegisterFilesArtifactArgs(filesArtifactIdStrsToUrls)
	if _, err := enclaveCtx.client.RegisterFilesArtifacts(context.Background(), args); err != nil {
		return stacktrace.Propagate(err, "An error occurred registering files artifacts: %+v", filesArtifactUrls)
	}
	return nil
}

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
func (enclaveCtx *EnclaveContext) AddService(
	serviceId services.ServiceID,
	containerConfigSupplier func(ipAddr string, sharedDirectory *services.SharedPath) (*services.ContainerConfig, error),
) (*services.ServiceContext, error) {

	serviceContext, err := enclaveCtx.AddServiceToPartition(
		serviceId,
		defaultPartitionId,
		containerConfigSupplier,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding service '%v' to the enclave in the default partition", serviceId)
	}

	return serviceContext, nil
}

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
func (enclaveCtx *EnclaveContext) AddServiceToPartition(
	serviceId services.ServiceID,
	partitionID PartitionID,
	containerConfigSupplier func(ipAddr string, sharedDirectory *services.SharedPath) (*services.ContainerConfig, error),
) (*services.ServiceContext, error) {

	ctx := context.Background()

	logrus.Trace("Registering new service ID with Kurtosis API...")
	registerServiceArgs := binding_constructors.NewRegisterServiceArgs(string(serviceId), string(partitionID))

	registerServiceResp, err := enclaveCtx.client.RegisterService(ctx, registerServiceArgs)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred registering service with ID '%v' with the Kurtosis API",
			serviceId)
	}
	// TODO Defer a removeService call here if the function exits unsuccessfully
	logrus.Trace("New service successfully registered with Kurtosis API")

	privateIpAddr := registerServiceResp.PrivateIpAddr
	relativeServiceDirpath := registerServiceResp.RelativeServiceDirpath

	sharedDirectory := enclaveCtx.getSharedDirectory(relativeServiceDirpath)

	logrus.Trace("Generating container config object using the container config supplier...")
	containerConfig, err := containerConfigSupplier(privateIpAddr, sharedDirectory)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred executing the container config supplier for service with ID '%v'",
			serviceId)
	}
	logrus.Trace("Container config object successfully generated")

	logrus.Tracef("Creating files artifact ID str -> mount dirpaths map...")
	artifactIdStrToMountDirpath := map[string]string{}
	for filesArtifactId, mountDirpath := range containerConfig.GetFilesArtifactMountpoints() {
		artifactIdStrToMountDirpath[string(filesArtifactId)] = mountDirpath
	}
	logrus.Trace("Successfully created files artifact ID str -> mount dirpaths map")

	logrus.Trace("Starting new service with Kurtosis API...")
	privatePorts := containerConfig.GetUsedPorts()
	privatePortsForApi := map[string]*kurtosis_core_rpc_api_bindings.Port{}
	for portId, portSpec := range privatePorts {
		privatePortsForApi[portId] = &kurtosis_core_rpc_api_bindings.Port{
			Number:   uint32(portSpec.GetNumber()),
			Protocol: kurtosis_core_rpc_api_bindings.Port_Protocol(portSpec.GetProtocol()),
		}
	}
	startServiceArgs := binding_constructors.NewStartServiceArgs(
		string(serviceId),
		containerConfig.GetImage(),
		privatePortsForApi,
		containerConfig.GetEntrypointOverrideArgs(),
		containerConfig.GetCmdOverrideArgs(),
		containerConfig.GetEnvironmentVariableOverrides(),
		serviceEnclaveDataDirMountpoint,
		artifactIdStrToMountDirpath,
	)
	resp, err := enclaveCtx.client.StartService(ctx, startServiceArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred starting the service with the Kurtosis API")
	}
	logrus.Trace("Successfully started service with Kurtosis API")

	serviceCtxPublicPorts, err := convertApiPortsToServiceContextPorts(resp.GetPublicPorts())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred converting the public ports returned by the API to ports usable by the service context")
	}

	serviceContext := services.NewServiceContext(
		enclaveCtx.client,
		serviceId,
		sharedDirectory,
		privateIpAddr,
		privatePorts,
		resp.GetPublicIpAddr(),
		serviceCtxPublicPorts,
	)

	return serviceContext, nil
}

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
func (enclaveCtx *EnclaveContext) GetServiceContext(serviceId services.ServiceID) (*services.ServiceContext, error) {
	getServiceInfoArgs := binding_constructors.NewGetServiceInfoArgs(string(serviceId))
	resp, err := enclaveCtx.client.GetServiceInfo(context.Background(), getServiceInfoArgs)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred when trying to get info for service '%v'",
			serviceId)
	}
	if resp.GetPrivateIpAddr() == "" {
		return nil, stacktrace.NewError(
			"Kurtosis API reported an empty private IP address for service '%v' - this should never happen, and is a bug with Kurtosis!",
			serviceId)
	}
	if resp.GetPublicIpAddr() == "" {
		return nil, stacktrace.NewError(
			"Kurtosis API reported an empty public IP address for service '%v' - this should never happen, and is a bug with Kurtosis!",
			serviceId)
	}

	relativeServiceDirpath := resp.GetRelativeServiceDirpath()
	if relativeServiceDirpath == "" {
		return nil, stacktrace.NewError(
			"Kurtosis API reported an empty relative service directory path for service '%v' - this should never happen, and is a bug with Kurtosis!",
			serviceId)
	}

	enclaveDataDirMountDirpathOnSvcContainer := resp.GetEnclaveDataDirMountDirpath()
	if enclaveDataDirMountDirpathOnSvcContainer == "" {
		return nil, stacktrace.NewError(
			"Kurtosis API reported an empty enclave data dir mount dirpath for service '%v' - this should never happen, and is a bug with Kurtosis!",
			serviceId)
	}

	sharedDirectory := enclaveCtx.getSharedDirectory(relativeServiceDirpath)

	serviceCtxPrivatePorts, err := convertApiPortsToServiceContextPorts(resp.GetPrivatePorts())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred converting the private ports returned by the API to ports usable by the service context")
	}
	serviceCtxPublicPorts, err := convertApiPortsToServiceContextPorts(resp.GetPublicPorts())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred converting the public ports returned by the API to ports usable by the service context")
	}

	serviceContext := services.NewServiceContext(
		enclaveCtx.client,
		serviceId,
		sharedDirectory,
		resp.GetPrivateIpAddr(),
		serviceCtxPrivatePorts,
		resp.GetPublicIpAddr(),
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
	args := binding_constructors.NewRemoveServiceArgs(string(serviceId), containerStopTimeoutSeconds)
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
	response, err := enclaveCtx.client.GetServices(context.Background(), &emptypb.Empty{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the service IDs in the enclave")
	}

	serviceIds := make(map[services.ServiceID]bool, len(response.GetServiceIds()))

	for key, _ := range response.GetServiceIds() {
		serviceId := services.ServiceID(key)
		if _, ok := serviceIds[serviceId]; !ok {
			serviceIds[serviceId] = true
		}
	}

	return serviceIds, nil
}

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
func (enclaveCtx *EnclaveContext) GetModules() (map[modules.ModuleID]bool, error) {
	response, err := enclaveCtx.client.GetModules(context.Background(), &emptypb.Empty{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the IDs of the modules in the enclave")
	}

	moduleIDs := make(map[modules.ModuleID]bool, len(response.GetModuleIds()))

	for key, _ := range response.GetModuleIds() {
		moduleID := modules.ModuleID(key)
		if _, ok := moduleIDs[moduleID]; !ok {
			moduleIDs[moduleID] = true
		}
	}

	return moduleIDs, nil
}

// ====================================================================================================
// 									   Private helper methods
// ====================================================================================================
func (enclaveCtx *EnclaveContext) getSharedDirectory(relativeServiceDirpath string) *services.SharedPath {

	absFilepathOnThisContainer := filepath.Join(enclaveCtx.enclaveDataDirpath, relativeServiceDirpath)
	absFilepathOnServiceContainer := filepath.Join(serviceEnclaveDataDirMountpoint, relativeServiceDirpath)

	sharedDirectory := services.NewSharedPath(absFilepathOnThisContainer, absFilepathOnServiceContainer)

	return sharedDirectory
}

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
