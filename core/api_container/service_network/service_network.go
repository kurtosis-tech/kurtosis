/* * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package service_network

import (
	"bytes"
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/api_container/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis/api_container/service_network/topology_types"
	"github.com/kurtosis-tech/kurtosis/api_container/service_network/user_service_launcher"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultPartitionId                   topology_types.PartitionID = "default"
	startingDefaultConnectionBlockStatus                            = false

	iproute2ContainerImage = "kurtosistech/iproute2"

	// We create two chains so that during modifications we can flush and rebuild one
	//  while the other one is live
	kurtosisIpTablesChain1 ipTablesChain = "KURTOSIS1"
	kurtosisIpTablesChain2 ipTablesChain = "KURTOSIS2"
	initialKurtosisIpTablesChain = kurtosisIpTablesChain1 // The Kurtosois chain that will be in use on service launch

	ipTablesCommand = "iptables"
	ipTablesInputChain = "INPUT"
	ipTablesOutputChain = "OUTPUT"
	ipTablesNewChainFlag = "-N"
	ipTablesInsertRuleFlag = "-I"
	ipTablesFlushChainFlag = "-F"
	ipTablesAppendRuleFlag  = "-A"
	ipTablesReplaceRuleFlag = "-R"
	ipTablesDropAction = "DROP"
	ipTablesFirstRuleIndex = 1	// iptables chains are 1-indexed
)

type ipTablesChain string

// We sleep forever because all the commands this container will run will be executed
//  via Docker exec
var ipRouteContainerCommand = []string{
	"sleep","infinity",
}

type containerInfo struct {
	containerId string
	ipAddr net.IP
}

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

	dockerNetworkId string

	freeIpAddrTracker *commons.FreeIpAddrTracker

	dockerManager *docker_manager.DockerManager

	userServiceLauncher *user_service_launcher.UserServiceLauncher

	topology *partition_topology.PartitionTopology

	// These are separate maps, rather than being bundled into a single containerInfo-valued map, because
	//  they're registered at different times (rather than in one atomic operation)
	serviceIps map[topology_types.ServiceID]net.IP
	serviceContainerIds map[topology_types.ServiceID]string

	// Tracks which Kurtosis chain is the primary chain, so we know
	//  which chain is in the background that we can flush and rebuild
	//  when we're changing iptables
	serviceIpTablesChainInUse map[topology_types.ServiceID]ipTablesChain

	sidecarContainerInfo map[topology_types.ServiceID]containerInfo
}

func NewServiceNetwork(
		isPartitioningEnabled bool,
		dockerNetworkId string,
		freeIpAddrTracker *commons.FreeIpAddrTracker,
		dockerManager *docker_manager.DockerManager,
		userServiceLauncher *user_service_launcher.UserServiceLauncher) *ServiceNetwork {
	defaultPartitionConnection := partition_topology.PartitionConnection{IsBlocked: startingDefaultConnectionBlockStatus}
	return &ServiceNetwork{
		isDestroyed: false,
		isPartitioningEnabled: isPartitioningEnabled,
		dockerNetworkId: dockerNetworkId,
		freeIpAddrTracker: freeIpAddrTracker,
		dockerManager: dockerManager,
		userServiceLauncher: userServiceLauncher,
		mutex:               &sync.Mutex{},
		topology:            partition_topology.NewPartitionTopology(
			defaultPartitionId,
			defaultPartitionConnection,
		),
		serviceIps: map[topology_types.ServiceID]net.IP{},
		serviceContainerIds: map[topology_types.ServiceID]string{},
		serviceIpTablesChainInUse: map[topology_types.ServiceID]ipTablesChain{},
		sidecarContainerInfo: map[topology_types.ServiceID]containerInfo{},
	}
}

/*
Completely repartitions the network, throwing away the old topology
 */
func (network *ServiceNetwork) Repartition(
		context context.Context,
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
	if err := network.updateIpTables(context, blocklists); err != nil {
		return stacktrace.Propagate(err, "An error occurred updating the IP tables to match the target blocklist after repartitioning")
	}
	return nil
}


// TODO Refactor this into smaller chunks - it's currently a very complex and tricky method
// TODO Add tests for this
/*
Creates the service with the given ID in the given partition

If partitionId is empty string, the default partition ID is used

Returns: The IP address of the new service
 */
func (network *ServiceNetwork) AddServiceInPartition(
		context context.Context,
		serviceId topology_types.ServiceID,
		imageName string,
		usedPorts map[nat.Port]bool,
		partitionId topology_types.PartitionID,
		ipPlaceholder string,
		startCmd []string,
		dockerEnvVars map[string]string,
		testVolumeMountDirpath string) (net.IP, error) {
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
			delete(network.serviceIps, serviceId)

			// NOTE: As of 2020-12-31, we don't actually have to undo the iptables modifications that we made to all
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
		if err := network.updateIpTables(context, blocklistsWithoutNewNode); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred updating the iptables of all the other services " +
				"before adding the node, meaning that the node wouldn't actually start in a partition")
		}
	}

	serviceContainerId, err := network.userServiceLauncher.Launch(
		context,
		serviceIp,
		imageName,
		usedPorts,
		ipPlaceholder,
		startCmd,
		dockerEnvVars,
		testVolumeMountDirpath)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the user service")
	}
	shouldLeaveServiceInTopology = true
	network.serviceContainerIds[serviceId] = serviceContainerId

	if network.isPartitioningEnabled {
		sidecarIp, err := network.freeIpAddrTracker.GetFreeIpAddr()
		if err != nil {
			return nil, stacktrace.Propagate(
				err,
				"An error occurred getting a free IP address for the networking sidecar container")
		}
		sidecarContainerId, err := network.dockerManager.CreateAndStartContainer(
			context,
			iproute2ContainerImage,
			network.dockerNetworkId,
			sidecarIp,
			map[docker_manager.ContainerCapability]bool{
				docker_manager.NetAdmin: true,
			},
			docker_manager.NewContainerNetworkMode(serviceContainerId),
			map[nat.Port]*nat.PortBinding{},
			ipRouteContainerCommand,
			map[string]string{}, // No environment variables
			map[string]string{}, // no bind mounts for services created via the Kurtosis API
			map[string]string{}, // No volume mounts either
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred starting the sidecar iproute container for modifying the service container's iptables")
		}
		network.sidecarContainerInfo[serviceId] = containerInfo{
			containerId: sidecarContainerId,
			ipAddr:      sidecarIp,
		}

		// As soon as we have the sidecar, we need to create the Kurtosis chain and insert it in first position
		//  on both the INPUT *and* the OUTPUT chains
		configureKurtosisChainsCommand := []string{
			ipTablesCommand,
			ipTablesNewChainFlag,
			string(kurtosisIpTablesChain1),
			"&&",
			ipTablesCommand,
			ipTablesNewChainFlag,
			string(kurtosisIpTablesChain2),
		}
		for _, chain := range []string{ipTablesInputChain, ipTablesOutputChain} {
			addKurtosisChainInFirstPositionCommand := []string{
				ipTablesCommand,
				ipTablesInsertRuleFlag,
				chain,
				strconv.Itoa(ipTablesFirstRuleIndex),
				"-j",
				string(initialKurtosisIpTablesChain),
			}
			configureKurtosisChainsCommand = append(configureKurtosisChainsCommand, "&&")
			configureKurtosisChainsCommand = append(
				configureKurtosisChainsCommand,
				addKurtosisChainInFirstPositionCommand...)
		}

		// We need to wrap this command with 'sh -c' because we're using '&&', and if we don't do this then
		//  iptables will think the '&&' is an argument for it and fail
		configureKurtosisChainShWrappedCommand := []string{
			"sh",
			"-c",
			strings.Join(configureKurtosisChainsCommand, " "),
		}

		logrus.Debugf("Running exec command to configure Kurtosis iptables chain: '%v'", configureKurtosisChainShWrappedCommand)
		execOutputBuf := &bytes.Buffer{}
		if err := network.dockerManager.RunExecCommand(
				context,
				sidecarContainerId,
			configureKurtosisChainShWrappedCommand,
				execOutputBuf); err !=  nil {
			logrus.Error("------------------ Kurtosis iptables chain-configuring exec command output --------------------")
			if _, err := io.Copy(logrus.StandardLogger().Out, execOutputBuf); err != nil {
				logrus.Errorf("An error occurred printing the exec logs: %v", err)
			}
			logrus.Error("---------------- End Kurtosis iptables chain-configuring exec command output --------------------")
			return nil, stacktrace.Propagate(err, "An error occurred running the exec command to configure iptables to use the custom Kurtosis chain")
		}
		network.serviceIpTablesChainInUse[serviceId] = initialKurtosisIpTablesChain

		// TODO Getting blocklists is an expensive call and, as of 2020-12-31, we do it twice - the solution is to make
		//  getting the blocklists not an expensive call
		blocklists, err := network.topology.GetBlocklists()
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting the blocklists to know what iptables " +
				"updates to apply on the new node")
		}
		newNodeBlocklist := blocklists[serviceId]
		updatesToApply := map[topology_types.ServiceID]*topology_types.ServiceIDSet{
			serviceId: newNodeBlocklist,
		}
		if err := network.updateIpTables(context, updatesToApply); err != nil {
			return nil, stacktrace.Propagate(err, "An error occured applying the iptables on the new node to partition it " +
				"off from other nodes")
		}
	}

	return serviceIp, nil
}

func (network *ServiceNetwork) RemoveService(
		context context.Context,
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

	// TODO Parallelize the shutdown of the service container and the sidecar container
	// Make a best-effort attempt to stop the service container
	logrus.Debugf("Removing service ID '%v' with container ID '%v'...", serviceId, serviceContainerId)
	if err := network.dockerManager.StopContainer(context, serviceContainerId, containerStopTimeout); err != nil {
		return stacktrace.Propagate(err, "An error occurred stopping the container with ID %v", serviceContainerId)
	}
	network.topology.RemoveService(serviceId)
	// TODO release the IP that the service got
	delete(network.serviceContainerIds, serviceId)
	delete(network.serviceIps, serviceId)
	logrus.Debugf("Successfully removed service with container ID %v", serviceContainerId)

	if network.isPartitioningEnabled {
		// NOTE: As of 2020-12-31, we don't need to update the iptables of the other services in the network to
		//  clear the now-removed service's IP because:
		// 	 a) nothing is using it so it doesn't do anything and
		//	 b) all service's iptables get overwritten on the next Add/Repartition call
		// If we ever do incremental iptables though, we'll need to fix all the other service's iptables here!

		sidecarContainerInfo, found := network.sidecarContainerInfo[serviceId]
		if !found {
			return stacktrace.NewError(
				"Couldn't find sidecar container ID for service '%v'; this is a code bug where the sidecar container ID didn't get stored at creation time",
				serviceId)

		}
		sidecarContainerId := sidecarContainerInfo.containerId

		// TODO Parallelize the shutdown of the sidecar container with the service container
		// Try to stop the sidecar container too
		logrus.Debugf("Removing sidecar container with container ID '%v'...", sidecarContainerId)
		// The sidecar container doesn't have any state and is in a 'sleep infinity' loop; it's okay to just kill
		if err := network.dockerManager.KillContainer(context, sidecarContainerId); err != nil {
			return stacktrace.Propagate(err, "An error occurred stopping the sidecar container with ID %v", sidecarContainerId)
		}
		// TODO release the IP that the service received
		delete(network.sidecarContainerInfo, serviceId)
		logrus.Debugf("Successfully removed sidecar container with container ID %v", sidecarContainerId)
	}
	return nil
}

// Stops all services that have been created by the API container, and renders the service network unusable
func (network *ServiceNetwork) Destroy(context context.Context, containerStopTimeout time.Duration) error {
	network.mutex.Lock()
	defer network.mutex.Unlock()
	if network.isDestroyed {
		return stacktrace.NewError("Cannot destroy the service network; it has already been destroyed")
	}

	containerStopErrors := []error{}

	// TODO parallelize this for faster shutdown
	logrus.Debugf("Making best-effort attempt to stop sidecar containers...")
	for serviceId, sidecarContainerInfo := range network.sidecarContainerInfo {
		sidecarContainerId := sidecarContainerInfo.containerId
		// Sidecar containers run 'sleep infinity' so it only wastes time to wait for graceful shutdown
		if err := network.dockerManager.KillContainer(context, sidecarContainerId); err != nil {
			wrappedErr := stacktrace.Propagate(
				err,
				"An error occurred stopping sidecar container with container ID '%v' for service '%s'",
				sidecarContainerId,
				serviceId)
			containerStopErrors = append(containerStopErrors, wrappedErr)
		}
	}
	logrus.Debugf("Made best-effort attempt to stop sidecar containers")

	// TODO parallelize this for faster shutdown
	logrus.Debugf("Making best-effort attempt to stop service containers...")
	for serviceId, serviceContainerId := range network.serviceContainerIds {
		// TODO set the stop timeout on the service itself
		if err := network.dockerManager.StopContainer(context, serviceContainerId, containerStopTimeout); err != nil {
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

// TODO write tests for me!!
/*
Updates the iptables of the services with the given IDs to match the target blocklists
 */
func (network *ServiceNetwork) updateIpTables(
		ctx context.Context,
		targetBlocklists map[topology_types.ServiceID]*topology_types.ServiceIDSet) error {
	// TODO Run the container updates in parallel, with the container being modified being the most important
	erroredServiceIdStrs := []string{}
	for serviceId, newBlocklist := range targetBlocklists {
		if err := network.updateIpTablesForService(ctx, serviceId, *newBlocklist); err != nil {
			logrus.Errorf("An error occurred updating the iptables for service '%v':", serviceId)
			logrus.Error(err)
			erroredServiceIdStrs = append(erroredServiceIdStrs, string(serviceId))
		}
	}
	if len(erroredServiceIdStrs) > 0 {
		return stacktrace.NewError(
			"Error(s) occurred updating the iptables for the following services: %v",
			strings.Join(erroredServiceIdStrs, ", "),
		)
	}
	return nil
}

// TODO Write tests for this, by extracting the logic to run exec commands on the sidecar into a separate, mockable
//  interface
func (network ServiceNetwork) updateIpTablesForService(ctx context.Context, serviceId topology_types.ServiceID, newBlocklist topology_types.ServiceIDSet) error {
	primaryChain := network.serviceIpTablesChainInUse[serviceId]
	var backgroundChain ipTablesChain
	if primaryChain == kurtosisIpTablesChain1 {
		backgroundChain = kurtosisIpTablesChain2
	} else if primaryChain == kurtosisIpTablesChain2 {
		backgroundChain = kurtosisIpTablesChain1
	} else {
		return stacktrace.NewError("Unrecognized iptables chain '%v' in use; this is a code bug", primaryChain)
	}

	sidecarContainerCmd, err := getSidecarContainerCommand(backgroundChain, newBlocklist, network.serviceIps)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the command to run in the sidecar container")
	}

	sidecarContainerInfo, found := network.sidecarContainerInfo[serviceId]
	if !found {
		return stacktrace.NewError(
			"Couldn't find a sidecar container info for Service ID '%v', which means there's no way " +
				"to update its iptables",
			serviceId,
		)
	}
	sidecarContainerId := sidecarContainerInfo.containerId

	logrus.Infof(
		"Running iptables command '%v' in sidecar container '%v' to update blocklist for service '%v'...",
		sidecarContainerCmd,
		sidecarContainerId,
		serviceId)
	execOutputBuf := &bytes.Buffer{}
	if err := network.dockerManager.RunExecCommand(ctx, sidecarContainerId, sidecarContainerCmd, execOutputBuf); err != nil {
		logrus.Error("-------------------- iptables blocklist-updating exec command output --------------------")
		if _, err := io.Copy(logrus.StandardLogger().Out, execOutputBuf); err != nil {
			logrus.Errorf("An error occurred printing the exec logs: %v", err)
		}
		logrus.Error("------------------ End iptables blocklist-updating exec command output --------------------")
		return stacktrace.Propagate(
			err,
			"An error occurred running iptables command '%v' in sidecar container '%v' to update the blocklist of service '%v'",
			sidecarContainerCmd,
			sidecarContainerId,
			serviceId)
	}
	network.serviceIpTablesChainInUse[serviceId] = backgroundChain
	logrus.Infof("Successfully updated blocklist for service '%v'", serviceId)
	return nil
}

/*
Given the new set of services that should be in the service's iptables, calculate the command that needs to be
	run in the sidecar container to make the service's iptables match the desired state.
 */
func getSidecarContainerCommand(
		backgroundChain ipTablesChain,
		newBlocklist topology_types.ServiceIDSet,
		serviceIps map[topology_types.ServiceID]net.IP) ([]string, error) {
	sidecarContainerCommand := []string{
		ipTablesCommand,
		ipTablesFlushChainFlag,
		string(backgroundChain),
	}

	if newBlocklist.Size() > 0 {
		ipsToBlockStrSlice := []string{}
		for _, serviceIdToBlock := range newBlocklist.Elems() {
			ipToBlock, found := serviceIps[serviceIdToBlock]
			if !found {
				return nil, stacktrace.NewError("Need to block the IP of target service ID '%v', but " +
					"the target service doesn't have an IP associated to it",
					serviceIdToBlock)
			}
			ipsToBlockStrSlice = append(ipsToBlockStrSlice, ipToBlock.String())
		}
		ipsToBlockCommaList := strings.Join(ipsToBlockStrSlice, ",")

		// As of 2020-12-31 the Kurtosis chains get used by both INPUT and OUTPUT intrinsic iptables chains,
		//  so we add rules to the Kurtosis chains to drop traffic both inbound and outbound
		for _, flag := range []string{"-s", "-d"} {
			// PERF NOTE: If it takes iptables a long time to insert all the rules, we could do the
			//  extra work leg work to calculate the diff and insert only what's needed
			addBlockedSourceIpsCommand := []string{
				ipTablesCommand,
				ipTablesAppendRuleFlag,
				string(backgroundChain),
				flag,
				ipsToBlockCommaList,
				"-j",
				ipTablesDropAction,
			}
			sidecarContainerCommand = append(sidecarContainerCommand, "&&")
			sidecarContainerCommand = append(sidecarContainerCommand, addBlockedSourceIpsCommand...)
		}
	}

	// Lastly, make sure to update which chain is being used for both INPUT and OUTPUT iptables
	for _, intrinsicChain := range []string{ipTablesInputChain, ipTablesOutputChain} {
		setBackgroundChainInFirstPositionCommand := []string{
			ipTablesCommand,
			ipTablesReplaceRuleFlag,
			intrinsicChain,
			strconv.Itoa(ipTablesFirstRuleIndex),
			"-j",
			string(backgroundChain),
		}
		sidecarContainerCommand = append(sidecarContainerCommand, "&&")
		sidecarContainerCommand = append(sidecarContainerCommand, setBackgroundChainInFirstPositionCommand...)
	}

	// Because the command contains '&&', we need to wrap this in 'sh -c' else iptables
	//  will think the '&&' is an argument intended for itself
	shWrappedCommand := []string{
		"sh",
		"-c",
		strings.Join(sidecarContainerCommand, " "),
	}
	return shWrappedCommand, nil
}
