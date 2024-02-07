package engine_manager

import (
	"context"
	"strings"
	"time"

	portal_constructors "github.com/kurtosis-tech/kurtosis-portal/api/golang/constructors"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/kurtosis_config_getter"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/portal_manager"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_cluster_setting"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/resolved_config"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/kurtosis/contexts-config-store/store"
	"github.com/kurtosis-tech/kurtosis/engine/launcher/engine_server_launcher"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	waitForEngineResponseTimeout = 5 * time.Second
	defaultClusterName           = resolved_config.DefaultDockerClusterName

	defaultEngineVersion          = ""
	waitUntilEngineStoppedTries   = 5
	waitUntilEngineStoppedCoolOff = 5 * time.Second

	doNotStartTheEngineInDebugModeForDefaultVersion = false
)

type EngineManager struct {
	kurtosisBackend                           backend_interface.KurtosisBackend
	shouldSendMetrics                         bool
	engineServerKurtosisBackendConfigSupplier engine_server_launcher.KurtosisBackendConfigSupplier
	clusterConfig                             *resolved_config.KurtosisClusterConfig
	onBastionHost                             bool
	enclaveEnvVars                            string
	allowedCORSOrigins                        *[]string
	// Make engine IP, port, and protocol configurable in the future
}

// TODO It's really weird that we have a context getting passed in to a constructor, but we have to do this
//
//	because we're currently creating the KurtosisBackend right here. The right way to fix this is have the
//	engine manager use the currently-set cluster information to:
//	1) check if it's Kubernetes or Docker
//	2) if it's Docker, try to start an engine in case one doesn't exist
//	3) if it's Kubernets, if creating the EngienClient fails then print the "you need to start a gateway" command
func NewEngineManager(ctx context.Context) (*EngineManager, error) {
	clusterSettingStore := kurtosis_cluster_setting.GetKurtosisClusterSettingStore()

	isClusterSet, err := clusterSettingStore.HasClusterSetting()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to check if cluster setting has been set.")
	}
	var clusterName string
	if !isClusterSet {
		// If the user has not yet set a cluster, use default
		clusterName = defaultClusterName
	} else {
		clusterName, err = clusterSettingStore.GetClusterSetting()
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to get cluster setting.")
		}
	}
	kurtosisConfig, err := getKurtosisConfig()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the Kurtosis config")
	}
	clusterConfig, err := kurtosis_config_getter.GetKurtosisClusterConfig()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the Kurtosis cluster config")
	}

	kurtosisBackend, err := clusterConfig.GetKurtosisBackend(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the Kurtosis backend for cluster '%v'", clusterName)
	}
	engineBackendConfigSupplier := clusterConfig.GetEngineBackendConfigSupplier()

	onBastionHost := false
	var enclaveEnvVars string
	currentContext, _ := store.GetContextsConfigStore().GetCurrentContext()
	if currentContext != nil {
		if store.IsRemote(currentContext) {
			onBastionHost = true
			enclaveEnvVars = currentContext.GetRemoteContextV0().GetEnvVars()
		}
	}

	return &EngineManager{
		kurtosisBackend:   kurtosisBackend,
		shouldSendMetrics: kurtosisConfig.GetShouldSendMetrics(),
		engineServerKurtosisBackendConfigSupplier: engineBackendConfigSupplier,
		clusterConfig:      clusterConfig,
		onBastionHost:      onBastionHost,
		enclaveEnvVars:     enclaveEnvVars,
		allowedCORSOrigins: nil,
	}, nil
}

// TODO This is a huge hack, that's only here temporarily because we have commands that use KurtosisBackend directly (they
//
//	should not), and EngineConsumingKurtosisCommand therefore needs to provide them with a KurtosisBackend. Once all our
//	commands only access the Kurtosis APIs, we can remove this.
func (manager *EngineManager) GetKurtosisBackend() backend_interface.KurtosisBackend {
	return manager.kurtosisBackend
}

// GetEngineStatus Returns:
//   - The engine status
//   - The host machine port bindings (not present if the engine is stopped)
//   - The engine version (only present if the engine is running)
func (manager *EngineManager) GetEngineStatus(
	ctx context.Context,
) (EngineStatus, *hostMachineIpAndPort, string, error) {
	runningEngineContainers, err := manager.kurtosisBackend.GetEngines(ctx, getRunningEnginesFilter())
	if err != nil {
		return "", nil, "", stacktrace.Propagate(err, "An error occurred getting Kurtosis engine containers")
	}

	numRunningEngineContainers := len(runningEngineContainers)
	if numRunningEngineContainers > 1 {
		return "", nil, "", stacktrace.NewError("Cannot report engine status because we found %v running Kurtosis engine containers; this is very strange as there should never be more than one", numRunningEngineContainers)
	} else if numRunningEngineContainers == 0 {
		return EngineStatus_Stopped, nil, "", nil
	}

	// TODO Replace this hacky method of defaulting to localhost:DefaultGrpcPort to get connected to the engine
	runningEngineIpAndPort := getDefaultKurtosisEngineLocalhostMachineIpAndPort()

	currentContext, err := store.GetContextsConfigStore().GetCurrentContext()
	if err == nil {
		if store.IsRemote(currentContext) {
			portalManager := portal_manager.NewPortalManager()
			if portalManager.IsReachable() {
				// Forward the remote engine port to the local machine
				portalClient := portalManager.GetClient()
				forwardEnginePortArgs := portal_constructors.NewForwardPortArgs(uint32(runningEngineIpAndPort.portNum), uint32(runningEngineIpAndPort.portNum), kurtosis_context.EngineRemoteEndpointType, &kurtosis_context.EnginePortTransportProtocol, &kurtosis_context.ForwardPortWaitUntilReady)
				if _, err := portalClient.ForwardPort(ctx, forwardEnginePortArgs); err != nil {
					return "", nil, "", stacktrace.Propagate(err, "Unable to forward the remote engine port to the local machine")
				}
			}
		}
	} else {
		logrus.Warnf("Unable to retrieve current Kurtosis context. This is not critical, it will assume using Kurtosis default context for now.")
	}

	engineClient, engineClientCloseFunc, err := getEngineClientFromHostMachineIpAndPort(runningEngineIpAndPort)
	if err != nil {
		return EngineStatus_ContainerRunningButServerNotResponding, runningEngineIpAndPort, "", nil
	}
	defer func() {
		if err = engineClientCloseFunc(); err != nil {
			logrus.Warnf("Error closing the engine client:\n'%v'", err)
		}
	}()
	logrus.Debugf("Successfully got engine client from host ip and port")

	engineInfo, err := getEngineInfoWithTimeout(ctx, engineClient)
	if err != nil {
		return EngineStatus_ContainerRunningButServerNotResponding, runningEngineIpAndPort, "", nil
	}

	logrus.Debugf("Successfully got engine info from client")
	return EngineStatus_Running, runningEngineIpAndPort, engineInfo.GetEngineVersion(), nil
}

// StartEngineIdempotentlyWithDefaultVersion Starts an engine if one doesn't exist already, and returns a client to it
func (manager *EngineManager) StartEngineIdempotentlyWithDefaultVersion(ctx context.Context, logLevel logrus.Level, poolSize uint8) (kurtosis_engine_rpc_api_bindings.EngineServiceClient, func() error, error) {
	status, maybeHostMachinePortBinding, engineVersion, err := manager.GetEngineStatus(ctx)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred retrieving the Kurtosis engine status, which is necessary for creating a connection to the engine")
	}
	logrus.Debugf("Engine status: '%v'", status)
	clusterType := manager.clusterConfig.GetClusterType()
	engineGuarantor := newEngineExistenceGuarantorWithDefaultVersion(
		ctx,
		maybeHostMachinePortBinding,
		manager.kurtosisBackend,
		manager.shouldSendMetrics,
		manager.engineServerKurtosisBackendConfigSupplier,
		logLevel,
		engineVersion,
		clusterType,
		manager.onBastionHost,
		poolSize,
		manager.enclaveEnvVars,
		manager.allowedCORSOrigins,
		doNotStartTheEngineInDebugModeForDefaultVersion,
	)
	// TODO Need to handle the Kubernetes case, where a gateway needs to be started after the engine is started but
	//  before we can return an EngineClient
	engineClient, engineClientCloseFunc, err := manager.startEngineWithGuarantor(ctx, status, engineGuarantor)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred starting the engine with the engine existence guarantor")
	}
	return engineClient, engineClientCloseFunc, nil
}

// StartEngineIdempotentlyWithCustomVersion Starts an engine if one doesn't exist already, and returns a client to it
func (manager *EngineManager) StartEngineIdempotentlyWithCustomVersion(ctx context.Context, engineImageVersionTag string, logLevel logrus.Level, poolSize uint8, shouldStartInDebugMode bool) (kurtosis_engine_rpc_api_bindings.EngineServiceClient, func() error, error) {
	status, maybeHostMachinePortBinding, engineVersion, err := manager.GetEngineStatus(ctx)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred retrieving the Kurtosis engine status, which is necessary for creating a connection to the engine")
	}

	clusterType := manager.clusterConfig.GetClusterType()
	engineGuarantor := newEngineExistenceGuarantorWithCustomVersion(
		ctx,
		maybeHostMachinePortBinding,
		manager.kurtosisBackend,
		manager.shouldSendMetrics,
		manager.engineServerKurtosisBackendConfigSupplier,
		engineImageVersionTag,
		logLevel,
		engineVersion,
		clusterType,
		manager.onBastionHost,
		poolSize,
		manager.enclaveEnvVars,
		manager.allowedCORSOrigins,
		shouldStartInDebugMode,
	)
	engineClient, engineClientCloseFunc, err := manager.startEngineWithGuarantor(ctx, status, engineGuarantor)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred starting the engine with the engine existence guarantor")
	}
	return engineClient, engineClientCloseFunc, nil
}

// StopEngineIdempotently Stops the engine if it's running, doing nothing if not
func (manager *EngineManager) StopEngineIdempotently(ctx context.Context) error {

	// TODO after 2022-07-08, when we're confident nobody is running enclaves/engines that use the bindmounted directory,
	//  add a step here that will delete the engine data dirpath if it exists on the host machine
	// host_machine_directories.GetEngineDataDirpath()

	// We execute the mechanism in three steps
	// 1- We stop the engine in order to execute the Kurtosis engine graceful stop (this block us to execute the backend destroy call directly without a stop call before)
	// 2- We wait until the engine was successfully stopped
	// 3- We destroy the engine for not letting any resource leak (e.g. Kubernetes namespace or Docker container)

	// First stop the engine in order to execute the graceful stop process inside the engine server
	runningFilters := getRunningEnginesFilter()
	successfulEngineGuids, erroredEngineGuids, err := manager.kurtosisBackend.StopEngines(ctx, runningFilters)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred stopping engines using filters '%+v'", runningFilters)
	}
	engineStopErrorStrs := []string{}
	for engineGuid, err := range erroredEngineGuids {
		if err != nil {
			wrappedErr := stacktrace.Propagate(
				err,
				"An error occurred stopping engine with GUID '%v'",
				engineGuid,
			)
			engineStopErrorStrs = append(engineStopErrorStrs, wrappedErr.Error())
		}
	}

	if len(engineStopErrorStrs) > 0 {
		return stacktrace.NewError(
			"One or more errors occurred stopping the engine(s):\n%v",
			strings.Join(
				engineStopErrorStrs,
				"\n\n",
			),
		)
	}

	logrus.Debugf("Stopped signal sent to engines %v", successfulEngineGuids)

	if err := manager.waitUntilEngineStoppedOrError(ctx); err != nil {
		return stacktrace.Propagate(err, "An error occurred waiting until the engine is stopped or an error happens during that process")
	}

	// Then, destroy the stopped engines, for not letting any resource leak
	stoppedFilters := getStoppedEnginesFilter()
	successfulDestroyedEngineGuids, erroredDestroyedEngineGuids, err := manager.kurtosisBackend.DestroyEngines(ctx, stoppedFilters)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred destroying the engines using filters '%+v'", stoppedFilters)
	}
	engineDestroyErrorStrs := []string{}
	for engineGuid, err := range erroredDestroyedEngineGuids {
		if err != nil {
			wrappedErr := stacktrace.Propagate(
				err,
				"An error occurred destroying engine with GUID '%v'",
				engineGuid,
			)
			engineDestroyErrorStrs = append(engineDestroyErrorStrs, wrappedErr.Error())
		}
	}

	if len(engineDestroyErrorStrs) > 0 {
		return stacktrace.NewError(
			"One or more errors occurred destroying the engine(s):\n%v",
			strings.Join(
				engineDestroyErrorStrs,
				"\n\n",
			),
		)
	}

	logrus.Debugf("Destroyed signal sent to engines %v", successfulDestroyedEngineGuids)
	return nil
}

// RestartEngineIdempotently restart the currently running engine.
// If a optionalVersionToUse string is passed, the new engine will be started on this version.
// If no optionalVersionToUse is passed, then the new engine will take the default version, unless
// restartEngineOnSameVersionIfAnyRunning is set to true in which case it will take the version of the currently
// running engine
func (manager *EngineManager) RestartEngineIdempotently(ctx context.Context, logLevel logrus.Level, optionalVersionToUse string, restartEngineOnSameVersionIfAnyRunning bool, poolSize uint8, shouldStartInDebugMode bool) (kurtosis_engine_rpc_api_bindings.EngineServiceClient, func() error, error) {
	var versionOfNewEngine string
	// We try to do our best to restart an engine on the same version the current on is on
	_, _, currentEngineVersion, err := manager.GetEngineStatus(ctx)
	if optionalVersionToUse != defaultEngineVersion || !restartEngineOnSameVersionIfAnyRunning {
		versionOfNewEngine = optionalVersionToUse
	} else {
		if err != nil {
			logrus.Warnf("Error getting current engine information before restarting it. A default engine will be started")
			// if useDefaultVersion = true, we were going to use the default version anyway
			versionOfNewEngine = defaultEngineVersion
		} else {
			versionOfNewEngine = currentEngineVersion
		}
	}
	logrus.Debugf("Restarting engine with version '%v', current engine version is '%v'", versionOfNewEngine, currentEngineVersion)

	if err := manager.StopEngineIdempotently(ctx); err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred stopping the engine currently running")
	}

	var engineClient kurtosis_engine_rpc_api_bindings.EngineServiceClient
	var engineClientCloseFunc func() error
	var restartEngineErr error
	if versionOfNewEngine != defaultEngineVersion {
		_, engineClientCloseFunc, restartEngineErr = manager.StartEngineIdempotentlyWithCustomVersion(ctx, versionOfNewEngine, logLevel, poolSize, shouldStartInDebugMode)
	} else {
		_, engineClientCloseFunc, restartEngineErr = manager.StartEngineIdempotentlyWithDefaultVersion(ctx, logLevel, poolSize)
	}
	if restartEngineErr != nil {
		return nil, nil, stacktrace.Propagate(restartEngineErr, "An error occurred starting a new engine")
	}
	return engineClient, engineClientCloseFunc, nil
}

// ====================================================================================================
//
//	Private Helper Functions
//
// ====================================================================================================
func (manager *EngineManager) startEngineWithGuarantor(ctx context.Context, currentStatus EngineStatus, engineGuarantor *engineExistenceGuarantor) (kurtosis_engine_rpc_api_bindings.EngineServiceClient, func() error, error) {
	if err := currentStatus.Accept(engineGuarantor); err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred guaranteeing that a Kurtosis engine is running")
	}
	hostMachinePortBinding := engineGuarantor.getPostVisitingHostMachineIpAndPort()

	currentContext, err := store.GetContextsConfigStore().GetCurrentContext()
	if err == nil {
		if store.IsRemote(currentContext) {
			portalManager := portal_manager.NewPortalManager()
			if portalManager.IsReachable() {
				// Forward the remote engine port to the local machine
				portalClient := portalManager.GetClient()
				forwardEnginePortArgs := portal_constructors.NewForwardPortArgs(uint32(hostMachinePortBinding.portNum), uint32(hostMachinePortBinding.portNum), kurtosis_context.EngineRemoteEndpointType, &kurtosis_context.EnginePortTransportProtocol, &kurtosis_context.ForwardPortWaitUntilReady)
				if _, err := portalClient.ForwardPort(ctx, forwardEnginePortArgs); err != nil {
					return nil, nil, stacktrace.Propagate(err, "Unable to forward the remote engine port to the local machine.")
				}
			}
		}
	} else {
		logrus.Warnf("Unable to retrieve current Kurtosis context. This is not critical, it will assume using Kurtosis default context for now.")
	}

	engineClient, clientCloseFunc, err := getEngineClientFromHostMachineIpAndPort(hostMachinePortBinding)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred connecting to the running engine; this is very strange and likely indicates a bug in the engine itself")
	}

	clusterType := manager.clusterConfig.GetClusterType()
	// If we're in docker, we can make a health check
	// In the kubernetes case, this health check will fail if the gateway isn't running
	if clusterType == resolved_config.KurtosisClusterType_Docker {
		// Final verification to ensure that the engine server is responding
		if _, err := getEngineInfoWithTimeout(ctx, engineClient); err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred connecting to the engine server after starting it ")
		}
	}

	if clusterType == resolved_config.KurtosisClusterType_Kubernetes {
		logrus.Infof("Engine running in Kubernetes cluster, to connect to the engine from outside the cluster run '%v %v' to open a local gateway to the engine", command_str_consts.KurtosisCmdStr, command_str_consts.GatewayCmdStr)
	}

	return engineClient, clientCloseFunc, nil
}

func getEngineClientFromHostMachineIpAndPort(hostMachineIpAndPort *hostMachineIpAndPort) (kurtosis_engine_rpc_api_bindings.EngineServiceClient, func() error, error) {
	url := hostMachineIpAndPort.GetURL()
	conn, err := grpc.Dial(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred dialling Kurtosis engine at URL '%v'", url)
	}
	engineClient := kurtosis_engine_rpc_api_bindings.NewEngineServiceClient(conn)
	return engineClient, conn.Close, nil
}

func getEngineInfoWithTimeout(ctx context.Context, client kurtosis_engine_rpc_api_bindings.EngineServiceClient) (*kurtosis_engine_rpc_api_bindings.GetEngineInfoResponse, error) {
	ctxWithTimeout, cancelFunc := context.WithTimeout(ctx, waitForEngineResponseTimeout)
	defer cancelFunc()
	engineInfo, err := client.GetEngineInfo(ctxWithTimeout, &emptypb.Empty{}, grpc.WaitForReady(true))
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Kurtosis engine server didn't return a response even with %v timeout",
			waitForEngineResponseTimeout,
		)
	}
	return engineInfo, nil
}

// getRunningEnginesFilter returns a filter for engines with status engine.EngineStatus_Running
func getRunningEnginesFilter() *engine.EngineFilters {
	return &engine.EngineFilters{
		GUIDs: nil,
		Statuses: map[container.ContainerStatus]bool{
			container.ContainerStatus_Running: true,
		},
	}
}

// getStoppedEnginesFilter returns a filter for engines with status engine.EngineStatus_Stopped
func getStoppedEnginesFilter() *engine.EngineFilters {
	return &engine.EngineFilters{
		GUIDs: nil,
		Statuses: map[container.ContainerStatus]bool{
			container.ContainerStatus_Stopped: true,
		},
	}
}

func getKurtosisConfig() (*resolved_config.KurtosisConfig, error) {
	configStore := kurtosis_config.GetKurtosisConfigStore()
	configProvider := kurtosis_config.NewKurtosisConfigProvider(configStore)
	kurtosisConfig, err := configProvider.GetOrInitializeConfig()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting or initializing the Kurtosis config")
	}
	return kurtosisConfig, nil
}

func (manager *EngineManager) waitUntilEngineStoppedOrError(ctx context.Context) error {
	var status EngineStatus
	var err error
	for i := 0; i < waitUntilEngineStoppedTries; i += 1 {
		status, _, _, err = manager.GetEngineStatus(ctx)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred checking the status of the engine")
		}
		if status == EngineStatus_Stopped {
			return nil
		}
		logrus.Debugf("Waiting engine to report stopped, currently reporting '%v'", status)
		time.Sleep(waitUntilEngineStoppedCoolOff)
	}
	return stacktrace.NewError("Engine did not report stopped status, last status reported was '%v'", status)
}
