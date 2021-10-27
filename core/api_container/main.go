/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package main

import (
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-client/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-client/golang/kurtosis_core_rpc_api_consts"
	"github.com/kurtosis-tech/kurtosis-core/api_container/server"
	"github.com/kurtosis-tech/kurtosis-core/api_container/server/external_container_store"
	"github.com/kurtosis-tech/kurtosis-core/api_container/server/module_store"
	"github.com/kurtosis-tech/kurtosis-core/api_container/server/module_store/module_launcher"
	"github.com/kurtosis-tech/kurtosis-core/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis-core/api_container/server/service_network/networking_sidecar"
	"github.com/kurtosis-tech/kurtosis-core/api_container/server/service_network/user_service_launcher"
	"github.com/kurtosis-tech/kurtosis-core/api_container/server/service_network/user_service_launcher/files_artifact_expander"
	"github.com/kurtosis-tech/kurtosis-core/commons"
	"github.com/kurtosis-tech/kurtosis-core/commons/api_container_docker_consts"
	"github.com/kurtosis-tech/kurtosis-core/commons/api_container_launcher"
	"github.com/kurtosis-tech/kurtosis-core/commons/enclave_data_volume"
	"github.com/kurtosis-tech/kurtosis-core/commons/object_labels_providers"
	"github.com/kurtosis-tech/kurtosis-core/commons/object_name_providers"
	minimal_grpc_server "github.com/kurtosis-tech/minimal-grpc-server/golang/server"
	"github.com/palantir/stacktrace"
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

	enclaveDataVol := enclave_data_volume.NewEnclaveDataVolume(api_container_docker_consts.EnclaveDataDirMountpoint)

	serviceNetwork, moduleStore, err := createServiceNetworkAndModuleStore(dockerManager, enclaveDataVol, freeIpAddrTracker, args)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the service network & module store")
	}

	//Creation of ApiContainerService
	apiContainerService, err := server.NewApiContainerService(
		enclaveDataVol,
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
		enclaveDataVol *enclave_data_volume.EnclaveDataVolume,
		freeIpAddrTracker *commons.FreeIpAddrTracker,
		args *api_container_launcher.APIContainerArgs) (service_network.ServiceNetwork, *module_store.ModuleStore, error) {
	enclaveId := args.EnclaveId
	enclaveObjNameProvider := object_name_providers.NewEnclaveObjectNameProvider(enclaveId)
	enclaveObjLabelsProvider := object_labels_providers.NewEnclaveObjectLabelsProvider(enclaveId)

	// TODO We don't want to have the artifact cache inside the volume anymore - it should be a separate volume, or on the local filesystem
	//  This is because, with Kurtosis interactive, it will need to be independent of executions of Kurtosis
	filesArtifactCache, err := enclaveDataVol.GetFilesArtifactCache()
	if err != nil {
		return nil, nil, stacktrace.Propagate(err,"An error occurred getting the files artifact cache")
	}

	dockerNetworkId := args.NetworkId
	isPartitioningEnabled := args.IsPartitioningEnabled

	filesArtifactExpander := files_artifact_expander.NewFilesArtifactExpander(
		args.EnclaveDataDirpathOnHostMachine,
		dockerManager,
		enclaveObjNameProvider,
		enclaveObjLabelsProvider,
		dockerNetworkId,
		freeIpAddrTracker,
		filesArtifactCache,
	)

	userServiceLauncher := user_service_launcher.NewUserServiceLauncher(
		dockerManager,
		enclaveObjNameProvider,
		enclaveObjLabelsProvider,
		freeIpAddrTracker,
		args.ShouldPublishPorts,
		filesArtifactExpander,
		args.EnclaveDataDirpathOnHostMachine,
	)

	networkingSidecarManager := networking_sidecar.NewStandardNetworkingSidecarManager(
		dockerManager,
		enclaveObjNameProvider,
		enclaveObjLabelsProvider,
		freeIpAddrTracker,
		dockerNetworkId)

	serviceNetwork := service_network.NewServiceNetworkImpl(
		isPartitioningEnabled,
		freeIpAddrTracker,
		dockerManager,
		dockerNetworkId,
		enclaveDataVol,
		userServiceLauncher,
		networkingSidecarManager)

	moduleLauncher := module_launcher.NewModuleLauncher(
		dockerManager,
		args.ApiContainerIpAddr,
		enclaveObjNameProvider,
		enclaveObjLabelsProvider,
		freeIpAddrTracker,
		args.ShouldPublishPorts,
		dockerNetworkId,
		args.EnclaveDataDirpathOnHostMachine,
	)

	moduleStore := module_store.NewModuleStore(dockerManager, moduleLauncher)

	return serviceNetwork, moduleStore, nil
}