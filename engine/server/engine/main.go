/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package main

import (
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-engine-server/api/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-engine-server/launcher/args"
	"github.com/kurtosis-tech/kurtosis-engine-server/launcher/engine_server_launcher"
	"github.com/kurtosis-tech/kurtosis-engine-server/server/engine/enclave_manager"
	"github.com/kurtosis-tech/kurtosis-engine-server/server/engine/server"
	metrics_client "github.com/kurtosis-tech/metrics-library/golang/lib/client"
	"github.com/kurtosis-tech/metrics-library/golang/lib/source"
	minimal_grpc_server "github.com/kurtosis-tech/minimal-grpc-server/golang/server"
	"github.com/kurtosis-tech/object-attributes-schema-lib/schema"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"os"
	"time"
)

const (
	successExitCode = 0
	failureExitCode = 1

	grpcServerStopGracePeriod = 5 * time.Second

	shouldFlushMetricsClientQueueOnEachEvent = false
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
	serverArgs, err := args.GetArgsFromEnv()
	if err != nil {
		return stacktrace.Propagate(err, "Couldn't retrieve engine server args from the environment")
	}

	logLevel, err := logrus.ParseLevel(serverArgs.LogLevelStr)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing the log level string '%v':", serverArgs.LogLevelStr)
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

	objAttrsProvider := schema.GetObjectAttributesProvider()
	enclaveManager := enclave_manager.NewEnclaveManager(
		dockerManager,
		objAttrsProvider,
		serverArgs.EngineDataDirpathOnHostMachine,
		engine_server_launcher.EngineDataDirpathOnEngineServerContainer,
	)

	metricsClient, err := metrics_client.CreateMetricsClient(source.KurtosisEngineSource, engine_server_launcher.KurtosisEngineVersion, serverArgs.MetricsUserID, serverArgs.DidUserAcceptSendingMetrics, shouldFlushMetricsClientQueueOnEachEvent)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the metrics client")
	}
	defer func() {
		if err := metricsClient.Close(); err != nil {
			logrus.Warnf("We tried to close the metrics client, but doing so threw an error:\n%v", err)
		}
	}()

	engineServerService := server.NewEngineServerService(serverArgs.ImageVersionTag, enclaveManager, metricsClient, serverArgs.MetricsUserID, serverArgs.DidUserAcceptSendingMetrics)

	engineServerServiceRegistrationFunc := func(grpcServer *grpc.Server) {
		kurtosis_engine_rpc_api_bindings.RegisterEngineServiceServer(grpcServer, engineServerService)
	}
	engineServer := minimal_grpc_server.NewMinimalGRPCServer(
		serverArgs.ListenPortNum,
		grpcServerStopGracePeriod,
		[]func(*grpc.Server){
			engineServerServiceRegistrationFunc,
		},
	)

	logrus.Info("Running server...")
	if err := engineServer.RunUntilInterrupted(); err != nil {
		return stacktrace.Propagate(err, "An error occurred running the server.")
	}
	return nil
}
