/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package main

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/backend_creator"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/kurtosis/core/launcher/api_container_launcher"
	"github.com/kurtosis-tech/kurtosis/engine/launcher/args"
	"github.com/kurtosis-tech/kurtosis/engine/launcher/args/kurtosis_backend_config"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/enclave_manager"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/server"
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
//
//	because this isn't the API container
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

	kurtosisBackend, err := getKurtosisBackend(ctx, serverArgs.KurtosisBackendType, backendConfig)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the Kurtosis backend for backend type '%v' and config '%+v'", serverArgs.KurtosisBackendType, backendConfig)
	}

	enclaveManager, err := getEnclaveManager(kurtosisBackend, serverArgs.KurtosisBackendType)
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

	var logsDatabaseClient centralized_logs.LogsDatabaseClient

	//TODO this is a hack until we completely finish the centralized logs for Kubernetes Backend
	//TODO Create the logs database client depending on the Kurtosis Backend type (momentarily)
	//TODO Then should be only one way, the one which uses the Loki's server
	if serverArgs.KurtosisBackendType == args.KurtosisBackendType_Docker {

		logsDatabase, err := kurtosisBackend.GetLogsDatabase(ctx)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting the logs database")
		}

		if logsDatabase == nil || logsDatabase.GetStatus() == container_status.ContainerStatus_Stopped {
			return stacktrace.NewError("It's not possible to run the engine serve due the current logs database container is not running")
		}

		privateLogsDatabaseAddress := fmt.Sprintf("%v:%v", logsDatabase.GetMaybePrivateIpAddr(), logsDatabase.GetPrivateHttpPort())

		logsDatabaseClient = centralized_logs.NewLokiLogsDatabaseClientWithDefaultHttpClient(privateLogsDatabaseAddress)

	} else {
		logsDatabaseClient = centralized_logs.NewKurtosisBackendLogClient(kurtosisBackend)
	}

	engineServerService := server.NewEngineServerService(serverArgs.ImageVersionTag, enclaveManager, metricsClient, serverArgs.MetricsUserID, serverArgs.DidUserAcceptSendingMetrics, logsDatabaseClient)

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

func getEnclaveManager(kurtosisBackend backend_interface.KurtosisBackend, kurtosisBackendType args.KurtosisBackendType) (*enclave_manager.EnclaveManager, error) {
	var apiContainerKurtosisBackendConfigSupplier api_container_launcher.KurtosisBackendConfigSupplier
	switch kurtosisBackendType {
	case args.KurtosisBackendType_Docker:
		apiContainerKurtosisBackendConfigSupplier = api_container_launcher.NewDockerKurtosisBackendConfigSupplier()
	case args.KurtosisBackendType_Kubernetes:
		apiContainerKurtosisBackendConfigSupplier = api_container_launcher.NewKubernetesKurtosisBackendConfigSupplier()
	default:
		return nil, stacktrace.NewError("Backend type '%v' was not recognized by engine server.", kurtosisBackendType.String())
	}

	enclaveManager := enclave_manager.NewEnclaveManager(kurtosisBackend, apiContainerKurtosisBackendConfigSupplier)

	return enclaveManager, nil
}


func getKurtosisBackend(ctx context.Context, kurtosisBackendType args.KurtosisBackendType, backendConfig interface{}) (backend_interface.KurtosisBackend, error) {
	var kurtosisBackend backend_interface.KurtosisBackend
	var err error
	switch kurtosisBackendType {
	case args.KurtosisBackendType_Docker:
		kurtosisBackend, err = backend_creator.GetLocalDockerKurtosisBackend(apiContainerModeArgsForKurtosisBackend)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting local Docker Kurtosis backend")
		}
	case args.KurtosisBackendType_Kubernetes:
		// Use this with more properties
		_, ok := (backendConfig).(kurtosis_backend_config.KubernetesBackendConfig)
		if !ok {
			return nil, stacktrace.NewError("Failed to cast cluster configuration interface to the appropriate type, even though Kurtosis backend type is '%v'", args.KurtosisBackendType_Kubernetes.String())
		}
		kurtosisBackend, err = lib.GetEngineServerKubernetesKurtosisBackend(ctx)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to get a Kubernetes backend")
		}
	default:
		return nil, stacktrace.NewError("Backend type '%v' was not recognized by engine server.", kurtosisBackendType.String())
	}

	return kurtosisBackend, nil
}