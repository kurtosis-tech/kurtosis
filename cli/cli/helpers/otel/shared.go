package otel

import (
	"context"

	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/resolved_config"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	defaultClickHouseImage = "clickhouse/clickhouse-server:26.3-alpine"
	defaultCollectorImage  = "otel/opentelemetry-collector-contrib:0.117.0"

	clickHouseHTTPPort   = uint16(8123)
	clickHouseNativePort = uint16(9000)

	collectorOTLPGRPCPort = uint16(4317)
	collectorOTLPHTTPPort = uint16(4318)
	collectorLokiPort     = uint16(3500)
	collectorHealthPort   = uint16(13133)

	unsupportedClusterTypeErrorMsg = "kurtosis otel is Docker-only; current cluster backend is '%v'."
)

var emptyDockerClientOpts = []client.Opt{}

type Endpoints struct {
	ClickHouseHTTPURL       string
	ClickHouseNativeAddress string
	CollectorOTLPGRPCURL    string
	CollectorOTLPHTTPURL    string
	CollectorLokiURL        string
}

func StartOtel(ctx context.Context, clusterType resolved_config.KurtosisClusterType) (*Endpoints, error) {
	switch clusterType {
	case resolved_config.KurtosisClusterType_Docker:
		dockerManager, err := docker_manager.CreateDockerManager(emptyDockerClientOpts)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating the Docker manager to start otel.")
		}
		endpoints, err := StartOtelInDocker(ctx, dockerManager)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred starting otel containers in Docker.")
		}
		return endpoints, nil
	case resolved_config.KurtosisClusterType_Podman, resolved_config.KurtosisClusterType_Kubernetes:
		return nil, stacktrace.NewError(unsupportedClusterTypeErrorMsg, clusterType.String())
	default:
		return nil, stacktrace.NewError("Unsupported cluster type: %v", clusterType.String())
	}
}

func StopOtel(ctx context.Context, clusterType resolved_config.KurtosisClusterType) error {
	switch clusterType {
	case resolved_config.KurtosisClusterType_Docker:
		dockerManager, err := docker_manager.CreateDockerManager(emptyDockerClientOpts)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred creating the Docker manager to stop otel.")
		}
		if err = StopOtelInDocker(ctx, dockerManager); err != nil {
			return stacktrace.Propagate(err, "An error occurred stopping otel containers in Docker.")
		}
	case resolved_config.KurtosisClusterType_Podman, resolved_config.KurtosisClusterType_Kubernetes:
		return stacktrace.NewError(unsupportedClusterTypeErrorMsg, clusterType.String())
	default:
		return stacktrace.NewError("Unsupported cluster type: %v", clusterType.String())
	}

	logrus.Info("Successfully stopped otel containers.")
	return nil
}

func NewLokiSink(lokiHost string) logs_aggregator.Sinks {
	return map[string]map[string]interface{}{
		"loki": {
			"type":     "loki",
			"endpoint": lokiHost,
			"healthcheck": map[string]bool{
				"enabled": false,
			},
			"encoding": map[string]string{
				"codec": "json",
			},
			"labels": map[string]string{
				"job": "kurtosis",
			},
		},
	}
}
