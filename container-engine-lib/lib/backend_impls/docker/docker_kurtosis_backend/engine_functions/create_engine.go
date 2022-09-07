package engine_functions

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/engine_functions/logs_components"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/engine_functions/logs_components/fluentbit"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/engine_functions/logs_components/loki"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/operation_parallelizer"
	"github.com/kurtosis-tech/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"reflect"
	"strings"
	"time"
)

const (
	maxWaitForEngineAvailabilityRetries         = 10
	timeBetweenWaitForEngineAvailabilityRetries = 1 * time.Second

	getAllEngineContainersOperationId operation_parallelizer.OperationID= "getAllEngineContainers"
	getAllLogsDatabaseContainersOperationId operation_parallelizer.OperationID= "getAllLogsDatabaseContainers"
	getAllLogsCollectorContainersOperationId operation_parallelizer.OperationID= "getAllLogsCollectorContainers"
)

func CreateEngine(
	ctx context.Context,
	imageOrgAndRepo string,
	imageVersionTag string,
	grpcPortNum uint16,
	grpcProxyPortNum uint16,
	logsCollectorHttpPortNumber uint16,
	envVars map[string]string,
	dockerManager *docker_manager.DockerManager,
	objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
) (
	*engine.Engine,
	error,
) {
	isThereEngineOrLogsComponentsContainersInTheCluster, existenceContainerIds,  err := isThereAnyOtherEngineOrLogsComponentsContainersInTheCluster(ctx, dockerManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred checking for engine containers and logs components containers existence")
	}
	canStartNewEngine := !isThereEngineOrLogsComponentsContainersInTheCluster
	if !canStartNewEngine {
		containerIdsStr := strings.Join(existenceContainerIds, ", ")
		return nil, stacktrace.NewError("No new engine won't be started because there exist an engine or logs component container in the cluster; the following containers with IDs '%v' should be removed before creating a new engine", containerIdsStr)
	}

	matchingNetworks, err := dockerManager.GetNetworksByName(ctx, consts.NameOfNetworkToStartEngineContainersIn)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred getting networks matching the network we want to start the engine in, '%v'",
			consts.NameOfNetworkToStartEngineContainersIn,
		)
	}
	numMatchingNetworks := len(matchingNetworks)
	if numMatchingNetworks == 0 && numMatchingNetworks > 1 {
		return nil, stacktrace.NewError(
			"Expected exactly one network matching the name of the network that we want to start the engine in, '%v', but got %v",
			consts.NameOfNetworkToStartEngineContainersIn,
			numMatchingNetworks,
		)
	}
	targetNetwork := matchingNetworks[0]
	targetNetworkId := targetNetwork.GetId()

	engineGuidStr, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred generating a UUID string for the engine")
	}
	engineGuid := engine.EngineGUID(engineGuidStr)

	//Declaring the centralized logs stack
	logsDatabaseContainer := loki.NewLokiLogDatabaseContainer()
	logsCollectorContainer := fluentbit.NewFluentbitLogsCollectorContainer()

	killCentralizedLogsComponentsContainersFunc, err := createCentralizedLogsComponents(
		ctx,
		engineGuid,
		targetNetworkId,
		targetNetwork.GetName(),
		logsCollectorHttpPortNumber,
		objAttrsProvider,
		dockerManager,
		logsDatabaseContainer,
		logsCollectorContainer,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the centralized logs components for the engine with GUID '%v' and network ID '%v'", engineGuid, targetNetworkId)
	}
	shouldKillCentralizedLogsComponentsContainers := true
	defer func() {
		if shouldKillCentralizedLogsComponentsContainers {
			killCentralizedLogsComponentsContainersFunc()
		}
	}()

	privateGrpcPortSpec, err := port_spec.NewPortSpec(grpcPortNum, consts.EnginePortProtocol)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the engine's private grpc port spec object using number '%v' and protocol '%v'",
			grpcPortNum,
			consts.EnginePortProtocol.String(),
		)
	}
	privateGrpcProxyPortSpec, err := port_spec.NewPortSpec(grpcProxyPortNum, consts.EnginePortProtocol)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the engine's private grpc proxy port spec object using number '%v' and protocol '%v'",
			grpcProxyPortNum,
			consts.EnginePortProtocol.String(),
		)
	}

	engineAttrs, err := objAttrsProvider.ForEngineServer(
		engineGuid,
		consts.KurtosisInternalContainerGrpcPortId,
		privateGrpcPortSpec,
		consts.KurtosisInternalContainerGrpcProxyPortId,
		privateGrpcProxyPortSpec,
	)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred getting the engine server container attributes using GUID '%v', grpc port num '%v', and "+
				"grpc proxy port num '%v'",
			engineGuid,
			grpcPortNum,
			grpcProxyPortNum,
		)
	}

	privateGrpcDockerPort, err := shared_helpers.TransformPortSpecToDockerPort(privateGrpcPortSpec)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred transforming the private grpc port spec to a Docker port")
	}
	privateGrpcProxyDockerPort, err := shared_helpers.TransformPortSpecToDockerPort(privateGrpcProxyPortSpec)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred transforming the private grpc proxy port spec to a Docker port")
	}

	usedPorts := map[nat.Port]docker_manager.PortPublishSpec{
		privateGrpcDockerPort:      docker_manager.NewManualPublishingSpec(grpcPortNum),
		privateGrpcProxyDockerPort: docker_manager.NewManualPublishingSpec(grpcProxyPortNum),
	}

	bindMounts := map[string]string{
		// Necessary so that the engine server can interact with the Docker engine
		consts.DockerSocketFilepath: consts.DockerSocketFilepath,
	}

	containerImageAndTag := fmt.Sprintf(
		"%v:%v",
		imageOrgAndRepo,
		imageVersionTag,
	)

	labelStrs := map[string]string{}
	for labelKey, labelValue := range engineAttrs.GetLabels() {
		labelStrs[labelKey.GetString()] = labelValue.GetString()
	}

	createAndStartArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(
		containerImageAndTag,
		engineAttrs.GetName().GetString(),
		targetNetworkId,
	).WithEnvironmentVariables(
		envVars,
	).WithBindMounts(
		bindMounts,
	).WithUsedPorts(
		usedPorts,
	).WithLabels(
		labelStrs,
	).Build()

	// Best-effort pull attempt
	if err = dockerManager.PullImage(ctx, containerImageAndTag); err != nil {
		logrus.Warnf("Failed to pull the latest version of engine server image '%v'; you may be running an out-of-date version", containerImageAndTag)
	}

	containerId, hostMachinePortBindings, err := dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred starting the Kurtosis engine container")
	}
	shouldKillEngineContainer := true
	defer func() {
		if shouldKillEngineContainer {
			// NOTE: We use the background context here so that the kill will still go off even if the reason for
			// the failure was the original context being cancelled
			if err := dockerManager.KillContainer(context.Background(), containerId); err != nil {
				logrus.Errorf(
					"Launching the engine server with GUID '%v' and container ID '%v' didn't complete successfully so we "+
						"tried to kill the container we started, but doing so exited with an error:\n%v",
					engineGuid,
					containerId,
					err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually stop engine server with GUID '%v'!!!!!!", engineGuid)
			}
		}
	}()

	if err := shared_helpers.WaitForPortAvailabilityUsingNetstat(
		ctx,
		dockerManager,
		containerId,
		privateGrpcPortSpec,
		maxWaitForEngineAvailabilityRetries,
		timeBetweenWaitForEngineAvailabilityRetries,
	); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the engine server's grpc port to become available")
	}

	// TODO UNCOMMENT THIS ONCE WE HAVE GRPC-PROXY WIRED UP!!
	/*
		if err := waitForPortAvailabilityUsingNetstat(ctx, backend.dockerManager, containerId, grpcProxyPortNum); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred waiting for the engine server's grpc proxy port to become available")
		}
	*/

	result, err := getEngineObjectFromContainerInfo(containerId, labelStrs, types.ContainerStatus_Running, hostMachinePortBindings)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating an engine object from container with GUID '%v'", containerId)
	}



	shouldKillEngineContainer = false
	shouldKillCentralizedLogsComponentsContainers = false
	return result, nil
}

// ====================================================================================================
// 									   Private helper methods
// ====================================================================================================
//TODO we can run it in parallel after the network creation, and we can wait before returning the EngineInfo object
func createCentralizedLogsComponents(
	ctx context.Context,
	engineGuid engine.EngineGUID,
	targetNetworkId string,
	targetNetworkName string,
	logsCollectorHttpPortNumber uint16,
	objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
	dockerManager *docker_manager.DockerManager,
	logsDatabaseContainer logs_components.LogsDatabaseContainer,
	logsCollectorContainer logs_components.LogsCollectorContainer,
) (func(), error) {

	logsDatabaseHost, logsDatabasePort, killLogsDatabaseContainerFunc, err := logsDatabaseContainer.CreateAndStart(
		ctx,
		consts.LogsDatabaseHttpPortId,
		engineGuid,
		targetNetworkId,
		targetNetworkName,
		objAttrsProvider,
		dockerManager,
	)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the logs database container with http port id '%v' for engine with GUID '%v' in Docker network with ID '%v'",
			consts.LogsDatabaseHttpPortId,
			engineGuid,
			targetNetworkId,
		)
	}
	shouldKillLogsDatabaseContainer := true
	defer func() {
		if shouldKillLogsDatabaseContainer {
			killLogsDatabaseContainerFunc()
		}
	}()

	killLogsCollectorContainerFunc, err := logsCollectorContainer.CreateAndStart(
		ctx,
		logsDatabaseHost,
		logsDatabasePort,
		logsCollectorHttpPortNumber,
		consts.LogsCollectorTcpPortId,
		consts.LogsCollectorHttpPortId,
		engineGuid,
		targetNetworkId,
		objAttrsProvider,
		dockerManager,
	)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred running the logs collector container with logs database host '%v', logs database port '%v', http port '%v', tcp port id '%v', and http port id '%v' for engine with GUID '%v' in Docker network with ID '%v'",
			logsDatabaseHost,
			logsDatabasePort,
			logsCollectorHttpPortNumber,
			consts.LogsCollectorTcpPortId,
			consts.LogsCollectorHttpPortId,
			engineGuid,
			targetNetworkId,
		)
	}
	shouldKillLogsCollectorContainer := true
	defer func() {
		if shouldKillLogsCollectorContainer {
			killLogsCollectorContainerFunc()
		}
	}()

	killCentralizedLogsComponentsContainersFunc := func() {
		killLogsDatabaseContainerFunc()
		killLogsCollectorContainerFunc()
	}

	shouldKillLogsDatabaseContainer = false
	shouldKillLogsCollectorContainer = false
	return killCentralizedLogsComponentsContainersFunc, nil
}

func isThereAnyOtherEngineOrLogsComponentsContainersInTheCluster(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
) (bool, []string, error){

	existentContainerIds := []string{}

	getAllEngineContainersOperation := func() (interface{}, error) {
		getAllEngineFilters := &engine.EngineFilters{}
		allEngineContainers, err := getMatchingEngines(ctx, getAllEngineFilters, dockerManager)
		if err != nil {
			return false, stacktrace.Propagate(err, "An error occurred getting all the engines using filters '%+v'", getAllEngineFilters)
		}

		allEngineContainerIDs := map[string]bool{}

		for containerId := range allEngineContainers{
			allEngineContainerIDs[containerId] = true
		}

		return allEngineContainers, nil
	}

	getLogsDatabaseContainerOperation := func() (interface{}, error) {

		logsDatabaseContainer, err := getLogsDatabaseContainer(ctx, dockerManager)
		if err != nil {
			return false, stacktrace.Propagate(err, "An error occurred getting the logs database container")
		}

		allLogsDatabaseContainerIDs := map[string]bool{
			logsDatabaseContainer.GetId(): true,
		}

		return allLogsDatabaseContainerIDs, nil
	}

	getLogsCollectorContainerOperation := func() (interface{}, error) {
		logsCollectorContainer, err := shared_helpers.GetLogsCollectorContainer(ctx, dockerManager)
		if err != nil {
			return false, stacktrace.Propagate(err, "An error occurred getting the logs collector container")
		}

		allLogsCollectorContainerIDs := map[string]bool{
			logsCollectorContainer.GetId(): true,
		}

		return allLogsCollectorContainerIDs, nil
	}

	allOperations := map[operation_parallelizer.OperationID]operation_parallelizer.Operation{
		getAllEngineContainersOperationId:        getAllEngineContainersOperation,
		getAllLogsDatabaseContainersOperationId:  getLogsDatabaseContainerOperation,
		getAllLogsCollectorContainersOperationId: getLogsCollectorContainerOperation,
	}

	successfulOperations, erroredOperations := operation_parallelizer.RunOperationsInParallel(allOperations)
	if len(erroredOperations) > 0 {
		return false, nil, stacktrace.NewError("An error occurred running these operations '%+v' in parallel\n Operations with errors: %+v", allOperations, erroredOperations)
	}

	for _, uncastedContainerIds := range successfulOperations {
		containerIdsValue := reflect.ValueOf(uncastedContainerIds)
		for _, containerIdValue := range containerIdsValue.MapKeys() {
			existentContainerIds = append(existentContainerIds, containerIdValue.String())
		}
	}

	if len(existentContainerIds) > 0 {
		return true, existentContainerIds, nil
	}
	return false, existentContainerIds, nil
}

