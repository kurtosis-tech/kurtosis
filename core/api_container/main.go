/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-client/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-client/golang/kurtosis_core_rpc_api_consts"
	"github.com/kurtosis-tech/kurtosis-core-launcher-lib/lib/api_container_docker_consts"
	v0 "github.com/kurtosis-tech/kurtosis-core-launcher-lib/lib/api_versions/v0"
	"github.com/kurtosis-tech/kurtosis/api_container/server"
	"github.com/kurtosis-tech/kurtosis/api_container/server/lambda_store"
	"github.com/kurtosis-tech/kurtosis/api_container/server/lambda_store/lambda_launcher"
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network/networking_sidecar"
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network/user_service_launcher"
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network/user_service_launcher/files_artifact_expander"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/kurtosis-tech/kurtosis/commons/container_own_id_finder"
	"github.com/kurtosis-tech/kurtosis/commons/enclave_data_volume"
	"github.com/kurtosis-tech/kurtosis/commons/object_name_providers"
	minimal_grpc_server "github.com/kurtosis-tech/minimal-grpc-server/golang/server"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
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

	err := runMain()
	if err != nil {
		logrus.Errorf("An error occurred when running the main function")
		fmt.Fprintln(logrus.StandardLogger().Out, err)
		os.Exit(failureExitCode)
	}
	os.Exit(successExitCode)

}

func runMain () error {
	args, err := v0.RetrieveV0LaunchAPIArgs()
	if err != nil {
		return stacktrace.Propagate(err, "Couldn't retrieve launch API args from the environment")
	}

	logLevel, err := logrus.ParseLevel(args.LogLevel)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing the log level string '%v':", args.LogLevel)
	}
	logrus.SetLevel(logLevel)

	dockerManager, err := createDockerManager()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the Docker manager")
	}
	ownContainerId, err := container_own_id_finder.GetOwnContainerIdByName(context.Background(), dockerManager, args.ContainerName)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting this container's ID")
	}
	defer func() {
		// This is a catch-all safeguard, to leave the network absolutely empty by the time the API container shuts down
		if err := disconnectExternalContainersAndKillEverythingElse(
				context.Background(),
				dockerManager,
				args.NetworkId,
				ownContainerId,
				args.ExternalMountedContainerIds); err != nil {
			// TODO propagate this to the user somehow - likely in the exit code of the API container
			logrus.Errorf("An error occurred when disconnecting external containers and killing everything else:")
			fmt.Fprintln(logrus.StandardLogger().Out, err)
		}
	}()

	enclaveDataVol := enclave_data_volume.NewEnclaveDataVolume(api_container_docker_consts.EnclaveDataVolumeMountpoint)

	serviceNetwork, lambdaStore, err := createServiceNetworkAndLambdaStore(dockerManager, enclaveDataVol, args)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the service network & Lambda store")
	}
	// TODO parallelize Lambda & service network destruction for perf
	defer func() {
		if err := serviceNetwork.Destroy(context.Background()); err != nil {
			logrus.Errorf("An error occurred while destroying the service network:")
			fmt.Fprintln(logrus.StandardLogger().Out, err)
		}
	}()
	defer func() {
		if err := lambdaStore.Destroy(context.Background()); err != nil {
			logrus.Errorf("An error occurred while destroying the Lambda store:")
			fmt.Fprintln(logrus.StandardLogger().Out, err)
		}
	}()

	//Creation of ApiContainerService
	apiContainerService, err := server.NewApiContainerService(enclaveDataVol, serviceNetwork, lambdaStore)
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
		logrus.Errorf("An error occurred running the server:")
		fmt.Fprintln(logrus.StandardLogger().Out, err)
		os.Exit(failureExitCode)
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

func createServiceNetworkAndLambdaStore(
		dockerManager *docker_manager.DockerManager,
		enclaveDataVol *enclave_data_volume.EnclaveDataVolume,
		args *v0.V0LaunchAPIArgs) (service_network.ServiceNetwork, *lambda_store.LambdaStore, error) {
	enclaveId := args.EnclaveId
	enclaveObjNameProvider := object_name_providers.NewEnclaveObjectNameProvider(enclaveId)
	_, parsedSubnetMask, err := net.ParseCIDR(args.SubnetMask)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred parsing subnet CIDR string '%v'", args.SubnetMask)
	}
	freeIpAddrTracker := commons.NewFreeIpAddrTracker(
		logrus.StandardLogger(),
		parsedSubnetMask,
		args.TakenIpAddrs,
	)

	// TODO We don't want to have the artifact cache inside the volume anymore - it should be a separate volume, or on the local filesystem
	//  This is because, with Kurtosis interactive, it will need to be independent of executions of Kurtosis
	filesArtifactCache, err := enclaveDataVol.GetFilesArtifactCache()
	if err != nil {
		return nil, nil, stacktrace.Propagate(err,"An error occurred getting the files artifact cache")
	}

	dockerNetworkId := args.NetworkId
	isPartitioningEnabled := args.IsPartitioningEnabled

	filesArtifactExpander := files_artifact_expander.NewFilesArtifactExpander(
		enclaveId,
		dockerManager,
		enclaveObjNameProvider,
		dockerNetworkId,
		freeIpAddrTracker,
		filesArtifactCache,
	)

	userServiceLauncher := user_service_launcher.NewUserServiceLauncher(
		dockerManager,
		enclaveObjNameProvider,
		freeIpAddrTracker,
		args.ShouldPublishPorts,
		filesArtifactExpander,
		enclaveId,
	)

	networkingSidecarManager := networking_sidecar.NewStandardNetworkingSidecarManager(
		dockerManager,
		enclaveObjNameProvider,
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

	lambdaLauncher := lambda_launcher.NewLambdaLauncher(
		dockerManager,
		args.ApiContainerIpAddr,
		enclaveObjNameProvider,
		freeIpAddrTracker,
		args.ShouldPublishPorts,
		dockerNetworkId,
		enclaveId,
	)

	lambdaStore := lambda_store.NewLambdaStore(dockerManager, lambdaLauncher)

	return serviceNetwork, lambdaStore, nil
}

func disconnectExternalContainersAndKillEverythingElse(
		ctx context.Context,
		dockerManager *docker_manager.DockerManager,
		networkId string,
		ownContainerId string,
		externalContainerIds map[string]bool) error {
	logrus.Debugf("Disconnecting external containers and killing everything else on network '%v'...", networkId)
	containerIds, err := dockerManager.GetContainerIdsConnectedToNetwork(ctx, networkId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the containers connected to network '%v', which is necessary for stopping them", networkId)
	}

	allContainerHandlingErrors := map[string]error{}
	for _, containerId := range containerIds {
		if containerId == ownContainerId {
			continue
		}

		var containerHandlingErr error = nil
		if _, found := externalContainerIds[containerId]; found {
			// We don't want to kill the external containers since we don't own them, but we need them gone from the network
			if err := dockerManager.DisconnectContainerFromNetwork(ctx, containerId, networkId); err != nil {
				containerHandlingErr = stacktrace.Propagate(err, "An error occurred disconnecting container '%v' from the network", containerId)
			}
		} else {
			if err := dockerManager.KillContainer(ctx, containerId); err != nil {
				containerHandlingErr = stacktrace.Propagate(err, "An error occurred killing container '%v'", containerId)
			}
		}
		if containerHandlingErr != nil {
			allContainerHandlingErrors[containerId] = containerHandlingErr
		}
	}

	if len(allContainerHandlingErrors) > 0 {
		errorStrs := []string{}
		for containerId, err := range allContainerHandlingErrors {
			strToAppend := fmt.Sprintf("An error occurred removing container '%v':\n%v", containerId, err.Error())
			errorStrs = append(errorStrs, strToAppend)
		}
		resultErrStr := strings.Join(errorStrs, "\n\n")
		return errors.New(resultErrStr)
	}
	logrus.Debugf("Successfully disconnected external containers and killed everything else on network '%v'", networkId)
	return nil
}
