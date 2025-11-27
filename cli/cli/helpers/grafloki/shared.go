package grafloki

import (
	"context"

	"github.com/dzobbe/PoTE-kurtosis/cli/cli/kurtosis_config/resolved_config"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	defaultLokiImage = "grafana/loki:3.4.2"
	lokiPort         = 3100

	defaultGrafanaImage = "grafana/grafana:11.6.0"
	grafanaPort         = 3000

	grafanaAuthAnonymousEnabledEnvVarKey   = "GF_AUTH_ANONYMOUS_ENABLED"
	grafanaAuthAnonymousEnabledEnvVarVal   = "true"
	grafanaAuthAnonymousOrgRoleEnvVarKey   = "GF_AUTH_ANONYMOUS_ORG_ROLE"
	grafanaAuthAnonymousOrgRoleEnvVarVal   = "Admin"
	grafanaSecurityAllowEmbeddingEnvVarKey = "GF_SECURITY_ALLOW_EMBEDDING"
	grafanaSecurityAllowEmbeddingEnvVarVal = "true"

	grafanaDatasourcesKey  = "datasources"
	grafanaDatasourcesPath = "/etc/grafana/provisioning/datasources"
)

type GrafanaDatasource struct {
	Name      string `yaml:"name"`
	Type_     string `yaml:"type"`
	Access    string `yaml:"access"`
	Url       string `yaml:"url"`
	IsDefault bool   `yaml:"isDefault"`
	Editable  bool   `yaml:"editable"`
}

type GrafanaDatasources struct {
	ApiVersion  int64               `yaml:"apiVersion"`
	Datasources []GrafanaDatasource `yaml:"datasources"`
}

func StartGrafloki(ctx context.Context, clusterType resolved_config.KurtosisClusterType, graflokiConfig resolved_config.GrafanaLokiConfig) (logs_aggregator.Sinks, string, error) {
	var lokiHost string
	var grafanaUrl string
	var err error
	switch clusterType {
	case resolved_config.KurtosisClusterType_Docker:
		dockerManager, err := docker_manager.CreateDockerManager(EmptyDockerClientOpts)
		if err != nil {
			return nil, "", stacktrace.Propagate(err, "An error occurred creating the Docker manager to start grafana and loki.")
		}
		lokiHost, grafanaUrl, err = StartGrafLokiInDocker(ctx, graflokiConfig, dockerManager)
		if err != nil {
			return nil, "", stacktrace.Propagate(err, "An error occurred starting Grafana and Loki in Docker.")
		}
	case resolved_config.KurtosisClusterType_Podman:
		podmanManager, err := docker_manager.CreatePodmanManager(EmptyDockerClientOpts)
		if err != nil {
			return nil, "", stacktrace.Propagate(err, "An error occurred creating the Podman Docker manager to start grafana and loki.")
		}
		lokiHost, grafanaUrl, err = StartGrafLokiInDocker(ctx, graflokiConfig, podmanManager)
		if err != nil {
			return nil, "", stacktrace.Propagate(err, "An error occurred starting Grafana and Loki in Docker.")
		}
	case resolved_config.KurtosisClusterType_Kubernetes:
		lokiHost, grafanaUrl, err = StartGrafLokiInKubernetes(ctx, graflokiConfig)
		if err != nil {
			return nil, "", stacktrace.Propagate(err, "An error occurred starting Grafana and Loki in Kubernetes.")
		}
	default:
		return nil, "", stacktrace.NewError("Unsupported cluster type: %v", clusterType.String())
	}

	// This matches the exact configurations here: https://vector.dev/docs/reference/configuration/sinks/loki/
	lokiSink := map[string]map[string]interface{}{
		"loki": {
			"type":     "loki",
			"endpoint": lokiHost,
			"encoding": map[string]string{
				"codec": "json",
			},
			"labels": map[string]string{
				"job": "kurtosis",
			},
		},
	}

	return lokiSink, grafanaUrl, nil
}

func StopGrafloki(ctx context.Context, clusterType resolved_config.KurtosisClusterType) error {
	switch clusterType {
	case resolved_config.KurtosisClusterType_Docker:
		dockerManager, err := docker_manager.CreateDockerManager(EmptyDockerClientOpts)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred creating the Docker manager to stop grafana and loki.")
		}
		err = StopGrafLokiInDocker(ctx, dockerManager)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred stopping Grafana and Loki containers in Docker.")
		}
	case resolved_config.KurtosisClusterType_Podman:
		podmanManager, err := docker_manager.CreatePodmanManager(EmptyDockerClientOpts)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred creating the Podman Docker manager to start grafana and loki.")
		}
		err = StopGrafLokiInDocker(ctx, podmanManager)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred stopping Grafana and Loki containers in Docker.")
		}
	case resolved_config.KurtosisClusterType_Kubernetes:
		err := StopGrafLokiInKubernetes(ctx)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred stopping Grafana and Loki containers in Kubernetes.")
		}
	default:
		return stacktrace.NewError("Unsupported cluster type: %v", clusterType.String())
	}

	logrus.Info("Successfully stopped Grafana and Loki.")
	return nil
}
