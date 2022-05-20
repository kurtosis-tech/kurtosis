/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package main

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/backend_creator"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis-core/launcher/api_container_launcher"
	"github.com/kurtosis-tech/kurtosis-engine-server/api/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-engine-server/launcher/args"
	"github.com/kurtosis-tech/kurtosis-engine-server/launcher/args/kurtosis_backend_config"
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
	ctx := context.Background()

	serverArgs, err := args.GetArgsFromEnv()
	if err != nil {
		return stacktrace.Propagate(err, "Couldn't retrieve engine server args from the environment")
	}

	logLevel, err := logrus.ParseLevel(serverArgs.LogLevelStr)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing the log level string '%v':", serverArgs.LogLevelStr)
	}
	logrus.SetLevel(logLevel)

	backendConfig := serverArgs.KurtosisBackendConfig
	if backendConfig == nil {
		return stacktrace.NewError("Backend configuration parameters are null - there must be backend configuration parameters.")
	}

	enclaveManager, err := getEnclaveManager(ctx, serverArgs.KurtosisBackendType, backendConfig)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to create an enclave manager for backend type '%v' and config '%+v'", serverArgs.KurtosisBackendType, backendConfig)
	}

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

func getEnclaveManager(ctx context.Context, kurtosisBackendType args.KurtosisBackendType, backendConfig interface{}) (*enclave_manager.EnclaveManager, error){
	var kurtosisBackend backend_interface.KurtosisBackend
	var err error
	var apiContainerKurtosisBackendConfigSupplier api_container_launcher.KurtosisBackendConfigSupplier
	switch kurtosisBackendType {
	case args.KurtosisBackendType_Docker:
		kurtosisBackend, err = backend_creator.GetLocalDockerKurtosisBackend(apiContainerModeArgsForKurtosisBackend)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting local Docker Kurtosis backend")
		}
		apiContainerKurtosisBackendConfigSupplier = api_container_launcher.NewDockerKurtosisBackendConfigSupplier()
	case args.KurtosisBackendType_Kubernetes:
		kubernetesBackendConfig, ok := (backendConfig).(kurtosis_backend_config.KubernetesBackendConfig)
		if !ok {
			return nil, stacktrace.NewError("Failed to cast cluster configuration interface to the appropriate type, even though Kurtosis backend type is '%v'", args.KurtosisBackendType_Kubernetes.String())
		}
		kurtosisBackend, err = lib.GetEngineServerKubernetesKurtosisBackend(
			ctx,
			kubernetesBackendConfig.StorageClass,
			kubernetesBackendConfig.EnclaveSizeInMegabytes,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to get a Kubernetes backend with storage class '%v' and enclave size (in MB) %d", kubernetesBackendConfig.StorageClass, kubernetesBackendConfig.EnclaveSizeInMegabytes)
		}
		apiContainerKurtosisBackendConfigSupplier = api_container_launcher.NewKubernetesKurtosisBackendConfigSupplier(kubernetesBackendConfig.StorageClass, kubernetesBackendConfig.EnclaveSizeInMegabytes)
	default:
		return nil, stacktrace.NewError("Backend type '%v' was not recognized by engine server.", kurtosisBackendType.String())
	}

	enclaveManager := enclave_manager.NewEnclaveManager(kurtosisBackend, apiContainerKurtosisBackendConfigSupplier)

	return enclaveManager, nil
}