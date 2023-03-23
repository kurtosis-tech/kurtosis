/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package service_network

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/shared_utils"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/files_artifacts_expansion"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
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
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"text/template"
)

const (
	filesArtifactExpansionDirsParentDirpath string = "/files-artifacts"

	// TODO This should be populated from the build flow that builds the files-artifacts-expander Docker image
	filesArtifactsExpanderImage string = "kurtosistech/files-artifacts-expander"

	minMemoryLimit              uint64 = 6 // Docker doesn't allow memory limits less than 6 megabytes
	defaultMemoryAllocMegabytes uint64 = 0

	folderPermissionForRenderedTemplates = 0755
	tempDirForRenderedTemplatesPrefix    = "temp-dir-for-rendered-templates-"

	ensureCompressedFileIsLesserThanGRPCLimit = false

	emptyCollectionLength        = 0
	exactlyOneShortenedUuidMatch = 1

	singleServiceStartupBatch = 1
)

var (
	// Guaranteed (by a unit test) to be a 1:1 mapping between API port protos and port spec protos
	apiContainerPortProtoToPortSpecPortProto = map[kurtosis_core_rpc_api_bindings.Port_TransportProtocol]port_spec.TransportProtocol{
		kurtosis_core_rpc_api_bindings.Port_TCP:  port_spec.TransportProtocol_TCP,
		kurtosis_core_rpc_api_bindings.Port_SCTP: port_spec.TransportProtocol_SCTP,
		kurtosis_core_rpc_api_bindings.Port_UDP:  port_spec.TransportProtocol_UDP,
	}

	emptyServiceNamesSetToUpdateAllConnections = map[service.ServiceName]bool{}
)

type storeFilesArtifactResult struct {
	err               error
	filesArtifactUuid enclave_data_directory.FilesArtifactUUID
}

// DefaultServiceNetwork is the in-memory representation of the service network that the API container will manipulate.
// To make any changes to the test network, this struct must be used.
type DefaultServiceNetwork struct {
	enclaveUuid enclave.EnclaveUUID

	apiContainerIpAddress   net.IP
	apiContainerGrpcPortNum uint16
	apiContainerVersion     string

	mutex *sync.Mutex // VERY IMPORTANT TO CHECK AT THE START OF EVERY METHOD!

	// Whether partitioning has been enabled for this particular test
	isPartitioningEnabled bool

	kurtosisBackend backend_interface.KurtosisBackend

	enclaveDataDir *enclave_data_directory.EnclaveDataDirectory

	topology *partition_topology.PartitionTopology

	// This is access in sub routine to start service in parallel. Hence, the lock right below
	// TODO: refactor this into its own class, or even better merge it with network topology into a super class
	//  that holds the complete description of the network in memory
	networkingSidecars  map[service.ServiceName]networking_sidecar.NetworkingSidecarWrapper
	networkSidecarsLock *sync.Mutex

	networkingSidecarManager networking_sidecar.NetworkingSidecarManager

	// Technically we SHOULD query the backend rather than ever storing any of this information, but we're able to get away with
	// this because the API container is the only client that modifies service state
	registeredServiceInfo map[service.ServiceName]*service.ServiceRegistration

	// This contains all service identifiers ever successfully created, this is append only
	allExistingAndHistoricalIdentifiers []*kurtosis_core_rpc_api_bindings.ServiceIdentifiers
}

func NewDefaultServiceNetwork(
	enclaveUuid enclave.EnclaveUUID,
	apiContainerIpAddr net.IP,
	apiContainerGrpcPortNum uint16,
	apiContainerVersion string,
	isPartitioningEnabled bool,
	kurtosisBackend backend_interface.KurtosisBackend,
	enclaveDataDir *enclave_data_directory.EnclaveDataDirectory,
	networkingSidecarManager networking_sidecar.NetworkingSidecarManager,
	enclaveDb *enclave_db.EnclaveDB,
) (*DefaultServiceNetwork, error) {
	networkTopology, err := partition_topology.NewPartitionTopology(
		partition_topology.DefaultPartitionId,
		partition_topology.ConnectionAllowed,
		enclaveDb,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the partition topology")
	}
	return &DefaultServiceNetwork{
		enclaveUuid:                         enclaveUuid,
		apiContainerIpAddress:               apiContainerIpAddr,
		apiContainerGrpcPortNum:             apiContainerGrpcPortNum,
		apiContainerVersion:                 apiContainerVersion,
		mutex:                               &sync.Mutex{},
		isPartitioningEnabled:               isPartitioningEnabled,
		kurtosisBackend:                     kurtosisBackend,
		enclaveDataDir:                      enclaveDataDir,
		topology:                            networkTopology,
		networkingSidecars:                  map[service.ServiceName]networking_sidecar.NetworkingSidecarWrapper{},
		networkSidecarsLock:                 &sync.Mutex{},
		networkingSidecarManager:            networkingSidecarManager,
		registeredServiceInfo:               map[service.ServiceName]*service.ServiceRegistration{},
		allExistingAndHistoricalIdentifiers: []*kurtosis_core_rpc_api_bindings.ServiceIdentifiers{},
	}, nil
}

/*
Completely repartitions the network, throwing away the old topology
*/
func (network *DefaultServiceNetwork) Repartition(
	ctx context.Context,
	newPartitionServices map[service_network_types.PartitionID]map[service.ServiceName]bool,
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

	if err := network.updateConnectionsFromTopology(ctx, emptyServiceNamesSetToUpdateAllConnections); err != nil {
		return stacktrace.Propagate(err, "Unable to update connections between the different partitions of the topology")
	}
	return nil
}

func (network *DefaultServiceNetwork) SetConnection(
	ctx context.Context,
	partition1 service_network_types.PartitionID,
	partition2 service_network_types.PartitionID,
	connection partition_topology.PartitionConnection,
) error {
	network.mutex.Lock()
	defer network.mutex.Unlock()
	isOperationSuccessful := false

	if !network.isPartitioningEnabled {
		return stacktrace.NewError("Cannot set connection; partitioning is not enabled")
	}

	currentPartitions, err := network.topology.GetPartitionServices()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while getting all partitions")
	}
	createdPartitionToRemoveIfFailure := map[service_network_types.PartitionID]bool{}
	for _, partition := range []service_network_types.PartitionID{partition1, partition2} {
		if _, found := currentPartitions[partition]; !found {
			logrus.Debugf("Setting connection between '%s' and '%s' but '%s' isn't registered as a partition yet. Creating it",
				partition1, partition2, partition)
			if err := network.topology.CreateEmptyPartitionWithDefaultConnection(partition); err != nil {
				return stacktrace.Propagate(err, "Partition '%v' creation failed", partition)
			}
			createdPartitionToRemoveIfFailure[partition] = true
		}
	}
	defer func() {
		if isOperationSuccessful {
			return
		}
		for partition := range createdPartitionToRemoveIfFailure {
			if err := network.topology.RemovePartition(partition); err != nil {
				logrus.Errorf("Partition '%s' was created as part of a SetConnection call, but due to a failure"+
					"it should be removed. Unfortunately, the removal failed for the following reason so the "+
					"partition will remain in place:\n%v", partition, err.Error())
			}
		}
	}()

	wasConnectionDefault, previousConnection, err := network.topology.GetPartitionConnection(partition1, partition2)
	if err != nil {
		return stacktrace.Propagate(err, "Unable to fetch current connection between '%s' and '%s'", partition1, partition2)
	}

	err = network.topology.SetConnection(partition1, partition2, connection)
	if err != nil {
		return stacktrace.Propagate(err, "Error setting the connection between '%s' and '%s'", partition1, partition2)
	}
	defer func() {
		if isOperationSuccessful {
			return
		}
		var resetConnectionErr error
		if wasConnectionDefault {
			resetConnectionErr = network.topology.UnsetConnection(partition1, partition2)
		} else {
			resetConnectionErr = network.topology.SetConnection(partition1, partition2, previousConnection)
		}
		if resetConnectionErr != nil {
			logrus.Errorf("A failure happened after setting the connection between '%s' and '%s', so it should "+
				"be reset to its previous value. Unfortunately, an error happened trying to set it back to its "+
				"previous value:\n%v", partition1, partition2, err.Error())
		}
	}()

	if err = network.updateConnectionsFromTopology(ctx, emptyServiceNamesSetToUpdateAllConnections); err != nil {
		return stacktrace.Propagate(err, "Unable to update connections between the different partitions of the topology")
	}
	isOperationSuccessful = true
	return nil
}

func (network *DefaultServiceNetwork) UnsetConnection(
	ctx context.Context,
	partition1 service_network_types.PartitionID,
	partition2 service_network_types.PartitionID,
) error {
	network.mutex.Lock()
	defer network.mutex.Unlock()
	isOperationSuccessful := false

	if !network.isPartitioningEnabled {
		return stacktrace.NewError("Cannot unset connection; partitioning is not enabled")
	}

	currentPartitions, err := network.topology.GetPartitionServices()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while getting all partitions")
	}
	for _, partition := range []service_network_types.PartitionID{partition1, partition2} {
		if _, found := currentPartitions[partition]; !found {
			logrus.Warnf("Unsetting connection between '%s' and '%s' but '%s' isn't registered as a partition yet. This will no-op",
				partition1, partition2, partition)
			return nil
		}
	}

	wasDefaultConnection, previousConnection, err := network.topology.GetPartitionConnection(partition1, partition2)
	if err != nil {
		return stacktrace.Propagate(err, "Unable to retrieve current connection between '%s' and '%s'", partition1, partition2)
	}
	if wasDefaultConnection {
		logrus.Debugf("Unsetting connection between '%s' and '%s' but connection was already the default. This will no-op",
			partition1, partition2)
		return nil
	}

	if err = network.topology.UnsetConnection(partition1, partition2); err != nil {
		return stacktrace.Propagate(err, "Unsetting connection between '%s' and '%s' failed", partition1, partition2)
	}
	defer func() {
		if isOperationSuccessful {
			return
		}
		if resetConnectionErr := network.topology.SetConnection(partition1, partition2, previousConnection); resetConnectionErr != nil {
			logrus.Errorf("An error happened resetting the connection between '%s' and '%s' and Kurtosis could not roll back the operation. Error was:\n%v", partition1, partition2, resetConnectionErr)
		}
	}()

	if err = network.updateConnectionsFromTopology(ctx, emptyServiceNamesSetToUpdateAllConnections); err != nil {
		return stacktrace.Propagate(err, "Unable to update connections between the different partitions of the topology")
	}
	isOperationSuccessful = true
	return nil
}

func (network *DefaultServiceNetwork) SetDefaultConnection(
	ctx context.Context,
	connection partition_topology.PartitionConnection,
) error {
	network.mutex.Lock()
	defer network.mutex.Unlock()
	isOperationSuccessful := false

	if !network.isPartitioningEnabled {
		return stacktrace.NewError("Cannot set default connection; partitioning is not enabled")
	}

	previousDefaultConnection := network.topology.GetDefaultConnection()

	network.topology.SetDefaultConnection(connection)
	defer func() {
		if isOperationSuccessful {
			return
		}
		network.topology.SetDefaultConnection(previousDefaultConnection)
	}()

	if err := network.updateConnectionsFromTopology(ctx, emptyServiceNamesSetToUpdateAllConnections); err != nil {
		return stacktrace.Propagate(err, "Unable to update connections between the different partitions of the topology")
	}
	isOperationSuccessful = true
	return nil
}

// StartService starts the service in the given partition in their own container
func (network *DefaultServiceNetwork) StartService(
	ctx context.Context,
	serviceName service.ServiceName,
	serviceConfig *kurtosis_core_rpc_api_bindings.ServiceConfig,
) (
	*service.Service,
	error,
) {
	serviceConfigMap := map[service.ServiceName]*kurtosis_core_rpc_api_bindings.ServiceConfig{
		serviceName: serviceConfig,
	}

	startedServices, serviceFailed, err := network.StartServices(ctx, serviceConfigMap, singleServiceStartupBatch)
	if err != nil {
		return nil, err
	}
	if failure, found := serviceFailed[serviceName]; found {
		return nil, failure
	}
	if startedService, found := startedServices[serviceName]; found {
		return startedService, nil
	}
	return nil, stacktrace.NewError("Service '%s' could not be started properly, but its state is unknown. This is a Kurtosis internal bug", serviceName)
}

// StartServices starts the services in the given partition in their own containers. It is a bulk operation, if a
// single service fails to start, the entire batch is rolled back.
//
// This is a bulk operation that follows a sequential approach with no parallelization yet.
// This function returns:
//   - successfulService - mapping of successful service ids to service objects with info about that service when the
//     entire batch of service could be started
//   - failedServices - mapping of failed service ids to errors causing those failures. As noted above, successful
//     services will be rolled back.
//   - error - when a broad and unexpected error happened.
func (network *DefaultServiceNetwork) StartServices(
	ctx context.Context,
	serviceConfigs map[service.ServiceName]*kurtosis_core_rpc_api_bindings.ServiceConfig,
	batchSize int,
) (
	map[service.ServiceName]*service.Service,
	map[service.ServiceName]error,
	error,
) {
	network.mutex.Lock()
	defer network.mutex.Unlock()
	batchSuccessfullyStarted := false
	startedServices := map[service.ServiceName]*service.Service{}
	failedServices := map[service.ServiceName]error{}

	// Save the services currently running in enclave for later
	currentlyRunningServicesInEnclave := map[service.ServiceName]bool{}
	for serviceName := range network.registeredServiceInfo {
		currentlyRunningServicesInEnclave[serviceName] = true
	}

	// We register all the services one by one
	serviceSuccessfullyRegistered := map[service.ServiceName]*service.ServiceRegistration{}
	servicesToStart := map[service.ServiceUUID]*kurtosis_core_rpc_api_bindings.ServiceConfig{}
	for serviceName, serviceConfig := range serviceConfigs {
		servicePartitionId := partition_topology.ParsePartitionId(serviceConfig.Subnetwork)
		serviceRegistration, err := network.registerService(ctx, serviceName, servicePartitionId)
		if err != nil {
			failedServices[serviceName] = stacktrace.Propagate(err, "Failed registering service with name: '%s'", serviceName)
		}
		serviceSuccessfullyRegistered[serviceName] = serviceRegistration
		servicesToStart[serviceRegistration.GetUUID()] = serviceConfig
	}
	defer func() {
		if batchSuccessfullyStarted {
			return
		}
		for serviceName := range serviceSuccessfullyRegistered {
			if err := network.unregisterService(ctx, serviceName); err != nil {
				logrus.Errorf("Error unregistering service '%s' from the service network. Error was: %v", serviceName, err)
			}
		}
	}()
	if len(failedServices) > 0 {
		return map[service.ServiceName]*service.Service{}, failedServices, nil
	}

	// We update the networking setup of the currently running services such that services starting won't be able
	// to communicate to services they should not communicate with.
	if network.isPartitioningEnabled && len(currentlyRunningServicesInEnclave) > 0 {
		if err := network.updateConnectionsFromTopology(ctx, currentlyRunningServicesInEnclave); err != nil {
			return nil, nil, stacktrace.Propagate(err, "Failure updating the network connections of the existing "+
				"services prior to starting the new services. Starting the following services will be aborted: %v. "+
				"Existing services in enclave: '%v'", serviceConfigs, currentlyRunningServicesInEnclave)
		}
	}

	startedServicesPerUuid, failedServicePerUuid := network.startRegisteredServices(ctx, servicesToStart, batchSize)

	for serviceName, serviceRegistration := range serviceSuccessfullyRegistered {
		serviceUuid := serviceRegistration.GetUUID()
		if serviceStartFailure, found := failedServicePerUuid[serviceUuid]; found {
			failedServices[serviceName] = serviceStartFailure
			continue
		}
		if startedService, found := startedServicesPerUuid[serviceUuid]; found {
			startedServices[serviceName] = startedService
			continue
		}
		if len(failedServicePerUuid) == 0 {
			// it is expected that if a service failed to be started, Kurtosis did not even try to start the others
			// and stopped midway. However, if failedServicePerUuid is empty, this is an internal bug
			failedServices[serviceName] = stacktrace.NewError("State of the service is unknown: %s. This is a Kurtosis internal bug", serviceName)
		}
	}
	defer func() {
		if batchSuccessfullyStarted {
			return
		}
		for serviceName, startedService := range startedServices {
			if err := network.destroyService(ctx, serviceName, startedService.GetRegistration().GetUUID()); err != nil {
				logrus.Errorf("One or more services failed to be started for this batch. Kurtosis tries to"+
					"roll back the entire batch, but failed destroying some services. Error was: %v", err)
			}
		}
	}()
	if len(failedServices) > 0 {
		return map[service.ServiceName]*service.Service{}, failedServices, nil
	}

	if len(startedServices) != len(serviceConfigs) {
		var requested []service.ServiceName
		for serviceName := range serviceConfigs {
			requested = append(requested, serviceName)
		}
		var result []service.ServiceName
		for serviceName := range startedServices {
			result = append(result, serviceName)
		}
		return nil, nil, stacktrace.NewError("This is a Kurtosis internal bug. The batch of services being started does not fit the number of services that were requested. (service started: '%v', requested: '%v')", result, requested)
	}

	for _, service := range startedServices {
		serviceNameStr := string(service.GetRegistration().GetName())
		serviceUuidStr := string(service.GetRegistration().GetUUID())
		shortenedUuidStr := uuid_generator.ShortenedUUIDString(serviceUuidStr)
		network.allExistingAndHistoricalIdentifiers = append(network.allExistingAndHistoricalIdentifiers, &kurtosis_core_rpc_api_bindings.ServiceIdentifiers{
			ServiceUuid:   serviceUuidStr,
			Name:          serviceNameStr,
			ShortenedUuid: shortenedUuidStr,
		})
	}

	batchSuccessfullyStarted = true
	return startedServices, map[service.ServiceName]error{}, nil
}

// UpdateService This is purely called from a Starlark context so this only works with Names
func (network *DefaultServiceNetwork) UpdateService(
	ctx context.Context,
	updateServiceConfigs map[service.ServiceName]*kurtosis_core_rpc_api_bindings.UpdateServiceConfig,
) (
	map[service.ServiceName]bool,
	map[service.ServiceName]error,
	error,
) {
	failedServicesPool := map[service.ServiceName]error{}
	successfullyUpdatedService := map[service.ServiceName]bool{}

	previousServicePartitions := map[service.ServiceName]service_network_types.PartitionID{}
	partitionCreatedDuringThisOperation := map[service_network_types.PartitionID]bool{}
	for serviceName, updateServiceConfig := range updateServiceConfigs {
		if updateServiceConfig.Subnetwork == nil {
			// nothing to do for this service
			continue
		}

		servicePartitions, err := network.topology.GetServicePartitions()
		if err != nil {
			failedServicesPool[serviceName] = stacktrace.Propagate(err, "An error occurred while fetching service partitions mapping for service '%v'", serviceName)
			continue
		}
		previousServicePartition, found := servicePartitions[serviceName]
		if !found {
			failedServicesPool[serviceName] = stacktrace.NewError("Error updating service '%s' as this service does not exist", serviceName)
			continue
		}
		previousServicePartitions[serviceName] = previousServicePartition

		newServicePartition := partition_topology.ParsePartitionId(updateServiceConfig.Subnetwork)
		if newServicePartition == previousServicePartition {
			// nothing to do for this service
			continue
		}

		partitionServices, err := network.topology.GetPartitionServices()
		if err != nil {
			failedServicesPool[serviceName] = stacktrace.Propagate(
				err,
				"Cannot update service '%v' as we tried to fetch existing partitions and failed",
				serviceName,
			)
			continue
		}

		if _, found = partitionServices[newServicePartition]; !found {
			logrus.Debugf("Partition with ID '%s' does not exist in current topology. Creating it to be able to "+
				"add service '%s' to it when it's created", newServicePartition, serviceName)
			if err := network.topology.CreateEmptyPartitionWithDefaultConnection(newServicePartition); err != nil {
				failedServicesPool[serviceName] = stacktrace.Propagate(
					err,
					"Cannot update service '%v' its new partition '%s' needed to be created and it failed",
					serviceName,
					newServicePartition,
				)
				continue
			}
			partitionCreatedDuringThisOperation[newServicePartition] = true
		}

		if err := network.moveServiceToPartitionInTopology(serviceName, newServicePartition); err != nil {
			failedServicesPool[serviceName] = stacktrace.Propagate(err, "Error updating service '%s' adding it to the new partition '%s'", serviceName, newServicePartition)
			continue
		}
	}
	defer func() {
		for serviceName, partitionIDToRollbackTo := range previousServicePartitions {
			if _, found := successfullyUpdatedService[serviceName]; found {
				continue
			}

			servicePartitions, err := network.topology.GetServicePartitions()
			if err != nil {
				logrus.Errorf("An error happened updating service '%s' and it needed to be moved back to partition '%s', but an error happened during this operation. Error was:\n%v", serviceName, partitionIDToRollbackTo, err)
				return
			}

			currentPartitionId, found := servicePartitions[serviceName]
			if !found {
				// service does not exist, nothing to roll back
				continue
			}
			if currentPartitionId == partitionIDToRollbackTo {
				// service is still in the partition it was before the call to UpdateService, nothing to roll back
				continue
			}
			// if service exists and is not in successfullyUpdatedService, roll it back to its previous partition
			if err := network.moveServiceToPartitionInTopology(serviceName, partitionIDToRollbackTo); err != nil {
				logrus.Errorf("An error happened updating service '%s' and it needed to be moved back to partition '%s', but an error happened during this operation. The service will be left in '%s'. Error was:\n%v", serviceName, partitionIDToRollbackTo, currentPartitionId, err)
			}
		}
		// finally, after all updates and roll-back have been performed, check for potentially empty partitions and remove them to keep the topology clean
		partitionServices, err := network.topology.GetPartitionServices()
		if err != nil {
			logrus.Errorf("Tried getting partition services to cleanup any empty partitions but failed.")
			return
		}
		for partitionID := range partitionCreatedDuringThisOperation {
			servicesInPartition, found := partitionServices[partitionID]
			if found && len(servicesInPartition) == 0 {
				if err := network.topology.RemovePartition(partitionID); err != nil {
					logrus.Errorf("Partition '%s' was left empty after a service update. It failed to be removes", partitionID)
				}
			}
		}
	}()

	if err := network.updateConnectionsFromTopology(ctx, emptyServiceNamesSetToUpdateAllConnections); err != nil {
		// successfullyUpdatedService is still empty here, so all services will be rolled back to their previous partition
		return nil, nil, stacktrace.Propagate(err, "Unable to update connections between the different partitions of the topology")
	}

	for serviceName := range updateServiceConfigs {
		if _, found := failedServicesPool[serviceName]; found {
			continue
		}
		successfullyUpdatedService[serviceName] = true
	}
	return successfullyUpdatedService, failedServicesPool, nil
}

func (network *DefaultServiceNetwork) RemoveService(
	ctx context.Context,
	serviceIdentifier string,
) (service.ServiceUUID, error) {
	network.mutex.Lock()
	defer network.mutex.Unlock()

	serviceName, err := network.getServiceNameForIdentifierUnlocked(serviceIdentifier)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while fetching name for service identifier '%v'", serviceIdentifier)
	}

	serviceToRemove, found := network.registeredServiceInfo[serviceName]
	if !found {
		return "", stacktrace.NewError("No service found with ID '%v'", serviceName)
	}
	serviceUuid := serviceToRemove.GetUUID()

	err = network.topology.RemoveService(serviceName)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while removing service '%v' from the network topology", serviceName)
	}

	network.cleanupInternalMapsUnlocked(serviceName)

	// We stop the service, rather than destroying it, so that we can keep logs around
	stopServiceFilters := &service.ServiceFilters{
		Names: nil,
		UUIDs: map[service.ServiceUUID]bool{
			serviceUuid: true,
		},
		Statuses: nil,
	}
	_, erroredUuids, err := network.kurtosisBackend.StopUserServices(ctx, network.enclaveUuid, stopServiceFilters)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred during the call to stop service '%v'", serviceUuid)
	}
	if err, found := erroredUuids[serviceUuid]; found {
		return "", stacktrace.Propagate(err, "An error occurred stopping service '%v'", serviceUuid)
	}

	sidecar, foundSidecar := network.networkingSidecars[serviceName]
	if network.isPartitioningEnabled && foundSidecar {
		// NOTE: As of 2020-12-31, we don't need to update the iptables of the other services in the network to
		//  clear the now-removed service's IP because:
		// 	 a) nothing is using it so it doesn't do anything and
		//	 b) all service's iptables get overwritten on the next Add/Repartition call
		// If we ever do incremental iptables though, we'll need to fix all the other service's iptables here!
		if err := network.networkingSidecarManager.Remove(ctx, sidecar); err != nil {
			return "", stacktrace.Propagate(err, "An error occurred destroying the sidecar for service with name '%v'", serviceName)
		}
		delete(network.networkingSidecars, serviceName)
		logrus.Debugf("Successfully removed sidecar attached to service with name '%v'", serviceName)
	}

	return serviceUuid, nil
}

// TODO we could switch this to be a bulk command; the backend would support it
func (network *DefaultServiceNetwork) PauseService(ctx context.Context, serviceIdentifier string) error {
	network.mutex.Lock()
	defer network.mutex.Unlock()

	serviceName, err := network.getServiceNameForIdentifierUnlocked(serviceIdentifier)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while getting service name for identifier '%v'", serviceIdentifier)
	}

	serviceObj, found := network.registeredServiceInfo[serviceName]
	if !found {
		return stacktrace.NewError("No service with name '%v' exists in the network", serviceName)
	}

	if err := network.kurtosisBackend.PauseService(ctx, network.enclaveUuid, serviceObj.GetUUID()); err != nil {
		return stacktrace.Propagate(err, "Failed to pause service '%v'", serviceIdentifier)
	}
	return nil
}

// TODO we could switch this to be a bulk command; the backend would support it
func (network *DefaultServiceNetwork) UnpauseService(ctx context.Context, serviceIdentifier string) error {
	network.mutex.Lock()
	defer network.mutex.Unlock()

	serviceName, err := network.getServiceNameForIdentifierUnlocked(serviceIdentifier)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while getting service name for identifier '%v'", serviceIdentifier)
	}

	serviceObj, found := network.registeredServiceInfo[serviceName]
	if !found {
		return stacktrace.NewError("No service with name '%v' exists in the network", serviceName)
	}

	if err := network.kurtosisBackend.UnpauseService(ctx, network.enclaveUuid, serviceObj.GetUUID()); err != nil {
		return stacktrace.Propagate(err, "Failed to unpause service '%v'", serviceIdentifier)
	}
	return nil
}

func (network *DefaultServiceNetwork) ExecCommand(ctx context.Context, serviceIdentifier string, command []string) (int32, string, error) {
	// NOTE: This will block all other operations while this command is running!!!! We might need to change this so it's
	// asynchronous
	network.mutex.Lock()
	defer network.mutex.Unlock()

	serviceName, err := network.getServiceNameForIdentifierUnlocked(serviceIdentifier)
	if err != nil {
		return 0, "", stacktrace.Propagate(err, "An error occurred while getting service name for identifier '%v'", serviceIdentifier)
	}

	serviceObj, found := network.registeredServiceInfo[serviceName]
	if !found {
		return 0, "", stacktrace.NewError(
			"Service '%v does not exist in the network",
			serviceIdentifier,
		)
	}

	// NOTE: This is a SYNCHRONOUS command, meaning that the entire network will be blocked until the command finishes
	// In the future, this will likely be insufficient

	serviceUuid := serviceObj.GetUUID()
	userServiceCommand := map[service.ServiceUUID][]string{
		serviceUuid: command,
	}

	successfulExecCommands, failedExecCommands, err := network.kurtosisBackend.RunUserServiceExecCommands(
		ctx,
		network.enclaveUuid,
		userServiceCommand)
	if err != nil {
		return 0, "", stacktrace.Propagate(
			err,
			"An error occurred calling kurtosis backend to exec command '%v' against service '%v'",
			command,
			serviceIdentifier)
	}
	if len(failedExecCommands) > 0 {
		serviceExecErrs := []string{}
		for serviceUUID, err := range failedExecCommands {
			wrappedErr := stacktrace.Propagate(
				err,
				"An error occurred attempting to run a command in a service with UUID `%v'",
				serviceUUID,
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

	execResult, isFound := successfulExecCommands[serviceUuid]
	if !isFound {
		return 0, "", stacktrace.NewError(
			"Unable to find result from running exec command '%v' against service '%v'",
			command,
			serviceUuid)
	}

	return execResult.GetExitCode(), execResult.GetOutput(), nil
}

func (network *DefaultServiceNetwork) HttpRequestService(ctx context.Context, serviceIdentifier string, portId string, method string, contentType string, endpoint string, body string) (*http.Response, error) {
	logrus.Debugf("Making a request '%v' '%v' '%v' '%v' '%v' '%v'", serviceIdentifier, portId, method, contentType, endpoint, body)
	service, getServiceErr := network.GetService(ctx, serviceIdentifier)
	if getServiceErr != nil {
		return nil, stacktrace.Propagate(getServiceErr, "An error occurred when getting service '%v' for HTTP request", serviceIdentifier)
	}
	port, found := service.GetPrivatePorts()[portId]
	if !found {
		return nil, stacktrace.NewError("An error occurred when getting port '%v' from service '%v' for HTTP request", serviceIdentifier, portId)
	}
	url := fmt.Sprintf("http://%v:%v%v", service.GetRegistration().GetPrivateIP(), port.GetNumber(), endpoint)
	if method == http.MethodPost {
		response, err := http.Post(url, contentType, strings.NewReader(body))
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred on POST HTTP request on '%v'", url)
		}
		return response, err
	} else if method == http.MethodGet {
		response, err := http.Get(url)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred on GET HTTP request on '%v'", url)
		}
		return response, err
	} else {
		return nil, stacktrace.NewError("An error occurred because %v is unsupported for HTTP request", method)
	}
}

func (network *DefaultServiceNetwork) GetService(ctx context.Context, serviceIdentifier string) (*service.Service, error) {
	network.mutex.Lock()
	defer network.mutex.Unlock()

	serviceName, err := network.getServiceNameForIdentifierUnlocked(serviceIdentifier)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while fetching name for service identifier '%v'", serviceIdentifier)
	}

	registration, found := network.registeredServiceInfo[serviceName]
	if !found {
		return nil, stacktrace.NewError("No service with name '%v' exists in network", serviceName)
	}
	serviceUuid := registration.GetUUID()

	getServiceFilters := &service.ServiceFilters{
		Names: nil,
		UUIDs: map[service.ServiceUUID]bool{
			registration.GetUUID(): true,
		},
		Statuses: nil,
	}
	matchingServices, err := network.kurtosisBackend.GetUserServices(ctx, network.enclaveUuid, getServiceFilters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting service '%v'", serviceUuid)
	}
	if len(matchingServices) == 0 {
		return nil, stacktrace.NewError(
			"A registration exists for service UUID '%v' but no service objects were found; this indicates that the service was "+
				"registered but not started",
			serviceUuid,
		)
	}
	if len(matchingServices) > 1 {
		return nil, stacktrace.NewError("Found multiple service objects matching UUID '%v'", serviceUuid)
	}
	serviceObj, found := matchingServices[serviceUuid]
	if !found {
		return nil, stacktrace.NewError("Found exactly one service object, but it didn't match expected UUID '%v'", serviceUuid)
	}

	return serviceObj, nil
}

func (network *DefaultServiceNetwork) GetServiceNames() map[service.ServiceName]bool {

	serviceNames := make(map[service.ServiceName]bool, len(network.registeredServiceInfo))

	for serviceName := range network.registeredServiceInfo {
		if _, ok := serviceNames[serviceName]; !ok {
			serviceNames[serviceName] = true
		}
	}
	return serviceNames
}

func (network *DefaultServiceNetwork) CopyFilesFromService(ctx context.Context, serviceIdentifier string, srcPath string, artifactName string) (enclave_data_directory.FilesArtifactUUID, error) {
	serviceName, err := network.getServiceNameForIdentifierUnlocked(serviceIdentifier)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while fetching name for service identifier '%v'", serviceIdentifier)
	}

	filesArtifactUuid, err := network.copyFilesFromServiceUnlocked(ctx, serviceName, srcPath, artifactName)
	if err != nil {
		return "", stacktrace.Propagate(err, "There was an error in copying files over to disk")
	}
	return filesArtifactUuid, nil
}

func (network *DefaultServiceNetwork) GetServiceRegistration(serviceName service.ServiceName) (*service.ServiceRegistration, bool) {
	network.mutex.Lock()
	defer network.mutex.Unlock()
	registration, found := network.registeredServiceInfo[serviceName]
	if !found {
		return nil, false
	}
	return registration, true
}

func (network *DefaultServiceNetwork) RenderTemplates(templatesAndDataByDestinationRelFilepath map[string]*kurtosis_core_rpc_api_bindings.RenderTemplatesToFilesArtifactArgs_TemplateAndData, artifactName string) (enclave_data_directory.FilesArtifactUUID, error) {
	filesArtifactUuid, err := network.renderTemplatesUnlocked(templatesAndDataByDestinationRelFilepath, artifactName)
	if err != nil {
		return "", stacktrace.Propagate(err, "There was an error in rendering templates to disk")
	}
	return filesArtifactUuid, nil
}

func (network *DefaultServiceNetwork) UploadFilesArtifact(data []byte, artifactName string) (enclave_data_directory.FilesArtifactUUID, error) {
	filesArtifactUuid, err := network.uploadFilesArtifactUnlocked(data, artifactName)
	if err != nil {
		return "", stacktrace.Propagate(err, "There was an error in uploading the files")
	}
	return filesArtifactUuid, nil
}

func (network *DefaultServiceNetwork) IsNetworkPartitioningEnabled() bool {
	return network.isPartitioningEnabled
}

func (network *DefaultServiceNetwork) GetExistingAndHistoricalServiceIdentifiers() []*kurtosis_core_rpc_api_bindings.ServiceIdentifiers {
	return network.allExistingAndHistoricalIdentifiers
}

// GetUniqueNameForFileArtifact : this will return unique artifact name after 5 retries, same as enclave id generator
func (network *DefaultServiceNetwork) GetUniqueNameForFileArtifact() (string, error) {
	filesArtifactStore, err := network.enclaveDataDir.GetFilesArtifactStore()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while getting files artifact store")
	}
	return filesArtifactStore.GenerateUniqueNameForFileArtifact(), nil
}

// ====================================================================================================
// 									   Private helper methods
// ====================================================================================================

// updateConnectionsFromTopology reads the current topology and updates the connections for the provided service names
// according to it.
// if serviceNames is empty, it updates the connection for all the services within the enclave
func (network *DefaultServiceNetwork) updateConnectionsFromTopology(ctx context.Context, serviceNames map[service.ServiceName]bool) error {
	availablePartitionConnectionConfigsPerServiceNames, err := network.topology.GetServicePartitionConnectionConfigByServiceName()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the packet loss configuration by service ID "+
			" to know what packet loss updates to apply")
	}

	var serviceNamesToUpdate map[service.ServiceName]bool
	if len(serviceNames) == emptyCollectionLength {
		// we add all the services currently stored in the topology to update everything
		serviceNamesToUpdate = map[service.ServiceName]bool{}
		for serviceName := range availablePartitionConnectionConfigsPerServiceNames {
			serviceNamesToUpdate[serviceName] = true
		}
	} else {
		serviceNamesToUpdate = serviceNames
	}

	// TODO: probably worth running those updates in parallel for enclave with a lot of services
	for serviceName := range serviceNamesToUpdate {
		otherServiceConnectionConfig, found := availablePartitionConnectionConfigsPerServiceNames[serviceName]
		if !found {
			return stacktrace.NewError("A service about to be updated could not be found in the connection config service map: '%s' (connection config service map was: '%v')", serviceName, availablePartitionConnectionConfigsPerServiceNames)
		}
		if err = updateTrafficControlConfiguration(ctx, serviceName, otherServiceConnectionConfig, network.registeredServiceInfo, network.networkingSidecars); err != nil {
			return stacktrace.Propagate(err, "An error occurred applying the traffic control configuration to partition off new nodes.")
		}
	}
	return nil
}

// Updates the traffic control configuration of the services with the given Names to match the target services packet loss configuration
// NOTE: This is not thread-safe, so it must be within a function that locks mutex!
func updateTrafficControlConfiguration(
	ctx context.Context,
	serviceName service.ServiceName,
	otherServiceConnectionConfigs map[service.ServiceName]*partition_topology.PartitionConnection,
	registeredServices map[service.ServiceName]*service.ServiceRegistration,
	networkingSidecars map[service.ServiceName]networking_sidecar.NetworkingSidecarWrapper,
) error {
	partitionConnectionConfigPerIpAddress := map[string]*partition_topology.PartitionConnection{}
	for connectedServiceName, partitionConnectionConfig := range otherServiceConnectionConfigs {
		connectedService, found := registeredServices[connectedServiceName]
		if !found {
			return stacktrace.NewError(
				"Service with name '%s' needs to update its connection configuration for service with name '%s', "+
					"but the latter doesn't have service registration info (i.e. an IP) associated with it",
				serviceName,
				connectedServiceName)
		}

		partitionConnectionConfigPerIpAddress[connectedService.GetPrivateIP().String()] = partitionConnectionConfig
	}

	sidecar, found := networkingSidecars[serviceName]
	if !found {
		return stacktrace.NewError(
			"Need to update qdisc configuration of service with name '%v', but the service doesn't have a sidecar",
			serviceName)
	}

	if err := sidecar.UpdateTrafficControl(ctx, partitionConnectionConfigPerIpAddress); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred updating the qdisc configuration for service '%v'",
			serviceName)
	}
	return nil
}

// registerService handles all the operations necessary to register a service before is can be started with
// startRegisteredService. If something fails along the way, the function takes care of rolling back the previous
// changes such that the enclave remains in the state before the call
// TODO: The approach is naive here as we register a single service, so it needs to be called within a loop
//
//	to register multiple services. We could leverage the fact that the BE handles registering multiple services
//	with a single call. For now, as registering a service is fairly low lift, it's fine this way
func (network *DefaultServiceNetwork) registerService(
	ctx context.Context,
	serviceName service.ServiceName,
	partitionId service_network_types.PartitionID,
) (
	*service.ServiceRegistration,
	error,
) {
	serviceSuccessfullyRegistered := false

	partitionServices, err := network.topology.GetPartitionServices()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting partition services")
	}

	if _, found := partitionServices[partitionId]; !found {
		logrus.Debugf("Paritition with ID '%s' does not exist in current topology. Creating it to be able to "+
			"add service '%s' to it when it's created", partitionId, serviceName)

		if err := network.topology.CreateEmptyPartitionWithDefaultConnection(partitionId); err != nil {
			return nil, stacktrace.Propagate(
				err,
				"Cannot register service '%s' because its partition '%s' failed to be created",
				serviceName,
				partitionId,
			)
		}
		// undo partition creation if starting the something fails downstream
		defer func() {
			if serviceSuccessfullyRegistered || partitionId == partition_topology.DefaultPartitionId {
				return
			}
			if err := network.topology.RemovePartition(partitionId); err != nil {
				logrus.Errorf("Paritition '%s' needs to be removed as it is empty, but its deletion failed with an unexpected error. Partition will remain in the topology. This is not critical but might be a sign of another more critical failure", partitionId)
			}
		}()
	}

	serviceToRegister := map[service.ServiceName]bool{
		serviceName: true,
	}
	servicesSuccessfullyRegistered, serviceFailedRegistration, err := network.kurtosisBackend.RegisterUserServices(ctx, network.enclaveUuid, serviceToRegister)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unexpected error happened registering service '%s'", serviceName)
	}
	if serviceRegistrationErr, found := serviceFailedRegistration[serviceName]; found {
		return nil, stacktrace.Propagate(serviceRegistrationErr, "Error registering service '%s'", serviceName)
	}
	serviceRegistration, found := servicesSuccessfullyRegistered[serviceName]
	if !found {
		return nil, stacktrace.NewError("Unexpected error while registering service '%s'. It was not flagged as neither failed nor successfully registered. This is a Kurtosis internal bug.", serviceName)
	}
	defer func() {
		if serviceSuccessfullyRegistered {
			return
		}
		serviceUuid := serviceRegistration.GetUUID()
		serviceToUnregister := map[service.ServiceUUID]bool{
			serviceUuid: true,
		}
		_, failedService, unexpectedErr := network.kurtosisBackend.UnregisterUserServices(ctx, network.enclaveUuid, serviceToUnregister)
		if unexpectedErr != nil {
			logrus.Errorf("An unexpected error happened unregistering service '%s' after it failed starting. It"+
				"is possible the service is still registered to the enclave.", serviceName)
			return
		}
		if unregisteringErr, found := failedService[serviceUuid]; found {
			logrus.Errorf("An error happened unregistering service '%s' after it failed starting. It"+
				"is possible the service is still registered to the enclave. The error was\n%v",
				serviceName, unregisteringErr.Error())
			return
		}
	}()

	network.registeredServiceInfo[serviceName] = serviceRegistration
	// remove service from the registered service map is something fails downstream
	defer func() {
		if serviceSuccessfullyRegistered {
			return
		}
		network.cleanupInternalMapsUnlocked(serviceName)
	}()

	err = network.addServiceToTopology(serviceName, partitionId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error adding service '%s' to partition '%s' in network topology", serviceName, partitionId)
	}
	logrus.Debugf("Successfully added service with name '%v' to topology", serviceName)
	// remove service from topology is something fails downstream
	defer func() {
		if serviceSuccessfullyRegistered {
			return
		}
		err = network.topology.RemoveService(serviceName)
		if err != nil {
			logrus.Errorf("An error occurred while removing service '%v' from the partition toplogy", serviceName)
		}
	}()
	serviceSuccessfullyRegistered = true
	return serviceRegistration, nil
}

// unregisterService is the opposite of register service. It cleans up everything is can to property unregister a
// service. It is expected that the service was properly registered.
// As registerService rolls back things if a failure happens halfway, we should never end up with a service
// half-registered, but it's worth calling out that this method with throw if called with such a service
func (network *DefaultServiceNetwork) unregisterService(ctx context.Context, serviceName service.ServiceName) error {
	servicePartitions, err := network.topology.GetServicePartitions()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while fetching service partitions mapping")
	}
	partitionId, partitionFound := servicePartitions[serviceName]
	err = network.topology.RemoveService(serviceName)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while removing service '%v' from the network topology", serviceName)
	}
	partitionServices, err := network.topology.GetPartitionServices()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while getting partition services")
	}
	if partitionFound && partitionId != partition_topology.DefaultPartitionId {
		if len(partitionServices[partitionId]) == 0 {
			if err := network.topology.RemovePartition(partitionId); err != nil {
				logrus.Warnf("Error removing partition '%s' as it was empty after removing service '%s'. "+
					"This is not critical but is unexpected. Error was: '%v'", partitionId, serviceName, err)
			}
		}
	}

	serviceRegistration, registrationFound := network.registeredServiceInfo[serviceName]
	if !registrationFound {
		return stacktrace.NewError("Unregistering a service that has not been properly registered should not happen: '%s'. This is a Kurtosis internal bug", serviceName)
	}

	network.cleanupInternalMapsUnlocked(serviceName)
	serviceUuid := serviceRegistration.GetUUID()
	serviceToUnregister := map[service.ServiceUUID]bool{
		serviceUuid: true,
	}
	_, failedService, unexpectedErr := network.kurtosisBackend.UnregisterUserServices(ctx, network.enclaveUuid, serviceToUnregister)
	if unexpectedErr != nil {
		return stacktrace.Propagate(unexpectedErr, "An unexpected error happened unregistering service '%s'. It "+
			"is possible the service is still registered to the enclave.", serviceName)
	}
	if unregisteringErr, found := failedService[serviceUuid]; found {
		return stacktrace.Propagate(unregisteringErr, "An error happened unregistering service '%s'. It"+
			"is possible the service is still registered to the enclave.",
			serviceName)
	}
	return nil
}

// startRegisteredService handles the logistic of starting a service in the relevant Kurtosis backend:
// Convert API ServiceConfig's to service.ServiceConfig's by:
// - converting API Ports to PortSpec's
// - converting files artifacts mountpoints to FilesArtifactsExpansion's'
// - passing down other data (eg. container image name, args, etc.)
// If network partitioning is enabled, it also takes care of starting the sidecar corresponding to this service
func (network *DefaultServiceNetwork) startRegisteredService(
	ctx context.Context,
	serviceUuid service.ServiceUUID,
	serviceConfigApi *kurtosis_core_rpc_api_bindings.ServiceConfig,
) (
	*service.Service,
	error,
) {
	serviceStartedSuccessfully := false
	var serviceConfig *service.ServiceConfig

	// Docker and K8s requires the minimum memory limit to be 6 megabytes to we make sure the allocation is at least that amount
	// But first, we check that it's not the default value, meaning the user potentially didn't even set it
	if serviceConfigApi.MemoryAllocationMegabytes != defaultMemoryAllocMegabytes && serviceConfigApi.MemoryAllocationMegabytes < minMemoryLimit {
		return nil, stacktrace.NewError("Memory allocation, `%d`, is too low. Kurtosis requires the memory limit to be at least `%d` megabytes for service with UUID '%v'.", serviceConfigApi.MemoryAllocationMegabytes, minMemoryLimit, serviceUuid)
	}

	// Convert ports
	privateServicePortSpecs, requestedPublicServicePortSpecs, err := convertAPIPortsToPortSpecs(serviceConfigApi.PrivatePorts, serviceConfigApi.PublicPorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while trying to convert public and private API ports to port specs for service with UUID '%v'", serviceUuid)
	}

	// Creates files artifacts expansions
	var filesArtifactsExpansion *files_artifacts_expansion.FilesArtifactsExpansion
	if len(serviceConfigApi.FilesArtifactMountpoints) == 0 {
		// Create service config with empty filesArtifactsExpansion if no files artifacts mountpoints for this service
		serviceConfig = service.NewServiceConfig(
			serviceConfigApi.ContainerImageName,
			privateServicePortSpecs,
			requestedPublicServicePortSpecs,
			serviceConfigApi.EntrypointArgs,
			serviceConfigApi.CmdArgs,
			serviceConfigApi.EnvVars,
			filesArtifactsExpansion,
			serviceConfigApi.CpuAllocationMillicpus,
			serviceConfigApi.MemoryAllocationMegabytes,
			serviceConfigApi.PrivateIpAddrPlaceholder)
	} else {
		filesArtifactsExpansions := []args.FilesArtifactExpansion{}
		expanderDirpathToUserServiceDirpathMap := map[string]string{}
		for mountpointOnUserService, filesArtifactIdentifier := range serviceConfigApi.FilesArtifactMountpoints {
			dirpathToExpandTo := path.Join(filesArtifactExpansionDirsParentDirpath, filesArtifactIdentifier)
			expansion := args.FilesArtifactExpansion{
				FilesIdentifier:   filesArtifactIdentifier,
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
			return nil, stacktrace.Propagate(err, "An error occurred creating files artifacts expander args for service with UUID '%s'", serviceUuid)
		}

		expanderEnvVars, err := args.GetEnvFromArgs(filesArtifactsExpanderArgs)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting files artifacts expander environment variables using args: %+v", filesArtifactsExpanderArgs)
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

		serviceConfig = service.NewServiceConfig(
			serviceConfigApi.ContainerImageName,
			privateServicePortSpecs,
			requestedPublicServicePortSpecs,
			serviceConfigApi.EntrypointArgs,
			serviceConfigApi.CmdArgs,
			serviceConfigApi.EnvVars,
			filesArtifactsExpansion,
			serviceConfigApi.CpuAllocationMillicpus,
			serviceConfigApi.MemoryAllocationMegabytes,
			serviceConfigApi.PrivateIpAddrPlaceholder)
	}

	// TODO(gb): make the backend also handle starting service sequentially to simplify the logic there as well
	serviceConfigMap := map[service.ServiceUUID]*service.ServiceConfig{
		serviceUuid: serviceConfig,
	}
	successfulServices, failedServices, err := network.kurtosisBackend.StartRegisteredUserServices(ctx, network.enclaveUuid, serviceConfigMap)
	if err != nil {
		return nil, err
	}
	if failedServiceErr, isFailed := failedServices[serviceUuid]; isFailed {
		return nil, failedServiceErr
	}

	startedService, isSuccessful := successfulServices[serviceUuid]
	if !isSuccessful {
		return nil, stacktrace.NewError("Service '%s' did not start properly but no error was thrown. This is a Kurtosis internal bug", serviceUuid)
	}
	defer func() {
		if serviceStartedSuccessfully {
			return
		}
		serviceToDestroyUuid := startedService.GetRegistration().GetUUID()
		userServiceFilters := &service.ServiceFilters{
			Names: nil,
			UUIDs: map[service.ServiceUUID]bool{
				serviceToDestroyUuid: true,
			},
			Statuses: nil,
		}
		_, failedToDestroyUuids, err := network.kurtosisBackend.DestroyUserServices(context.Background(), network.enclaveUuid, userServiceFilters)
		if err != nil {
			logrus.Errorf("Attempted to destroy the services with UUIDs '%v' but had no success. You must manually destroy the services! The following error had occurred:\n'%v'", serviceToDestroyUuid, err)
			return
		}
		if failedToDestroyErr, found := failedToDestroyUuids[serviceToDestroyUuid]; found {
			logrus.Errorf("Attempted to destroy the services with UUIDs '%v' but had no success. You must manually destroy the services! The following error had occurred:\n'%v'", serviceToDestroyUuid, failedToDestroyErr)
		}
	}()

	// if partition is enabled, create a sidecar associated with this service
	if network.isPartitioningEnabled {
		if err := network.createSidecarAndAddToMap(ctx, startedService); err != nil {
			return nil, stacktrace.Propagate(err, "Error creating sidecar for service '%s'", serviceUuid)
		}
		serviceNameSet := map[service.ServiceName]bool{
			startedService.GetRegistration().GetName(): true,
		}
		// update the connection for this service only
		if err := network.updateConnectionsFromTopology(ctx, serviceNameSet); err != nil {
			return nil, stacktrace.Propagate(err, "Error updating the networking rules for this service '%s' (UUID: '%s')", startedService.GetRegistration().GetName(), serviceUuid)
		}
		logrus.Debugf("Successfully created sidecars for service with ID '%v'", serviceUuid)
	}

	serviceStartedSuccessfully = true
	return startedService, nil
}

// destroyService is the opposite of startRegisteredService. It removes a started service from the enclave. Note that it does not
// take care of unregistering the service. For this, unregisterService should be called
// Similar to unregisterService, it is expected that the service passed to destroyService has been properly started.
// the function might fail if the service is half-started
// Note: the function also takes care of destroying any networking sidecar associated with the service
func (network *DefaultServiceNetwork) destroyService(ctx context.Context, serviceName service.ServiceName, serviceUuid service.ServiceUUID) error {
	var errorResult error
	// deleting the service first
	userServiceFilters := &service.ServiceFilters{
		Names: nil,
		UUIDs: map[service.ServiceUUID]bool{
			serviceUuid: true,
		},
		Statuses: nil,
	}
	_, failedToDestroyUuids, err := network.kurtosisBackend.DestroyUserServices(context.Background(), network.enclaveUuid, userServiceFilters)
	if err != nil {
		errorResult = stacktrace.Propagate(err, "Attempted to destroy the services with UUID '%v' but had no success. You must manually destroy the services. Kurtosis will now try to remove its sidecar if it exists but might it fail as well.", serviceUuid)
	}
	if failedToDestroyErr, found := failedToDestroyUuids[serviceUuid]; found {
		errorResult = stacktrace.Propagate(failedToDestroyErr, "Attempted to destroy the services with UUID '%v' but had no success. You must manually destroy the services. Kurtosis will now try to remove its sidecar if it exists but might it fail as well.", serviceUuid)
	}

	// deleting the sidecar
	networkingSidecar, found := network.networkingSidecars[serviceName]
	if found {
		delete(network.networkingSidecars, serviceName)
		err = network.networkingSidecarManager.Remove(ctx, networkingSidecar)
		if errorResult == nil && err != nil {
			errorResult = stacktrace.Propagate(err, "Attempted to clean up the sidecar for service with name '%s' but an error occurred.", serviceName)
		}
	}
	return errorResult
}

// startRegisteredServices starts multiple services in parallel
//
// It iterates over all the services to start and kicks off a go subroutine for each of them.
// The subroutine will block until it can write to concurrencyControlChan. concurrencyControlChan is a simple buffered
// channel that can contain a max of 4 values. It's a common way in go to control concurrency to make sure no more than
// X subroutine is running at the same time.
//
// Once the for loops has started all the subroutine, we use a WaitGroup for this method to block until all subroutines
// have completed
//
// The subroutine accounts for its result populating the startedServices and failedServices maps (which are be accessed
// behind a mutex as those are not concurrent maps) and finishes by release a permit from the WaitGroup
func (network *DefaultServiceNetwork) startRegisteredServices(
	ctx context.Context,
	serviceConfigs map[service.ServiceUUID]*kurtosis_core_rpc_api_bindings.ServiceConfig,
	batchSize int,
) (map[service.ServiceUUID]*service.Service, map[service.ServiceUUID]error) {
	wg := sync.WaitGroup{}

	concurrencyControlChan := make(chan bool, batchSize)
	defer close(concurrencyControlChan)

	startedServices := map[service.ServiceUUID]*service.Service{}
	failedServices := map[service.ServiceUUID]error{}
	mapWriteMutex := sync.Mutex{}

	// async kick off all the routines one by one
	for serviceUuid, serviceConfig := range serviceConfigs {
		serviceToStartUuid := serviceUuid
		serviceToStartConfig := serviceConfig

		if len(failedServices) > 0 {
			// stop scheduling more service start
			// as one already failed, the full batch will be reverted anyway so no need to continue any further
			break
		}
		// The concurrencyControlChan will block if the buffer is currently full, i.e. if maxConcurrentServiceStart
		// subroutines are already running in the background
		concurrencyControlChan <- true
		wg.Add(1)
		go func() {
			defer func() {
				// at the end, make sure the subroutine releases one permit from the WaitGroup, and make sure to
				// also pop a value from the concurrencyControlChan to allow any potentially waiting subroutine to
				// start
				wg.Done()
				<-concurrencyControlChan
			}()
			logrus.Debugf("Starting service '%s'", serviceToStartUuid)
			startedService, err := network.startRegisteredService(ctx, serviceToStartUuid, serviceToStartConfig)
			mapWriteMutex.Lock()
			defer mapWriteMutex.Unlock()
			if err != nil {
				failedServices[serviceToStartUuid] = err
				logrus.Debugf("Service '%s' could not start due to some errors", serviceToStartUuid)
			} else {
				startedServices[serviceToStartUuid] = startedService
				logrus.Debugf("Service '%s' started successfully", serviceToStartUuid)
			}
		}()
	}

	// wait for all subroutines to complete and return
	wg.Wait()
	return startedServices, failedServices
}

// This method is not thread safe. Only call this from a method where there is a mutex lock on the network.
func (network *DefaultServiceNetwork) copyFilesFromServiceUnlocked(ctx context.Context, serviceName service.ServiceName, srcPath string, artifactName string) (enclave_data_directory.FilesArtifactUUID, error) {
	serviceObj, found := network.registeredServiceInfo[serviceName]
	if !found {
		return "", stacktrace.NewError("Cannot copy files from service '%v' because it does not exist in the network", serviceName)
	}
	serviceUuid := serviceObj.GetUUID()

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
		filesArtifactUuid, storeFileErr := store.StoreFile(pipeReader, artifactName)
		storeFilesArtifactResultChan <- storeFilesArtifactResult{
			err:               storeFileErr,
			filesArtifactUuid: filesArtifactUuid,
		}
	}()

	if err := network.gzipAndPushTarredFileBytesToOutput(ctx, pipeWriter, serviceUuid, srcPath); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred gzip'ing and pushing tar'd file bytes to the pipe")
	}

	storeFileResult := <-storeFilesArtifactResultChan
	if storeFileResult.err != nil {
		return "", stacktrace.Propagate(
			err,
			"An error occurred storing files from path '%v' on service '%v' in in the files artifact store",
			srcPath,
			serviceUuid,
		)
	}

	return storeFileResult.filesArtifactUuid, nil
}

func (network *DefaultServiceNetwork) gzipAndPushTarredFileBytesToOutput(
	ctx context.Context,
	output io.WriteCloser,
	serviceUuid service.ServiceUUID,
	srcPathOnContainer string,
) error {
	defer output.Close()

	// Need to compress the TAR bytes on our side, since we're not guaranteedj
	gzippingOutput := gzip.NewWriter(output)
	defer gzippingOutput.Close()

	if err := network.kurtosisBackend.CopyFilesFromUserService(ctx, network.enclaveUuid, serviceUuid, srcPathOnContainer, gzippingOutput); err != nil {
		return stacktrace.Propagate(err, "An error occurred copying source '%v' from user service with UUID '%v' in enclave with UUID '%v'", srcPathOnContainer, serviceUuid, network.enclaveUuid)
	}

	return nil
}

// This method is not thread safe. Only call this from a method where there is a mutex lock on the network.
func (network *DefaultServiceNetwork) addServiceToTopology(serviceName service.ServiceName, partitionID service_network_types.PartitionID) error {
	if err := network.topology.AddService(serviceName, partitionID); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred adding service with name '%v' to partition '%v' in the topology",
			serviceName,
			partitionID,
		)
	}
	shouldRemoveFromTopology := true
	defer func() {
		if shouldRemoveFromTopology {
			err := network.topology.RemoveService(serviceName)
			if err != nil {
				logrus.Errorf("An error occurred while removing service '%v' from the partition toplogy", serviceName)
			}
		}
	}()

	shouldRemoveFromTopology = false
	return nil
}

func (network *DefaultServiceNetwork) moveServiceToPartitionInTopology(serviceName service.ServiceName, partitionID service_network_types.PartitionID) error {
	isOperationSuccessful := false
	servicePartitions, err := network.topology.GetServicePartitions()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while fetching service partitions mapping")
	}
	serviceCurrentPartition, found := servicePartitions[serviceName]
	if !found {
		return stacktrace.NewError("Service with name '%s' not found in the topology", serviceName)
	}
	err = network.topology.RemoveService(serviceName)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while removing service '%v' from the network topology", serviceName)
	}
	defer func() {
		if isOperationSuccessful {
			return
		}
		if err := network.topology.AddService(serviceName, serviceCurrentPartition); err != nil {
			logrus.Errorf("Service '%s' could not be moved to partition '%s'. It should have been rolled back to its previous partition '%s' but this operation failed", serviceName, partitionID, serviceCurrentPartition)
			return
		}
	}()
	if err := network.topology.AddService(serviceName, partitionID); err != nil {
		return stacktrace.Propagate(err, "Error moving service '%s' to its new partition '%s'", serviceName, partitionID)
	}
	isOperationSuccessful = true
	return nil
}

// This method is not thread safe. Only call this from a method where there is a mutex lock on the network.
func (network *DefaultServiceNetwork) createSidecarAndAddToMap(ctx context.Context, service *service.Service) error {
	serviceRegistration := service.GetRegistration()
	serviceUUID := serviceRegistration.GetUUID()
	serviceName := serviceRegistration.GetName()

	sidecar, err := network.networkingSidecarManager.Add(ctx, serviceUUID)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred adding the networking sidecar for service `%v`", serviceName)
	}
	shouldRemoveSidecarFromManager := true
	defer func() {
		if shouldRemoveSidecarFromManager {
			err = network.networkingSidecarManager.Remove(ctx, sidecar)
			if err != nil {
				logrus.Errorf("Attempted to remove network sidecar during cleanup for service '%v' but failed", serviceName)
			}
		}
	}()

	network.networkSidecarsLock.Lock()
	network.networkingSidecars[serviceName] = sidecar
	shouldRemoveSidecarFromMap := true
	network.networkSidecarsLock.Unlock()
	defer func() {
		network.networkSidecarsLock.Lock()
		defer network.networkSidecarsLock.Unlock()
		if shouldRemoveSidecarFromMap {
			delete(network.networkingSidecars, serviceName)
		}
	}()

	if err := sidecar.InitializeTrafficControl(ctx); err != nil {
		return stacktrace.Propagate(err, "An error occurred initializing the newly-created networking-sidecar-traffic-control-qdisc-configuration for service `%v`", serviceName)
	}

	shouldRemoveSidecarFromMap = false
	shouldRemoveSidecarFromManager = false
	return nil
}

// This method is not thread safe. Only call this from a method where there is a mutex lock on the network.
func (network *DefaultServiceNetwork) renderTemplatesUnlocked(templatesAndDataByDestinationRelFilepath map[string]*kurtosis_core_rpc_api_bindings.RenderTemplatesToFilesArtifactArgs_TemplateAndData, artifactName string) (enclave_data_directory.FilesArtifactUUID, error) {
	tempDirForRenderedTemplates, err := os.MkdirTemp("", tempDirForRenderedTemplatesPrefix)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while creating a temp dir for rendered templates '%v'", tempDirForRenderedTemplates)
	}
	defer os.RemoveAll(tempDirForRenderedTemplates)

	for destinationRelFilepath, templateAndData := range templatesAndDataByDestinationRelFilepath {
		templateAsAString := templateAndData.Template
		templateDataAsJson := templateAndData.DataAsJson

		templateDataJsonAsBytes := []byte(templateDataAsJson)
		templateDataJsonReader := bytes.NewReader(templateDataJsonAsBytes)

		// We don't use standard json.Unmarshal as that converts large integers to floats
		// Using this custom decoder we get the json.Number representation which is closer to other json implementations
		// This talks about the issue further https://github.com/square/go-jose/issues/351#issuecomment-847193900
		decoder := json.NewDecoder(templateDataJsonReader)
		decoder.UseNumber()

		var templateData interface{}
		if err = decoder.Decode(&templateData); err != nil {
			return "", stacktrace.Propagate(err, "An error occurred while decoding the template data json '%v' for file '%v'", templateDataAsJson, destinationRelFilepath)
		}

		destinationFilepath := path.Join(tempDirForRenderedTemplates, destinationRelFilepath)
		if err = renderTemplateToFile(templateAsAString, templateData, destinationFilepath); err != nil {
			return "", stacktrace.Propagate(err, "There was an error in rendering template for file '%v'", destinationRelFilepath)
		}
	}

	compressedFile, err := shared_utils.CompressPath(tempDirForRenderedTemplates, ensureCompressedFileIsLesserThanGRPCLimit)
	if err != nil {
		return "", stacktrace.Propagate(err, "There was an error compressing dir '%v'", tempDirForRenderedTemplates)
	}

	store, err := network.enclaveDataDir.GetFilesArtifactStore()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while getting files artifact store")
	}
	filesArtifactUuid, err := store.StoreFile(bytes.NewReader(compressedFile), artifactName)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while storing the file '%v' in the files artifact store", compressedFile)
	}
	shouldDeleteFilesArtifact := true
	defer func() {
		if shouldDeleteFilesArtifact {
			if err = store.RemoveFile(string(filesArtifactUuid)); err != nil {
				logrus.Errorf("We tried to clean up the files artifact '%v' we had stored but failed:\n%v", artifactName, err)
			}
		}
	}()

	shouldDeleteFilesArtifact = false
	return filesArtifactUuid, nil
}

// This method is not thread safe. Only call this from a method where there is a mutex lock on the network.
func (network *DefaultServiceNetwork) uploadFilesArtifactUnlocked(data []byte, artifactName string) (enclave_data_directory.FilesArtifactUUID, error) {
	reader := bytes.NewReader(data)

	filesArtifactStore, err := network.enclaveDataDir.GetFilesArtifactStore()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while getting files artifact store")
	}

	filesArtifactUuid, err := filesArtifactStore.StoreFile(reader, artifactName)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while trying to store files.")
	}

	return filesArtifactUuid, nil
}

// This isn't thread safe and must be called from a thread safe context
func (network *DefaultServiceNetwork) cleanupInternalMapsUnlocked(serviceName service.ServiceName) {
	_, found := network.registeredServiceInfo[serviceName]
	if !found {
		return
	}
	delete(network.registeredServiceInfo, serviceName)
}

// This isn't thread safe and must be called from a thread safe context
func (network *DefaultServiceNetwork) getServiceNameForIdentifierUnlocked(serviceIdentifier string) (service.ServiceName, error) {
	maybeServiceUuid := service.ServiceUUID(serviceIdentifier)
	serviceUuidToServiceName := map[service.ServiceUUID]service.ServiceName{}
	serviceShortenedUuidToServiceName := map[string][]service.ServiceName{}

	for serviceName, registration := range network.registeredServiceInfo {
		serviceUuid := registration.GetUUID()
		serviceShortenedUuid := uuid_generator.ShortenedUUIDString(string(serviceUuid))
		serviceUuidToServiceName[serviceUuid] = serviceName
		serviceShortenedUuidToServiceName[serviceShortenedUuid] = append(serviceShortenedUuidToServiceName[serviceShortenedUuid], serviceName)
	}

	if serviceName, found := serviceUuidToServiceName[maybeServiceUuid]; found {
		return serviceName, nil
	}

	maybeShortenedUuid := serviceIdentifier
	if serviceNames, found := serviceShortenedUuidToServiceName[maybeShortenedUuid]; found {
		if len(serviceNames) == exactlyOneShortenedUuidMatch {
			return serviceNames[0], nil
		} else {
			return "", stacktrace.NewError("Found multiple matching service names '%v' for shortened uuid '%v'. Please be more specific", serviceNames, maybeShortenedUuid)
		}
	}

	maybeServiceName := service.ServiceName(serviceIdentifier)
	if _, found := network.registeredServiceInfo[maybeServiceName]; found {
		return maybeServiceName, nil
	}

	return "", stacktrace.NewError("Couldn't find a matching service name for identifier '%v'", serviceIdentifier)
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
	apiProto := port.GetTransportProtocol()
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

	result, err := port_spec.NewPortSpec(portNumUint16, portSpecProto, port.GetMaybeApplicationProtocol())
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

func renderTemplateToFile(templateAsAString string, templateData interface{}, destinationFilepath string) error {
	parsedTemplate, err := template.New(path.Base(destinationFilepath)).Parse(templateAsAString)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred in parsing the template string '%v'", destinationFilepath)
	}

	// Creat all parent directories to account for nesting
	destinationFileDir := path.Dir(destinationFilepath)
	if err = os.MkdirAll(destinationFileDir, folderPermissionForRenderedTemplates); err != nil {
		return stacktrace.Propagate(err, "There was an error in creating the parent directory '%v' to write the file '%v' into.", destinationFileDir, destinationFilepath)
	}

	renderedTemplateFile, err := os.Create(destinationFilepath)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while creating temporary file to render template into for file '%v'.", destinationFilepath)
	}
	defer renderedTemplateFile.Close()

	if err = parsedTemplate.Execute(renderedTemplateFile, templateData); err != nil {
		return stacktrace.Propagate(err, "An error occurred while writing the rendered template to destination '%v'", destinationFilepath)
	}
	return nil
}
