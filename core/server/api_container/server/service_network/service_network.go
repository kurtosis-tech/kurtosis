/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package service_network

import (
	"compress/gzip"
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifact_expansion"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/service_network/files_artifact_expander"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/service_network/networking_sidecar"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/kurtosis-core/server/commons/enclave_data_directory"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	defaultPartitionId                       service_network_types.PartitionID = "default"
	startingDefaultConnectionPacketLossValue                                   = 0
)

type storeFilesArtifactResult struct {
	filesArtifactId service.FilesArtifactID
	err             error
}

/*
This is the in-memory representation of the service network that the API container will manipulate. To make
	any changes to the test network, this struct must be used.
*/
type ServiceNetwork struct {
	enclaveId enclave.EnclaveID

	mutex *sync.Mutex // VERY IMPORTANT TO CHECK AT THE START OF EVERY METHOD!

	// Whether partitioning has been enabled for this particular test
	isPartitioningEnabled bool

	kurtosisBackend backend_interface.KurtosisBackend

	enclaveDataDir *enclave_data_directory.EnclaveDataDirectory

	filesArtifactExpander    *files_artifact_expander.FilesArtifactExpander

	topology *partition_topology.PartitionTopology

	networkingSidecars map[service.ServiceID]networking_sidecar.NetworkingSidecarWrapper

	networkingSidecarManager networking_sidecar.NetworkingSidecarManager

	// Technically we SHOULD query the backend rather than ever storing any of this information, but we're able to get away with
	// this because the API container is the only client that modifies service state
	registeredServiceInfo map[service.ServiceID]*service.ServiceRegistration
}

func NewServiceNetwork(
	enclaveId enclave.EnclaveID,
	isPartitioningEnabled bool,
	kurtosisBackend backend_interface.KurtosisBackend,
	enclaveDataDir *enclave_data_directory.EnclaveDataDirectory,
	networkingSidecarManager networking_sidecar.NetworkingSidecarManager,
	filesArtifactExpander *files_artifact_expander.FilesArtifactExpander,
) *ServiceNetwork {
	defaultPartitionConnection := partition_topology.PartitionConnection{
		PacketLossPercentage: startingDefaultConnectionPacketLossValue,
	}
	return &ServiceNetwork{
		enclaveId:             enclaveId,
		mutex:                 &sync.Mutex{},
		isPartitioningEnabled: isPartitioningEnabled,
		kurtosisBackend:       kurtosisBackend,
		enclaveDataDir:        enclaveDataDir,
		filesArtifactExpander: filesArtifactExpander,
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

// Registers a service for use with the network (creating the IPs and so forth), but doesn't start it
// If the partition ID is empty, registers the service with the default partition
func (network ServiceNetwork) RegisterService(
	ctx context.Context,
	serviceId service.ServiceID,
	partitionId service_network_types.PartitionID,
) (net.IP, error) {
	// TODO extract this into a wrapper function that can be wrapped around every service call (so we don't forget)
	network.mutex.Lock()
	defer network.mutex.Unlock()

	if _, found := network.registeredServiceInfo[serviceId]; found {
		return nil, stacktrace.NewError(
			"Cannot register service '%v' because it already exists in the network",
			serviceId,
		)
	}

	if partitionId == "" {
		partitionId = defaultPartitionId
	}
	if _, found := network.topology.GetPartitionServices()[partitionId]; !found {
		return nil, stacktrace.NewError(
			"No partition with ID '%v' exists in the current partition topology",
			partitionId,
		)
	}

	userService, err := network.kurtosisBackend.RegisterUserService(
		ctx,
		network.enclaveId,
		serviceId,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred registering service with ID '%v'", serviceId)
	}
	shouldDestroyService := true
	defer func() {
		if shouldDestroyService {
			network.destroyServiceBestEffortAfterRegistrationFailure(userService.GetGUID())
		}
	}()

	network.registeredServiceInfo[serviceId] = userService
	shouldRemoveFromServiceMap := true
	defer func() {
		if shouldRemoveFromServiceMap {
			delete(network.registeredServiceInfo, serviceId)
		}
	}()

	if err := network.topology.AddService(serviceId, partitionId); err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred adding service with ID '%v' to partition '%v' in the topology",
			serviceId,
			partitionId,
		)
	}
	shouldRemoveTopologyAddition := true
	defer func() {
		if shouldRemoveTopologyAddition {
			network.topology.RemoveService(serviceId)
		}
	}()

	shouldDestroyService = false
	shouldRemoveFromServiceMap = false
	shouldRemoveTopologyAddition = false
	return userService.GetPrivateIP(), nil
}

// TODO add tests for this
/*
Starts a previously-registered but not-started service by creating it in a container

Returns:
	Mapping of port-used-by-service -> port-on-the-Docker-host-machine where the user can make requests to the port
		to access the port. If a used port doesn't have a host port bound, then the value will be nil.
*/
func (network *ServiceNetwork) StartService(
	ctx context.Context,
	serviceId service.ServiceID,
	imageName string,
	privatePorts map[string]*port_spec.PortSpec,
	entrypointArgs []string,
	cmdArgs []string,
	dockerEnvVars map[string]string,
	filesArtifactMountDirpaths map[service.FilesArtifactID]string,
) (
	resultServiceGuid service.ServiceGUID,
	resultMaybePublicIpAddr net.IP, // Will be nil if the service doesn't declare any private ports
	resultPublicPorts map[string]*port_spec.PortSpec,
	resultErr error,
) {
	// TODO extract this into a wrapper function that can be wrapped around every service call (so we don't forget)
	network.mutex.Lock()
	defer network.mutex.Unlock()

	registration, found := network.registeredServiceInfo[serviceId]
	if !found {
		return "", nil, nil, stacktrace.NewError("Cannot start service; no registration exists for service with ID '%v'", serviceId)
	}
	serviceGuid := registration.GetGUID()

	// When partitioning is enabled, there's a race condition where:
	//   a) we need to start the service before we can launch the sidecar but
	//   b) we can't modify the qdisc configuration until the sidecar container is launched.
	// This means that there's a period of time at startup where the container might not be partitioned. We solve
	//  this by setting the packet loss config of the new service in the already-existing services' qdisc.
	// This means that when the new service is launched, even if its own qdisc isn't yet updated, all the services
	//  it would communicate are already dropping traffic to it before it even starts.
	if network.isPartitioningEnabled {
		servicePacketLossConfigurationsByServiceID, err := network.topology.GetServicePacketLossConfigurationsByServiceID()
		if err != nil {
			return "", nil, nil, stacktrace.Propagate(err, "An error occurred getting the packet loss configuration by service ID "+
				" to know what packet loss updates to apply on the new node")
		}

		servicesPacketLossConfigurationsWithoutNewNode := map[service.ServiceID]map[service.ServiceID]float32{}
		for serviceIdInTopology, otherServicesPacketLossConfigs := range servicePacketLossConfigurationsByServiceID {
			if serviceId == serviceIdInTopology {
				continue
			}
			servicesPacketLossConfigurationsWithoutNewNode[serviceIdInTopology] = otherServicesPacketLossConfigs
		}

		if err := updateTrafficControlConfiguration(
			ctx,
			servicesPacketLossConfigurationsWithoutNewNode,
			network.registeredServiceInfo,
			network.networkingSidecars,
		); err != nil {
			return "", nil, nil, stacktrace.Propagate(
				err,
				"An error occurred updating the traffic control configuration of all the other services "+
					 "before adding the new service, meaning that the service wouldn't actually start in a partition",
			)
		}
		// TODO defer an undo somehow???
	}

	userService, err := network.startService(
		ctx,
		serviceGuid,
		imageName,
		privatePorts,
		entrypointArgs,
		cmdArgs,
		dockerEnvVars,
		filesArtifactMountDirpaths,
	)
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(
			err,
			"An error occurred starting service '%v'",
			serviceId,
		)
	}
	// NOTE: There's no real way to defer-undo the service start

	if network.isPartitioningEnabled {
		sidecar, err := network.networkingSidecarManager.Add(ctx, registration.GetGUID())
		if err != nil {
			return "", nil, nil, stacktrace.Propagate(err, "An error occurred adding the networking sidecar")
		}
		network.networkingSidecars[serviceId] = sidecar

		if err := sidecar.InitializeTrafficControl(ctx); err != nil {
			return "", nil, nil, stacktrace.Propagate(err, "An error occurred initializing the newly-created networking-sidecar-traffic-control-qdisc-configuration")
		}

		// TODO Getting packet loss configuration by service ID is an expensive call and, as of 2021-11-23, we do it twice - the solution is to make
		//  Getting packet loss configuration by service ID not an expensive call
		servicePacketLossConfigurationsByServiceID, err := network.topology.GetServicePacketLossConfigurationsByServiceID()
		if err != nil {
			return "", nil, nil, stacktrace.Propagate(err, "An error occurred getting the packet loss configuration by service ID "+
				" to know what packet loss updates to apply on the new node")
		}
		newNodeServicePacketLossConfiguration := servicePacketLossConfigurationsByServiceID[serviceId]
		updatesToApply := map[service.ServiceID]map[service.ServiceID]float32{
			serviceId: newNodeServicePacketLossConfiguration,
		}
		if err := updateTrafficControlConfiguration(ctx, updatesToApply, network.registeredServiceInfo, network.networkingSidecars); err != nil {
			return "", nil, nil, stacktrace.Propagate(err, "An error occurred applying the traffic control configuration on the new node to partition it "+
				"off from other nodes")
		}
	}

	return serviceGuid, userService.GetMaybePublicIP(), userService.GetMaybePublicPorts(), nil
}

func (network *ServiceNetwork) RemoveService(
	ctx context.Context,
	serviceId service.ServiceID,
	containerStopTimeout time.Duration,
) error {
	network.mutex.Lock()
	defer network.mutex.Unlock()

	serviceToRemove, found := network.registeredServiceInfo[serviceId]
	if !found {
		return stacktrace.NewError("No service found with ID '%v'", serviceId)
	}
	serviceGuid := serviceToRemove.GetGUID()

	network.topology.RemoveService(serviceId)

	delete(network.registeredServiceInfo, serviceId)

	// We stop the service, rather than destroying it, so that we can keep logs around
	stopServiceFilters := &service.ServiceFilters{
		GUIDs:    map[service.ServiceGUID]bool{
			serviceGuid: true,
		},
	}
	_, erroredGuids, err := network.kurtosisBackend.StopUserServices(ctx, network.enclaveId, stopServiceFilters)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred during the call to stop service '%v'", serviceGuid)
	}
	if err, found := erroredGuids[serviceGuid]; found {
		return stacktrace.Propagate(err, "An error occurred stopping service '%v'", serviceGuid)
	}

	sidecar, foundSidecar := network.networkingSidecars[serviceId]
	if network.isPartitioningEnabled && foundSidecar {
		// NOTE: As of 2020-12-31, we don't need to update the iptables of the other services in the network to
		//  clear the now-removed service's IP because:
		// 	 a) nothing is using it so it doesn't do anything and
		//	 b) all service's iptables get overwritten on the next Add/Repartition call
		// If we ever do incremental iptables though, we'll need to fix all the other service's iptables here!
		if err := network.networkingSidecarManager.Remove(ctx, sidecar); err != nil {
			return stacktrace.Propagate(err, "An error occurred destroying the sidecar for service with ID '%v'", serviceId)
		}
		delete(network.networkingSidecars, serviceId)
		logrus.Debugf("Successfully removed sidecar attached to service with ID '%v'", serviceId)
	}

	return nil
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
		return stacktrace.Propagate(err,"Failed to pause service '%v'", serviceId)
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
		return stacktrace.Propagate(err,"Failed to unpause service '%v'", serviceId)
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
		GUIDs:    map[service.ServiceGUID]bool{
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
			"A registration exists for service GUID '%v' but no service objects were found; this indicates that the service was " +
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

func (network *ServiceNetwork) CopyFilesFromService(ctx context.Context, serviceId service.ServiceID, srcPath string) (service.FilesArtifactID, error) {
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
		filesArtifactId, storeFileErr := store.StoreFile(pipeReader)
		storeFilesArtifactResultChan <- storeFilesArtifactResult{
			filesArtifactId: filesArtifactId,
			err:             storeFileErr,
		}
	}()

	if err := network.gzipAndPushTarredFileBytesToOutput(ctx, pipeWriter, serviceGuid, srcPath); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred gzip'ing and pushing tar'd file bytes to the pipe")
	}

	storeFileResult := <- storeFilesArtifactResultChan
	if storeFileResult.err != nil {
		return "", stacktrace.Propagate(
			err,
			"An error occurred storing files from path '%v' on service '%v' in in the files artifact store",
			srcPath,
			serviceGuid,
		)
	}

	return storeFileResult.filesArtifactId, nil
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

func (network *ServiceNetwork) destroyServiceBestEffortAfterRegistrationFailure(
	serviceGuid service.ServiceGUID,
) {
	destroyServiceFilters := &service.ServiceFilters{
		GUIDs: map[service.ServiceGUID]bool{
			serviceGuid: true,
		},
	}
	// Use background context in case the input one is cancelled
	_, erroredRegistrations, err := network.kurtosisBackend.DestroyUserServices(context.Background(), network.enclaveId, destroyServiceFilters)
	var errToPrint error
	if err != nil {
		errToPrint = err
	} else if destroyErr, found := erroredRegistrations[serviceGuid]; found {
		errToPrint = destroyErr
	}
	if errToPrint != nil {
		logrus.Warnf(
			"Registering service with ID '%v' didn't complete successfully so we tried to destroy the " +
				"service that we created, but doing so threw an error:\n%v",
			serviceGuid,
			errToPrint,
		)
		logrus.Warnf(
			"!!! ACTION REQUIRED !!! You'll need to manually destroy service with GUID '%v'!!!",
			serviceGuid,
		)
	}
}

func (network *ServiceNetwork) startService(
	ctx context.Context,
	serviceGuid service.ServiceGUID,
	imageName string,
	privatePorts map[string]*port_spec.PortSpec,
	entrypointArgs []string,
	cmdArgs []string,
	envVars map[string]string,
	// Mapping of UUIDs of previously-registered files artifacts -> mountpoints on the container
	// being launched
	filesArtifactUuidsToMountpoints map[service.FilesArtifactID]string,
) (
	resultUserService *service.Service,
	resultErr error,
) {
	usedArtifactUuidSet := map[service.FilesArtifactID]bool{}
	for artifactUuid := range filesArtifactUuidsToMountpoints {
		usedArtifactUuidSet[artifactUuid] = true
	}

	// First expand the files artifacts into volumes, so that any errors get caught early
	// NOTE: if users don't need to investigate the volume contents, we could keep track of the volumes we create
	//  and delete them at the end of the test to keep things cleaner
	artifactUuidsToExpansionGUIDs, err := network.filesArtifactExpander.ExpandArtifacts(ctx, serviceGuid, usedArtifactUuidSet)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred expanding the requested files artifacts into volumes")
	}

	artifactVolumeMounts := map[files_artifact_expansion.FilesArtifactExpansionGUID]string{}
	for artifactUuid, mountpoint := range filesArtifactUuidsToMountpoints {
		artifactExpansionGUID, found := artifactUuidsToExpansionGUIDs[artifactUuid]
		if !found {
			return nil, stacktrace.NewError(
				"Even though we declared that we need files artifact '%v' to be expanded, no expansion containing the "+
					"expanded contents was found; this is a bug in Kurtosis",
				artifactUuid,
			)
		}
		artifactVolumeMounts[artifactExpansionGUID] = mountpoint
	}

	launchedUserService, err := network.kurtosisBackend.StartUserService(
		ctx,
		network.enclaveId,
		serviceGuid,
		imageName,
		privatePorts,
		entrypointArgs,
		cmdArgs,
		envVars,
		artifactVolumeMounts,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred starting service '%v' with image '%v'", serviceGuid, imageName)
	}
	return launchedUserService, nil
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