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

	// Container-internal listen ports (what each service binds to inside its container).
	clickHouseHTTPPort   = uint16(8123)
	clickHouseNativePort = uint16(9000)

	collectorOTLPGRPCPort = uint16(4317)
	collectorOTLPHTTPPort = uint16(4318)
	collectorLokiPort     = uint16(3500)
	collectorHealthPort   = uint16(13133)

	// Host-published ports. Deliberately non-default so the side containers do not
	// collide with a developer's own ClickHouse (8123) or OTLP collector (4317/4318)
	// already bound on the Docker host. Must match the ethereum-package's
	// ENGINE_OTEL_* constants in main.star.
	clickHouseHTTPHostPort    = uint16(18123)
	collectorOTLPGRPCHostPort = uint16(14317)
	collectorOTLPHTTPHostPort = uint16(14318)

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
	case resolved_config.KurtosisClusterType_Kubernetes:
		endpoints, err := StartOtelInK8s(ctx)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred starting otel resources in Kubernetes.")
		}
		return endpoints, nil
	case resolved_config.KurtosisClusterType_Podman:
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
	case resolved_config.KurtosisClusterType_Kubernetes:
		if err := StopOtelInK8s(ctx); err != nil {
			return stacktrace.Propagate(err, "An error occurred stopping otel resources in Kubernetes.")
		}
	case resolved_config.KurtosisClusterType_Podman:
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
