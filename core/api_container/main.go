/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/kurtosis/api_container/api_container_rpc_api/api_container_rpc_api_consts"
	"github.com/kurtosis-tech/kurtosis/api_container/api_container_rpc_api/bindings"
	api_container_env_var_values2 "github.com/kurtosis-tech/kurtosis/api_container/docker_api/api_container_env_var_values"
	"github.com/kurtosis-tech/kurtosis/api_container/docker_api/api_container_mountpoints"
	"github.com/kurtosis-tech/kurtosis/api_container/server"
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network/container_name_provider"
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network/networking_sidecar"
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network/user_service_launcher"
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network/user_service_launcher/files_artifact_expander"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/kurtosis-tech/kurtosis/commons/docker_constants"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/kurtosis-tech/kurtosis/commons/free_host_port_binding_supplier"
	"github.com/kurtosis-tech/kurtosis/commons/logrus_log_levels"
	"github.com/kurtosis-tech/kurtosis/commons/suite_execution_volume"
	minimal_grpc_server "github.com/kurtosis-tech/minimal-grpc-server/server"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"os"
	"strings"
	"time"
)

const (
	successExitCode = 0
	failureExitCode = 1

	grpcServerStopGracePeriod = 5 * time.Second
)

func main() {
	// NOTE: we'll want to change the ForceColors to false if we ever want structured logging
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})

	logLevelArg := flag.String(
		"log-level",
		"info",
		fmt.Sprintf(
			"Log level to use for the API container (%v)",
			strings.Join(logrus_log_levels.GetAcceptableLogLevelStrs(), "|"),
		),
	)

	// NOTE: We take this in as JSON so that it's easy to modify the params without needing th emodify the Dockerfile
	paramsJsonArg := flag.String(
		"params-json",
		"",
		"JSON string containing the params to the API container",
	)

	flag.Parse()

	logLevel, err := logrus.ParseLevel(*logLevelArg)
	if err != nil {
		logrus.Errorf("An error occurred parsing the log level string '%v':", *logLevelArg)
		fmt.Fprintln(logrus.StandardLogger().Out, err)
		os.Exit(failureExitCode)
	}
	logrus.SetLevel(logLevel)

	suiteExecutionVolume := suite_execution_volume.NewSuiteExecutionVolume(api_container_mountpoints.SuiteExecutionVolumeMountDirpath)
	paramsJson := *paramsJsonArg

	apiContainerService, err := createApiContainerService(suiteExecutionVolume, paramsJson)
	if err != nil {
		logrus.Errorf("An error occurred creating the API container service using params JSON '%v':", paramsJson)
		fmt.Fprintln(logrus.StandardLogger().Out, err)
		os.Exit(failureExitCode)
	}
	apiContainerServiceRegistrationFunc := func(grpcServer *grpc.Server) {
		bindings.RegisterApiContainerServiceServer(grpcServer, apiContainerService)
	}
	apiContainerServer := minimal_grpc_server.NewMinimalGRPCServer(
		api_container_rpc_api_consts.ListenPort,
		api_container_rpc_api_consts.ListenProtocol,
		grpcServerStopGracePeriod,
		[]func(*grpc.Server){
			apiContainerServiceRegistrationFunc,
		},
	)

	logrus.Info("Running server...")
	if err := apiContainerServer.Run(); err != nil {
		logrus.Errorf("An error occurred running the server:")
		fmt.Fprintln(logrus.StandardLogger().Out, err)
		os.Exit(failureExitCode)
	}
	os.Exit(successExitCode)
}

func createApiContainerService(
		// TODO Rename this so it's not tied to testing
		suiteExecutionVolume *suite_execution_volume.SuiteExecutionVolume,
		paramsJsonStr string) (*server.ApiContainerService, error) {
	paramsJsonBytes := []byte(paramsJsonStr)
	var args api_container_env_var_values2.ApiContainerArgs
	if err := json.Unmarshal(paramsJsonBytes, &args); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred deserializing the args JSON '%v'", paramsJsonStr)
	}

	dockerManager, err := createDockerManager()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the Docker manager")
	}

	containerNameElemsProvider := container_name_provider.NewContainerNameElementsProvider(args.EnclaveNameElems)

	freeIpAddrTracker, err := commons.NewFreeIpAddrTracker(
		logrus.StandardLogger(),
		args.SubnetMask,
		args.TakenIpAddrs,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the free IP address tracker")
	}

	enclaveDirectory, err := suiteExecutionVolume.GetEnclaveDirectory(args.EnclaveNameElems)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the enclave directory using elems '%+v'", args.EnclaveNameElems)
	}

	// TODO We don't want to have the artifact cache inside the volume anymore - it should be a separate volume, or on the local filesystem
	artifactCache, err := suiteExecutionVolume.GetArtifactCache()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the artifact cache")
	}

	var hostPortBindingSupplier *free_host_port_binding_supplier.FreeHostPortBindingSupplier = nil
	if args.HostPortBindingSupplierParams != nil {
		hostPortSupplierParams := args.HostPortBindingSupplierParams
		supplier, err := free_host_port_binding_supplier.NewFreeHostPortBindingSupplier(
			docker_constants.HostMachineDomainInsideContainer,
			hostPortSupplierParams.InterfaceIp,
			hostPortSupplierParams.Protocol,
			hostPortSupplierParams.PortRangeStart,
			hostPortSupplierParams.PortRangeEnd,
			hostPortSupplierParams.TakenPorts,
		)
		if err != nil {
			return nil, stacktrace.Propagate(
				err,
				"Host port binding supplier params were non-null, but an error occurred creating the host port binding supplier",
			)
		}
		hostPortBindingSupplier = supplier
	}

	serviceNetwork := createServiceNetwork(
		args.EnclaveNameElems,
		args.SuiteExecutionVolumeName,
		containerNameElemsProvider,
		artifactCache,
		enclaveDirectory,
		dockerManager,
		freeIpAddrTracker,
		args.NetworkId,
		args.IsPartitioningEnabled,
		hostPortBindingSupplier)

	result := server.NewApiContainerService(dockerManager, serviceNetwork)

	return result, nil
}

func createDockerManager() (*docker_manager.DockerManager, error) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not initialize a Docker client from the environment")
	}

	dockerManager := docker_manager.NewDockerManager(logrus.StandardLogger(), dockerClient)
	return dockerManager, nil
}

func createServiceNetwork(
		filesArtifactExpansionVolumeNamePrefixElems []string,
		suiteExecutionVolName string,
		containerNameElemsProvider *container_name_provider.ContainerNameElementsProvider,
		artifactCache *suite_execution_volume.ArtifactCache,
		enclaveDirectory *suite_execution_volume.EnclaveDirectory,
		dockerManager *docker_manager.DockerManager,
		freeIpAddrTracker *commons.FreeIpAddrTracker,
		dockerNetworkId string,
		isPartitioningEnabled bool,
		hostPortBindingSupplier *free_host_port_binding_supplier.FreeHostPortBindingSupplier) *service_network.ServiceNetwork {

	filesArtifactExpander := files_artifact_expander.NewFilesArtifactExpander(
		suiteExecutionVolName,
		dockerManager,
		containerNameElemsProvider,
		dockerNetworkId,
		freeIpAddrTracker)

	userServiceLauncher := user_service_launcher.NewUserServiceLauncher(
		filesArtifactExpansionVolumeNamePrefixElems,
		dockerManager,
		containerNameElemsProvider,
		freeIpAddrTracker,
		hostPortBindingSupplier,
		artifactCache,
		filesArtifactExpander,
		dockerNetworkId,
		suiteExecutionVolName)

	networkingSidecarManager := networking_sidecar.NewStandardNetworkingSidecarManager(
		dockerManager,
		containerNameElemsProvider,
		freeIpAddrTracker,
		dockerNetworkId)

	serviceNetwork := service_network.NewServiceNetwork(
		isPartitioningEnabled,
		freeIpAddrTracker,
		dockerManager,
		enclaveDirectory,
		userServiceLauncher,
		networkingSidecarManager)

	return serviceNetwork
}
