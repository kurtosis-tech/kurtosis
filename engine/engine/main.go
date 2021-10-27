/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package main

import (
	"encoding/json"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/golang/kurtosis_engine_rpc_api_consts"
	"github.com/kurtosis-tech/kurtosis-engine-server/engine/enclave_manager"
	"github.com/kurtosis-tech/kurtosis-engine-server/engine/kurtosis_engine_server_docker_api"
	"github.com/kurtosis-tech/kurtosis-engine-server/engine/server"
	minimal_grpc_server "github.com/kurtosis-tech/minimal-grpc-server/golang/server"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
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
		logrus.Errorf("An error occurred when running the main function")
		fmt.Fprintln(logrus.StandardLogger().Out, err)
		os.Exit(failureExitCode)
	}
	os.Exit(successExitCode)
}

func runMain () error {
	serializedArgsStr, found := os.LookupEnv(kurtosis_engine_server_docker_api.SerializedArgsEnvVar)
	if !found {
		return stacktrace.NewError(
			"Expected environment variable '%v' containing serialized args to the engine container, but none was found",
			kurtosis_engine_server_docker_api.SerializedArgsEnvVar,
		)
	}
	var args kurtosis_engine_server_docker_api.EngineServerArgs
	if err := json.Unmarshal([]byte(serializedArgsStr), &args); err != nil {
		return stacktrace.Propagate(err, "An error occurred deserializing serialized args string '%v'", serializedArgsStr)
	}

	logLevel, err := logrus.ParseLevel(args.LogLevelStr)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing log level string '%v'", args.LogLevelStr)
	}
	logrus.SetLevel(logLevel)

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the Docker client")
	}
	dockerManager := docker_manager.NewDockerManager(
		logrus.StandardLogger(),
		dockerClient,
	)
	enclaveManager := enclave_manager.NewEnclaveManager(dockerManager, args.EngineDataDirpathOnHostMachine, kurtosis_engine_server_docker_api.EngineDataDirpathOnEngineContainer)

	engineServerService := server.NewEngineServerService(enclaveManager)

	engineServerServiceRegistrationFunc := func(grpcServer *grpc.Server) {
		kurtosis_engine_rpc_api_bindings.RegisterEngineServiceServer(grpcServer, engineServerService)
	}
	engineServer := minimal_grpc_server.NewMinimalGRPCServer(
		uint32(kurtosis_engine_rpc_api_consts.ListenPort),
		kurtosis_engine_rpc_api_consts.ListenProtocol,
		grpcServerStopGracePeriod,
		[]func(*grpc.Server){
			engineServerServiceRegistrationFunc,
		},
	)

	logrus.Info("Running server...")
	if err := engineServer.Run(); err != nil {
		return stacktrace.Propagate(err, "An error occurred running the server.")
	}
	return nil
}
