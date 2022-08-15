/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package service_network

import (
	"compress/gzip"
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifacts_expansion"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis-core/api/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-core/files_artifacts_expander/args"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/service_network/networking_sidecar"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/kurtosis-core/server/commons/enclave_data_directory"
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
	filesArtifactsExpanderImage = "kurtosistech/kurtosis-files-artifacts-expander"

	minMemoryLimit = 6 // Docker doesn't allow memory limits less than 6 megabytes
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

// Registers services for use within the network (creating the IPs and so forth), but doesn't start them
// If the partition ID is empty, registers the services with the default partition.
//
// This is a bulk operation that follows a funnel/rollback approach.
// This means that when an error occurs, for an indvidiual operation (service in this case), we add it to a set of
// failed service ids to errors, and return that to the client of this function. At the sametime we rollback/undo any
// resources that were created during the failed operation, thus narrowing the funnel of operations
// that were successful. Thus, this function:
// Returns:
//	- successfulService - mapping of successful service ids to private ip address of the service
//	- failedServices - mapping of failed service ids to errors causing those failures
//	- error	- if error occurred with bulk operation as a whole
func (network ServiceNetwork) RegisterServices(
	ctx context.Context,
	serviceIDs map[service.ServiceID]bool,
	partitionID service_network_types.PartitionID,
) (
	resultSuccessfulServiceRegistrations map[service.ServiceID]net.IP,
	resultFailedServiceRegistrations map[service.ServiceID]error,
	resultErr error,
) {
	// TODO extract this into a wrapper function that can be wrapped around every service call (so we don't forget)
	network.mutex.Lock()
	defer network.mutex.Unlock()
	failedServicePool := map[service.ServiceID]error{}
	serviceIDsToRegister := map[service.ServiceID]bool{}
	for id, _ := range serviceIDs {
		serviceIDsToRegister[id] = true
	}

	for serviceID, _ := range serviceIDs {
		if _, found := network.registeredServiceInfo[serviceID]; found {
			failedServicePool[serviceID] = stacktrace.NewError(
				"Cannot register service '%v' because it already exists in the network",
				serviceID,
			)
			delete(serviceIDsToRegister, serviceID)
		}
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

	successfulRegistrations, failedRegistrations, err := network.kurtosisBackend.RegisterUserServices(
		ctx,
		network.enclaveId,
		serviceIDsToRegister,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred registering services with IDs '%v'", serviceIDs)
	}
	// Defer an undo to all the successful registrations in case an error occurs in future phases
	shouldRemoveServices := map[service.ServiceID]bool{}
	for serviceID, serviceRegistration := range successfulRegistrations {
		shouldRemoveServices[serviceID] = true
		defer func() {
			if shouldRemoveServices[serviceID] {
				err = network.destroyServiceAfterFailure(context.Background(), serviceRegistration.GetGUID())
				if err != nil {
					failedServicePool[serviceID] = stacktrace.Propagate(err,
						"WARNING: Attempted to remove service '%v' after it failed to register but an error occurredwhile doing so." +
						"Must destroy service manually!", serviceID)
				}
			}
		}()
	}
	for serviceID, registrationErr := range failedRegistrations {
		failedServicePool[serviceID] = stacktrace.Propagate(registrationErr, "Failed to register service with ID '%v'", serviceID)
	}

	successfulServiceIPs := map[service.ServiceID]net.IP{}
	for serviceID, serviceRegistration := range successfulRegistrations {
		network.registeredServiceInfo[serviceID] = serviceRegistration
		shouldRemoveFromServicesMap := true
		defer func() {
			if shouldRemoveFromServicesMap {
				delete(network.registeredServiceInfo, serviceID)
			}
		}()

		if err := network.topology.AddService(serviceID, partitionID); err != nil {
			failedServicePool[serviceID] = stacktrace.Propagate(
				err,
				"An error occurred adding service with ID '%v' to partition '%v' in the topology",
				serviceID,
				partitionID,
			)
			continue
		}
		shouldRemoveFromTopology := true
		defer func() {
			if shouldRemoveFromTopology {
				network.topology.RemoveService(serviceID)
			}
		}()

		shouldRemoveFromServicesMap = false
		shouldRemoveFromTopology = false
		privateIP := serviceRegistration.GetPrivateIP()
		successfulServiceIPs[serviceID] = privateIP
	}

	// Do not remove services that were successful
	for serviceID, _ := range successfulServiceIPs {
		shouldRemoveServices[serviceID] = false
	}
	return successfulServiceIPs, failedServicePool, nil
}

// Starts a previously-registered but not-started service by creating it in a container.
//
// This is a bulk operation that follows a funnel/rollback approach.
// This means that when an error occurs, for an indvidiual operation (service in this case), we add it to a set of
// failed service ids to errors, and return that to the client of this function. At the sametime we rollback/undo any
// resources that were created during the failed operation, thus narrowing the funnel of operations
// that were successful. Thus, this function:
// Returns:
//	- successfulService - mapping of successful service ids to service objects with info about that service
//	- failedServices - mapping of failed service ids to errors causing those failures
//	- error	- if error occurred with bulk operation as a whole
func(network *ServiceNetwork) StartServices(
	ctx context.Context,
	apiServiceConfigs map[service.ServiceID]*kurtosis_core_rpc_api_bindings.ServiceConfig,
) (
	resultSuccessfulServices map[service.ServiceID]*service.Service,
	resultFailedServices map[service.ServiceID]error,
	resultErr error,
) {
	// TODO extract this into a wrapper function that can be wrapped around every service call (so we don't forget)
	network.mutex.Lock()
	defer network.mutex.Unlock()
	failedServicesPool := map[service.ServiceID]error{}

	serviceGUIDsToIDs := map[service.ServiceGUID]service.ServiceID{}
	serviceGUIDsToConfigs := map[service.ServiceGUID]*kurtosis_core_rpc_api_bindings.ServiceConfig{}
	for serviceID, apiServiceConfig := range apiServiceConfigs {
		registration, found := network.registeredServiceInfo[serviceID]
		if !found {
			failedServicesPool[serviceID] = stacktrace.NewError("Cannot start service; no registration exists for service with ID '%v'", serviceID)
			continue
		}

		guid := registration.GetGUID()
		serviceGUIDsToIDs[guid] = serviceID
		serviceGUIDsToConfigs[guid] = apiServiceConfig
	}

	// When partitioning is enabled, there's a race condition where:
	//   a) we need to start the services before we can launch the sidecar but
	//   b) we can't modify the qdisc configurations until the sidecar container is launched.
	// This means that there's a period of time at startup where the containers might not be partitioned. We solve
	// this by setting the packet loss config of the new services in the already-existing services' qdisc.
	// This means that when the new services are launched, even if their own qdisc isn't yet updated, all the services
	// it would communicate are already dropping traffic to it before it even starts.
	if network.isPartitioningEnabled {
		// The services being started should have been added to the network topology when they were registered
		servicePacketLossConfigurationsByServiceID, err := network.topology.GetServicePacketLossConfigurationsByServiceID()
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred getting the packet loss configuration by service ID "+
				" to know what packet loss updates to apply on the new node")
		}

		servicesPacketLossConfigurationsWithoutNewNodes := map[service.ServiceID]map[service.ServiceID]float32{}
		for serviceIdInTopology, otherServicesPacketLossConfigs := range servicePacketLossConfigurationsByServiceID {
			if _, found := apiServiceConfigs[serviceIdInTopology]; found {
				continue
			}
			servicesPacketLossConfigurationsWithoutNewNodes[serviceIdInTopology] = otherServicesPacketLossConfigs
		}

		if err := updateTrafficControlConfiguration(
			ctx,
			servicesPacketLossConfigurationsWithoutNewNodes,
			network.registeredServiceInfo,
			network.networkingSidecars,
		); err != nil {
			return nil, nil, stacktrace.Propagate(
				err,
				"An error occurred updating the traffic control configuration of all the other services "+
					"before adding the new service, meaning that the service wouldn't actually start in a partition",
			)
		}
	}

	successfulServiceGUIDs, failedServiceGUIDs, err := network.startServices(ctx, serviceGUIDsToConfigs)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred attempting to add services to the service network.")
	}
	// TODO We SHOULD defer an undo to undo the service-starting resource that we did here, but we don't have a way to just undo
	// that and leave the registration intact (since we only have RemoveService as of 2022-08-11, but that also deletes the registration,
	//which would mean deleting a resource we don't own here)

	// Convert from GUID maps to ID maps
	successfulServices := map[service.ServiceID]*service.Service{}
	for serviceGUID, serviceInfo := range successfulServiceGUIDs {
		serviceID, found := serviceGUIDsToIDs[serviceGUID]
		if !found {
			return nil, nil, stacktrace.NewError("Could not find a service ID associated with the service GUID '%v'. This should not happen and is a bug in Kurtosis.", serviceGUID)
		}
		successfulServices[serviceID] = serviceInfo
	}
	// Add services that failed to start to failed service pool
	for serviceGUID, serviceErr := range failedServiceGUIDs {
		serviceID, found := serviceGUIDsToIDs[serviceGUID]
		if !found {
			return nil, nil, stacktrace.NewError("Could not find a service ID associated with the service GUID '%v'. This should not happen and is a bug in Kurtosis.", serviceGUID)
		}
		failedServicesPool[serviceID] = serviceErr
	}

	if network.isPartitioningEnabled {
		// TODO Getting packet loss configuration by service ID is an expensive call and, as of 2021-11-23, we do it twice - the solution is to make
		//  Getting packet loss configuration by service ID not an expensive call
		servicePacketLossConfigurationsByServiceID, err := network.topology.GetServicePacketLossConfigurationsByServiceID()
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred getting the packet loss configuration by service ID "+
				" to know what packet loss updates to apply on the new node")
		}
		updatesToApply := map[service.ServiceID]map[service.ServiceID]float32{}

		// In the initial phase, network packets of services in the network were blocked from services that were about to be started.
		// Here, we are now blocking off successfully started services from the rest of the network to further guarantee network partitioning.
		// We don't undo the blocking off of failed services by the rest of the network because the services in the network are blocking traffic
		// from containers that don't exist anyways.
		for serviceID, serviceInfo := range successfulServices {
			serviceRegistration := serviceInfo.GetRegistration()
			serviceGUID := serviceRegistration.GetGUID()

			sidecar, err := network.networkingSidecarManager.Add(ctx, serviceGUID)
			if err != nil {
				failedServicesPool[serviceID] = stacktrace.Propagate(err, "An error occurred adding the networking sidecar for service `%v`", serviceID)
				delete(successfulServices, serviceID)
				continue
			}
			// TODO: When atomicity is implemented for StartServices, add a defer to undo creation of sidecar in case of failure

			network.networkingSidecars[serviceID] = sidecar
			// TODO: When atomicity is implemented for StartServices, add a defer to undo addition of sidecar to map

			if err := sidecar.InitializeTrafficControl(ctx); err != nil {
				failedServicesPool[serviceID] = stacktrace.Propagate(err, "An error occurred initializing the newly-created networking-sidecar-traffic-control-qdisc-configuration for service `%v`", serviceID)
				delete(successfulServices, serviceID)
				continue
			}

			newNodeServicePacketLossConfiguration := servicePacketLossConfigurationsByServiceID[serviceID]
			updatesToApply[serviceID] = newNodeServicePacketLossConfiguration
		}

		if err := updateTrafficControlConfiguration(ctx, updatesToApply, network.registeredServiceInfo, network.networkingSidecars); err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred applying the traffic control configuration on the new nodes to partition them "+
				"off from other nodes")
		}
	}

	return successfulServices, failedServicesPool, nil
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
		return nil, stacktrace.Propagate(
			err,
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

/*
func newServiceGUID(serviceID service.ServiceID) service.ServiceGUID {
	suffix := current_time_str_provider.GetCurrentTimeStr()
	return service.ServiceGUID(string(serviceID) + "-" + suffix)
}

func getServiceByServiceGUIDFilter(serviceGUID service.ServiceGUID) *service.ServiceFilters {
	return &service.ServiceFilters{
		GUIDs: map[service.ServiceGUID]bool{
			serviceGUID: true,
		},
	}
}

func gzipCompressFile(readCloser io.Reader) (resultFilepath string, resultErr error) {
	useDefaultDirectoryArg := ""
	withoutPatternArg := ""
	tgzFile, err := ioutil.TempFile(useDefaultDirectoryArg,withoutPatternArg)
	if err != nil {
		return "", stacktrace.Propagate(err,
			"There was an error creating a temporary file")
	}
	defer tgzFile.Close()

	gzipCompressingWriter := gzip.NewWriter(tgzFile)
	defer gzipCompressingWriter.Close()

	tarGzipFileFilepath := tgzFile.Name()
	if _, err := io.Copy(gzipCompressingWriter, readCloser); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred copying content to file '%v'", tarGzipFileFilepath)
	}

	return tarGzipFileFilepath, nil
}
*/

func (network *ServiceNetwork) destroyServiceAfterFailure(ctx context.Context, serviceGUID service.ServiceGUID) error {
	destroyServiceFilters := &service.ServiceFilters{
		GUIDs: map[service.ServiceGUID]bool{
			serviceGUID: true,
		},
	}
	// Use background context in case the input one is cancelled
	_, erroredRegistrations, err := network.kurtosisBackend.DestroyUserServices(ctx, network.enclaveId, destroyServiceFilters)
	var errToReturn error
	if err != nil {
		errToReturn = err
	} else if destroyErr, found := erroredRegistrations[serviceGUID]; found {
		errToReturn = destroyErr
	}
	if errToReturn != nil {
		return stacktrace.NewError("Attempted to destroy the service with GUID'%v', but doing so threw an error:\n%v", serviceGUID, errToReturn)
	}
	return nil
}

func (network *ServiceNetwork) startServices(
	ctx context.Context,
	APIServiceConfigs map[service.ServiceGUID]*kurtosis_core_rpc_api_bindings.ServiceConfig,
) (
	resultSuccessfulServices map[service.ServiceGUID]*service.Service,
	resultFailedServices map[service.ServiceGUID]error,
	resultErr error,
) {
	failedServicesPool := map[service.ServiceGUID]error{}
	serviceConfigs := map[service.ServiceGUID]*service.ServiceConfig{}

	// Convert API ServiceConfig's to service.ServiceConfig's by:
	// - converting API Ports to PortSpec's
	// - converting files artifacts mountpoints to FilesArtifactsExpansion's'
	// - passing down other data (eg. container image name, args, etc.)
	for guid, config := range APIServiceConfigs {
		// Docker and K8s requires the minimum memory limit to be 6 megabytes to we make sure the allocation is at least that amount
		// But first, we check that it's not the default value, meaning the user potentially didn't even set it
		if config.MemoryAllocationMegabytes != defaultMemoryAllocMegabytes && config.MemoryAllocationMegabytes < minMemoryLimit {
			failedServicesPool[guid] = stacktrace.NewError("Memory allocation, `%d`, is too low. Kurtosis requires the memory limit to be at least `%d` megabytes for service with GUID '%v'.", config.MemoryAllocationMegabytes, minMemoryLimit, guid)
			continue
		}

		// Convert ports
		privateServicePortSpecs, requestedPublicServicePortSpecs, err := convertAPIPortsToPortSpecs(config.PrivatePorts, config.PublicPorts)
		if err != nil {
			failedServicesPool[guid] = stacktrace.Propagate(err, "An error occurred while trying to convert public and private API ports to port specs for service with GUID '%v'", guid)
			continue
		}

		// Creates files artifacts expansions
		var filesArtifactsExpansion *files_artifacts_expansion.FilesArtifactsExpansion
		if len(config.FilesArtifactMountpoints) == 0 {
			// Create service config with empty filesArtifactsExpansion if no files artifacts mountpoints for this service
			serviceConfigs[guid] = service.NewServiceConfig(
				config.ContainerImageName,
				privateServicePortSpecs,
				requestedPublicServicePortSpecs,
				config.EntrypointArgs,
				config.CmdArgs,
				config.EnvVars,
				filesArtifactsExpansion,
				config.CpuAllocationMillicpus,
				config.MemoryAllocationMegabytes)
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
			failedServicesPool[guid] = stacktrace.Propagate(err, "An error occurred creating files artifacts expander args for service `%v`", guid)
			continue
		}
		expanderEnvVars, err := args.GetEnvFromArgs(filesArtifactsExpanderArgs)
		if err != nil {
			failedServicesPool[guid] =  stacktrace.Propagate(err, "An error occurred getting files artifacts expander environment variables using args: %+v", filesArtifactsExpanderArgs)
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

		serviceConfigs[guid] = service.NewServiceConfig(
			config.ContainerImageName,
			privateServicePortSpecs,
			requestedPublicServicePortSpecs,
			config.EntrypointArgs,
			config.CmdArgs,
			config.EnvVars,
			filesArtifactsExpansion,
			config.CpuAllocationMillicpus,
			config.MemoryAllocationMegabytes)
	}

	successfulServices, failedServices, err := network.kurtosisBackend.StartUserServices(ctx, network.enclaveId, serviceConfigs)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred starting user services in underlying container engine.")
	}

	// Add services that failed to start to failed services pool
	for guid, serviceErr := range failedServices {
		failedServicesPool[guid] = serviceErr
	}

	return successfulServices, failedServices, nil
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