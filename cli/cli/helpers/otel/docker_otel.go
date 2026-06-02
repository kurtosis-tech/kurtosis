package otel

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	ClickHouseContainerLabel = "kurtosis-otel-clickhouse"
	CollectorContainerLabel  = "kurtosis-otel-collector"

	bridgeNetworkName = "bridge"

	clickHouseInitSQLMountPath = "/docker-entrypoint-initdb.d/init.sql"
	collectorConfigMountPath   = "/etc/otelcol/config.yaml"

	clickHouseMaxAvailabilityRetries    = 60
	clickHouseAvailabilityRetryInterval = 1 * time.Second

	tempFilePerms = 0644
)

var ClickHouseContainerNamePrefix = fmt.Sprintf("%v-", ClickHouseContainerLabel)
var CollectorContainerNamePrefix = fmt.Sprintf("%v-", CollectorContainerLabel)

var clickHouseContainerLabels = map[string]string{
	docker_label_key.ContainerTypeDockerLabelKey.GetString(): ClickHouseContainerLabel,
}
var collectorContainerLabels = map[string]string{
	docker_label_key.ContainerTypeDockerLabelKey.GetString(): CollectorContainerLabel,
}

//go:embed files/init.sql
var clickHouseInitSQL string

//go:embed files/collector-bootstrap-config.yaml
var collectorBootstrapConfig string

func StartOtelInDocker(ctx context.Context, dockerManager *docker_manager.DockerManager) (*Endpoints, error) {
	clickHouseContainer, err := getContainerByLabel(ctx, dockerManager, clickHouseContainerLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting ClickHouse container by labels: %v.", clickHouseContainerLabels)
	}
	collectorContainer, err := getContainerByLabel(ctx, dockerManager, collectorContainerLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting collector container by labels: %v.", collectorContainerLabels)
	}

	if clickHouseContainerNeedsRecreation(clickHouseContainer) {
		logrus.Infof("Existing otel ClickHouse container is stale. Recreating it...")
		if err := dockerManager.RemoveContainer(ctx, clickHouseContainer.GetName()); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred removing stale otel ClickHouse container '%v'.", clickHouseContainer.GetName())
		}
		clickHouseContainer = nil
	}

	var clickHouseHTTPURL string
	var clickHouseNativeAddress string
	clickHouseWasCreated := false
	removeClickHouseContainerFunc := func() {}
	if clickHouseContainer == nil {
		logrus.Infof("No running otel ClickHouse container found. Creating it...")
		clickHouseHTTPURL, clickHouseNativeAddress, removeClickHouseContainerFunc, err = createClickHouseContainer(ctx, dockerManager)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating otel ClickHouse container.")
		}
		clickHouseWasCreated = true
		defer func() {
			if clickHouseWasCreated {
				removeClickHouseContainerFunc()
			}
		}()
	} else {
		clickHouseHTTPURL, clickHouseNativeAddress, err = getClickHouseEndpointsOnBridgeNetwork(ctx, dockerManager, clickHouseContainer)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting ClickHouse endpoints from running container.")
		}
	}

	if collectorContainerNeedsRecreation(collectorContainer, clickHouseWasCreated) {
		logrus.Infof("Existing otel collector container is stale. Recreating it...")
		if err := dockerManager.RemoveContainer(ctx, collectorContainer.GetName()); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred removing stale otel collector container '%v'.", collectorContainer.GetName())
		}
		collectorContainer = nil
	}

	var collectorOTLPGRPCURL string
	var collectorOTLPHTTPURL string
	var collectorLokiURL string
	collectorWasCreated := false
	removeCollectorContainerFunc := func() {}
	if collectorContainer == nil {
		logrus.Infof("No running otel collector container found. Creating it...")
		collectorOTLPGRPCURL, collectorOTLPHTTPURL, collectorLokiURL, removeCollectorContainerFunc, err = createCollectorContainer(ctx, dockerManager, clickHouseNativeAddress)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating otel collector container.")
		}
		collectorWasCreated = true
		defer func() {
			if collectorWasCreated {
				removeCollectorContainerFunc()
			}
		}()
	} else {
		collectorOTLPGRPCURL, collectorOTLPHTTPURL, collectorLokiURL, err = getCollectorEndpointsOnBridgeNetwork(ctx, dockerManager, collectorContainer)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting collector endpoints from running container.")
		}
	}

	clickHouseWasCreated = false
	collectorWasCreated = false
	return &Endpoints{
		ClickHouseHTTPURL:       clickHouseHTTPURL,
		ClickHouseNativeAddress: clickHouseNativeAddress,
		CollectorOTLPGRPCURL:    collectorOTLPGRPCURL,
		CollectorOTLPHTTPURL:    collectorOTLPHTTPURL,
		CollectorLokiURL:        collectorLokiURL,
	}, nil
}

func createClickHouseContainer(ctx context.Context, dockerManager *docker_manager.DockerManager) (string, string, func(), error) {
	bridgeNetworkId, err := dockerManager.GetNetworkIdByName(ctx, bridgeNetworkName)
	if err != nil {
		return "", "", nil, stacktrace.Propagate(err, "An error occurred getting Docker network id by Name: %v", bridgeNetworkName)
	}

	initSQLFilepath, err := writeTempFile("otel-clickhouse-init-*.sql", clickHouseInitSQL)
	if err != nil {
		return "", "", nil, stacktrace.Propagate(err, "An error occurred writing ClickHouse init SQL.")
	}

	clickHouseUuid, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return "", "", nil, stacktrace.Propagate(err, "An error occurred generating a uuid for otel ClickHouse.")
	}
	clickHouseContainerName := fmt.Sprintf("%v%v", ClickHouseContainerNamePrefix, clickHouseUuid)
	clickHouseArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(defaultClickHouseImage, clickHouseContainerName, bridgeNetworkId).
		WithUsedPorts(map[nat.Port]docker_manager.PortPublishSpec{
			nat.Port(strconv.Itoa(int(clickHouseHTTPPort)) + "/tcp"):   docker_manager.NewManualPublishingSpec(clickHouseHTTPHostPort),
			nat.Port(strconv.Itoa(int(clickHouseNativePort)) + "/tcp"): docker_manager.NewNoPublishingSpec(),
		}).
		WithEnvironmentVariables(map[string]string{
			"CLICKHOUSE_DB":                        "otel",
			"CLICKHOUSE_DEFAULT_ACCESS_MANAGEMENT": "1",
		}).
		WithBindMounts(map[string]string{
			initSQLFilepath: clickHouseInitSQLMountPath,
		}).
		WithFetchingLatestImageIfMissing().
		WithRestartPolicy(docker_manager.RestartOnFailure).
		WithNetworkMode(bridgeNetworkName).
		WithLabels(clickHouseContainerLabels).
		Build()

	clickHouseContainerId, _, err := dockerManager.CreateAndStartContainer(ctx, clickHouseArgs)
	if err != nil {
		return "", "", nil, stacktrace.Propagate(err, "An error occurred creating '%v' container.", clickHouseContainerName)
	}
	shouldDestroyClickHouseContainer := true
	removeClickHouseContainerFunc := func() {
		err := dockerManager.RemoveContainer(ctx, clickHouseContainerId)
		if err != nil {
			logrus.Warnf("Attempted to remove otel ClickHouse container after an error occurred creating it but an error occurred removing it.")
			logrus.Warnf("Manually remove otel ClickHouse container with id: %v", clickHouseContainerId)
		}
	}
	defer func() {
		if shouldDestroyClickHouseContainer {
			removeClickHouseContainerFunc()
		}
	}()
	logrus.Infof("otel ClickHouse container started.")
	if err := waitForClickHouseAvailability(ctx, dockerManager, clickHouseContainerId); err != nil {
		return "", "", nil, stacktrace.Propagate(err, "An error occurred waiting for otel ClickHouse container to become available.")
	}

	clickHouseHTTPURL, clickHouseNativeAddress, err := getEndpointsOnBridgeNetwork(ctx, dockerManager, clickHouseContainerId, clickHouseHTTPPort, clickHouseNativePort)
	if err != nil {
		return "", "", nil, stacktrace.Propagate(err, "An error occurred getting ClickHouse endpoints on network '%v'.", bridgeNetworkName)
	}

	shouldDestroyClickHouseContainer = false
	return clickHouseHTTPURL, clickHouseNativeAddress, removeClickHouseContainerFunc, nil
}

func createCollectorContainer(ctx context.Context, dockerManager *docker_manager.DockerManager, clickHouseNativeAddress string) (string, string, string, func(), error) {
	bridgeNetworkId, err := dockerManager.GetNetworkIdByName(ctx, bridgeNetworkName)
	if err != nil {
		return "", "", "", nil, stacktrace.Propagate(err, "An error occurred getting Docker network id by Name: %v", bridgeNetworkName)
	}

	collectorConfig := strings.ReplaceAll(collectorBootstrapConfig, "{{CLICKHOUSE_NATIVE_ADDRESS}}", clickHouseNativeAddress)
	configFilepath, err := writeTempFile("otel-collector-bootstrap-*.yaml", collectorConfig)
	if err != nil {
		return "", "", "", nil, stacktrace.Propagate(err, "An error occurred writing otel collector bootstrap config.")
	}

	collectorUuid, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return "", "", "", nil, stacktrace.Propagate(err, "An error occurred generating a uuid for otel collector.")
	}
	collectorContainerName := fmt.Sprintf("%v%v", CollectorContainerNamePrefix, collectorUuid)
	collectorArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(defaultCollectorImage, collectorContainerName, bridgeNetworkId).
		WithCmdArgs([]string{fmt.Sprintf("--config=%v", collectorConfigMountPath)}).
		WithUsedPorts(map[nat.Port]docker_manager.PortPublishSpec{
			nat.Port(strconv.Itoa(int(collectorOTLPGRPCPort)) + "/tcp"): docker_manager.NewManualPublishingSpec(collectorOTLPGRPCHostPort),
			nat.Port(strconv.Itoa(int(collectorOTLPHTTPPort)) + "/tcp"): docker_manager.NewManualPublishingSpec(collectorOTLPHTTPHostPort),
			nat.Port(strconv.Itoa(int(collectorLokiPort)) + "/tcp"):     docker_manager.NewNoPublishingSpec(),
			nat.Port(strconv.Itoa(int(collectorHealthPort)) + "/tcp"):   docker_manager.NewNoPublishingSpec(),
		}).
		WithBindMounts(map[string]string{
			configFilepath: collectorConfigMountPath,
		}).
		WithFetchingLatestImageIfMissing().
		WithRestartPolicy(docker_manager.RestartOnFailure).
		WithNetworkMode(bridgeNetworkName).
		WithLabels(collectorContainerLabels).
		Build()

	collectorContainerId, _, err := dockerManager.CreateAndStartContainer(ctx, collectorArgs)
	if err != nil {
		return "", "", "", nil, stacktrace.Propagate(err, "An error occurred creating '%v' container.", collectorContainerName)
	}
	shouldDestroyCollectorContainer := true
	removeCollectorContainerFunc := func() {
		err := dockerManager.RemoveContainer(ctx, collectorContainerId)
		if err != nil {
			logrus.Warnf("Attempted to remove otel collector container after an error occurred creating it but an error occurred removing it.")
			logrus.Warnf("Manually remove otel collector container with id: %v", collectorContainerId)
		}
	}
	defer func() {
		if shouldDestroyCollectorContainer {
			removeCollectorContainerFunc()
		}
	}()
	logrus.Infof("otel collector container started.")

	collectorOTLPGRPCURL, collectorOTLPHTTPURL, collectorLokiURL, err := getCollectorEndpointsOnBridgeNetworkByContainerId(ctx, dockerManager, collectorContainerId)
	if err != nil {
		return "", "", "", nil, stacktrace.Propagate(err, "An error occurred getting collector endpoints on network '%v'.", bridgeNetworkName)
	}

	shouldDestroyCollectorContainer = false
	return collectorOTLPGRPCURL, collectorOTLPHTTPURL, collectorLokiURL, removeCollectorContainerFunc, nil
}

func getClickHouseEndpointsOnBridgeNetwork(ctx context.Context, dockerManager *docker_manager.DockerManager, clickHouseContainer *types.Container) (string, string, error) {
	return getEndpointsOnBridgeNetwork(ctx, dockerManager, clickHouseContainer.GetId(), clickHouseHTTPPort, clickHouseNativePort)
}

func getCollectorEndpointsOnBridgeNetwork(ctx context.Context, dockerManager *docker_manager.DockerManager, collectorContainer *types.Container) (string, string, string, error) {
	return getCollectorEndpointsOnBridgeNetworkByContainerId(ctx, dockerManager, collectorContainer.GetId())
}

func getCollectorEndpointsOnBridgeNetworkByContainerId(ctx context.Context, dockerManager *docker_manager.DockerManager, collectorContainerId string) (string, string, string, error) {
	bridgeNetworkIpAddr, err := dockerManager.GetContainerIPOnNetwork(ctx, collectorContainerId, bridgeNetworkName)
	if err != nil {
		return "", "", "", stacktrace.Propagate(err, "An error occurred getting container '%v' ip address on network '%v'.", collectorContainerId, bridgeNetworkName)
	}
	return fmt.Sprintf("http://%v:%v", bridgeNetworkIpAddr, collectorOTLPGRPCPort), fmt.Sprintf("http://%v:%v", bridgeNetworkIpAddr, collectorOTLPHTTPPort), fmt.Sprintf("http://%v:%v", bridgeNetworkIpAddr, collectorLokiPort), nil
}

func getEndpointsOnBridgeNetwork(ctx context.Context, dockerManager *docker_manager.DockerManager, containerId string, firstPort uint16, secondPort uint16) (string, string, error) {
	bridgeNetworkIpAddr, err := dockerManager.GetContainerIPOnNetwork(ctx, containerId, bridgeNetworkName)
	if err != nil {
		return "", "", stacktrace.Propagate(err, "An error occurred getting container '%v' ip address on network '%v'.", containerId, bridgeNetworkName)
	}
	return fmt.Sprintf("http://%v:%v", bridgeNetworkIpAddr, firstPort), fmt.Sprintf("%v:%v", bridgeNetworkIpAddr, secondPort), nil
}

func clickHouseContainerNeedsRecreation(clickHouseContainer *types.Container) bool {
	if clickHouseContainer == nil {
		return false
	}
	return !hasExpectedHostPortBinding(clickHouseContainer.GetHostPortBindings(), clickHouseHTTPPort, clickHouseHTTPHostPort)
}

func collectorContainerNeedsRecreation(collectorContainer *types.Container, clickHouseWasCreated bool) bool {
	if collectorContainer == nil {
		return false
	}
	if clickHouseWasCreated {
		return true
	}
	hostPortBindings := collectorContainer.GetHostPortBindings()
	return !hasExpectedHostPortBinding(hostPortBindings, collectorOTLPGRPCPort, collectorOTLPGRPCHostPort) || !hasExpectedHostPortBinding(hostPortBindings, collectorOTLPHTTPPort, collectorOTLPHTTPHostPort)
}

func hasExpectedHostPortBinding(hostPortBindings map[nat.Port]*nat.PortBinding, containerPort uint16, hostPort uint16) bool {
	binding, found := hostPortBindings[nat.Port(strconv.Itoa(int(containerPort))+"/tcp")]
	return found && binding != nil && binding.HostPort == strconv.Itoa(int(hostPort))
}

func waitForClickHouseAvailability(ctx context.Context, dockerManager *docker_manager.DockerManager, clickHouseContainerId string) error {
	for i := 0; i < clickHouseMaxAvailabilityRetries; i++ {
		exitCode, err := dockerManager.RunUserServiceExecCommands(ctx, clickHouseContainerId, "", []string{"clickhouse-client", "--query", "SELECT 1"}, io.Discard)
		if err == nil && exitCode == 0 {
			return nil
		}
		if i < clickHouseMaxAvailabilityRetries-1 {
			time.Sleep(clickHouseAvailabilityRetryInterval)
		}
	}
	return stacktrace.NewError("otel ClickHouse container did not become available after %v retries.", clickHouseMaxAvailabilityRetries)
}

func StopOtelInDocker(ctx context.Context, dockerManager *docker_manager.DockerManager) error {
	collectorContainer, err := getContainerByLabel(ctx, dockerManager, collectorContainerLabels)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting collector container by labels: %v", collectorContainerLabels)
	}
	if collectorContainer != nil {
		if err := dockerManager.RemoveContainer(ctx, collectorContainer.GetName()); err != nil {
			return stacktrace.Propagate(err, "An error occurred removing otel collector container '%v'", collectorContainer.GetName())
		}
	}

	clickHouseContainer, err := getContainerByLabel(ctx, dockerManager, clickHouseContainerLabels)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting ClickHouse container by labels: %v", clickHouseContainerLabels)
	}
	if clickHouseContainer != nil {
		if err := dockerManager.RemoveContainer(ctx, clickHouseContainer.GetName()); err != nil {
			return stacktrace.Propagate(err, "An error occurred removing otel ClickHouse container '%v'", clickHouseContainer.GetName())
		}
	}

	return nil
}

func getContainerByLabel(ctx context.Context, dockerManager *docker_manager.DockerManager, containerLabels map[string]string) (*types.Container, error) {
	containers, err := dockerManager.GetContainersByLabels(ctx, containerLabels, false)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting container by labels '%+v'.", containerLabels)
	}
	if len(containers) > 1 {
		return nil, stacktrace.NewError("More than one container with labels '%v' found.", containerLabels)
	}
	if len(containers) == 0 {
		return nil, nil
	}
	return containers[0], nil
}

func writeTempFile(pattern string, contents string) (string, error) {
	tmpFile, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating temp file.")
	}
	if _, err := tmpFile.WriteString(contents); err != nil {
		closeErr := tmpFile.Close()
		if closeErr != nil {
			logrus.Warnf("Error closing temp file after write failure:\n'%v'", closeErr)
		}
		return "", stacktrace.Propagate(err, "An error occurred writing temp file.")
	}
	if err := tmpFile.Close(); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred closing temp file.")
	}
	if err := os.Chmod(tmpFile.Name(), tempFilePerms); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred setting temp file permissions.")
	}
	return tmpFile.Name(), nil
}
