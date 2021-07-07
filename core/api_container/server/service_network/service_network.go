/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package service_network

import (
	"bytes"
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis-client/golang/core_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network/service_network_types"
	"net"
	"time"
)

type ServiceNetwork interface {
	/*
	Completely repartitions the network, throwing away the old topology
	*/
	Repartition(
		ctx context.Context,
		newPartitionServices map[service_network_types.PartitionID]*service_network_types.ServiceIDSet,
		newPartitionConnections map[service_network_types.PartitionConnectionID]partition_topology.PartitionConnection,
		newDefaultConnection partition_topology.PartitionConnection,
	) error

	// Registers a service for use with the network (creating the IPs and so forth), but doesn't start it
	// If the partition ID is empty, registers the service with the default partition
	RegisterService(
		serviceId service_network_types.ServiceID,
		partitionId service_network_types.PartitionID,
	) (net.IP, error)

	// Generates files in a location in the suite execution volume allocated to the given service
	GenerateFiles(
		serviceId service_network_types.ServiceID,
		filesToGenerate map[string]*core_api_bindings.FileGenerationOptions,
	) (map[string]string, error)

	// Copies files from the static file cache to the given service's filespace
	// Returns a mapping of static_file_id -> filepath_relative_to_suite_ex_vol_root
	LoadStaticFiles(
		serviceId service_network_types.ServiceID,
		staticFileIdKeys map[string]bool,
	) (map[string]string, error)

	// TODO add tests for this
	/*
	Starts a previously-registered but not-started service by creating it in a container

	Returns:
		Mapping of port-used-by-service -> port-on-the-Docker-host-machine where the user can make requests to the port
			to access the port. If a used port doesn't have a host port bound, then the value will be nil.
	*/
	StartService(
		ctx context.Context,
		serviceId service_network_types.ServiceID,
		imageName string,
		usedPorts map[nat.Port]bool,
		entrypointArgs []string,
		cmdArgs []string,
		dockerEnvVars map[string]string,
		suiteExecutionVolMntDirpath string,
		filesArtifactMountDirpaths map[string]string,
	) (map[nat.Port]*nat.PortBinding, error)

	RemoveService(
		ctx context.Context,
		serviceId service_network_types.ServiceID,
		containerStopTimeout time.Duration,
	) error

	ExecCommand(
		ctx context.Context,
		serviceId service_network_types.ServiceID,
		command []string,
	) (int32, *bytes.Buffer, error)

	GetServiceIP(serviceId service_network_types.ServiceID) (net.IP, error)
}
