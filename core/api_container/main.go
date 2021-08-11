/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/kurtosis-client/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-client/golang/kurtosis_core_rpc_api_consts"
	"github.com/kurtosis-tech/kurtosis/api_container/docker_api/api_container_env_var_values"
	"github.com/kurtosis-tech/kurtosis/api_container/docker_api/api_container_env_vars"
	"github.com/kurtosis-tech/kurtosis/api_container/docker_api/api_container_mountpoints"
	"github.com/kurtosis-tech/kurtosis/api_container/server"
	"github.com/kurtosis-tech/kurtosis/api_container/server/lambda_store"
	"github.com/kurtosis-tech/kurtosis/api_container/server/lambda_store/lambda_launcher"
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network/networking_sidecar"
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network/user_service_launcher"
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network/user_service_launcher/files_artifact_expander"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/kurtosis-tech/kurtosis/commons/enclave_data_volume"
	"github.com/kurtosis-tech/kurtosis/commons/object_name_providers"
	minimal_grpc_server "github.com/kurtosis-tech/minimal-grpc-server/server"
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

	defaultContainerStopTimeout = 1

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
	logLevelStr, found := os.LookupEnv(api_container_env_vars.LogLevelEnvVar)
	if !found {
		return stacktrace.NewError("No log level environment variable '%v' defined", api_container_env_vars.LogLevelEnvVar)
	}
	paramsJsonStr, found := os.LookupEnv(api_container_env_vars.ParamsJsonEnvVar)
	if !found {
		return stacktrace.NewError("No custom params JSON environment variable '%v' defined", api_container_env_vars.ParamsJsonEnvVar)
	}

	logLevel, err := logrus.ParseLevel(logLevelStr)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing the log level string '%v':", logLevelStr)
	}
	logrus.SetLevel(logLevel)

	enclaveDataVol := enclave_data_volume.NewEnclaveDataVolume(api_container_mountpoints.EnclaveDataVolumeMountpoint)

	serviceNetwork, lambdaStore, err := createServiceNetworkAndLambdaStore(enclaveDataVol, paramsJsonStr)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the service network & Lambda store")
	}
	defer func() {
		if err := serviceNetwork.Destroy(context.Background(), defaultContainerStopTimeout); err != nil {
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

func createServiceNetworkAndLambdaStore(enclaveDataVol *enclave_data_volume.EnclaveDataVolume, paramsJsonStr string) (service_network.ServiceNetwork, *lambda_store.LambdaStore, error) {
	paramsJsonBytes := []byte(paramsJsonStr)
	var args api_container_env_var_values.ApiContainerArgs
	if err := json.Unmarshal(paramsJsonBytes, &args); err != nil {
		return nil, nil, stacktrace.Propagate(err,"An error occurred deserializing the args JSON '%v'", paramsJsonStr)
	}

	dockerManager, err := createDockerManager()
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred creating the Docker manager")
	}

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
		dockerNetworkId,
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

	lambdaStore := lambda_store.NewLambdaStore(lambdaLauncher)

	return serviceNetwork, lambdaStore, nil
}
