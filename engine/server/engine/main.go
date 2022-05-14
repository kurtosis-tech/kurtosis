/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package main

import (
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/backend_creator"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis-engine-server/api/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-engine-server/launcher/args"
	"github.com/kurtosis-tech/kurtosis-engine-server/launcher/args/kurtosis_backend_type"
	"github.com/kurtosis-tech/kurtosis-engine-server/launcher/args/kurtosis_cluster_config"
	"github.com/kurtosis-tech/kurtosis-engine-server/server/engine/enclave_manager"
	"github.com/kurtosis-tech/kurtosis-engine-server/server/engine/server"
	metrics_client "github.com/kurtosis-tech/metrics-library/golang/lib/client"
	"github.com/kurtosis-tech/metrics-library/golang/lib/source"
	minimal_grpc_server "github.com/kurtosis-tech/minimal-grpc-server/golang/server"
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

// Nil indicates that the KurtosisBackend should not operate in API container mode, which is appropriate here
//  because this isn't the API container
var apiContainerModeArgsForKurtosisBackend *backend_creator.APIContainerModeArgs = nil

type doNothingMetricsClientCallback struct{}

func (d doNothingMetricsClientCallback) Success()          {}
func (d doNothingMetricsClientCallback) Failure(err error) {}

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

func runMain() error {
	serverArgs, err := args.GetArgsFromEnv()
	if err != nil {
		return stacktrace.Propagate(err, "Couldn't retrieve engine server args from the environment")
	}

	logLevel, err := logrus.ParseLevel(serverArgs.LogLevelStr)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing the log level string '%v':", serverArgs.LogLevelStr)
	}
	logrus.SetLevel(logLevel)

	kurtosisBackendType := serverArgs.KurtosisBackendType
	var kurtosisBackend backend_interface.KurtosisBackend
	switch kurtosisBackendType {
	case kurtosis_backend_type.Docker:
		kurtosisBackend, err = backend_creator.GetLocalDockerKurtosisBackend(apiContainerModeArgsForKurtosisBackend)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting a Kurtosis backend connected to local Docker")
		}
	case kurtosis_backend_type.Kubernetes:
		clusterConfig := serverArgs.KurtosisClusterConfig
		if clusterConfig == nil {
			return stacktrace.NewError("Kurtosis backend type is '%v' but cluster configuration parameters are null.", kurtosis_backend_type.Kubernetes.String())
		}
		kubernetesClusterConfig, ok := (*clusterConfig).(kurtosis_cluster_config.KurtosisClusterConfig)
		if !ok {
			return stacktrace.NewError("Failed to cast cluster configuration interface to KurtosisClusterConfig, even though Kurtosis backend type is '%v'", kurtosis_backend_type.Kubernetes.String())
		}
		kurtosisBackend, err = lib.GetLocalKubernetesKurtosisBackend(kubernetesClusterConfig.StorageClass, kubernetesClusterConfig.EnclaveSizeInGigabytes)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to build a Kubernetes backend with storage class '%v' and enclave size (in GB) %d", kubernetesClusterConfig.StorageClass, kubernetesClusterConfig.EnclaveSizeInGigabytes)
		}
	}

	enclaveManager := enclave_manager.NewEnclaveManager(kurtosisBackend)

	metricsClient, metricsClientCloseFunc, err := metrics_client.CreateMetricsClient(
		source.KurtosisEngineSource,
		serverArgs.ImageVersionTag,
		serverArgs.MetricsUserID,
		serverArgs.DidUserAcceptSendingMetrics,
		shouldFlushMetricsClientQueueOnEachEvent,
		doNothingMetricsClientCallback{},
	)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the metrics client")
	}
	defer func() {
		if err := metricsClientCloseFunc(); err != nil {
			logrus.Warnf("We tried to close the metrics client, but doing so threw an error:\n%v", err)
		}
	}()

	engineServerService := server.NewEngineServerService(serverArgs.ImageVersionTag, enclaveManager, metricsClient, serverArgs.MetricsUserID, serverArgs.DidUserAcceptSendingMetrics)

	engineServerServiceRegistrationFunc := func(grpcServer *grpc.Server) {
		kurtosis_engine_rpc_api_bindings.RegisterEngineServiceServer(grpcServer, engineServerService)
	}
	engineServer := minimal_grpc_server.NewMinimalGRPCServer(
		serverArgs.GrpcListenPortNum,
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
