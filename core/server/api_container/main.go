/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package main

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/interpretation_time_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/starlark_run"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages/git_package_content_provider"
	"net"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/backend_creator"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/configs"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/kurtosis/core/launcher/args"
	"github.com/kurtosis-tech/kurtosis/core/launcher/args/kurtosis_backend_config"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_plan_persistence"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/commons/enclave_data_directory"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/analytics_logger"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/metrics_client"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/source"
	minimal_grpc_server "github.com/kurtosis-tech/minimal-grpc-server/golang/server"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"go.starlark.net/starlark"
	"google.golang.org/grpc"
)

const (
	successExitCode = 0
	failureExitCode = 1

	grpcServerStopGracePeriod = 5 * time.Second

	forceColors   = true
	fullTimestamp = true

	logMethodAlongWithLogLine = true
	functionPathSeparator     = "."
	emptyFunctionName         = ""

	shouldFlushMetricsClientQueueOnEachEvent = false
)

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

	clusterConfig := serverArgs.KurtosisBackendConfig
	if clusterConfig == nil {
		return stacktrace.NewError("Kurtosis backend type is '%v' but cluster configuration parameters are null.", args.KurtosisBackendType_Kubernetes.String())
	}

	repositoriesDirPath, tempDirectoriesDirPath, githubAuthDirPath, enclaveDatabaseDirpath, err := enclaveDataDir.GetEnclaveDataDirectoryPaths()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting directory paths of the enclave data directory.")
	}

	enclaveDb, err := enclave_db.GetOrCreateEnclaveDatabase(enclaveDatabaseDirpath)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while getting the enclave db")
	}

	filesArtifactStore, err := enclaveDataDir.GetFilesArtifactStore()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the files artifact store")
	}

	githubAuthProvider := git_package_content_provider.NewGitHubPackageAuthProvider(githubAuthDirPath)
	gitPackageContentProvider := git_package_content_provider.NewGitPackageContentProvider(repositoriesDirPath, tempDirectoriesDirPath, githubAuthProvider, enclaveDb)

	// TODO Extract into own function
	var kurtosisBackend backend_interface.KurtosisBackend
	switch serverArgs.KurtosisBackendType {
	case args.KurtosisBackendType_Docker:
		apiContainerModeArgs := &backend_creator.APIContainerModeArgs{
			Context:        ctx,
			EnclaveID:      enclave.EnclaveUUID(serverArgs.EnclaveUUID),
			APIContainerIP: ownIpAddress,
			IsProduction:   serverArgs.IsProductionEnclave,
		}
		kurtosisBackend, err = backend_creator.GetDockerKurtosisBackend(apiContainerModeArgs, configs.NoRemoteBackendConfig)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting local Docker Kurtosis backend")
		}
	case args.KurtosisBackendType_Kubernetes:
		// TODO Use this value when we have fields for the API container
		clusterConfigK8s, ok := (clusterConfig).(kurtosis_backend_config.KubernetesBackendConfig)
		if !ok {
			return stacktrace.NewError(
				"Failed to cast untyped cluster configuration object '%+v' to the appropriate type, even though "+
					"Kurtosis backend type is '%v'",
				clusterConfig,
				args.KurtosisBackendType_Kubernetes.String(),
			)
		}
		// TODO wrap up APIContainerModeArgs if the parameter list keeps on going up (currently just IsProductionEnclave)
		kurtosisBackend, err = kubernetes_kurtosis_backend.GetApiContainerBackend(ctx, clusterConfigK8s.StorageClass, serverArgs.IsProductionEnclave)
		if err != nil {
			return stacktrace.Propagate(
				err,
				"An error occurred getting Kurtosis Kubernetes backend for APIC",
			)
		}
	default:
		return stacktrace.NewError("Backend type '%v' was not recognized by API container.", serverArgs.KurtosisBackendType.String())
	}

	serviceNetwork, err := createServiceNetwork(kurtosisBackend, enclaveDataDir, serverArgs, ownIpAddress, enclaveDb)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the service network")
	}

	logger := logrus.StandardLogger()
	metricsClient, closeClientFunc, err := metrics_client.CreateMetricsClient(
		metrics_client.NewMetricsClientCreatorOption(
			source.KurtosisCoreSource,
			serverArgs.Version,
			serverArgs.MetricsUserID,
			serverArgs.KurtosisBackendType.String(),
			serverArgs.DidUserAcceptSendingMetrics,
			shouldFlushMetricsClientQueueOnEachEvent,
			metrics_client.DoNothingMetricsClientCallback{},
			analytics_logger.ConvertLogrusLoggerToAnalyticsLogger(logger),
			serverArgs.IsCI,
			serverArgs.CloudUserID,
			serverArgs.CloudInstanceID,
		),
	)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the metrics client")
	}
	defer func() {
		if err := closeClientFunc(); err != nil {
			logrus.Warnf("We tried to close the metrics client, but doing so threw an error:\n%v", err)
		}
	}()

	starlarkValueSerde := createStarlarkValueSerde()
	runtimeValueStore, err := runtime_value_store.CreateRuntimeValueStore(starlarkValueSerde, enclaveDb)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the runtime value store")
	}

	interpretationTimeValueStore, err := interpretation_time_value_store.CreateInterpretationTimeValueStore(enclaveDb, starlarkValueSerde)
	if err != nil {
		return stacktrace.Propagate(err, "an error occurred while creating the interpretation time value store")
	}

	// Load the current enclave plan, in case the enclave is being restarted
	enclavePlan, err := enclave_plan_persistence.Load(enclaveDb)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred loading stored enclave plan")
	}

	// TODO: Consolidate Interpreter, Validator and Executor into a single interface
	startosisInterpreter := startosis_engine.NewStartosisInterpreter(serviceNetwork, gitPackageContentProvider, runtimeValueStore, starlarkValueSerde, serverArgs.EnclaveEnvVars, interpretationTimeValueStore)
	startosisRunner := startosis_engine.NewStartosisRunner(
		startosisInterpreter,
		startosis_engine.NewStartosisValidator(&kurtosisBackend, serviceNetwork, filesArtifactStore),
		startosis_engine.NewStartosisExecutor(starlarkValueSerde, runtimeValueStore, enclavePlan, enclaveDb))

	starlarkRunRepository, err := starlark_run.GetOrCreateNewStarlarkRunRepository(enclaveDb)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the starlark run repository")
	}

	//Creation of ApiContainerService
	restartPolicy := kurtosis_core_rpc_api_bindings.RestartPolicy_NEVER
	if serverArgs.IsProductionEnclave {
		restartPolicy = kurtosis_core_rpc_api_bindings.RestartPolicy_ALWAYS
	}
	apiContainerService, err := server.NewApiContainerService(
		filesArtifactStore,
		serviceNetwork,
		startosisRunner,
		startosisInterpreter,
		gitPackageContentProvider,
		restartPolicy,
		metricsClient,
		githubAuthProvider,
		starlarkRunRepository,
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
	enclaveDb *enclave_db.EnclaveDB,
) (service_network.ServiceNetwork, error) {
	enclaveIdStr := args.EnclaveUUID
	enclaveUuid := enclave.EnclaveUUID(enclaveIdStr)

	apiContainerInfo := service_network.NewApiContainerInfo(
		ownIpAddress,
		args.GrpcListenPortNum,
		args.Version,
	)

	serviceNetwork, err := service_network.NewDefaultServiceNetwork(
		enclaveUuid,
		apiContainerInfo,
		kurtosisBackend,
		enclaveDataDir,
		enclaveDb,
	)

	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the default service network")
	}
	return serviceNetwork, nil
}

func createStarlarkValueSerde() *kurtosis_types.StarlarkValueSerde {
	starlarkThread := &starlark.Thread{
		Name:       "starlark-serde-thread",
		Print:      nil,
		Load:       nil,
		OnMaxSteps: nil,
		Steps:      0,
	}
	starlarkEnv := startosis_engine.Predeclared()
	builtins := startosis_engine.KurtosisTypeConstructors()
	for _, builtin := range builtins {
		starlarkEnv[builtin.Name()] = builtin
	}
	return kurtosis_types.NewStarlarkValueSerde(starlarkThread, starlarkEnv)
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
