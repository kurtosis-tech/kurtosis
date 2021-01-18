/* * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package service_network

import (
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/api_container/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis/api_container/service_network/networking_sidecar"
	"github.com/kurtosis-tech/kurtosis/api_container/service_network/topology_types"
	"github.com/kurtosis-tech/kurtosis/api_container/service_network/user_service_launcher"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	defaultPartitionId                   topology_types.PartitionID = "default"
	startingDefaultConnectionBlockStatus                            = false
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

	userServiceLauncher *user_service_launcher.UserServiceLauncher

	topology *partition_topology.PartitionTopology

	// These are separate maps, rather than being bundled into a single containerInfo-valued map, because
	//  they're registered at different times (rather than in one atomic operation)
	serviceIps map[topology_types.ServiceID]net.IP
	serviceContainerIds map[topology_types.ServiceID]string

	networkingSidecars map[topology_types.ServiceID]networking_sidecar.NetworkingSidecar

	networkingSidecarManager networking_sidecar.NetworkingSidecarManager
}

func NewServiceNetwork(
		isPartitioningEnabled bool,
		freeIpAddrTracker *commons.FreeIpAddrTracker,
		dockerManager *docker_manager.DockerManager,
		userServiceLauncher *user_service_launcher.UserServiceLauncher,
		networkingSidecarManager networking_sidecar.NetworkingSidecarManager) *ServiceNetwork {
	defaultPartitionConnection := partition_topology.PartitionConnection{IsBlocked: startingDefaultConnectionBlockStatus}
	return &ServiceNetwork{
		isDestroyed: false,
		isPartitioningEnabled: isPartitioningEnabled,
		freeIpAddrTracker: freeIpAddrTracker,
		dockerManager: dockerManager,
		userServiceLauncher: userServiceLauncher,
		mutex:               &sync.Mutex{},
		topology:            partition_topology.NewPartitionTopology(
			defaultPartitionId,
			defaultPartitionConnection,
		),
		serviceIps:               map[topology_types.ServiceID]net.IP{},
		serviceContainerIds:      map[topology_types.ServiceID]string{},
		networkingSidecars:       map[topology_types.ServiceID]networking_sidecar.NetworkingSidecar{},
		networkingSidecarManager: networkingSidecarManager,
	}
}

/*
Completely repartitions the network, throwing away the old topology
 */
func (network *ServiceNetwork) Repartition(
		ctx context.Context,
		newPartitionServices map[topology_types.PartitionID]*topology_types.ServiceIDSet,
		newPartitionConnections map[topology_types.PartitionConnectionID]partition_topology.PartitionConnection,
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


// TODO Add tests for this
/*
Creates the service with the given ID in the given partition

If partitionId is empty string, the default partition ID is used

Returns: The IP address of the new service
 */
func (network *ServiceNetwork) AddServiceInPartition(
		ctx context.Context,
		serviceId topology_types.ServiceID,
		imageName string,
		usedPorts map[nat.Port]bool,
		partitionId topology_types.PartitionID,
		ipPlaceholder string,
		startCmd []string,
		dockerEnvVars map[string]string,
		testVolumeMountDirpath string,
		filesArtifactMountDirpaths map[string]string) (net.IP, error) {
	// TODO extract this into a wrapper function that can be wrapped around every service call (so we don't forget)
	network.mutex.Lock()
	defer network.mutex.Unlock()
	if network.isDestroyed {
		return nil, stacktrace.NewError("Cannot add service; the service network has been destroyed")
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

	serviceIp, err := network.freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred when getting an IP to give the container running the new service with Docker image '%v'",
			imageName)
	}
	logrus.Debugf("Giving new service the following IP: %v", serviceIp.String())

	// When partitioning is enabled, there's a race condition where:
	//   a) we need to start the service before we can launch the sidecar but
	//   b) we can't modify the iptables until the sidecar container is launched.
	// This means that there's a period of time at startup where the container might not be partitioned. We solve
	//  this by blocking the new service's IP in the already-existing services' iptables BEFORE we start the service.
	// This means that when the new service is launched, even if its own iptables aren't yet updated, all the services
	//  it would communicate are already dropping traffic from it.
	// The unfortunate result is that we need to add the service to the PartitionTopology before the container or
	//  its sidecar is even started, which means we need to roll back if an error occurred during startup.
	if err := network.topology.AddService(serviceId, partitionId); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding service '%v' to partition '%v' in the topology", serviceId, partitionId)
	}
	network.serviceIps[serviceId] = serviceIp
	shouldLeaveServiceInTopology := false
	defer func() {
		if !shouldLeaveServiceInTopology {
			network.topology.RemoveService(serviceId)
			network.freeIpAddrTracker.ReleaseIpAddr(serviceIp)
			delete(network.serviceIps, serviceId)

			// NOTE: As of 2020-12-31, we don't actually have to undo the iptables modifications that we'll make to all
			//  the other services, because their iptables will be completely overwritten by the next
			//  Add/Repartition event. If we ever make iptables updates incremental though, we WILL need to
			//  undo the iptables we added here!
		}
	}()

	// Tell all the other services to ignore the soon-to-be-started node, so that when it starts
	//  they absolutely won't communicate with it
	if network.isPartitioningEnabled {
		blocklists, err := network.topology.GetBlocklists()
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting the blocklists for updating iptables before the " +
				"node was added, which means we can't add the node because we can't partition it away properly")
		}
		blocklistsWithoutNewNode := map[topology_types.ServiceID]*topology_types.ServiceIDSet{}
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

	serviceContainerId, err := network.userServiceLauncher.Launch(
		ctx,
		serviceId,
		serviceIp,
		imageName,
		usedPorts,
		ipPlaceholder,
		startCmd,
		dockerEnvVars,
		testVolumeMountDirpath,
		filesArtifactMountDirpaths)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the user service")
	}
	shouldLeaveServiceInTopology = true
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
		updatesToApply := map[topology_types.ServiceID]*topology_types.ServiceIDSet{
			serviceId: newNodeBlocklist,
		}
		if err := updateIpTables(ctx, updatesToApply, network.serviceIps, network.networkingSidecars); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred applying the iptables on the new node to partition it " +
				"off from other nodes")
		}
	}

	return serviceIp, nil
}

func (network *ServiceNetwork) RemoveService(
		ctx context.Context,
		serviceId topology_types.ServiceID,
		containerStopTimeout time.Duration) error {
	network.mutex.Lock()
	defer network.mutex.Unlock()
	if network.isDestroyed {
		return stacktrace.NewError("Cannot remove service; the service network has been destroyed")
	}

	serviceContainerId, found := network.serviceContainerIds[serviceId]
	if !found {
		return stacktrace.NewError("Unknown service '%v'", serviceId)
	}
	serviceIp, found := network.serviceIps[serviceId]
	if !found {
		return stacktrace.NewError("No IP found for service '%v'", serviceId)
	}

	// TODO PERF: Parallelize the shutdown of the service container and the sidecar container
	// Make a best-effort attempt to stop the service container
	logrus.Debugf("Removing service ID '%v' with container ID '%v'...", serviceId, serviceContainerId)
	if err := network.dockerManager.StopContainer(ctx, serviceContainerId, containerStopTimeout); err != nil {
		return stacktrace.Propagate(err, "An error occurred stopping the container with ID %v", serviceContainerId)
	}
	network.topology.RemoveService(serviceId)
	network.freeIpAddrTracker.ReleaseIpAddr(serviceIp)
	delete(network.serviceContainerIds, serviceId)
	delete(network.serviceIps, serviceId)
	logrus.Debugf("Successfully removed service with container ID %v", serviceContainerId)

	if network.isPartitioningEnabled {
		// NOTE: As of 2020-12-31, we don't need to update the iptables of the other services in the network to
		//  clear the now-removed service's IP because:
		// 	 a) nothing is using it so it doesn't do anything and
		//	 b) all service's iptables get overwritten on the next Add/Repartition call
		// If we ever do incremental iptables though, we'll need to fix all the other service's iptables here!

		sidecar, found := network.networkingSidecars[serviceId]
		if !found {
			return stacktrace.NewError(
				"Couldn't find sidecar container for service '%v'; this is a code bug where the sidecar container ID didn't get stored at creation time",
				serviceId)

		}

		if err := network.networkingSidecarManager.Destroy(ctx, sidecar); err != nil {
			return stacktrace.Propagate(err, "An error occurred destroying the sidecar for service with ID '%v'", serviceId)
		}
		delete(network.networkingSidecars, serviceId)
		logrus.Debugf("Successfully removed sidecar attached to service with ID '%v'", serviceId)
	}
	return nil
}

// Stops all services that have been created by the API container, and renders the service network unusable
func (network *ServiceNetwork) Destroy(ctx context.Context, containerStopTimeout time.Duration) error {
	network.mutex.Lock()
	defer network.mutex.Unlock()
	if network.isDestroyed {
		return stacktrace.NewError("Cannot destroy the service network; it has already been destroyed")
	}

	containerStopErrors := []error{}

	// TODO PERF: parallelize this for faster shutdown
	logrus.Debugf("Making best-effort attempt to stop sidecar containers...")
	for serviceId, sidecar := range network.networkingSidecars {
		if err := network.networkingSidecarManager.Destroy(ctx, sidecar); err != nil {
			wrappedErr := stacktrace.Propagate(
				err,
				"An error occurred destroying sidecar container for service with ID '%v'",
				serviceId)
			containerStopErrors = append(containerStopErrors, wrappedErr)
		}
	}
	logrus.Debugf("Made best-effort attempt to stop sidecar containers")

	// TODO PERF: parallelize this for faster shutdown
	logrus.Debugf("Making best-effort attempt to stop service containers...")
	for serviceId, serviceContainerId := range network.serviceContainerIds {
		// TODO set the stop timeout on the service itself
		if err := network.dockerManager.StopContainer(ctx, serviceContainerId, containerStopTimeout); err != nil {
			wrappedErr := stacktrace.Propagate(
				err,
				"An error occurred stopping container for service '%v' with container ID '%v'",
				serviceId,
				serviceContainerId)
			containerStopErrors = append(containerStopErrors, wrappedErr)
		}
	}
	logrus.Debugf("Made best-effort attempt to stop service containers")

	network.isDestroyed = true

	if len(containerStopErrors) > 0 {
		errorStrs := []string{}
		for _, err := range containerStopErrors {
			errStr := err.Error()
			errorStrs = append(errorStrs, errStr)
		}
		joinedErrStrings := strings.Join(errorStrs, "\n")
		return stacktrace.NewError(
			"One or more error(s) occurred stopping the services in the test network " +
				"during service network destruction:\n%s",
			joinedErrStrings)
	}

	return nil
}

// ====================================================================================================
// 									   Private helper methods
// ====================================================================================================
/*
Updates the iptables of the services with the given IDs to match the target blocklists

NOTE: This is not thread-safe, so it must be within a function that locks mutex!
 */
func updateIpTables(
		ctx context.Context,
		targetBlocklists map[topology_types.ServiceID]*topology_types.ServiceIDSet,
		serviceIps map[topology_types.ServiceID]net.IP,
		networkingSidecars map[topology_types.ServiceID]networking_sidecar.NetworkingSidecar) error {
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

