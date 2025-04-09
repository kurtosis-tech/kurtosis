package grafloki

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/resolved_config"
	"github.com/kurtosis-tech/kurtosis/cli/cli/out"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	lokiImage = "grafana/loki:3.4.2"
	lokiPort  = 3100

	grafanaImage = "grafana/grafana:11.6.0"
	grafanaPort  = 3000

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

func StartGrafloki(ctx context.Context, clusterType resolved_config.KurtosisClusterType, graflokiConfig resolved_config.GrafanaLoki) (logs_aggregator.Sinks, error) {
	var lokiHost string
	var grafanaUrl string
	var err error
	switch clusterType {
	case resolved_config.KurtosisClusterType_Docker:
		lokiHost, grafanaUrl, err = StartGrafLokiInDocker(ctx, graflokiConfig)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred starting Grafana and Loki in Docker.")
		}
	case resolved_config.KurtosisClusterType_Kubernetes:
		lokiHost, grafanaUrl, err = StartGrafLokiInKubernetes(ctx, graflokiConfig)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred starting Grafana and Loki in Kubernetes.")
		}
	default:
		return nil, stacktrace.NewError("Unsupported cluster type: %v", clusterType.String())
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
	out.PrintOutLn(fmt.Sprintf("Grafana running at %v", grafanaUrl))

	return lokiSink, nil
}
