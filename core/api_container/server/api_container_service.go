/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package server

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/api_container/api_container_rpc_api/bindings"
	"github.com/kurtosis-tech/kurtosis/api_container/server/module_store"
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
	"time"
)

const (
	// Custom-set max size for logs coming back from docker exec.
	// Protobuf sets a maximum of 2GB for responses, in interest of keeping performance sane
	// we pick a reasonable limit of 10MB on log responses for docker exec.
	// See: https://stackoverflow.com/questions/34128872/google-protobuf-maximum-size/34186672
	maxLogOutputSizeBytes = 10 * 1024 * 1024
)

type ApiContainerService struct {
	dockerManager   *docker_manager.DockerManager
	serviceNetwork  *service_network.ServiceNetwork
	modules 		*module_store.ModuleStore
}

func NewApiContainerService(dockerManager *docker_manager.DockerManager, serviceNetwork *service_network.ServiceNetwork) *ApiContainerService {
	return &ApiContainerService{dockerManager: dockerManager, serviceNetwork: serviceNetwork}
}

func (service ApiContainerService) LoadModule(ctx context.Context, args *bindings.LoadModuleArgs) (*bindings.LoadModuleResponse, error) {
	moduleId, moduleIpAddr, err := service.modules.LoadModule(ctx, args.ContainerImage, args.ParamsJson)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred loading module with container image '%v' and params JSON '%v'", args.ContainerImage, args.ParamsJson)
	}
	result := &bindings.LoadModuleResponse{
		ModuleId: string(moduleId),
		IpAddr:   moduleIpAddr.String(),
	}
	return result, nil
}

func (service ApiContainerService) RegisterService(ctx context.Context, args *bindings.RegisterServiceArgs) (*bindings.RegisterServiceResponse, error) {
	serviceId := service_network_types.ServiceID(args.ServiceId)
	partitionId := service_network_types.PartitionID(args.PartitionId)

	ip, err := service.serviceNetwork.RegisterService(serviceId, partitionId)
	if err != nil {
		// TODO IP: Leaks internal information about API container
		return nil, stacktrace.Propagate(err, "An error occurred registering service '%v' in the service network", serviceId)
	}

	return &bindings.RegisterServiceResponse{
		IpAddr:                          ip.String(),
	}, nil
}

func (service ApiContainerService) GenerateFiles(ctx context.Context, args *bindings.GenerateFilesArgs) (*bindings.GenerateFilesResponse, error) {
	serviceId := service_network_types.ServiceID(args.ServiceId)
	filesToGenerate := args.FilesToGenerate
	generatedFileRelativeFilepaths, err := service.serviceNetwork.GenerateFiles(serviceId, filesToGenerate)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred generating files for service '%v'", serviceId)
	}
	return &bindings.GenerateFilesResponse{
		GeneratedFileRelativeFilepaths: generatedFileRelativeFilepaths,
	}, nil
}

func (service ApiContainerService) StartService(ctx context.Context, args *bindings.StartServiceArgs) (*bindings.StartServiceResponse, error) {
	logrus.Debugf("Received request to start service with the following args: %+v", args)

	usedPorts := map[nat.Port]bool{}
	portObjToPortSpecStr := map[nat.Port]string{}
	for portSpecStr := range args.UsedPorts {
		// NOTE: this function, frustratingly, doesn't return an error on failure - just emptystring
		protocol, portNumberStr := nat.SplitProtoPort(portSpecStr)
		if protocol == "" {
			return nil, stacktrace.NewError(
				"Could not split port specification string '%s' into protocol & number strings",
				portSpecStr)
		}
		portObj, err := nat.NewPort(protocol, portNumberStr)
		if err != nil {
			// TODO IP: Leaks internal information about the API container
			return nil, stacktrace.Propagate(
				err,
				"An error occurred constructing a port object out of protocol '%v' and port number string '%v'",
				protocol,
				portNumberStr)
		}
		usedPorts[portObj] = true
		portObjToPortSpecStr[portObj] = portSpecStr
	}

	serviceId := service_network_types.ServiceID(args.ServiceId)

	hostPortBindings, err := service.serviceNetwork.StartService(
		ctx,
		serviceId,
		args.DockerImage,
		usedPorts,
		args.EntrypointArgs,
		args.CmdArgs,
		args.DockerEnvVars,
		args.SuiteExecutionVolMntDirpath,
		args.FilesArtifactMountDirpaths)
	if err != nil {
		// TODO IP: Leaks internal information about the API container
		return nil, stacktrace.Propagate(err, "An error occurred starting the service in the service network")
	}

	// We strip out ports with nil host port bindings to make it easier to iterate over this map on the client side
	responseHostPortBindings := map[string]*bindings.PortBinding{}
	for portObj, hostPortBinding := range hostPortBindings {
		portSpecStr, found := portObjToPortSpecStr[portObj]
		if !found {
			return nil, stacktrace.NewError(
				"Found a port object, %+v, that doesn't correspond to a spec string as passed in via the args; this is very strange!",
				portObj,
			)
		}
		if hostPortBinding != nil {
			responseBinding := &bindings.PortBinding{
				InterfaceIp:   hostPortBinding.HostIP,
				InterfacePort: hostPortBinding.HostPort,
			}
			responseHostPortBindings[portSpecStr] = responseBinding
		}
	}
	response := bindings.StartServiceResponse{
		UsedPortsHostPortBindings: responseHostPortBindings,
	}

	serviceStartLoglineSuffix := ""
	if len(responseHostPortBindings) > 0 {
		serviceStartLoglineSuffix = fmt.Sprintf(
			" with the following service-port-to-host-port bindings: %+v",
			responseHostPortBindings,
		)
	}
	logrus.Infof("Started service '%v'%v", serviceId, serviceStartLoglineSuffix)

	return &response, nil
}

func (service ApiContainerService) RemoveService(ctx context.Context, args *bindings.RemoveServiceArgs) (*emptypb.Empty, error) {
	serviceId := service_network_types.ServiceID(args.ServiceId)

	containerStopTimeoutSeconds := args.ContainerStopTimeoutSeconds
	containerStopTimeout := time.Duration(containerStopTimeoutSeconds) * time.Second

	if err := service.serviceNetwork.RemoveService(ctx, serviceId, containerStopTimeout); err != nil {
		// TODO IP: Leaks internal information about the API container
		return nil, stacktrace.Propagate(err, "An error occurred removing service with ID '%v'", serviceId)
	}
	return &emptypb.Empty{}, nil
}

func (service ApiContainerService) Repartition(ctx context.Context, args *bindings.RepartitionArgs) (*emptypb.Empty, error) {
	// No need to check for dupes here - that happens at the lowest-level call to ServiceNetwork.Repartition (as it should)
	partitionServices := map[service_network_types.PartitionID]*service_network_types.ServiceIDSet{}
	for partitionIdStr, servicesInPartition := range args.PartitionServices {
		partitionId := service_network_types.PartitionID(partitionIdStr)
		serviceIdSet := service_network_types.NewServiceIDSet()
		for serviceIdStr := range servicesInPartition.ServiceIdSet {
			serviceId := service_network_types.ServiceID(serviceIdStr)
			serviceIdSet.AddElem(serviceId)
		}
		partitionServices[partitionId] = serviceIdSet
	}

	partitionConnections := map[service_network_types.PartitionConnectionID]partition_topology.PartitionConnection{}
	for partitionAStr, partitionBToConnection := range args.PartitionConnections {
		partitionAId := service_network_types.PartitionID(partitionAStr)
		for partitionBStr, connectionInfo := range partitionBToConnection.ConnectionInfo {
			partitionBId := service_network_types.PartitionID(partitionBStr)
			partitionConnectionId := *service_network_types.NewPartitionConnectionID(partitionAId, partitionBId)
			if _, found := partitionConnections[partitionConnectionId]; found {
				return nil, stacktrace.NewError(
					"Partition connection '%v' <-> '%v' was defined twice (possibly in reverse order)",
					partitionAId,
					partitionBId)
			}
			partitionConnection := partition_topology.PartitionConnection{
				IsBlocked: connectionInfo.IsBlocked,
			}
			partitionConnections[partitionConnectionId] = partitionConnection
		}
	}

	defaultConnectionInfo := args.DefaultConnection
	defaultConnection := partition_topology.PartitionConnection{
		IsBlocked: defaultConnectionInfo.IsBlocked,
	}

	if err := service.serviceNetwork.Repartition(
		ctx,
		partitionServices,
		partitionConnections,
		defaultConnection); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred repartitioning the test network")
	}
	return &emptypb.Empty{}, nil
}

func (service ApiContainerService) ExecCommand(ctx context.Context, args *bindings.ExecCommandArgs) (*bindings.ExecCommandResponse, error) {
	serviceIdStr := args.ServiceId
	serviceId := service_network_types.ServiceID(serviceIdStr)
	command := args.CommandArgs
	exitCode, logOutput, err := service.serviceNetwork.ExecCommand(ctx, serviceId, command)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred running exec command '%v' against service '%v' in the service network",
			command,
			serviceId)
	}
	logOutputSize := logOutput.Len()
	if logOutputSize > maxLogOutputSizeBytes {
		return nil, stacktrace.NewError("Log output from docker exec command %+v was %v bytes, but maximum size allowed by Kurtosis is %v",
			command,
			logOutputSize,
			maxLogOutputSizeBytes,
		)
	}
	resp := &bindings.ExecCommandResponse{
		ExitCode: exitCode,
		LogOutput: logOutput.Bytes(),
	}
	return resp, nil
}

