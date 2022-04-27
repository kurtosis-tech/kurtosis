/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package service_network

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/free-ip-addr-tracker-lib/lib"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/service_network/networking_sidecar"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/service_network/user_service_launcher"
	"github.com/kurtosis-tech/kurtosis-core/server/commons/current_time_str_provider"
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

// Information that gets created with a service's registration
type serviceRegistrationInfo struct {
	serviceGUID      service.ServiceGUID
	privateIpAddr    net.IP
	serviceDirectory *enclave_data_directory.ServiceDirectory
}

// Information that gets created when a container is started for a service
type serviceRunInfo struct {
	// Service's container ID
	serviceGUID service.ServiceGUID

	// Where the enclave data dir is bind-mounted on the service
	enclaveDataDirMntDirpath string

	// NOTE: When we want to make restart-able enclaves, we'll need to read these values from the container every time
	//  we need them (rather than storing them in-memory on the API container, which means the API container can't be restarted)
	privatePorts      map[string]*port_spec.PortSpec
	maybePublicIpAddr net.IP                         // Can be nil if the service doesn't declare any private ports
	maybePublicPorts  map[string]*port_spec.PortSpec // Will be nil if the service doesn't declare any private ports
}

/*
This is the in-memory representation of the service network that the API container will manipulate. To make
	any changes to the test network, this struct must be used.
*/
type ServiceNetworkImpl struct {
	enclaveId enclave.EnclaveID

	// When the network is destroyed, all requests will fail
	// This ensures that when the initializer tells the API container to destroy everything, the still-running
	//  testsuite can't create more work
	isDestroyed bool // VERY IMPORTANT TO CHECK AT THE START OF EVERY METHOD!

	mutex *sync.Mutex // VERY IMPORTANT TO CHECK AT THE START OF EVERY METHOD!

	// Whether partitioning has been enabled for this particular test
	isPartitioningEnabled bool

	freeIpAddrTracker *lib.FreeIpAddrTracker

	kurtosisBackend backend_interface.KurtosisBackend

	enclaveDataDir *enclave_data_directory.EnclaveDataDirectory

	userServiceLauncher *user_service_launcher.UserServiceLauncher

	topology *partition_topology.PartitionTopology

	// These are separate maps, rather than being bundled into a single containerInfo-valued map, because
	//  they're registered at different times (rather than in one atomic operation)
	serviceRegistrationInfo map[service.ServiceID]serviceRegistrationInfo
	serviceRunInfo          map[service.ServiceID]serviceRunInfo

	networkingSidecars map[service.ServiceID]networking_sidecar.NetworkingSidecarWrapper

	networkingSidecarManager networking_sidecar.NetworkingSidecarManager
}

func NewServiceNetworkImpl(
	enclaveId enclave.EnclaveID,
	isPartitioningEnabled bool,
	freeIpAddrTracker *lib.FreeIpAddrTracker,
	kurtosisBackend backend_interface.KurtosisBackend,
	enclaveDataDir *enclave_data_directory.EnclaveDataDirectory,
	userServiceLauncher *user_service_launcher.UserServiceLauncher,
	networkingSidecarManager networking_sidecar.NetworkingSidecarManager) *ServiceNetworkImpl {
	defaultPartitionConnection := partition_topology.PartitionConnection{PacketLossPercentage: startingDefaultConnectionPacketLossValue}
	return &ServiceNetworkImpl{
		enclaveId:             enclaveId,
		isDestroyed:           false,
		isPartitioningEnabled: isPartitioningEnabled,
		freeIpAddrTracker:     freeIpAddrTracker,
		kurtosisBackend:       kurtosisBackend,
		enclaveDataDir:        enclaveDataDir,
		userServiceLauncher:   userServiceLauncher,
		mutex:                 &sync.Mutex{},
		topology: partition_topology.NewPartitionTopology(
			defaultPartitionId,
			defaultPartitionConnection,
		),
		serviceRegistrationInfo:  map[service.ServiceID]serviceRegistrationInfo{},
		serviceRunInfo:           map[service.ServiceID]serviceRunInfo{},
		networkingSidecars:       map[service.ServiceID]networking_sidecar.NetworkingSidecarWrapper{},
		networkingSidecarManager: networkingSidecarManager,
	}
}

/*
Completely repartitions the network, throwing away the old topology
*/
func (network *ServiceNetworkImpl) Repartition(
	ctx context.Context,
	newPartitionServices map[service_network_types.PartitionID]map[service.ServiceID]bool,
	newPartitionConnections map[service_network_types.PartitionConnectionID]partition_topology.PartitionConnection,
	newDefaultConnection partition_topology.PartitionConnection) error {
	network.mutex.Lock()
	defer network.mutex.Unlock()
	if network.isDestroyed {
		return stacktrace.NewError("Cannot repartition; the service network has been destroyed")
	}

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

	if err := updateTrafficControlConfiguration(ctx, servicePacketLossConfigurationsByServiceID, network.serviceRegistrationInfo, network.networkingSidecars); err != nil {
		return stacktrace.Propagate(err, "An error occurred updating the traffic control configuration to match the target service packet loss configurations after repartitioning")
	}
	return nil
}

// Registers a service for use with the network (creating the IPs and so forth), but doesn't start it
// If the partition ID is empty, registers the service with the default partition
func (network ServiceNetworkImpl) RegisterService(
	serviceId service.ServiceID,
	partitionId service_network_types.PartitionID,
) (net.IP, string, error) {
	// TODO extract this into a wrapper function that can be wrapped around every service call (so we don't forget)
	network.mutex.Lock()
	defer network.mutex.Unlock()
	if network.isDestroyed {
		return nil, "", stacktrace.NewError("Cannot register service with ID '%v'; the service network has been destroyed", serviceId)
	}

	if strings.TrimSpace(string(serviceId)) == "" {
		return nil, "", stacktrace.NewError("Service ID cannot be empty or whitespace")
	}

	if _, found := network.serviceRegistrationInfo[serviceId]; found {
		return nil, "", stacktrace.NewError("Cannot register service with ID '%v'; a service with that ID already exists", serviceId)
	}

	if partitionId == "" {
		partitionId = defaultPartitionId
	}
	if _, found := network.topology.GetPartitionServices()[partitionId]; !found {
		return nil, "", stacktrace.NewError(
			"No partition with ID '%v' exists in the current partition topology",
			partitionId,
		)
	}

	ip, err := network.freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return nil, "", stacktrace.Propagate(err, "An error occurred getting an IP for service with ID '%v'", serviceId)
	}
	shouldFreeIpAddr := true
	defer func() {
		// To keep our bookkeeping correct, if an error occurs later we need to back out the IP-adding that we do now
		if shouldFreeIpAddr {
			network.freeIpAddrTracker.ReleaseIpAddr(ip)
		}
	}()
	logrus.Debugf("Giving service '%v' IP '%v'", serviceId, ip.String())

	serviceGuid := newServiceGUID(serviceId)
	serviceDirectory, err := network.enclaveDataDir.GetServiceDirectory(serviceGuid)
	if err != nil {
		return nil, "", stacktrace.Propagate(err, "An error occurred creating a new service directory for service with GUID '%v'", serviceGuid)
	}

	newServiceRegistrationInfo := serviceRegistrationInfo{
		serviceGUID:      serviceGuid,
		privateIpAddr:    ip,
		serviceDirectory: serviceDirectory,
	}

	network.serviceRegistrationInfo[serviceId] = newServiceRegistrationInfo
	shouldUndoRegistrationInfoAdd := true
	defer func() {
		// If an error occurs, the service ID won't be used so we need to delete it from the map
		if shouldUndoRegistrationInfoAdd {
			delete(network.serviceRegistrationInfo, serviceId)
		}
	}()

	if err := network.topology.AddService(serviceId, partitionId); err != nil {
		return nil, "", stacktrace.Propagate(
			err,
			"An error occurred adding service with ID '%v' to partition '%v' in the topology",
			serviceId,
			partitionId,
		)
	}

	shouldFreeIpAddr = false
	shouldUndoRegistrationInfoAdd = false
	return ip, serviceDirectory.GetDirpathRelativeToDataDirRoot(), nil
}

// TODO add tests for this
/*
Starts a previously-registered but not-started service by creating it in a container

Returns:
	Mapping of port-used-by-service -> port-on-the-Docker-host-machine where the user can make requests to the port
		to access the port. If a used port doesn't have a host port bound, then the value will be nil.
*/
func (network *ServiceNetworkImpl) StartService(
	ctx context.Context,
	serviceId service.ServiceID,
	imageName string,
	privatePorts map[string]*port_spec.PortSpec,
	entrypointArgs []string,
	cmdArgs []string,
	dockerEnvVars map[string]string,
	enclaveDataDirMntDirpath string,
	// TODO REMOVE
	oldFilesArtifactMountDirpaths map[service.FilesArtifactID]string,
	filesArtifactMountDirpaths map[service.FilesArtifactID]string,
) (
	resultMaybePublicIpAddr net.IP, // Will be nil if the service doesn't declare any private ports
	resultPublicPorts map[string]*port_spec.PortSpec,
	resultErr error,
) {
	// TODO extract this into a wrapper function that can be wrapped around every service call (so we don't forget)
	network.mutex.Lock()
	defer network.mutex.Unlock()
	if network.isDestroyed {
		return nil, nil, stacktrace.NewError("Cannot start container for service with ID '%v'; the service network has been destroyed", serviceId)
	}

	registrationInfo, registrationInfoFound := network.serviceRegistrationInfo[serviceId]
	if !registrationInfoFound {
		return nil, nil, stacktrace.NewError("Cannot start container for service with ID '%v'; no service with that ID has been registered", serviceId)
	}
	if _, found := network.serviceRunInfo[serviceId]; found {
		return nil, nil, stacktrace.NewError("Cannot start container for service with ID '%v'; that service ID already has run information associated with it", serviceId)
	}
	serviceGuid := registrationInfo.serviceGUID
	serviceIpAddr := registrationInfo.privateIpAddr

	// When partitioning is enabled, there's a race condition where:
	//   a) we need to start the service before we can launch the sidecar but
	//   b) we can't modify the qdisc configuration until the sidecar container is launched.
	// This means that there's a period of time at startup where the container might not be partitioned. We solve
	//  this by setting the packet loss config of the new service in the already-existing services' qdisc.
	// This means that when the new service is launched, even if its own qdisc isn't yet updated, all the services
	//  it would communicate are already dropping traffic to it.
	if network.isPartitioningEnabled {
		servicePacketLossConfigurationsByServiceID, err := network.topology.GetServicePacketLossConfigurationsByServiceID()
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred getting the packet loss configuration by service ID "+
				" to know what packet loss updates to apply on the new node")
		}

		servicesPacketLossConfigurationsWithoutNewNode := map[service.ServiceID]map[service.ServiceID]float32{}
		for serviceIdInTopology, otherServicesPacketLossConfigs := range servicePacketLossConfigurationsByServiceID {
			if serviceId == serviceIdInTopology {
				continue
			}
			servicesPacketLossConfigurationsWithoutNewNode[serviceIdInTopology] = otherServicesPacketLossConfigs
		}

		if err := updateTrafficControlConfiguration(ctx, servicesPacketLossConfigurationsWithoutNewNode, network.serviceRegistrationInfo, network.networkingSidecars); err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred updating the traffic control configuration of all the other services "+
				"before adding the node, meaning that the node wouldn't actually start in a partition")
		}
	}

	userService, err := network.userServiceLauncher.Launch(
		ctx,
		serviceGuid,
		serviceId,
		network.enclaveId,
		serviceIpAddr,
		imageName,
		privatePorts,
		entrypointArgs,
		cmdArgs,
		dockerEnvVars,
		enclaveDataDirMntDirpath,
		oldFilesArtifactMountDirpaths,
		filesArtifactMountDirpaths,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(
			err,
			"An error occurred creating the user service",
		)
	}

	// TODO Restart-able enclaves: Don't store any service information in memory; instead, get it all from the KurtosisBackend when required
	serviceGUID := userService.GetGUID()
	maybeServicePublicIpAddr := userService.GetMaybePublicIP()
	maybeServicePublicPorts := userService.GetMaybePublicPorts()

	runInfo := serviceRunInfo{
		serviceGUID:              serviceGUID,
		enclaveDataDirMntDirpath: enclaveDataDirMntDirpath,
		privatePorts:             privatePorts,
		maybePublicIpAddr:        maybeServicePublicIpAddr,
		maybePublicPorts:         maybeServicePublicPorts,
	}
	network.serviceRunInfo[serviceId] = runInfo

	if network.isPartitioningEnabled {
		sidecar, err := network.networkingSidecarManager.Add(ctx, serviceGuid)
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred adding the networking sidecar")
		}
		network.networkingSidecars[serviceId] = sidecar

		if err := sidecar.InitializeTrafficControl(ctx); err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred initializing the newly-created networking-sidecar-traffic-control-qdisc-configuration")
		}

		// TODO Getting packet loss configuration by service ID is an expensive call and, as of 2021-11-23, we do it twice - the solution is to make
		//  Getting packet loss configuration by service ID not an expensive call
		servicePacketLossConfigurationsByServiceID, err := network.topology.GetServicePacketLossConfigurationsByServiceID()
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred getting the packet loss configuration by service ID "+
				" to know what packet loss updates to apply on the new node")
		}
		newNodeServicePacketLossConfiguration := servicePacketLossConfigurationsByServiceID[serviceId]
		updatesToApply := map[service.ServiceID]map[service.ServiceID]float32{
			serviceId: newNodeServicePacketLossConfiguration,
		}
		if err := updateTrafficControlConfiguration(ctx, updatesToApply, network.serviceRegistrationInfo, network.networkingSidecars); err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred applying the traffic control configuration on the new node to partition it "+
				"off from other nodes")
		}
	}

	return maybeServicePublicIpAddr, maybeServicePublicPorts, nil
}

func (network *ServiceNetworkImpl) RemoveService(
	ctx context.Context,
	serviceId service.ServiceID,
	containerStopTimeout time.Duration,
) error {
	// TODO switch to a wrapper function
	network.mutex.Lock()
	defer network.mutex.Unlock()
	if network.isDestroyed {
		return stacktrace.NewError("Cannot remove service; the service network has been destroyed")
	}

	if err := network.removeServiceWithoutMutex(ctx, serviceId, containerStopTimeout); err != nil {
		return stacktrace.Propagate(err, "An error occurred removing service with ID '%v'", serviceId)
	}
	return nil
}

func (network *ServiceNetworkImpl) ExecCommand(
	ctx context.Context,
	serviceId service.ServiceID,
	command []string,
) (int32, string, error) {
	// NOTE: This will block all other operations while this command is running!!!! We might need to change this so it's
	// asynchronous
	network.mutex.Lock()
	defer network.mutex.Unlock()
	if network.isDestroyed {
		return 0, "", stacktrace.NewError("Cannot run exec command; the service network has been destroyed")
	}

	runInfo, found := network.serviceRunInfo[serviceId]
	if !found {
		return 0, "", stacktrace.NewError(
			"Could not run exec command '%v' against service '%v'; no container has been created for the service yet",
			command,
			serviceId)
	}

	// NOTE: This is a SYNCHRONOUS command, meaning that the entire network will be blocked until the command finishes
	// In the future, this will likely be insufficient

	serviceGuid := runInfo.serviceGUID
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

func (network *ServiceNetworkImpl) GetServiceRegistrationInfo(serviceId service.ServiceID) (
	privateIpAddr net.IP,
	relativeServiceDirpath string,
	resultErr error,
) {
	network.mutex.Lock()
	defer network.mutex.Unlock()
	if network.isDestroyed {
		return nil, "", stacktrace.NewError("Cannot get registration info for service '%v'; the service network has been destroyed", serviceId)
	}

	registrationInfo, found := network.serviceRegistrationInfo[serviceId]
	if !found {
		return nil, "", stacktrace.NewError("No registration information found for service with ID '%v'", serviceId)
	}

	return registrationInfo.privateIpAddr, registrationInfo.serviceDirectory.GetDirpathRelativeToDataDirRoot(), nil
}


func (network *ServiceNetworkImpl) GetServiceRunInfo(serviceId service.ServiceID) (
	privatePorts map[string]*port_spec.PortSpec,
	publicIpAddr net.IP,
	publicPorts map[string]*port_spec.PortSpec,
	enclaveDataDirMntDirpath string,
	resultErr error,
) {
	network.mutex.Lock()
	defer network.mutex.Unlock()
	if network.isDestroyed {
		return nil, nil, nil, "", stacktrace.NewError("Cannot get run info for service '%v'; the service network has been destroyed", serviceId)
	}

	runInfo, found := network.serviceRunInfo[serviceId]
	if !found {
		return nil, nil, nil, "", stacktrace.NewError("No run information found for service with ID '%v'", serviceId)
	}
	return runInfo.privatePorts, runInfo.maybePublicIpAddr, runInfo.maybePublicPorts, runInfo.enclaveDataDirMntDirpath, nil
}

func (network *ServiceNetworkImpl) GetServiceIDs() map[service.ServiceID]bool {

	serviceIDs := make(map[service.ServiceID]bool, len(network.serviceRunInfo))

	for serviceId := range network.serviceRunInfo {
		if _, ok := serviceIDs[serviceId]; !ok {
			serviceIDs[serviceId] = true
		}
	}
	return serviceIDs
}

func (network *ServiceNetworkImpl) CopyFromService(ctx context.Context, serviceId service.ServiceID, srcPath string) (
	io.ReadCloser,
	error,
) {
	runInfo, foundRunInfo := network.serviceRunInfo[serviceId]
	if !foundRunInfo {
		return nil, stacktrace.NewError("No run information found for service with ID '%v'", serviceId)

	}
	serviceGuid := runInfo.serviceGUID

	readCloser, err := network.kurtosisBackend.CopyFromUserService(ctx, network.enclaveId, serviceGuid, srcPath)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred copying source '%v' from user service with GUID '%v' in enclave with ID '%v'", srcPath, serviceGuid, network.enclaveId)
	}

	return readCloser, nil
}

// ====================================================================================================
// 									   Private helper methods
// ====================================================================================================
func (network *ServiceNetworkImpl) removeServiceWithoutMutex(
	ctx context.Context,
	serviceId service.ServiceID,
	containerStopTimeout time.Duration) error {

	registrationInfo, foundRegistrationInfo := network.serviceRegistrationInfo[serviceId]
	if !foundRegistrationInfo {
		return stacktrace.NewError("No registration info found for service '%v'", serviceId)
	}
	network.topology.RemoveService(serviceId)
	delete(network.serviceRegistrationInfo, serviceId)

	// TODO PERF: Parallelize the shutdown of the service container and the sidecar container
	runInfo, foundRunInfo := network.serviceRunInfo[serviceId]
	if foundRunInfo {
		serviceGUID := runInfo.serviceGUID
		// Make a best-effort attempt to stop the service container
		logrus.Debugf("Stopping service with GUID '%v' for service ID '%v'...", serviceGUID, serviceId)
		_, failedToStopServiceErrs, err := network.kurtosisBackend.StopUserServices(
			ctx,
			getServiceByServiceGUIDFilter(serviceGUID),
		)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred calling the backend to stop service with GUID '%v'", serviceGUID)
		}
		if len(failedToStopServiceErrs) > 0 {
			serviceStopErrs := []string{}
			for serviceGUID, err := range failedToStopServiceErrs {
				wrappedErr := stacktrace.Propagate(
					err,
					"An error occurred stopping service with GUID `%v'",
					serviceGUID,
				)
				serviceStopErrs = append(serviceStopErrs, wrappedErr.Error())
			}
			return stacktrace.NewError(
				"One or more errors occurred stopping the service(s): \n%v",
				strings.Join(
					serviceStopErrs,
					"\n\n",
				),
			)
		}
		delete(network.serviceRunInfo, serviceId)
		logrus.Debugf("Successfully stopped service GUID '%v'", serviceGUID)
	}
	network.freeIpAddrTracker.ReleaseIpAddr(registrationInfo.privateIpAddr)

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

/*
Updates the traffic control configuration of the services with the given IDs to match the target services packet loss configuration

NOTE: This is not thread-safe, so it must be within a function that locks mutex!
*/
func updateTrafficControlConfiguration(
	ctx context.Context,
	targetServicePacketLossConfigs map[service.ServiceID]map[service.ServiceID]float32,
	serviceRegistrationInfo map[service.ServiceID]serviceRegistrationInfo,
	networkingSidecars map[service.ServiceID]networking_sidecar.NetworkingSidecarWrapper,
) error {

	// TODO PERF: Run the container updates in parallel, with the container being modified being the most important

	for serviceId, allOtherServicesPacketLossConfigurations := range targetServicePacketLossConfigs {
		allPacketLossPercentageForIpAddresses := map[string]float32{}
		for otherServiceId, otherServicePacketLossPercentage := range allOtherServicesPacketLossConfigurations {

			infoForService, found := serviceRegistrationInfo[otherServiceId]
			if !found {
				return stacktrace.NewError(
					"Service with ID '%v' needs to add packet loss configuration for service with ID '%v', but the latter "+
						"doesn't have service registration info (i.e. an IP) associated with it",
					serviceId,
					otherServiceId)
			}

			allPacketLossPercentageForIpAddresses[infoForService.privateIpAddr.String()] = otherServicePacketLossPercentage
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
