/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package main

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/backend_creator"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/core/launcher/args"
	"github.com/kurtosis-tech/kurtosis/core/launcher/args/kurtosis_backend_config"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/networking_sidecar"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/commons/enclave_data_directory"
	metrics_client "github.com/kurtosis-tech/metrics-library/golang/lib/client"
	"github.com/kurtosis-tech/metrics-library/golang/lib/source"
	minimal_grpc_server "github.com/kurtosis-tech/minimal-grpc-server/golang/server"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
	"os"
	"path"
	"runtime"
	"strings"
	"time"
)

const (
	successExitCode = 0
	failureExitCode = 1

	grpcServerStopGracePeriod = 5 * time.Second

	shouldFlushMetricsClientQueueOnEachEvent = false

	forceColors   = true
	fullTimestamp = true

	logMethodAlongWithLogLine = true
	functionPathSeparator     = "."
	emptyFunctionName         = ""
)

type doNothingMetricsClientCallback struct{}

func (d doNothingMetricsClientCallback) Success()          {}
func (d doNothingMetricsClientCallback) Failure(err error) {}

func main() {
	// This allows the filename & function to be reported
	logrus.SetReportCaller(logMethodAlongWithLogLine)
	// NOTE: we'll want to change the ForceColors to false if we ever want structured logging
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:               forceColors,
		DisableColors:             false,
		ForceQuote:                false,
		DisableQuote:              false,
		EnvironmentOverrideColors: false,
		DisableTimestamp:          false,
		FullTimestamp:             fullTimestamp,
		TimestampFormat:           "",
		DisableSorting:            false,
		SortingFunc:               nil,
		DisableLevelTruncation:    false,
		PadLevelText:              false,
		QuoteEmptyFields:          false,
		FieldMap:                  nil,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			fullFunctionPath := strings.Split(f.Function, functionPathSeparator)
			functionName := fullFunctionPath[len(fullFunctionPath)-1]
			_, filename := path.Split(f.File)
			return emptyFunctionName, formatFilenameFunctionForLogs(filename, functionName)
		},
	})

	err := runMain()
	if err != nil {
		logrus.Errorf("An error occurred when running the main function:")
		fmt.Fprintln(logrus.StandardLogger().Out, err)
		os.Exit(failureExitCode)
	}
	os.Exit(successExitCode)

}

func runMain() error {
	ctx := context.Background()

	serverArgs, ownIpAddress, err := args.GetArgsFromEnv()
	if err != nil {
		return stacktrace.Propagate(err, "Couldn't retrieve API container args from the environment")
	}

	logLevel, err := logrus.ParseLevel(serverArgs.LogLevel)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing the log level string '%v':", serverArgs.LogLevel)
	}
	logrus.SetLevel(logLevel)

	enclaveDataDir := enclave_data_directory.NewEnclaveDataDirectory(serverArgs.EnclaveDataVolumeDirpath)

	filesArtifactStore, err := enclaveDataDir.GetFilesArtifactStore()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the files artifact store")
	}

	clusterConfig := serverArgs.KurtosisBackendConfig
	if clusterConfig == nil {
		return stacktrace.NewError("Kurtosis backend type is '%v' but cluster configuration parameters are null.", args.KurtosisBackendType_Kubernetes.String())
	}

	gitPackageContentProvider, err := enclaveDataDir.GetGitPackageContentProvider()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while creating the Git module content provider")
	}

	// TODO Extract into own function
	var kurtosisBackend backend_interface.KurtosisBackend
	switch serverArgs.KurtosisBackendType {
	case args.KurtosisBackendType_Docker:
		apiContainerModeArgs := &backend_creator.APIContainerModeArgs{
			Context:        ctx,
			EnclaveID:      enclave.EnclaveID(serverArgs.EnclaveId),
			APIContainerIP: ownIpAddress,
		}
		kurtosisBackend, err = backend_creator.GetLocalDockerKurtosisBackend(apiContainerModeArgs)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting local Docker Kurtosis backend")
		}
	case args.KurtosisBackendType_Kubernetes:
		// TODO Use this value when we have fields for the API container
		_, ok := (clusterConfig).(kurtosis_backend_config.KubernetesBackendConfig)
		if !ok {
			return stacktrace.NewError(
				"Failed to cast untyped cluster configuration object '%+v' to the appropriate type, even though "+
					"Kurtosis backend type is '%v'",
				clusterConfig,
				args.KurtosisBackendType_Kubernetes.String(),
			)
		}
		kurtosisBackend, err = lib.GetApiContainerKubernetesKurtosisBackend(ctx)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to get a Kubernetes backend")
		}
	default:
		return stacktrace.NewError("Backend type '%v' was not recognized by API container.", serverArgs.KurtosisBackendType.String())
	}

	serviceNetwork, err := createServiceNetwork(kurtosisBackend, enclaveDataDir, serverArgs, ownIpAddress)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the service network & module store")
	}

	metricsClient, closeClientFunc, err := metrics_client.CreateMetricsClient(
		source.KurtosisCoreSource,
		serverArgs.Version,
		serverArgs.MetricsUserID,
		serverArgs.DidUserAcceptSendingMetrics,
		shouldFlushMetricsClientQueueOnEachEvent,
		doNothingMetricsClientCallback{},
	)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the metrics client")
	}
	defer func() {
		if err := closeClientFunc(); err != nil {
			logrus.Warnf("We tried to close the metrics client, but doing so threw an error:\n%v", err)
		}
	}()

	// TODO: Consolidate Interpreter, Validator and Executor into a single interface
	startosisRunner := startosis_engine.NewStartosisRunner(
		startosis_engine.NewStartosisInterpreter(serviceNetwork, gitPackageContentProvider, runtime_value_store.NewRuntimeValueStore()),
		startosis_engine.NewStartosisValidator(&kurtosisBackend, serviceNetwork),
		startosis_engine.NewStartosisExecutor())

	//Creation of ApiContainerService
	apiContainerService, err := server.NewApiContainerService(
		filesArtifactStore,
		serviceNetwork,
		startosisRunner,
		metricsClient,
		gitPackageContentProvider,
	)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the API container service")
	}

	apiContainerServiceRegistrationFunc := func(grpcServer *grpc.Server) {
		kurtosis_core_rpc_api_bindings.RegisterApiContainerServiceServer(grpcServer, apiContainerService)
	}
	apiContainerServer := minimal_grpc_server.NewMinimalGRPCServer(
		serverArgs.GrpcListenPortNum,
		grpcServerStopGracePeriod,
		[]func(*grpc.Server){
			apiContainerServiceRegistrationFunc,
		},
	)

	logrus.Info("Running server...")
	if err := apiContainerServer.RunUntilInterrupted(); err != nil {
		return stacktrace.Propagate(err, "An error occurred running the API container server")
	}

	return nil
}

func createServiceNetwork(
	kurtosisBackend backend_interface.KurtosisBackend,
	enclaveDataDir *enclave_data_directory.EnclaveDataDirectory,
	args *args.APIContainerArgs,
	ownIpAddress net.IP,
) (service_network.ServiceNetwork, error) {
	enclaveIdStr := args.EnclaveId
	enclaveId := enclave.EnclaveID(enclaveIdStr)

	/*
		filesArtifactStore, err := enclaveDataDir.GetFilesArtifactStore()
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred getting the files artifact store")
		}
	*/

	isPartitioningEnabled := args.IsPartitioningEnabled

	/*
		filesArtifactExpander := files_artifact_expander.NewFilesArtifactExpander(
			kurtosisBackend,
			enclaveObjAttrsProvider,
			enclaveId,
			filesArtifactStore,
		)
	*/

	networkingSidecarManager := networking_sidecar.NewStandardNetworkingSidecarManager(
		kurtosisBackend,
		enclaveId)

	serviceNetwork := service_network.NewDefaultServiceNetwork(
		enclaveId,
		ownIpAddress,
		args.GrpcListenPortNum,
		args.Version,
		isPartitioningEnabled,
		kurtosisBackend,
		enclaveDataDir,
		networkingSidecarManager,
	)
	return serviceNetwork, nil
}

func formatFilenameFunctionForLogs(filename string, functionName string) string {
	var output strings.Builder
	output.WriteString("[")
	output.WriteString(filename)
	output.WriteString(":")
	output.WriteString(functionName)
	output.WriteString("]")
	return output.String()
}
