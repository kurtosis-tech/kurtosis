package engine_manager

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/kurtosis_config_getter"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_cluster_setting"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/resolved_config"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_database"
	"github.com/kurtosis-tech/kurtosis/engine/launcher/engine_server_launcher"
	"github.com/kurtosis-tech/object-attributes-schema-lib/schema"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"strings"
	"time"
)

const (
	waitForEngineResponseTimeout    = 5 * time.Second
	defaultClusterName              = resolved_config.DefaultDockerClusterName
	defaultHttpLogsCollectorPortNum = uint16(9712)
)

// Unfortunately, Docker doesn't have constants for the protocols it supports declared
var objAttrsSchemaPortProtosToDockerPortProtos = map[schema.PortProtocol]string{
	schema.PortProtocol_TCP:  "tcp",
	schema.PortProtocol_SCTP: "sctp",
	schema.PortProtcol_UDP:   "udp",
}

type EngineManager struct {
	kurtosisBackend                           backend_interface.KurtosisBackend
	shouldSendMetrics                         bool
	engineServerKurtosisBackendConfigSupplier engine_server_launcher.KurtosisBackendConfigSupplier
	clusterConfig                             *resolved_config.KurtosisClusterConfig
	// Make engine IP, port, and protocol configurable in the future
}

// TODO It's really weird that we have a context getting passed in to a constructor, but we have to do this
//  because we're currently creating the KurtosisBackend right here. The right way to fix this is have the
//  engine manager use the currently-set cluster information to:
//  1) check if it's Kubernetes or Docker
//  2) if it's Docker, try to start an engine in case one doesn't exist
//  3) if it's Kubernets, if creating the EngienClient fails then print the "you need to start a gateway" command
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
		return nil, stacktrace.Propagate(err, "E")
	}

	kurtosisBackend, err := clusterConfig.GetKurtosisBackend(ctx)
	if err != nil {
		return nil, stacktrace.NewError("An error occurred getting the Kurtosis backend for cluster '%v'", clusterName)
	}
	engineBackendConfigSupplier := clusterConfig.GetEngineBackendConfigSupplier()

	return &EngineManager{
		kurtosisBackend:   kurtosisBackend,
		shouldSendMetrics: kurtosisConfig.GetShouldSendMetrics(),
		engineServerKurtosisBackendConfigSupplier: engineBackendConfigSupplier,
		clusterConfig: clusterConfig,
	}, nil
}

// TODO This is a huge hack, that's only here temporarily because we have commands that use KurtosisBackend directly (they
//  should not), and EngineConsumingKurtosisCommand therefore needs to provide them with a KurtosisBackend. Once all our
//  commands only access the Kurtosis APIs, we can remove this.
func (manager *EngineManager) GetKurtosisBackend() backend_interface.KurtosisBackend {
	return manager.kurtosisBackend
}

/*
Returns:
	- The engine status
	- The host machine port bindings (not present if the engine is stopped)
	- The engine version (only present if the engine is running)
*/
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

	engineClient, engineClientCloseFunc, err := getEngineClientFromHostMachineIpAndPort(runningEngineIpAndPort)
	if err != nil {
		return EngineStatus_ContainerRunningButServerNotResponding, runningEngineIpAndPort, "", nil
	}
	defer func() {
		if err = engineClientCloseFunc(); err != nil {
			logrus.Warnf("Error closing the engine client:\n'%v'", err)
		}
	}()

	engineInfo, err := getEngineInfoWithTimeout(ctx, engineClient)
	if err != nil {
		return EngineStatus_ContainerRunningButServerNotResponding, runningEngineIpAndPort, "", nil
	}

	return EngineStatus_Running, runningEngineIpAndPort, engineInfo.GetEngineVersion(), nil
}

// Starts an engine if one doesn't exist already, and returns a client to it
func (manager *EngineManager) StartEngineIdempotentlyWithDefaultVersion(ctx context.Context, logLevel logrus.Level) (kurtosis_engine_rpc_api_bindings.EngineServiceClient, func() error, error) {
	status, maybeHostMachinePortBinding, engineVersion, err := manager.GetEngineStatus(ctx)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred retrieving the Kurtosis engine status, which is necessary for creating a connection to the engine")
	}
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
	)
	// TODO Need to handle the Kubernetes case, where a gateway needs to be started after the engine is started but
	//  before we can return an EngineClient
	engineClient, engineClientCloseFunc, err := manager.startEngineWithGuarantor(ctx, status, engineGuarantor)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred starting the engine with the engine existence guarantor")
	}
	return engineClient, engineClientCloseFunc, nil
}

// Starts an engine if one doesn't exist already, and returns a client to it
func (manager *EngineManager) StartEngineIdempotentlyWithCustomVersion(ctx context.Context, engineImageVersionTag string, logLevel logrus.Level) (kurtosis_engine_rpc_api_bindings.EngineServiceClient, func() error, error) {
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
	)
	engineClient, engineClientCloseFunc, err := manager.startEngineWithGuarantor(ctx, status, engineGuarantor)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred starting the engine with the engine existence guarantor")
	}
	return engineClient, engineClientCloseFunc, nil
}

// Stops the engine if it's running, doing nothing if not
func (manager *EngineManager) StopEngineIdempotently(ctx context.Context) error {

	// TODO after 2022-07-08, when we're confident nobody is running enclaves/engines that use the bindmounted directory,
	//  add a step here that will delete the engine data dirpath if it exists on the host machine
	// host_machine_directories.GetEngineDataDirpath()

	_, erroredEngineGuids, err := manager.kurtosisBackend.StopEngines(ctx, getRunningEnginesFilter())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred stopping ")
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
	clusterType := manager.clusterConfig.GetClusterType()

	//TODO This is a temporary hack we should remove it when centralized logs be implemented in the KubernetesBackend
	if clusterType == resolved_config.KurtosisClusterType_Docker {
		if err = manager.destroyCentralizedLogsComponents(ctx); err != nil {
			return stacktrace.Propagate(err, "An error occurred destroying the centralized logs components")
		}
	}

	return nil
}

// ====================================================================================================
//                                       Private Helper Functions
// ====================================================================================================
func (manager *EngineManager) startEngineWithGuarantor(ctx context.Context, currentStatus EngineStatus, engineGuarantor *engineExistenceGuarantor) (kurtosis_engine_rpc_api_bindings.EngineServiceClient, func() error, error) {
	if err := currentStatus.Accept(engineGuarantor); err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred guaranteeing that a Kurtosis engine is running")
	}
	hostMachinePortBinding := engineGuarantor.getPostVisitingHostMachineIpAndPort()

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

func (manager *EngineManager) destroyCentralizedLogsComponents(ctx context.Context) error {

	logsCollectorFilters := &logs_collector.LogsCollectorFilters{}
	logsDatabaseFilters := &logs_database.LogsDatabaseFilters{}

	if err := manager.kurtosisBackend.DestroyLogsCollector(ctx, logsCollectorFilters); err != nil {
		return stacktrace.Propagate(err, "An error occurred destroying the logs collector")
	}
	if err := manager.kurtosisBackend.DestroyLogsDatabase(ctx, logsDatabaseFilters); err != nil {
		return stacktrace.Propagate(err, "An error occurred destroying the logs collector")
	}
	return nil
}

func getEngineClientFromHostMachineIpAndPort(hostMachineIpAndPort *hostMachineIpAndPort) (kurtosis_engine_rpc_api_bindings.EngineServiceClient, func() error, error) {
	url := hostMachineIpAndPort.GetURL()
	conn, err := grpc.Dial(url, grpc.WithInsecure())
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
		Statuses: map[container_status.ContainerStatus]bool{
			container_status.ContainerStatus_Running: true,
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
