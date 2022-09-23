/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package service_network

import (
	"compress/gzip"
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/files_artifacts_expansion"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/api/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/core/files_artifacts_expander/args"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/networking_sidecar"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/kurtosis/core/server/commons/enclave_data_directory"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"math"
	"net"
	"path"
	"strings"
	"sync"
)

const (
	defaultPartitionId                       service_network_types.PartitionID = "default"
	startingDefaultConnectionPacketLossValue                                   = 0

	filesArtifactExpansionDirsParentDirpath = "/files-artifacts"

	// TODO This should be populated from the build flow that builds the files-artifacts-expander Docker image
	filesArtifactsExpanderImage = "kurtosistech/files-artifacts-expander"

	minMemoryLimit              = 6 // Docker doesn't allow memory limits less than 6 megabytes
	defaultMemoryAllocMegabytes = 0
)

// Guaranteed (by a unit test) to be a 1:1 mapping between API port protos and port spec protos
var apiContainerPortProtoToPortSpecPortProto = map[kurtosis_core_rpc_api_bindings.Port_Protocol]port_spec.PortProtocol{
	kurtosis_core_rpc_api_bindings.Port_TCP:  port_spec.PortProtocol_TCP,
	kurtosis_core_rpc_api_bindings.Port_SCTP: port_spec.PortProtocol_SCTP,
	kurtosis_core_rpc_api_bindings.Port_UDP:  port_spec.PortProtocol_UDP,
}

type storeFilesArtifactResult struct {
	filesArtifactUuid enclave_data_directory.FilesArtifactUUID
	err               error
}

/*
This is the in-memory representation of the service network that the API container will manipulate. To make

	any changes to the test network, this struct must be used.
*/
type ServiceNetwork struct {
	enclaveId enclave.EnclaveID

	apiContainerIpAddress   net.IP
	apiContainerGrpcPortNum uint16
	apiContainerVersion     string

	mutex *sync.Mutex // VERY IMPORTANT TO CHECK AT THE START OF EVERY METHOD!

	// Whether partitioning has been enabled for this particular test
	isPartitioningEnabled bool

	kurtosisBackend backend_interface.KurtosisBackend

	enclaveDataDir *enclave_data_directory.EnclaveDataDirectory

	topology *partition_topology.PartitionTopology

	networkingSidecars map[service.ServiceID]networking_sidecar.NetworkingSidecarWrapper

	networkingSidecarManager networking_sidecar.NetworkingSidecarManager

	// Technically we SHOULD query the backend rather than ever storing any of this information, but we're able to get away with
	// this because the API container is the only client that modifies service state
	registeredServiceInfo map[service.ServiceID]*service.ServiceRegistration
}

func NewServiceNetwork(
	enclaveId enclave.EnclaveID,
	apiContainerIpAddr net.IP,
	apiContainerGrpcPortNum uint16,
	apiContainerVersion string,
	isPartitioningEnabled bool,
	kurtosisBackend backend_interface.KurtosisBackend,
	enclaveDataDir *enclave_data_directory.EnclaveDataDirectory,
	networkingSidecarManager networking_sidecar.NetworkingSidecarManager,
) *ServiceNetwork {
	defaultPartitionConnection := partition_topology.PartitionConnection{
		PacketLossPercentage: startingDefaultConnectionPacketLossValue,
	}
	return &ServiceNetwork{
		enclaveId:               enclaveId,
		apiContainerIpAddress:   apiContainerIpAddr,
		apiContainerGrpcPortNum: apiContainerGrpcPortNum,
		apiContainerVersion:     apiContainerVersion,
		mutex:                   &sync.Mutex{},
		isPartitioningEnabled:   isPartitioningEnabled,
		kurtosisBackend:         kurtosisBackend,
		enclaveDataDir:          enclaveDataDir,
		topology: partition_topology.NewPartitionTopology(
			defaultPartitionId,
			defaultPartitionConnection,
		),
		networkingSidecars:       map[service.ServiceID]networking_sidecar.NetworkingSidecarWrapper{},
		networkingSidecarManager: networkingSidecarManager,
		registeredServiceInfo:    map[service.ServiceID]*service.ServiceRegistration{},
	}
}

/*
Completely repartitions the network, throwing away the old topology
*/
func (network *ServiceNetwork) Repartition(
	ctx context.Context,
	newPartitionServices map[service_network_types.PartitionID]map[service.ServiceID]bool,
	newPartitionConnections map[service_network_types.PartitionConnectionID]partition_topology.PartitionConnection,
	newDefaultConnection partition_topology.PartitionConnection,
) error {
	network.mutex.Lock()
	defer network.mutex.Unlock()

	if !network.isPartitioningEnabled {
		return stacktrace.NewError("Cannot repartition; partitioning is not enabled")
	}

	if err := network.topology.Repartition(newPartitionServices, newPartitionConnections, newDefaultConnection); err != nil {
		return stacktrace.Propagate(err, "An error occurred repartitioning the network topology")
	}

	servicePacketLossConfigurationsByServiceID, err := network.topology.GetServicePacketLossConfigurationsByServiceID()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the packet loss configuration by service ID "+
			" after repartition, meaning that no partitions are actually being enforced!")
	}

	if err := updateTrafficControlConfiguration(ctx, servicePacketLossConfigurationsByServiceID, network.registeredServiceInfo, network.networkingSidecars); err != nil {
		return stacktrace.Propagate(err, "An error occurred updating the traffic control configuration to match the target service packet loss configurations after repartitioning")
	}
	return nil
}

// Starts the services in the given partition in their own containers
//
// This is a bulk operation that follows a funnel/rollback approach.
// This means that when an error occurs, for an indvidiual operation (service in this case), we add it to a set of
// failed service ids to errors, and return that to the client of this function. At the sametime we rollback/undo any
// resources that were created during the failed operation, thus narrowing the funnel of operations
// that were successful. Thus, this function:
// Returns:
//   - successfulService - mapping of successful service ids to service objects with info about that service
//   - failedServices - mapping of failed service ids to errors causing those failures
//   - error	- if error occurred with bulk operation as a whole
func (network *ServiceNetwork) StartServices(
	ctx context.Context,
	serviceConfigs map[service.ServiceID]*kurtosis_core_rpc_api_bindings.ServiceConfig,
	partitionID service_network_types.PartitionID,
) (
	resultSuccessfulServices map[service.ServiceID]*service.Service,
	resultFailedServices map[service.ServiceID]error,
	resultErr error,
) {
	// TODO extract this into a wrapper function that can be wrapped around every service call (so we don't forget)
	network.mutex.Lock()
	defer network.mutex.Unlock()
	failedServicesPool := map[service.ServiceID]error{}

	servicesToStart := map[service.ServiceID]*kurtosis_core_rpc_api_bindings.ServiceConfig{}
	for serviceID, config := range serviceConfigs {
		if _, found := network.registeredServiceInfo[serviceID]; found {
			failedServicesPool[serviceID] = stacktrace.NewError(
				"Cannot start service '%v' because it already exists in the network",
				serviceID,
			)
			continue
		}
		servicesToStart[serviceID] = config
	}

	if partitionID == "" {
		partitionID = defaultPartitionId
	}
	if _, found := network.topology.GetPartitionServices()[partitionID]; !found {
		return nil, nil, stacktrace.NewError(
			"No partition with ID '%v' exists in the current partition topology",
			partitionID,
		)
	}

	successfulStarts, failedStarts, err := network.startServices(ctx, servicesToStart)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred attempting to add services to the service network.")
	}
	serviceGUIDsToRemove := map[service.ServiceGUID]bool{}
	for _, serviceInfo := range successfulStarts {
		guid := serviceInfo.GetRegistration().GetGUID()
		serviceGUIDsToRemove[guid] = true
	}
	// defer undo all to destroy services that fail in a later phase
	defer func() {
		if len(serviceGUIDsToRemove) == 0 {
			return
		}
		userServiceFilters := &service.ServiceFilters{
			GUIDs: serviceGUIDsToRemove,
		}
		_, failedToDestroyGUIDs, err := network.kurtosisBackend.DestroyUserServices(context.Background(), network.enclaveId, userServiceFilters)
		if err != nil {
			logrus.Errorf("Attempted to destroy all services with GUIDs '%v' together but had no success. You must manually destroy the services! The following error had occurred:\n'%v'", serviceGUIDsToRemove, err)
			return
		}
		if len(failedToDestroyGUIDs) == 0 {
			return
		}
		destroyFailuresAccountedFor := 0
		for serviceID, serviceInfo := range successfulStarts {
			guid := serviceInfo.GetRegistration().GetGUID()
			destroyErr, found := failedToDestroyGUIDs[guid]
			if !found {
				continue
			}
			logrus.Errorf("Failed to destroy the service '%v' after it failed to start. You must manually destroy the service! The following error had occurred:\n'%v'", serviceID, destroyErr)
			destroyFailuresAccountedFor += 1
		}
		if destroyFailuresAccountedFor != len(failedToDestroyGUIDs) {
			logrus.Errorf("Couldn't propagate all failures while destroying services. This is a bug in Kurtosis. Accounted for '%v', expected '%v' services.", destroyFailuresAccountedFor, len(failedToDestroyGUIDs))
		}
	}()

	// We add all the successfully started services to a list of services that will be processed in later phases
	servicesToProcessFurther := map[service.ServiceID]*service.Service{}
	for serviceID, serviceInfo := range successfulStarts {
		servicesToProcessFurther[serviceID] = serviceInfo
	}
	for serviceID, err := range failedStarts {
		failedServicesPool[serviceID] = err
	}

	for serviceID, serviceInfo := range servicesToProcessFurther {
		network.registeredServiceInfo[serviceID] = serviceInfo.GetRegistration()
	}
	servicesToRemoveFromRegistrationMap := map[service.ServiceID]bool{}
	for serviceID := range servicesToProcessFurther {
		servicesToRemoveFromRegistrationMap[serviceID] = true
	}
	defer func() {
		if len(servicesToRemoveFromRegistrationMap) == 0 {
			return
		}
		for serviceID := range servicesToRemoveFromRegistrationMap {
			delete(network.registeredServiceInfo, serviceID)
		}
	}()

	for serviceID := range servicesToProcessFurther {
		err = network.addServiceToTopology(serviceID, partitionID)
		if err != nil {
			failedServicesPool[serviceID] = stacktrace.Propagate(err, "An error occurred adding service '%v' to the topology", serviceID)
			delete(servicesToProcessFurther, serviceID)
			continue
		}
		logrus.Debugf("Successfully added service with ID '%v' to topology", serviceID)
	}
	serviceIDsForTopologyCleanup := map[service.ServiceID]bool{}
	for serviceID := range servicesToProcessFurther {
		serviceIDsForTopologyCleanup[serviceID] = true
	}
	defer func() {
		if len(serviceIDsForTopologyCleanup) == 0 {
			return
		}
		for serviceID := range serviceIDsForTopologyCleanup {
			network.topology.RemoveService(serviceID)
		}
	}()

	// TODO Fix race condition
	// There is race condition here.
	// 1. We first start the target services
	// 2. Then we create the sidecars for the target services
	// 3. Only then we block access between the target services & the rest of the world (both ways)
	// Between 1 & 3 the target & others can speak to each other if they choose to (eg: run a port scan)
	sidecarsToCleanUp := map[service.ServiceID]bool{}
	if network.isPartitioningEnabled {
		for serviceID, serviceInfo := range servicesToProcessFurther {
			err = network.createSidecarAndAddToMap(ctx, serviceInfo)
			if err != nil {
				failedServicesPool[serviceID] = stacktrace.Propagate(err, "An error occurred while adding networking sidecar for service '%v'", serviceID)
				delete(servicesToProcessFurther, serviceID)
				continue
			}
			logrus.Debugf("Successfully created sidecars for service with ID '%v'", serviceID)
		}
		for serviceID := range servicesToProcessFurther {
			sidecarsToCleanUp[serviceID] = true
		}
		// This defer-undo undoes resources created by `createSidecarAndAddToMap` in the reverse order of creation
		defer func() {
			if len(sidecarsToCleanUp) == 0 {
				return
			}
			for serviceID := range sidecarsToCleanUp {
				networkingSidecar, found := network.networkingSidecars[serviceID]
				if !found {
					logrus.Errorf("Tried cleaning up sidecar for service with ID '%v' but couldn't retrieve it from the cache. This is a Kurtosis bug.", serviceID)
					continue
				}
				err = network.networkingSidecarManager.Remove(ctx, networkingSidecar)
				if err != nil {
					logrus.Errorf("Attempted to clean up the sidecar for service with ID '%v' but the following error occurred:\n'%v'", serviceID, err)
					continue
				}
				delete(network.networkingSidecars, serviceID)
			}
		}()

		servicePacketLossConfigurationsByServiceID, err := network.topology.GetServicePacketLossConfigurationsByServiceID()
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred getting the packet loss configuration by service ID "+
				" to know what packet loss updates to apply")
		}

		// We apply all the configurations. We can't filter to source/target being a service started in this method call as we'd miss configurations between existing services.
		// The updates completely replace the tables, and we can't lose partitioning between existing services
		if err := updateTrafficControlConfiguration(ctx, servicePacketLossConfigurationsByServiceID, network.registeredServiceInfo, network.networkingSidecars); err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred applying the traffic control configuration to partition off new nodes.")
		}
		logrus.Debugf("Successfully applied qdisc configurations")
		// We don't need to undo the traffic control changes because in the worst case existing nodes have entries in their traffic control for IP addresses that don't resolve to any containers.
	}

	// All processing is done so the services can be marked successful
	successfulServicePool := map[service.ServiceID]*service.Service{}
	for serviceID, serviceInfo := range servicesToProcessFurther {
		successfulServicePool[serviceID] = serviceInfo
	}
	logrus.Infof("Sueccesfully started services '%v' and failed '%v' in the service network", successfulServicePool, failedServicesPool)
	for serviceID, serviceInfo := range successfulServicePool {
		guid := serviceInfo.GetRegistration().GetGUID()
		delete(serviceGUIDsToRemove, guid)
		delete(servicesToRemoveFromRegistrationMap, serviceID)
		delete(serviceIDsForTopologyCleanup, serviceID)
		delete(sidecarsToCleanUp, serviceID)
	}
	return successfulServicePool, failedServicesPool, nil
}

func (network *ServiceNetwork) RemoveService(
	ctx context.Context,
	serviceId service.ServiceID,
) (service.ServiceGUID, error) {
	network.mutex.Lock()
	defer network.mutex.Unlock()

	serviceToRemove, found := network.registeredServiceInfo[serviceId]
	if !found {
		return "", stacktrace.NewError("No service found with ID '%v'", serviceId)
	}
	serviceGuid := serviceToRemove.GetGUID()

	network.topology.RemoveService(serviceId)

	delete(network.registeredServiceInfo, serviceId)

	// We stop the service, rather than destroying it, so that we can keep logs around
	stopServiceFilters := &service.ServiceFilters{
		GUIDs: map[service.ServiceGUID]bool{
			serviceGuid: true,
		},
	}
	_, erroredGuids, err := network.kurtosisBackend.StopUserServices(ctx, network.enclaveId, stopServiceFilters)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred during the call to stop service '%v'", serviceGuid)
	}
	if err, found := erroredGuids[serviceGuid]; found {
		return "", stacktrace.Propagate(err, "An error occurred stopping service '%v'", serviceGuid)
	}

	sidecar, foundSidecar := network.networkingSidecars[serviceId]
	if network.isPartitioningEnabled && foundSidecar {
		// NOTE: As of 2020-12-31, we don't need to update the iptables of the other services in the network to
		//  clear the now-removed service's IP because:
		// 	 a) nothing is using it so it doesn't do anything and
		//	 b) all service's iptables get overwritten on the next Add/Repartition call
		// If we ever do incremental iptables though, we'll need to fix all the other service's iptables here!
		if err := network.networkingSidecarManager.Remove(ctx, sidecar); err != nil {
			return "", stacktrace.Propagate(err, "An error occurred destroying the sidecar for service with ID '%v'", serviceId)
		}
		delete(network.networkingSidecars, serviceId)
		logrus.Debugf("Successfully removed sidecar attached to service with ID '%v'", serviceId)
	}

	return serviceGuid, nil
}

// TODO we could switch this to be a bulk command; the backend would support it
func (network *ServiceNetwork) PauseService(
	ctx context.Context,
	serviceId service.ServiceID,
) error {
	network.mutex.Lock()
	defer network.mutex.Unlock()

	serviceObj, found := network.registeredServiceInfo[serviceId]
	if !found {
		return stacktrace.NewError("No service with ID '%v' exists in the network", serviceId)
	}

	if err := network.kurtosisBackend.PauseService(ctx, network.enclaveId, serviceObj.GetGUID()); err != nil {
		return stacktrace.Propagate(err, "Failed to pause service '%v'", serviceId)
	}
	return nil
}

// TODO we could switch this to be a bulk command; the backend would support it
func (network *ServiceNetwork) UnpauseService(
	ctx context.Context,
	serviceId service.ServiceID,
) error {
	network.mutex.Lock()
	defer network.mutex.Unlock()

	serviceObj, found := network.registeredServiceInfo[serviceId]
	if !found {
		return stacktrace.NewError("No service with ID '%v' exists in the network", serviceId)
	}

	if err := network.kurtosisBackend.UnpauseService(ctx, network.enclaveId, serviceObj.GetGUID()); err != nil {
		return stacktrace.Propagate(err, "Failed to unpause service '%v'", serviceId)
	}
	return nil
}

func (network *ServiceNetwork) ExecCommand(
	ctx context.Context,
	serviceId service.ServiceID,
	command []string,
) (int32, string, error) {
	// NOTE: This will block all other operations while this command is running!!!! We might need to change this so it's
	// asynchronous
	network.mutex.Lock()
	defer network.mutex.Unlock()

	serviceObj, found := network.registeredServiceInfo[serviceId]
	if !found {
		return 0, "", stacktrace.NewError(
			"Service '%v does not exist in the network",
			serviceId,
		)
	}

	// NOTE: This is a SYNCHRONOUS command, meaning that the entire network will be blocked until the command finishes
	// In the future, this will likely be insufficient

	serviceGuid := serviceObj.GetGUID()
	userServiceCommand := map[service.ServiceGUID][]string{
		serviceGuid: command,
	}

	successfulExecCommands, failedExecCommands, err := network.kurtosisBackend.RunUserServiceExecCommands(
		ctx,
		network.enclaveId,
		userServiceCommand)
	if err != nil {
		return 0, "", stacktrace.Propagate(
			err,
			"An error occurred calling kurtosis backend to exec command '%v' against service '%v'",
			command,
			serviceId)
	}
	if len(failedExecCommands) > 0 {
		serviceExecErrs := []string{}
		for serviceGUID, err := range failedExecCommands {
			wrappedErr := stacktrace.Propagate(
				err,
				"An error occurred attempting to run a command in a service with GUID `%v'",
				serviceGUID,
			)
			serviceExecErrs = append(serviceExecErrs, wrappedErr.Error())
		}
		return 0, "", stacktrace.NewError(
			"One or more errors occurred attempting to exec command(s) in the service(s): \n%v",
			strings.Join(
				serviceExecErrs,
				"\n\n",
			),
		)
	}

	execResult, isFound := successfulExecCommands[serviceGuid]
	if !isFound {
		return 0, "", stacktrace.NewError(
			"Unable to find result from running exec command '%v' against service '%v'",
			command,
			serviceGuid)
	}

	return execResult.GetExitCode(), execResult.GetOutput(), nil
}

func (network *ServiceNetwork) GetService(ctx context.Context, serviceId service.ServiceID) (
	*service.Service,
	error,
) {
	network.mutex.Lock()
	defer network.mutex.Unlock()

	registration, found := network.registeredServiceInfo[serviceId]
	if !found {
		return nil, stacktrace.NewError("No service with ID '%v' exists", serviceId)
	}
	serviceGuid := registration.GetGUID()

	getServiceFilters := &service.ServiceFilters{
		GUIDs: map[service.ServiceGUID]bool{
			registration.GetGUID(): true,
		},
	}
	matchingServices, err := network.kurtosisBackend.GetUserServices(ctx, network.enclaveId, getServiceFilters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting service '%v'", serviceGuid)
	}
	if len(matchingServices) == 0 {
		return nil, stacktrace.NewError(
			"A registration exists for service GUID '%v' but no service objects were found; this indicates that the service was "+
				"registered but not started",
			serviceGuid,
		)
	}
	if len(matchingServices) > 1 {
		return nil, stacktrace.NewError("Found multiple service objects matching GUID '%v'", serviceGuid)
	}
	serviceObj, found := matchingServices[serviceGuid]
	if !found {
		return nil, stacktrace.NewError("Found exactly one service object, but it didn't match expected GUID '%v'", serviceGuid)
	}

	return serviceObj, nil
}

func (network *ServiceNetwork) GetServiceIDs() map[service.ServiceID]bool {

	serviceIDs := make(map[service.ServiceID]bool, len(network.registeredServiceInfo))

	for serviceId := range network.registeredServiceInfo {
		if _, ok := serviceIDs[serviceId]; !ok {
			serviceIDs[serviceId] = true
		}
	}
	return serviceIDs
}

func (network *ServiceNetwork) CopyFilesFromService(ctx context.Context, serviceId service.ServiceID, srcPath string) (enclave_data_directory.FilesArtifactUUID, error) {
	serviceObj, found := network.registeredServiceInfo[serviceId]
	if !found {
		return "", stacktrace.NewError("Cannot copy files from service '%v' because it does not exist in the network", serviceId)
	}
	serviceGuid := serviceObj.GetGUID()

	store, err := network.enclaveDataDir.GetFilesArtifactStore()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting the files artifact store")
	}

	pipeReader, pipeWriter := io.Pipe()
	defer pipeReader.Close()
	defer pipeWriter.Close()

	storeFilesArtifactResultChan := make(chan storeFilesArtifactResult)
	go func() {
		defer pipeReader.Close()

		//And finally pass it the .tgz file to the artifact file store
		filesArtifactUuId, storeFileErr := store.StoreFile(pipeReader)
		storeFilesArtifactResultChan <- storeFilesArtifactResult{
			filesArtifactUuid: filesArtifactUuId,
			err:               storeFileErr,
		}
	}()

	if err := network.gzipAndPushTarredFileBytesToOutput(ctx, pipeWriter, serviceGuid, srcPath); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred gzip'ing and pushing tar'd file bytes to the pipe")
	}

	storeFileResult := <-storeFilesArtifactResultChan
	if storeFileResult.err != nil {
		return "", stacktrace.Propagate(
			err,
			"An error occurred storing files from path '%v' on service '%v' in in the files artifact store",
			srcPath,
			serviceGuid,
		)
	}

	return storeFileResult.filesArtifactUuid, nil
}

// ====================================================================================================
// 									   Private helper methods
// ====================================================================================================
/*
Updates the traffic control configuration of the services with the given IDs to match the target services packet loss configuration

NOTE: This is not thread-safe, so it must be within a function that locks mutex!
*/
func updateTrafficControlConfiguration(
	ctx context.Context,
	targetServicePacketLossConfigs map[service.ServiceID]map[service.ServiceID]float32,
	services map[service.ServiceID]*service.ServiceRegistration,
	networkingSidecars map[service.ServiceID]networking_sidecar.NetworkingSidecarWrapper,
) error {

	// TODO PERF: Run the container updates in parallel, with the container being modified being the most important

	for serviceId, allOtherServicesPacketLossConfigurations := range targetServicePacketLossConfigs {
		allPacketLossPercentageForIpAddresses := map[string]float32{}
		for otherServiceId, otherServicePacketLossPercentage := range allOtherServicesPacketLossConfigurations {
			otherService, found := services[otherServiceId]
			if !found {
				return stacktrace.NewError(
					"Service with ID '%v' needs to add packet loss configuration for service with ID '%v', but the latter "+
						"doesn't have service registration info (i.e. an IP) associated with it",
					serviceId,
					otherServiceId)
			}

			allPacketLossPercentageForIpAddresses[otherService.GetPrivateIP().String()] = otherServicePacketLossPercentage
		}

		sidecar, found := networkingSidecars[serviceId]
		if !found {
			return stacktrace.NewError(
				"Need to update qdisc configuration of service with ID '%v', but the service doesn't have a sidecar",
				serviceId)
		}

		if err := sidecar.UpdateTrafficControl(ctx, allPacketLossPercentageForIpAddresses); err != nil {
			return stacktrace.Propagate(
				err,
				"An error occurred updating the qdisc configuration for service '%v'",
				serviceId)
		}
	}
	return nil
}

func (network *ServiceNetwork) startServices(
	ctx context.Context,
	APIServiceConfigs map[service.ServiceID]*kurtosis_core_rpc_api_bindings.ServiceConfig,
) (
	resultSuccessfulServices map[service.ServiceID]*service.Service,
	resultFailedServices map[service.ServiceID]error,
	resultErr error,
) {
	failedServicesPool := map[service.ServiceID]error{}
	serviceConfigs := map[service.ServiceID]*service.ServiceConfig{}

	// Convert API ServiceConfig's to service.ServiceConfig's by:
	// - converting API Ports to PortSpec's
	// - converting files artifacts mountpoints to FilesArtifactsExpansion's'
	// - passing down other data (eg. container image name, args, etc.)
	for serviceID, config := range APIServiceConfigs {
		// Docker and K8s requires the minimum memory limit to be 6 megabytes to we make sure the allocation is at least that amount
		// But first, we check that it's not the default value, meaning the user potentially didn't even set it
		if config.MemoryAllocationMegabytes != defaultMemoryAllocMegabytes && config.MemoryAllocationMegabytes < minMemoryLimit {
			failedServicesPool[serviceID] = stacktrace.NewError("Memory allocation, `%d`, is too low. Kurtosis requires the memory limit to be at least `%d` megabytes for service with ID '%v'.", config.MemoryAllocationMegabytes, minMemoryLimit, serviceID)
			continue
		}

		// Convert ports
		privateServicePortSpecs, requestedPublicServicePortSpecs, err := convertAPIPortsToPortSpecs(config.PrivatePorts, config.PublicPorts)
		if err != nil {
			failedServicesPool[serviceID] = stacktrace.Propagate(err, "An error occurred while trying to convert public and private API ports to port specs for service with ID '%v'", serviceID)
			continue
		}

		// Creates files artifacts expansions
		var filesArtifactsExpansion *files_artifacts_expansion.FilesArtifactsExpansion
		if len(config.FilesArtifactMountpoints) == 0 {
			// Create service config with empty filesArtifactsExpansion if no files artifacts mountpoints for this service
			serviceConfigs[serviceID] = service.NewServiceConfig(
				config.ContainerImageName,
				privateServicePortSpecs,
				requestedPublicServicePortSpecs,
				config.EntrypointArgs,
				config.CmdArgs,
				config.EnvVars,
				filesArtifactsExpansion,
				config.CpuAllocationMillicpus,
				config.MemoryAllocationMegabytes,
				config.PrivateIpAddrPlaceholder)
			continue
		}

		filesArtifactsExpansions := []args.FilesArtifactExpansion{}
		expanderDirpathToUserServiceDirpathMap := map[string]string{}
		for filesArtifactUUIDStr, mountpointOnUserService := range config.FilesArtifactMountpoints {
			dirpathToExpandTo := path.Join(filesArtifactExpansionDirsParentDirpath, filesArtifactUUIDStr)
			expansion := args.FilesArtifactExpansion{
				FilesArtifactUUID: filesArtifactUUIDStr,
				DirPathToExpandTo: dirpathToExpandTo,
			}
			filesArtifactsExpansions = append(filesArtifactsExpansions, expansion)

			expanderDirpathToUserServiceDirpathMap[dirpathToExpandTo] = mountpointOnUserService
		}

		filesArtifactsExpanderArgs, err := args.NewFilesArtifactsExpanderArgs(
			network.apiContainerIpAddress.String(),
			network.apiContainerGrpcPortNum,
			filesArtifactsExpansions,
		)
		if err != nil {
			failedServicesPool[serviceID] = stacktrace.Propagate(err, "An error occurred creating files artifacts expander args for service `%v`", serviceID)
			continue
		}
		expanderEnvVars, err := args.GetEnvFromArgs(filesArtifactsExpanderArgs)
		if err != nil {
			failedServicesPool[serviceID] = stacktrace.Propagate(err, "An error occurred getting files artifacts expander environment variables using args: %+v", filesArtifactsExpanderArgs)
			continue
		}

		expanderImageAndTag := fmt.Sprintf(
			"%v:%v",
			filesArtifactsExpanderImage,
			network.apiContainerVersion,
		)

		filesArtifactsExpansion = &files_artifacts_expansion.FilesArtifactsExpansion{
			ExpanderImage:                     expanderImageAndTag,
			ExpanderEnvVars:                   expanderEnvVars,
			ExpanderDirpathsToServiceDirpaths: expanderDirpathToUserServiceDirpathMap,
		}

		serviceConfigs[serviceID] = service.NewServiceConfig(
			config.ContainerImageName,
			privateServicePortSpecs,
			requestedPublicServicePortSpecs,
			config.EntrypointArgs,
			config.CmdArgs,
			config.EnvVars,
			filesArtifactsExpansion,
			config.CpuAllocationMillicpus,
			config.MemoryAllocationMegabytes,
			config.PrivateIpAddrPlaceholder)
	}

	successfulServices, failedServices, err := network.kurtosisBackend.StartUserServices(ctx, network.enclaveId, serviceConfigs)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred starting user services in underlying container engine.")
	}

	// Add services that failed to start to failed services pool
	for serviceID, serviceErr := range failedServices {
		failedServicesPool[serviceID] = serviceErr
	}

	return successfulServices, failedServicesPool, nil
}

func (network *ServiceNetwork) gzipAndPushTarredFileBytesToOutput(
	ctx context.Context,
	output io.WriteCloser,
	serviceGuid service.ServiceGUID,
	srcPathOnContainer string,
) error {
	defer output.Close()

	// Need to compress the TAR bytes on our side, since we're not guaranteedj
	gzippingOutput := gzip.NewWriter(output)
	defer gzippingOutput.Close()

	if err := network.kurtosisBackend.CopyFilesFromUserService(ctx, network.enclaveId, serviceGuid, srcPathOnContainer, gzippingOutput); err != nil {
		return stacktrace.Propagate(err, "An error occurred copying source '%v' from user service with GUID '%v' in enclave with ID '%v'", srcPathOnContainer, serviceGuid, network.enclaveId)
	}

	return nil
}

// This method is not thread safe. Only call this from a method where there is a mutex lock on the network.
func (network *ServiceNetwork) addServiceToTopology(serviceID service.ServiceID, partitionID service_network_types.PartitionID) error {
	if err := network.topology.AddService(serviceID, partitionID); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred adding service with ID '%v' to partition '%v' in the topology",
			serviceID,
			partitionID,
		)
	}
	shouldRemoveFromTopology := true
	defer func() {
		if shouldRemoveFromTopology {
			network.topology.RemoveService(serviceID)
		}
	}()

	shouldRemoveFromTopology = false
	return nil
}

// This method is not thread safe. Only call this from a method where there is a mutex lock on the network.
func (network *ServiceNetwork) createSidecarAndAddToMap(ctx context.Context, service *service.Service) error {
	serviceRegistration := service.GetRegistration()
	serviceGUID := serviceRegistration.GetGUID()
	serviceID := serviceRegistration.GetID()

	sidecar, err := network.networkingSidecarManager.Add(ctx, serviceGUID)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred adding the networking sidecar for service `%v`", serviceID)
	}
	shouldRemoveSidecarFromManager := true
	defer func() {
		if shouldRemoveSidecarFromManager {
			err = network.networkingSidecarManager.Remove(ctx, sidecar)
			if err != nil {
				logrus.Errorf("Attempted to remove network sidecar during cleanup for service '%v' but failed", serviceID)
			}
		}
	}()

	network.networkingSidecars[serviceID] = sidecar
	shouldRemoveSidecarFromMap := true
	defer func() {
		if shouldRemoveSidecarFromMap {
			delete(network.networkingSidecars, serviceID)
		}
	}()

	if err := sidecar.InitializeTrafficControl(ctx); err != nil {
		return stacktrace.Propagate(err, "An error occurred initializing the newly-created networking-sidecar-traffic-control-qdisc-configuration for service `%v`", serviceID)
	}

	shouldRemoveSidecarFromMap = false
	shouldRemoveSidecarFromManager = false
	return nil
}

func convertAPIPortsToPortSpecs(
	privateAPIPorts map[string]*kurtosis_core_rpc_api_bindings.Port,
	publicAPIPorts map[string]*kurtosis_core_rpc_api_bindings.Port,
) (
	resultPrivatePortSpecs map[string]*port_spec.PortSpec,
	resultPublicPortSpecs map[string]*port_spec.PortSpec,
	resultErr error,
) {
	privatePortSpecs := map[string]*port_spec.PortSpec{}
	for portID, privateAPIPort := range privateAPIPorts {
		privatePortSpec, err := transformApiPortToPortSpec(privateAPIPort)
		if err != nil {
			return nil, nil, stacktrace.NewError("An error occurred transforming the API port for private port '%v' into a port spec port", portID)
		}
		privatePortSpecs[portID] = privatePortSpec
	}

	//TODO this is a huge hack to temporarily enable static ports for NEAR until we have a more productized solution
	if len(publicAPIPorts) > 0 {
		err := checkPrivateAndPublicPortsAreOneToOne(privateAPIPorts, publicAPIPorts)
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "Provided public and private ports are not one to one.")
		}
	}

	publicPortSpecs := map[string]*port_spec.PortSpec{}
	for portID, publicAPIPort := range publicAPIPorts {
		publicPortSpec, err := transformApiPortToPortSpec(publicAPIPort)
		if err != nil {
			return nil, nil, stacktrace.NewError("An error occurred transforming the API port for public port '%v' into a port spec port", portID)
		}
		publicPortSpecs[portID] = publicPortSpec
	}
	//TODO Finished the huge hack to temporarily enable static ports for NEAR
	return privatePortSpecs, publicPortSpecs, nil
}

func transformApiPortToPortSpec(port *kurtosis_core_rpc_api_bindings.Port) (*port_spec.PortSpec, error) {
	portNumUint32 := port.GetNumber()
	apiProto := port.GetProtocol()
	if portNumUint32 > math.MaxUint16 {
		return nil, stacktrace.NewError(
			"API port num '%v' is bigger than max allowed port spec port num '%v'",
			portNumUint32,
			math.MaxUint16,
		)
	}
	portNumUint16 := uint16(portNumUint32)
	portSpecProto, found := apiContainerPortProtoToPortSpecPortProto[apiProto]
	if !found {
		return nil, stacktrace.NewError("Couldn't find a port spec proto for API port proto '%v'; this should never happen, and is a bug in Kurtosis!", apiProto.String())
	}
	result, err := port_spec.NewPortSpec(portNumUint16, portSpecProto)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating port spec object with port num '%v' and protocol '%v'",
			portNumUint16,
			portSpecProto,
		)
	}
	return result, nil
}

// Ensure that provided [privatePorts] and [publicPorts] are one to one by checking:
// - There is a matching publicPort for every portID in privatePorts
// - There are the same amount of private and public ports
// If error is nil, the public and private ports are one to one.
func checkPrivateAndPublicPortsAreOneToOne(privatePorts map[string]*kurtosis_core_rpc_api_bindings.Port, publicPorts map[string]*kurtosis_core_rpc_api_bindings.Port) error {
	if len(privatePorts) != len(publicPorts) {
		return stacktrace.NewError("The received private ports length and the public ports length are not equal. Received '%v' private ports and '%v' public ports", len(privatePorts), len(publicPorts))
	}

	for portID, privatePortSpec := range privatePorts {
		if _, found := publicPorts[portID]; !found {
			return stacktrace.NewError("Expected to receive public port with ID '%v' bound to private port number '%v', but it was not found", portID, privatePortSpec.GetNumber())
		}
	}
	return nil
}
