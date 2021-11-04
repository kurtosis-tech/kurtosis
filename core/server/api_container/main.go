/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package main

import (
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-core/api/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-core/api/golang/kurtosis_core_rpc_api_consts"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/external_container_store"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/module_store"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/module_store/module_launcher"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/service_network/networking_sidecar"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/service_network/user_service_launcher"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/service_network/user_service_launcher/files_artifact_expander"
	"github.com/kurtosis-tech/kurtosis-core/server/commons"
	"github.com/kurtosis-tech/kurtosis-core/server/commons/api_container_docker_consts"
	"github.com/kurtosis-tech/kurtosis-core/server/commons/api_container_launcher"
	"github.com/kurtosis-tech/kurtosis-core/server/commons/enclave_data_directory"
	minimal_grpc_server "github.com/kurtosis-tech/minimal-grpc-server/golang/server"
	"github.com/kurtosis-tech/object-attributes-schema-lib/schema"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
	"os"
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

	err := runMain()
	if err != nil {
		logrus.Errorf("An error occurred when running the main function:")
		fmt.Fprintln(logrus.StandardLogger().Out, err)
		os.Exit(failureExitCode)
	}
	os.Exit(successExitCode)

}

func runMain () error {
	args, err := api_container_launcher.RetrieveAPIContainerArgs()
	if err != nil {
		return stacktrace.Propagate(err, "Couldn't retrieve launch API args from the environment")
	}

	logLevel, err := logrus.ParseLevel(args.LogLevel)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing the log level string '%v':", args.LogLevel)
	}
	logrus.SetLevel(logLevel)

	_, parsedSubnetMask, err := net.ParseCIDR(args.SubnetMask)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing subnet CIDR string '%v'", args.SubnetMask)
	}
	freeIpAddrTracker := commons.NewFreeIpAddrTracker(
		logrus.StandardLogger(),
		parsedSubnetMask,
		args.TakenIpAddrs,
	)

	externalContainerStore := external_container_store.NewExternalContainerStore(freeIpAddrTracker)

	dockerManager, err := createDockerManager()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the Docker manager")
	}

	enclaveDataDir := enclave_data_directory.NewEnclaveDataDirectory(api_container_docker_consts.EnclaveDataDirMountpoint)

	serviceNetwork, moduleStore, err := createServiceNetworkAndModuleStore(dockerManager, enclaveDataDir, freeIpAddrTracker, args)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the service network & module store")
	}

	//Creation of ApiContainerService
	apiContainerService, err := server.NewApiContainerService(
		enclaveDataDir,
		externalContainerStore,
		serviceNetwork,
		moduleStore,
	)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the API container service")
	}

	apiContainerServiceRegistrationFunc := func(grpcServer *grpc.Server) {
		kurtosis_core_rpc_api_bindings.RegisterApiContainerServiceServer(grpcServer, apiContainerService)
	}
	apiContainerServer := minimal_grpc_server.NewMinimalGRPCServer(
		kurtosis_core_rpc_api_consts.ListenPort,
		kurtosis_core_rpc_api_consts.ListenProtocol,
		grpcServerStopGracePeriod,
		[]func(*grpc.Server){
			apiContainerServiceRegistrationFunc,
		},
	)

	logrus.Info("Running server...")
	if err := apiContainerServer.Run(); err != nil {
		return stacktrace.Propagate(err, "An error occurred running the API container server")
	}

	return nil
}

func createDockerManager() (*docker_manager.DockerManager, error) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not initialize a Docker client from the environment")
	}

	dockerManager := docker_manager.NewDockerManager(logrus.StandardLogger(), dockerClient)
	return dockerManager, nil
}

func createServiceNetworkAndModuleStore(
		dockerManager *docker_manager.DockerManager,
		enclaveDataDir *enclave_data_directory.EnclaveDataDirectory,
		freeIpAddrTracker *commons.FreeIpAddrTracker,
		args *api_container_launcher.APIContainerArgs) (service_network.ServiceNetwork, *module_store.ModuleStore, error) {
	enclaveId := args.EnclaveId

	objAttrsProvider := schema.GetObjectAttributesProvider()
	enclaveObjAttrsProvider := objAttrsProvider.ForEnclave(enclaveId)

	// TODO We don't want to have the artifact cache inside the enclave data dir anymore - it should prob be a separate directory local filesystem
	//  This is because, with Kurtosis interactive, it will need to be independent of executions of Kurtosis
	filesArtifactCache, err := enclaveDataDir.GetFilesArtifactCache()
	if err != nil {
		return nil, nil, stacktrace.Propagate(err,"An error occurred getting the files artifact cache")
	}

	dockerNetworkId := args.NetworkId
	isPartitioningEnabled := args.IsPartitioningEnabled

	filesArtifactExpander := files_artifact_expander.NewFilesArtifactExpander(
		args.EnclaveDataDirpathOnHostMachine,
		dockerManager,
		enclaveObjAttrsProvider,
		dockerNetworkId,
		freeIpAddrTracker,
		filesArtifactCache,
	)

	userServiceLauncher := user_service_launcher.NewUserServiceLauncher(
		dockerManager,
		enclaveObjAttrsProvider,
		freeIpAddrTracker,
		args.ShouldPublishPorts,
		filesArtifactExpander,
		args.EnclaveDataDirpathOnHostMachine,
	)

	networkingSidecarManager := networking_sidecar.NewStandardNetworkingSidecarManager(
		dockerManager,
		enclaveObjAttrsProvider,
		freeIpAddrTracker,
		dockerNetworkId)

	serviceNetwork := service_network.NewServiceNetworkImpl(
		isPartitioningEnabled,
		freeIpAddrTracker,
		dockerManager,
		dockerNetworkId,
		enclaveDataDir,
		userServiceLauncher,
		networkingSidecarManager)

	moduleLauncher := module_launcher.NewModuleLauncher(
		dockerManager,
		args.ApiContainerIpAddr,
		enclaveObjAttrsProvider,
		freeIpAddrTracker,
		args.ShouldPublishPorts,
		dockerNetworkId,
		args.EnclaveDataDirpathOnHostMachine,
	)

	moduleStore := module_store.NewModuleStore(dockerManager, moduleLauncher)

	return serviceNetwork, moduleStore, nil
}
