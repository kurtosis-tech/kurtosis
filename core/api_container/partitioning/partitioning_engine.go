/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package partitioning

import (
	"bytes"
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/api_container/api"
	"github.com/kurtosis-tech/kurtosis/api_container/partitioning/testnet_topology"
	"github.com/kurtosis-tech/kurtosis/api_container/user_service_launcher"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"strings"
	"sync"
)

const (
	defaultPartitionId                   PartitionID = "default"
	startingDefaultConnectionBlockStatus             = false

	iproute2ContainerImage = "kurtosistech/iproute2"

	// We'll create a new chain at the 0th spot for every service
	kurtosisIpTablesChain = "KURTOSIS"

	ipTablesCommand = "iptables"
	ipTablesInputChain = "INPUT"
	ipTablesNewChainFlag = "-N"
	ipTablesInsertRuleFlag = "-I"
	ipTablesFlushChainFlag = "-F"
	ipTablesAppendRuleFlag  = "-A"
	ipTablesDropAction = "DROP"
)

type PartitionID string

/**
This is the engine
 */
type PartitioningEngine struct {
	// Whether partitioning has been enabled for this particular test
	isPartitioningEnabled bool

	dockerNetworkId string

	freeIpAddrTracker *commons.FreeIpAddrTracker

	dockerManager *commons.DockerManager

	userServiceLauncher *user_service_launcher.UserServiceLauncher

	mutex *sync.Mutex

	topology *testnet_topology.TestnetTopology

	// == Per-service info ==================================================================
	serviceIps map[api.ServiceID]net.IP

	serviceContainerIds map[api.ServiceID]string

	sidecarContainerIps map[api.ServiceID]net.IP

	// Mapping of Service ID -> the sidecar iproute container that we'll use to manipulate the
	//  container's networking
	sidecarContainerIds map[api.ServiceID]string

	// Mapping of serviceID -> set of serviceIDs tracking what's currently being dropped in the INPUT chain of the service
	ipTablesBlocks map[api.ServiceID]*ServiceIDSet
}

func NewPartitioningEngine(
		isPartitioningEnabled bool,
		dockerNetworkId string,
		freeIpAddrTracker *commons.FreeIpAddrTracker,
		dockerManager *commons.DockerManager,
		userServiceLauncher *user_service_launcher.UserServiceLauncher) *PartitioningEngine {
	defaultPartitionConnection := testnet_topology.PartitionConnection{IsBlocked: startingDefaultConnectionBlockStatus}
	return &PartitioningEngine{
		isPartitioningEnabled: isPartitioningEnabled,
		dockerNetworkId: dockerNetworkId,
		freeIpAddrTracker: freeIpAddrTracker,
		dockerManager: dockerManager,
		userServiceLauncher: userServiceLauncher,
		mutex:               &sync.Mutex{},
		topology:            testnet_topology.NewTestnetTopology(
			defaultPartitionId,
			defaultPartitionConnection,
		),
		serviceIps:          map[api.ServiceID]net.IP{},
		sidecarContainerIds: map[api.ServiceID]string{},
	}
}

/*
Completely repartitions the network, throwing away the old topology
 */
func (engine *PartitioningEngine) Repartition(
		context context.Context,
		newPartitionServices map[PartitionID]*ServiceIDSet,
		newPartitionConnections map[testnet_topology.PartitionConnectionID]testnet_topology.PartitionConnection,
		newDefaultConnection testnet_topology.PartitionConnection) error {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()

	if !engine.isPartitioningEnabled {
		return stacktrace.NewError("Cannot repartition; partitioning is not enabled")
	}

	if err := engine.topology.Repartition(newPartitionServices, newPartitionConnections, newDefaultConnection); err != nil {
		return stacktrace.Propagate(err, "An error occurred repartitioning the network topology")
	}
	if err := engine.updateIpTables(context); err != nil {
		return stacktrace.Propagate(err, "An error occurred updating the IP tables to match the target blocklist after repartitioning")
	}
	return nil
}


/*
Creates the service with the given ID in the given partition

If partitionId is empty string, the default partition ID is used

Returns: The IP address of the new service
 */
func (engine *PartitioningEngine) CreateServiceInPartition(
		context context.Context,
		serviceId api.ServiceID,
		imageName string,
		usedPorts map[nat.Port]bool,
		partitionId PartitionID,
		ipPlaceholder string,
		startCmd []string,
		dockerEnvVars map[string]string,
		testVolumeMountDirpath string) (net.IP, error) {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()

	if partitionId == "" {
		partitionId = defaultPartitionId
	}

	if _, found := engine.topology.GetPartitionServices()[partitionId]; !found {
		return nil, stacktrace.NewError(
			"No partition with ID '%v' exists in the current partition topology",
			partitionId,
		)
	}

	// TODO Modify this to take in an IP, to kill the race condition with the service starting & partition application
	serviceContainerId, serviceIp, err := engine.userServiceLauncher.Launch(
		context,
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
	// TODO this is very error-prone - combine them into a single object so it's atomic!
	engine.serviceIps[serviceId] = serviceIp
	engine.serviceContainerIds[serviceId] = serviceContainerId
	if err := engine.topology.AddService(serviceId, partitionId); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding service '%v' to partition '%v'", serviceId, partitionId)
	}

	if engine.isPartitioningEnabled {
		sidecarIp, err := engine.freeIpAddrTracker.GetFreeIpAddr()
		if err != nil {
			return nil, stacktrace.Propagate(
				err,
				"An error occurred getting a free IP address for the networking sidecar container")
		}
		sidecarContainerId, err := engine.dockerManager.CreateAndStartContainer(
			context,
			iproute2ContainerImage,
			engine.dockerNetworkId,
			sidecarIp,
			map[commons.ContainerCapability]bool{
				commons.NetAdmin: true,
			},
			commons.NewContainerNetworkMode(serviceContainerId),
			map[nat.Port]*nat.PortBinding{},
			[]string{"sleep","infinity"},  // We sleep forever since iptables stuff gets executed via 'exec'
			map[string]string{}, // No environment variables
			map[string]string{}, // no bind mounts for services created via the Kurtosis API
			map[string]string{}, // No volume mounts either
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred starting the sidecar iproute container for modifying the service container's iptables")
		}
		// TODO combine these into an object so it's one atomic operation
		engine.sidecarContainerIps[serviceId] = sidecarIp
		engine.sidecarContainerIds[serviceId] = sidecarContainerId

		// As soon as we have the sidecar, we need to create the Kurtosis chain and insert it in first position on the INPUT chain
		configureKurtosisChainCommand := []string{
			ipTablesCommand,
			ipTablesNewChainFlag,
			kurtosisIpTablesChain,
			"&&",
			ipTablesCommand,
			ipTablesInsertRuleFlag,
			ipTablesInputChain,
			"1",  // We want to insert the Kurtosis chain first, so it always runs
			"-j",
			kurtosisIpTablesChain,
		}
		if err := engine.dockerManager.RunExecCommand(
				context,
				sidecarContainerId,
				configureKurtosisChainCommand,
				logrus.StandardLogger().Out); err !=  nil {
			return nil, stacktrace.Propagate(err, "An error occurred configuring iptables to use the custom Kurtosis chain")
		}

		// TODO Right now, there's a period of time between user service container launch, and the recalculation of
		//  the blocklist and the application of iptables to the user's container
		//  This means there's a race condition period of time where the service container will be able to talk to everyone!
		//  The fix is to, before starting the service, apply the blocklists to every other node
		//  That way, even with the race condition, the other nodes won't accept traffic from the new node
		if err := engine.updateIpTables(context); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred updating IP tables after adding service '%v'", serviceId)
		}
	}

	return serviceIp, nil
}

// TODO write tests for me!!
/*
Gets the latest target blocklists from the topology and makes sure iptables matches
 */
func (engine *PartitioningEngine) updateIpTables(context context.Context) error {
	targetBlocklists, err := engine.topology.GetBlocklists()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the current blocklists")
	}

	toUpdate, err := getServicesNeedingIpTablesUpdates(engine.ipTablesBlocks, targetBlocklists)

	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the services that need iptables updated")
	}

	sidecarContainerCmds, err := getSidecarContainerCommands(toUpdate, engine.serviceIps)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the sidecar container commands for the " +
			"services that need iptables updates")
	}

	// TODO Run the container updates in parallel, with the container being modified being the most important
	for serviceId, command := range sidecarContainerCmds {
		sidecarContainerId, found := engine.sidecarContainerIds[serviceId]
		if !found {
			// TODO maybe start one if we can't find it?
			return stacktrace.NewError(
				"Couldn't find a sidecar container for Service ID '%v', which means there's no way " +
					"to update its iptables",
				serviceId,
			)
		}

		logrus.Info("Running iptables command '%v' in sidecar container '%v' to update blocklist for service '%v'...")
		commandLogsBuf := &bytes.Buffer{}
		if err := engine.dockerManager.RunExecCommand(context, sidecarContainerId, command, commandLogsBuf); err != nil {
			return stacktrace.Propagate(
				err,
				"An error occurred running iptables command '%v' in sidecar container '%v' to update the blocklist of service '%v'",
				command,
				sidecarContainerId,
				serviceId)
		}
		logrus.Info("Successfully updated blocklist for service '%v'")
		logrus.Info("---------------------------------- Command Logs ---------------------------------------")
		if _, err := io.Copy(logrus.StandardLogger().Out, commandLogsBuf); err != nil {
			logrus.Error("An error occurred printing the exec logs: %v", err)
		}
		logrus.Info("---------------------------------End Command Logs -------------------------------------")
	}

	// Defensive copy when we store
	blockListToStore := map[api.ServiceID]*ServiceIDSet{}
	for serviceId, newBlockedServicesForId := range targetBlocklists {
		blockListToStore[serviceId] = newBlockedServicesForId.Copy()
	}
	engine.ipTablesBlocks = blockListToStore
	return nil
}

// TODO Write tests for me!!
/*
Compares the target state of the world with the current state of the world, and returns only a list of
	the services that need to be updated.
 */
func getServicesNeedingIpTablesUpdates(
		currentBlockedServices map[api.ServiceID]*ServiceIDSet,
		newBlockedServices map[api.ServiceID]*ServiceIDSet) (map[api.ServiceID]*ServiceIDSet, error) {
	result := map[api.ServiceID]*ServiceIDSet{}
	for serviceId, newBlockedServicesForId := range newBlockedServices {
		if newBlockedServicesForId.Contains(serviceId) {
			return nil, stacktrace.NewError("Requested for service ID '%v' to block itself!", serviceId)
		}

		// To avoid unnecessary Docker work, we won't update any iptables if the result would be the same
		//  as the current state
		currentBlockedServicesForId, found := currentBlockedServices[serviceId]
		if !found {
			currentBlockedServicesForId = NewServiceIDSet()
		}

		noChangesNeeded := newBlockedServicesForId.Equals(currentBlockedServicesForId)
		if noChangesNeeded {
			continue
		}

		result[serviceId] = newBlockedServicesForId
	}
	return result, nil
}

// TODO write tests for me!!
/*
Given a list of updates that need to happen to a service's iptables, a map of serviceID -> commands that
	will be executed on the sidecar Docker container for the service

Args:
	toUpdate: A mapping of serviceID -> set of serviceIDs to block
 */
func getSidecarContainerCommands(
		toUpdate map[api.ServiceID]*ServiceIDSet,
		serviceIps map[api.ServiceID]net.IP) (map[api.ServiceID][]string, error) {
	result := map[api.ServiceID][]string{}


	// TODO We build two separate commands - flush the Kurtosis iptables chain, and then populate it with new stuff
	//  This means that there's a (very small) window of time where the iptables aren't blocked
	//  To fix this, we should really have two Kurtosis chains, and while one is running build the other one and
	//  then switch over in one atomic operation.
	for serviceId, newBlockedServicesForId := range toUpdate {
		// TODO If this is performance-inhibitive, we could do the extra work to insert only what's needed
		// When modifying a service's iptables, we always want to flush the old and set the new, rather
		//  than trying to update
		sidecarContainerCommand := []string{
			ipTablesCommand,
			ipTablesFlushChainFlag,
			kurtosisIpTablesChain,
		}

		if newBlockedServicesForId.Size() >= 0 {
			ipsToBlockStrSlice := []string{}
			for _, serviceIdToBlock := range newBlockedServicesForId.Elems() {
				ipToBlock, found := serviceIps[serviceIdToBlock]
				if !found {
					return nil, stacktrace.NewError("Service ID '%v' needs to block the IP of target service ID '%v', but " +
						"the target service doesn't have an IP associated to it",
						serviceId,
						serviceIdToBlock)
				}
				ipsToBlockStrSlice = append(ipsToBlockStrSlice, ipToBlock.String())
			}
			ipsToBlockCommaList := strings.Join(ipsToBlockStrSlice, ",")

			addBlockedIpsCommand := []string{
				ipTablesCommand,
				ipTablesAppendRuleFlag,
				kurtosisIpTablesChain,
				"-s",
				ipsToBlockCommaList,
				"-j",
				ipTablesDropAction,
			}
			sidecarContainerCommand = append(sidecarContainerCommand, "&&")
			sidecarContainerCommand = append(sidecarContainerCommand, addBlockedIpsCommand...)
		}
		result[serviceId] = sidecarContainerCommand
	}
	return result, nil
}