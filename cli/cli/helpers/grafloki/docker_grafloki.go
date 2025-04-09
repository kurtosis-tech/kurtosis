package grafloki

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/go-yaml/yaml"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service_user"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	LokiContainerName    = "kurtosis-loki"
	GrafanaContainerName = "kurtosis-grafana"
	lokiReadinessPath    = "/ready"

	bridgeNetworkName = "bridge"
	localhostAddr     = "127.0.0.1"
	rootUserUid       = 0
)

var EmptyDockerClientOpts = []client.Opt{}
var lokiContainerLabels = map[string]string{
	docker_label_key.ContainerTypeDockerLabelKey.GetString(): LokiContainerName,
}
var grafanaContainerLabels = map[string]string{
	docker_label_key.ContainerTypeDockerLabelKey.GetString(): GrafanaContainerName,
}

func StartGrafLokiInDocker(ctx context.Context) (string, string, error) {
	dockerManager, err := docker_manager.CreateDockerManager(EmptyDockerClientOpts)
	if err != nil {
		return "", "", stacktrace.Propagate(err, "An error occurred creating the docker manager to start grafana and loki.")
	}

	var lokiHost string
	doesGrafanaAndLokiExist, lokiHost, err := checkGrafanaAndLokiContainerExistence(ctx, dockerManager, lokiContainerLabels, grafanaContainerLabels)
	if err != nil {
		return "", "", stacktrace.Propagate(err, "An error occurred checking if Grafana and Loki exist.")
	}

	if !doesGrafanaAndLokiExist {
		logrus.Infof("No running Grafana and Loki containers found. Creating them...")
		lokiHost, err = createGrafanaAndLokiContainers(ctx, dockerManager)
		if err != nil {
			return "", "", stacktrace.Propagate(err, "An error occurred creating Grafana and Loki containers.")
		}
	}

	grafanaUrl := fmt.Sprintf("http://%v:%v", localhostAddr, grafanaPort)
	return lokiHost, grafanaUrl, nil
}

func createGrafanaAndLokiContainers(ctx context.Context, dockerManager *docker_manager.DockerManager) (string, error) {
	lokiNatPort := nat.Port(strconv.Itoa(lokiPort) + "/tcp")
	grafanaNatPort := nat.Port(strconv.Itoa(grafanaPort) + "/tcp")

	bridgeNetworkId, err := dockerManager.GetNetworkIdByName(ctx, bridgeNetworkName)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting Docker network id by Name: %v", bridgeNetworkName)
	}

	lokiArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(lokiImage, LokiContainerName, bridgeNetworkId).
		WithUsedPorts(map[nat.Port]docker_manager.PortPublishSpec{
			lokiNatPort: docker_manager.NewManualPublishingSpec(lokiPort),
		}).
		WithFetchingLatestImageIfMissing().
		WithRestartPolicy(docker_manager.RestartOnFailure).
		WithNetworkMode(bridgeNetworkName).
		WithLabels(lokiContainerLabels).
		Build()
	lokiContainerId, _, err := dockerManager.CreateAndStartContainer(ctx, lokiArgs)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating '%v' container.", LokiContainerName)
	}
	shouldDestroyLokiContainer := true
	defer func() {
		if shouldDestroyLokiContainer {
			err := dockerManager.RemoveContainer(ctx, lokiContainerId)
			if err != nil {
				logrus.Warnf("Attempted to remove Loki container after an error occurred creating it but an error occurred removing it.")
				logrus.Warnf("Manually remove Loki container with id: %v", lokiContainerId)
			}
		}
	}()
	logrus.Infof("Loki container started.")

	lokiBridgeNetworkIpAddr, err := dockerManager.GetContainerIPOnNetwork(ctx, lokiContainerId, bridgeNetworkName)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting container '%v' ip address on network '%v'.", lokiContainerId, bridgeNetworkName)
	}

	lokiBridgeNetworkIpAddress := fmt.Sprintf("http://%v:%v", lokiBridgeNetworkIpAddr, lokiPort)
	lokiHostNetworkIpAddress := fmt.Sprintf("http://%v:%v", localhostAddr, lokiPort)
	if err := waitForLokiReadiness(lokiHostNetworkIpAddress, lokiReadinessPath); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred waiting for Loki container to become ready.")
	}

	grafanaDatasource := &GrafanaDatasources{
		ApiVersion: int64(1),
		Datasources: []GrafanaDatasource{
			{
				Name:      LokiContainerName,
				Type_:     "loki",
				Access:    "proxy",
				Url:       lokiBridgeNetworkIpAddress,
				IsDefault: true,
				Editable:  true,
			},
		}}
	grafanaDatasourceYaml, err := yaml.Marshal(grafanaDatasource)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred serializing Grafana datasource to yaml: %v", grafanaDatasourceYaml)
	}

	tmpFile, err := os.CreateTemp("", "grafana-datasource-*.yaml")
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating temp datasource config.")
	}
	defer tmpFile.Close()
	if _, err := tmpFile.WriteString(string(grafanaDatasourceYaml)); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred writing config.")
	}

	root := service_user.NewServiceUser(rootUserUid)
	grafanaArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(grafanaImage, GrafanaContainerName, bridgeNetworkId).
		WithUsedPorts(map[nat.Port]docker_manager.PortPublishSpec{
			grafanaNatPort: docker_manager.NewManualPublishingSpec(grafanaPort),
		}).
		WithEnvironmentVariables(map[string]string{
			grafanaAuthAnonymousEnabledEnvVarKey:   grafanaAuthAnonymousEnabledEnvVarVal,
			grafanaSecurityAllowEmbeddingEnvVarKey: grafanaSecurityAllowEmbeddingEnvVarVal,
			grafanaAuthAnonymousOrgRoleEnvVarKey:   grafanaAuthAnonymousOrgRoleEnvVarVal,
		}).
		WithBindMounts(map[string]string{
			tmpFile.Name(): fmt.Sprintf("%v/loki.yaml", grafanaDatasourcesPath),
		}).
		WithFetchingLatestImageIfMissing().
		WithRestartPolicy(docker_manager.RestartOnFailure).
		WithNetworkMode(bridgeNetworkName).
		WithLabels(grafanaContainerLabels).
		WithUser(root).
		Build()
	grafanaContainerId, _, err := dockerManager.CreateAndStartContainer(ctx, grafanaArgs)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error creating creating '%v' container.", GrafanaContainerName)
	}
	shouldDestroyGrafanaContainer := true
	defer func() {
		if shouldDestroyGrafanaContainer {
			err := dockerManager.RemoveContainer(ctx, grafanaContainerId)
			if err != nil {
				logrus.Warnf("Attempted to remove Grafana container after an error occurred creating it but an error occurred removing it.")
				logrus.Warnf("Manually remove Grafana container with id: %v", grafanaContainerId)
			}
		}
	}()
	logrus.Infof("Grafana container started.")

	shouldDestroyLokiContainer = false
	shouldDestroyGrafanaContainer = false
	return lokiBridgeNetworkIpAddress, nil
}

func waitForLokiReadiness(lokiHost string, readyPath string) error {
	const (
		retryDelay  = 1 * time.Second
		maxAttempts = 30
	)
	url := lokiHost + readyPath
	for i := 0; i < maxAttempts; i++ {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(retryDelay)
	}
	return stacktrace.NewError("%v did not become ready after %v attempts", lokiHost, maxAttempts)
}

func checkGrafanaAndLokiContainerExistence(ctx context.Context, dockerManager *docker_manager.DockerManager, lokiContainerLabels map[string]string, grafanaContainerLabels map[string]string) (bool, string, error) {
	existsLoki := false
	existsGrafana := false
	var lokiBridgeNetworkIpAddress string

	lokiContainer, err := getContainerByLabel(ctx, dockerManager, lokiContainerLabels)
	if err != nil {
		return false, "", stacktrace.Propagate(err, "An error occurred getting Loki container by labels: %v.", lokiContainerLabels)
	}
	if lokiContainer != nil {
		existsLoki = true
		lokiBridgeNetworkIpAddress, err = dockerManager.GetContainerIPOnNetwork(ctx, lokiContainer.GetId(), bridgeNetworkName)
		if err != nil {
			return false, "", stacktrace.Propagate(err, "An error occurred getting IP of Loki container on network: %v", bridgeNetworkName)
		}
	}

	grafanaContainer, err := getContainerByLabel(ctx, dockerManager, grafanaContainerLabels)
	if err != nil {
		return false, "", stacktrace.Propagate(err, "An error occurred getting Grafana container by labels: %v.", grafanaContainerLabels)
	}
	if grafanaContainer != nil {
		existsGrafana = true
	}

	return existsLoki && existsGrafana, fmt.Sprintf("http://%v:%v", lokiBridgeNetworkIpAddress, lokiPort), nil
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

func StopGrafLokiInDocker(ctx context.Context) error {
	dockerManager, err := docker_manager.CreateDockerManager(EmptyDockerClientOpts)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating Docker manager.")
	}
	if err := dockerManager.RemoveContainer(ctx, GrafanaContainerName); err != nil {
		return stacktrace.Propagate(err, "An error occurred removing Grafana container '%v'", GrafanaContainerName)
	}
	if err := dockerManager.RemoveContainer(ctx, LokiContainerName); err != nil {
		return stacktrace.Propagate(err, "An error occurred removing Loki container '%v'", GrafanaContainerName)
	}
	return nil
}
