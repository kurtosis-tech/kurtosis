/* * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package service_network

import (
	"bytes"
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/api_container/rpc_api/bindings"
	networking_sidecar2 "github.com/kurtosis-tech/kurtosis/api_container/service_network/networking_sidecar"
	partition_topology2 "github.com/kurtosis-tech/kurtosis/api_container/service_network/partition_topology"
	service_network_types2 "github.com/kurtosis-tech/kurtosis/api_container/service_network/service_network_types"
	user_service_launcher2 "github.com/kurtosis-tech/kurtosis/api_container/service_network/user_service_launcher"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/kurtosis-tech/kurtosis/commons/suite_execution_volume"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	defaultPartitionId                   service_network_types2.PartitionID = "default"
	startingDefaultConnectionBlockStatus                                    = false
)

/*
This is the in-memory representation of the service network that the API container will manipulate. To make
	any changes to the test network, this struct must be used.
 */
type ServiceNetwork struct {
	// When the network is destroyed, all requests will fail
	// This ensures that when the initializer tells the API container to destroy everything, the still-running
	//  testsuite can't create more work
	isDestroyed bool   	// VERY IMPORTANT TO CHECK AT THE START OF EVERY METHOD!

	mutex *sync.Mutex	// VERY IMPORTANT TO CHECK AT THE START OF EVERY METHOD!

	// Whether partitioning has been enabled for this particular test
	isPartitioningEnabled bool

	freeIpAddrTracker *commons.FreeIpAddrTracker

	dockerManager *docker_manager.DockerManager

	testExecutionDirectory *suite_execution_volume.TestExecutionDirectory

	userServiceLauncher *user_service_launcher2.UserServiceLauncher

	topology *partition_topology2.PartitionTopology

	// These are separate maps, rather than being bundled into a single containerInfo-valued map, because
	//  they're registered at different times (rather than in one atomic operation)
	serviceIps map[service_network_types2.ServiceID]net.IP
	serviceContainerIds map[service_network_types2.ServiceID]string

	networkingSidecars map[service_network_types2.ServiceID]networking_sidecar2.NetworkingSidecar

	networkingSidecarManager networking_sidecar2.NetworkingSidecarManager
}

func NewServiceNetwork(
		isPartitioningEnabled bool,
		freeIpAddrTracker *commons.FreeIpAddrTracker,
		dockerManager *docker_manager.DockerManager,
		testExecutionDirectory *suite_execution_volume.TestExecutionDirectory,
		userServiceLauncher *user_service_launcher2.UserServiceLauncher,
		networkingSidecarManager networking_sidecar2.NetworkingSidecarManager) *ServiceNetwork {
	defaultPartitionConnection := partition_topology2.PartitionConnection{IsBlocked: startingDefaultConnectionBlockStatus}
	return &ServiceNetwork{
		isDestroyed: false,
		isPartitioningEnabled: isPartitioningEnabled,
		freeIpAddrTracker: freeIpAddrTracker,
		dockerManager: dockerManager,
		testExecutionDirectory: testExecutionDirectory,
		userServiceLauncher: userServiceLauncher,
		mutex:               &sync.Mutex{},
		topology:            partition_topology2.NewPartitionTopology(
			defaultPartitionId,
			defaultPartitionConnection,
		),
		serviceIps:               map[service_network_types2.ServiceID]net.IP{},
		serviceContainerIds:      map[service_network_types2.ServiceID]string{},
		networkingSidecars:       map[service_network_types2.ServiceID]networking_sidecar2.NetworkingSidecar{},
		networkingSidecarManager: networkingSidecarManager,
	}
}

/*
Completely repartitions the network, throwing away the old topology
 */
func (network *ServiceNetwork) Repartition(
		ctx context.Context,
		newPartitionServices map[service_network_types2.PartitionID]*service_network_types2.ServiceIDSet,
		newPartitionConnections map[service_network_types2.PartitionConnectionID]partition_topology2.PartitionConnection,
		newDefaultConnection partition_topology2.PartitionConnection) error {
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
	blocklists, err := network.topology.GetBlocklists()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the blocklists after repartition, meaning that " +
			"no partitions are actually being enforced!")
	}
	if err := updateIpTables(ctx, blocklists, network.serviceIps, network.networkingSidecars); err != nil {
		return stacktrace.Propagate(err, "An error occurred updating the IP tables to match the target blocklist after repartitioning")
	}
	return nil
}

// Registers a service for use with the network (creating the IPs and so forth), but doesn't start it
// If the partition ID is empty, registers the service with the default partition
func (network ServiceNetwork) RegisterService(
		serviceId service_network_types2.ServiceID,
		partitionId service_network_types2.PartitionID) (net.IP, error) {
	// TODO extract this into a wrapper function that can be wrapped around every service call (so we don't forget)
	network.mutex.Lock()
	defer network.mutex.Unlock()
	if network.isDestroyed {
		return nil, stacktrace.NewError("Cannot register service with ID '%v'; the service network has been destroyed", serviceId)
	}

	if strings.TrimSpace(string(serviceId)) == "" {
		return nil, stacktrace.NewError("Service ID cannot be empty or whitespace")
	}

	_, found := network.serviceIps[serviceId]
	if found {
		return nil, stacktrace.NewError("Cannot register service with ID '%v'; a service with that ID already exists", serviceId)
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

	ip, err := network.freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting an IP for service with ID '%v'", serviceId)
	}
	logrus.Debugf("Giving service '%v' IP '%v'", serviceId, ip.String())
	network.serviceIps[serviceId] = ip
	shouldUndoServiceIpAdd := true
	// To keep our bookkeeping correct, if an error occurs later we need to back out the IP-adding that we do now
	defer func() {
		if shouldUndoServiceIpAdd {
			delete(network.serviceIps, serviceId)
			network.freeIpAddrTracker.ReleaseIpAddr(ip)
		}
	}()

	if err := network.topology.AddService(serviceId, partitionId); err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred adding service with ID '%v' to partition '%v' in the topology",
			serviceId,
			partitionId)
	}
	shouldUndoServiceIpAdd = false

	return ip, nil
}

// Generates files in a location in the suite execution volume allocated to the given service
func (network *ServiceNetwork) GenerateFiles(
		serviceId service_network_types2.ServiceID,
		filesToGenerate map[string]*bindings.FileGenerationOptions) (map[string]string, error) {
	// TODO extract this into a wrapper function that can be wrapped around every service call (so we don't forget)
	network.mutex.Lock()
	defer network.mutex.Unlock()
	if network.isDestroyed {
		return nil, stacktrace.NewError("Cannot register service with ID '%v'; the service network has been destroyed", serviceId)
	}

	serviceDirectory, err := network.testExecutionDirectory.GetServiceDirectory(string(serviceId))
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the service directory inside the test execution volume for service with ID '%v'", serviceId)
	}

	generatedFilesRelativeFilepaths := map[string]string{}
	for userCreatedFileKey, fileGenerationOptions := range filesToGenerate {
		fileTypeToGenerate := fileGenerationOptions.GetFileTypeToGenerate()
		switch fileTypeToGenerate {
		case bindings.FileGenerationOptions_FILE:
			file, err := serviceDirectory.GetFile(userCreatedFileKey)
			if err != nil {
				return nil, stacktrace.Propagate(err, "An error occurred creating file '%v' for service with ID '%v'", userCreatedFileKey, serviceId)
			}
			generatedFilesRelativeFilepaths[userCreatedFileKey] = file.GetFilepathRelativeToVolRoot()
			logrus.Debugf("Created generated file '%v' at '%v'", userCreatedFileKey, file.GetFilepathRelativeToVolRoot())
		default:
			return nil, stacktrace.NewError(
				"Could not generate file '%v'; unrecognized file type '%v'; this is a bug in Kurtosis",
				userCreatedFileKey,
				fileTypeToGenerate,
			)
		}
	}
	return generatedFilesRelativeFilepaths, nil
}

// TODO add tests for this
/*
Starts a previously-registered but not-started service by creating it in a container

Returns:
	Mapping of port-used-by-service -> port-on-the-Docker-host-machine where the user can make requests to the port
		to access the port. This will be empty if no ports are bound.
 */
func (network *ServiceNetwork) StartService(
		ctx context.Context,
		serviceId service_network_types2.ServiceID,
		imageName string,
		usedPorts map[nat.Port]bool,
		entrypointArgs []string,
		cmdArgs []string,
		dockerEnvVars map[string]string,
		suiteExecutionVolMntDirpath string,
		filesArtifactMountDirpaths map[string]string) (map[nat.Port]*nat.PortBinding, error) {
	// TODO extract this into a wrapper function that can be wrapped around every service call (so we don't forget)
	network.mutex.Lock()
	defer network.mutex.Unlock()
	if network.isDestroyed {
		return nil, stacktrace.NewError("Cannot start container for service with ID '%v'; the service network has been destroyed", serviceId)
	}

	serviceIpAddr, foundIp := network.serviceIps[serviceId]
	if !foundIp {
		return nil, stacktrace.NewError("Cannot start container for service with ID '%v'; no service with that ID has been registered", serviceId)
	}
	if _, found := network.serviceContainerIds[serviceId]; found {
		return nil, stacktrace.NewError("Cannot start container for service with ID '%v'; that service ID already has a container associated with it", serviceId)
	}

	// When partitioning is enabled, there's a race condition where:
	//   a) we need to start the service before we can launch the sidecar but
	//   b) we can't modify the iptables until the sidecar container is launched.
	// This means that there's a period of time at startup where the container might not be partitioned. We solve
	//  this by blocking the new service's IP in the already-existing services' iptables BEFORE we start the service.
	// This means that when the new service is launched, even if its own iptables aren't yet updated, all the services
	//  it would communicate are already dropping traffic from it.
	if network.isPartitioningEnabled {
		blocklists, err := network.topology.GetBlocklists()
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting the blocklists for updating iptables before the " +
				"node was added, which means we can't add the node because we can't partition it away properly")
		}
		blocklistsWithoutNewNode := map[service_network_types2.ServiceID]*service_network_types2.ServiceIDSet{}
		for serviceInTopologyId, servicesToBlock := range blocklists {
			if serviceId == serviceInTopologyId {
				continue
			}
			blocklistsWithoutNewNode[serviceInTopologyId] = servicesToBlock
		}
		if err := updateIpTables(ctx, blocklistsWithoutNewNode, network.serviceIps, network.networkingSidecars); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred updating the iptables of all the other services " +
				"before adding the node, meaning that the node wouldn't actually start in a partition")
		}
	}

	serviceContainerId, hostPortBindings, err := network.userServiceLauncher.Launch(
		ctx,
		serviceId,
		serviceIpAddr,
		imageName,
		usedPorts,
		entrypointArgs,
		cmdArgs,
		dockerEnvVars,
		suiteExecutionVolMntDirpath,
		filesArtifactMountDirpaths)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the user service container")
	}
	network.serviceContainerIds[serviceId] = serviceContainerId

	if network.isPartitioningEnabled {
		sidecar, err := network.networkingSidecarManager.Create(ctx, serviceId, serviceContainerId)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating the networking sidecar container")
		}
		network.networkingSidecars[serviceId] = sidecar

		if err := sidecar.InitializeIpTables(ctx); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred initializing the newly-created sidecar container iptables")
		}

		// TODO Getting blocklists is an expensive call and, as of 2020-12-31, we do it twice - the solution is to make
		//  getting the blocklists not an expensive call (see also https://github.com/kurtosis-tech/kurtosis-core/issues/123 )
		blocklists, err := network.topology.GetBlocklists()
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting the blocklists to know what iptables " +
				"updates to apply on the new node")
		}
		newNodeBlocklist := blocklists[serviceId]
		updatesToApply := map[service_network_types2.ServiceID]*service_network_types2.ServiceIDSet{
			serviceId: newNodeBlocklist,
		}
		if err := updateIpTables(ctx, updatesToApply, network.serviceIps, network.networkingSidecars); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred applying the iptables on the new node to partition it " +
				"off from other nodes")
		}
	}

	return hostPortBindings, nil
}

func (network *ServiceNetwork) RemoveService(
		ctx context.Context,
		serviceId service_network_types2.ServiceID,
		containerStopTimeout time.Duration) error {
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

func (network *ServiceNetwork) ExecCommand(
		ctx context.Context,
		serviceId service_network_types2.ServiceID,
		command []string) (int32, *bytes.Buffer, error) {
	// NOTE: This will block all other operations while this command is running!!!! We might need to change this so it's
	// asynchronous
	network.mutex.Lock()
	defer network.mutex.Unlock()
	if network.isDestroyed {
		return 0, nil, stacktrace.NewError("Cannot run exec command; the service network has been destroyed")
	}

	containerId, found := network.serviceContainerIds[serviceId]
	if !found {
		return 0, nil, stacktrace.NewError(
			"Could not run exec command '%v' against service '%v'; no container has been created for the service yet",
			command,
			serviceId)
	}

	// NOTE: This is a SYNCHRONOUS command, meaning that the entire network will be blocked until the command finishes
	// In the future, this will likely be insufficient

	execOutputBuf := &bytes.Buffer{}
	exitCode, err := network.dockerManager.RunExecCommand(ctx, containerId, command, execOutputBuf)
	if err != nil {
		return 0, nil, stacktrace.Propagate(
			err,
			"An error occurred running exec command '%v' against service '%v'",
			command,
			serviceId)
	}
	return exitCode, execOutputBuf, nil
}

// ====================================================================================================
// 									   Private helper methods
// ====================================================================================================
func (network *ServiceNetwork) removeServiceWithoutMutex(
		ctx context.Context,
		serviceId service_network_types2.ServiceID,
		containerStopTimeout time.Duration) error {
	serviceIp, foundIpAddr := network.serviceIps[serviceId]
	if !foundIpAddr {
		return stacktrace.NewError("No IP found for service '%v'", serviceId)
	}
	network.topology.RemoveService(serviceId)
	delete(network.serviceIps, serviceId)

	// TODO PERF: Parallelize the shutdown of the service container and the sidecar container
	serviceContainerId, foundContainerId := network.serviceContainerIds[serviceId]
	if foundContainerId {
		// Make a best-effort attempt to stop the service container
		logrus.Debugf("Stopping container ID '%v' for service ID '%v'...", serviceContainerId, serviceId)
		if err := network.dockerManager.StopContainer(ctx, serviceContainerId, containerStopTimeout); err != nil {
			return stacktrace.Propagate(err, "An error occurred stopping the container with ID %v", serviceContainerId)
		}
		delete(network.serviceContainerIds, serviceId)
		logrus.Debugf("Successfully stopped container ID")
	}
	network.freeIpAddrTracker.ReleaseIpAddr(serviceIp)

	sidecar, foundSidecar := network.networkingSidecars[serviceId]
	if network.isPartitioningEnabled && foundSidecar {
		// NOTE: As of 2020-12-31, we don't need to update the iptables of the other services in the network to
		//  clear the now-removed service's IP because:
		// 	 a) nothing is using it so it doesn't do anything and
		//	 b) all service's iptables get overwritten on the next Add/Repartition call
		// If we ever do incremental iptables though, we'll need to fix all the other service's iptables here!
		if err := network.networkingSidecarManager.Destroy(ctx, sidecar); err != nil {
			return stacktrace.Propagate(err, "An error occurred destroying the sidecar for service with ID '%v'", serviceId)
		}
		delete(network.networkingSidecars, serviceId)
		logrus.Debugf("Successfully removed sidecar attached to service with ID '%v'", serviceId)
	}

	return nil
}

/*
Updates the iptables of the services with the given IDs to match the target blocklists

NOTE: This is not thread-safe, so it must be within a function that locks mutex!
 */
func updateIpTables(
		ctx context.Context,
		targetBlocklists map[service_network_types2.ServiceID]*service_network_types2.ServiceIDSet,
		serviceIps map[service_network_types2.ServiceID]net.IP,
		networkingSidecars map[service_network_types2.ServiceID]networking_sidecar2.NetworkingSidecar) error {
	// TODO PERF: Run the container updates in parallel, with the container being modified being the most important
	for serviceId, newBlocklist := range targetBlocklists {
		allIpsToBlock := []net.IP{}
		for _, serviceIdToBlock := range newBlocklist.Elems() {
			ipToBlock, found := serviceIps[serviceIdToBlock]
			if !found {
				return stacktrace.NewError(
					"Service with ID '%v' needs to block service with ID '%v', but the latter " +
						"doesn't have an IP associated with it",
					serviceId,
					serviceIdToBlock)
			}
			allIpsToBlock = append(allIpsToBlock, ipToBlock)
		}

		sidecar, found := networkingSidecars[serviceId]
		if !found {
			return stacktrace.NewError(
				"Need to update iptables of service with ID '%v', but the service doesn't have a sidecar",
				serviceId)
		}
		if err := sidecar.UpdateIpTables(ctx, allIpsToBlock); err != nil {
			return stacktrace.Propagate(
				err,
				"An error occurred updating the iptables for service '%v'",
				serviceId)
		}
	}
	return nil
}


