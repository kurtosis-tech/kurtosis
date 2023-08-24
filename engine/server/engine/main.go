/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package main

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings/kurtosis_engine_rpc_api_bindingsconnect"
	connect_server "github.com/kurtosis-tech/kurtosis/connect-server"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/backend_creator"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/configs"
	"github.com/kurtosis-tech/kurtosis/core/launcher/api_container_launcher"
	em_api "github.com/kurtosis-tech/kurtosis/enclave-manager/server"
	"github.com/kurtosis-tech/kurtosis/engine/launcher/args"
	"github.com/kurtosis-tech/kurtosis/engine/launcher/args/kurtosis_backend_config"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/enclave_manager"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/server"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
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
	webappPortAddr            = ":9711"

	remoteBackendConfigFilename = "remote_backend_config.json"
	pathToStaticFolder          = "/run/webapp"
	indexPath                   = "index.html"
)

// Nil indicates that the KurtosisBackend should not operate in API container mode, which is appropriate here
//
//	because this isn't the API container
var apiContainerModeArgsForKurtosisBackend *backend_creator.APIContainerModeArgs = nil

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

	backendConfig := serverArgs.KurtosisLocalBackendConfig
	if backendConfig == nil {
		return stacktrace.NewError("Backend configuration parameters are null - there must be backend configuration parameters.")
	}

	var remoteBackendConfigMaybe *configs.KurtosisRemoteBackendConfig
	if serverArgs.OnBastionHost {
		// Read remote backend config from the local filesystem
		remoteBackendConfigPath := filepath.Join(consts.EngineConfigLocalDir, remoteBackendConfigFilename)
		remoteBackendConfigBytes, err := os.ReadFile(remoteBackendConfigPath)
		if err != nil {
			return stacktrace.Propagate(err, "The remote backend config '%s' cannot be found", remoteBackendConfigPath)
		}
		remoteBackendConfigMaybe, err = configs.NewRemoteBackendConfigFromJSON(remoteBackendConfigBytes)
		if err != nil {
			return stacktrace.Propagate(err, "The remote backend config '%s' is not valid JSON", remoteBackendConfigPath)
		}
	}

	kurtosisBackend, err := getKurtosisBackend(ctx, serverArgs.KurtosisBackendType, backendConfig, remoteBackendConfigMaybe)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the Kurtosis backend for backend type '%v' and config '%+v'", serverArgs.KurtosisBackendType, backendConfig)
	}

	enclaveManager, err := getEnclaveManager(kurtosisBackend, serverArgs.KurtosisBackendType, serverArgs.ImageVersionTag, serverArgs.PoolSize, serverArgs.EnclaveEnvVars)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to create an enclave manager for backend type '%v' and config '%+v'", serverArgs.KurtosisBackendType, backendConfig)
	}

	// osFs is a wrapper around disk
	osFs := persistent_volume.NewOsVolumeFilesystem()
	logsDatabaseClient := persistent_volume.NewPersistentVolumeLogsDatabaseClient(kurtosisBackend, osFs)

	go func() {
		fileServer := http.FileServer(http.Dir(pathToStaticFolder))
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path, err := filepath.Abs(r.URL.Path)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			path = filepath.Join(pathToStaticFolder, path)

			_, err = os.Stat(path)
			if os.IsNotExist(err) {
				// file does not exist, serve index.html
				http.ServeFile(w, r, filepath.Join(pathToStaticFolder, indexPath))
				return
			} else if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			fileServer.ServeHTTP(w, r)
		})

		err := http.ListenAndServe(webappPortAddr, handler)
		if err != nil {
			logrus.Debugf("error while starting the webapp: \n%v", err)
		}
	}()

	go func() {
		err = em_api.RunEnclaveManagerApiServer()
		if err != nil {
			logrus.Fatal("an error occurred while processing the auth settings, exiting!", err)
			fmt.Fprintln(logrus.StandardLogger().Out, err)
			os.Exit(failureExitCode)
		}
	}()

	engineConnectServer := server.NewEngineConnectServerService(serverArgs.ImageVersionTag, enclaveManager, serverArgs.MetricsUserID, serverArgs.DidUserAcceptSendingMetrics, logsDatabaseClient)
	apiPath, handler := kurtosis_engine_rpc_api_bindingsconnect.NewEngineServiceHandler(engineConnectServer)

	logrus.Info("Running server...")
	engineHttpServer := connect_server.NewConnectServer(serverArgs.GrpcListenPortNum, grpcServerStopGracePeriod, handler, apiPath)
	if err := engineHttpServer.RunServerUntilInterruptedWithCors(cors.AllowAll()); err != nil {
		return stacktrace.Propagate(err, "An error occurred running the server.")
	}
	return nil
}

func getEnclaveManager(
	kurtosisBackend backend_interface.KurtosisBackend,
	kurtosisBackendType args.KurtosisBackendType,
	engineVersion string,
	poolSize uint8,
	enclaveEnvVars string,
) (*enclave_manager.EnclaveManager, error) {
	var apiContainerKurtosisBackendConfigSupplier api_container_launcher.KurtosisBackendConfigSupplier
	switch kurtosisBackendType {
	case args.KurtosisBackendType_Docker:
		apiContainerKurtosisBackendConfigSupplier = api_container_launcher.NewDockerKurtosisBackendConfigSupplier()
	case args.KurtosisBackendType_Kubernetes:
		apiContainerKurtosisBackendConfigSupplier = api_container_launcher.NewKubernetesKurtosisBackendConfigSupplier()
	default:
		return nil, stacktrace.NewError("Backend type '%v' was not recognized by engine server.", kurtosisBackendType.String())
	}

	enclaveManager, err := enclave_manager.CreateEnclaveManager(
		kurtosisBackend,
		kurtosisBackendType,
		apiContainerKurtosisBackendConfigSupplier,
		engineVersion,
		poolSize,
		enclaveEnvVars,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating enclave manager for backend type '%+v' using pool-size '%v' and engine version '%v'", kurtosisBackendType, poolSize, engineVersion)
	}

	return enclaveManager, nil
}

func getKurtosisBackend(ctx context.Context, kurtosisBackendType args.KurtosisBackendType, backendConfig interface{}, remoteBackendConfigMaybe *configs.KurtosisRemoteBackendConfig) (backend_interface.KurtosisBackend, error) {
	var kurtosisBackend backend_interface.KurtosisBackend
	var err error
	switch kurtosisBackendType {
	case args.KurtosisBackendType_Docker:
		kurtosisBackend, err = backend_creator.GetDockerKurtosisBackend(apiContainerModeArgsForKurtosisBackend, remoteBackendConfigMaybe)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting local Docker Kurtosis backend")
		}
	case args.KurtosisBackendType_Kubernetes:
		if remoteBackendConfigMaybe != nil {
			return nil, stacktrace.NewError("Using a Remote Kurtosis Backend isn't allowed with Kubernetes. " +
				"Either switch to a local only context to use Kubernetes or switch the cluster to Docker to " +
				"connect to a remote Kurtosis backend")
		}
		// Use this with more properties
		_, ok := (backendConfig).(kurtosis_backend_config.KubernetesBackendConfig)
		if !ok {
			return nil, stacktrace.NewError("Failed to cast cluster configuration interface to the appropriate type, even though Kurtosis backend type is '%v'", args.KurtosisBackendType_Kubernetes.String())
		}
		kurtosisBackend, err = kubernetes_kurtosis_backend.GetEngineServerBackend(ctx)
		if err != nil {
			return nil, stacktrace.Propagate(
				err,
				"An error occurred getting Kurtosis Kubernetes backend for engine",
			)
		}
	default:
		return nil, stacktrace.NewError("Backend type '%v' was not recognized by engine server.", kurtosisBackendType.String())
	}

	return kurtosisBackend, nil
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
