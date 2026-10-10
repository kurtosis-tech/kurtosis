package kubernetes_kurtosis_backend

import (
	"context"
	"os"
	"strings"
	"sync"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/logs_aggregator_functions"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/logs_aggregator_functions/implementations/vector"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/logs_collector_functions"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/logs_collector_functions/implementations/fluentbit"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

// An out-of-band-deployed engine (e.g. Helm/GitOps) skips the CreateEngine flow that
// sets up the logs components, so ensureLogsComponents self-bootstraps them on demand.
const (
	bootstrapAggregatorHTTPPort = uint16(8686)
	bootstrapCollectorTCPPort   = uint16(9712)
	bootstrapCollectorHTTPPort  = uint16(9713)
	engineNamespaceFile         = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
	fallbackEngineNS            = "kurtosis-engine"
)

var ensureLogsComponentsMutex sync.Mutex

// ownEngineNamespace returns the namespace the engine pod runs in (read from the
// in-cluster ServiceAccount), falling back to the conventional engine namespace.
func ownEngineNamespace() string {
	data, err := os.ReadFile(engineNamespaceFile)
	if err != nil || len(strings.TrimSpace(string(data))) == 0 {
		return fallbackEngineNS
	}
	return strings.TrimSpace(string(data))
}

// ensureLogsComponents creates the logs aggregator + collector if missing. No-op
// when both already exist (the healthy case), so existing behavior is unchanged.
func (backend *KubernetesKurtosisBackend) ensureLogsComponents(ctx context.Context) error {
	ensureLogsComponentsMutex.Lock()
	defer ensureLogsComponentsMutex.Unlock()

	agg, err := logs_aggregator_functions.GetLogsAggregator(ctx, backend.kubernetesManager)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred checking for an existing logs aggregator")
	}
	if agg == nil {
		logrus.Info("No logs aggregator found; self-bootstrapping it...")
		agg, _, err = logs_aggregator_functions.CreateLogsAggregator(
			ctx,
			ownEngineNamespace(),
			vector.NewVectorLogsAggregatorResourcesManager(),
			bootstrapAggregatorHTTPPort,
			logs_aggregator.Sinks{}, // default sink (matches the CreateEngine default)
			false,                   // shouldEnablePersistentVolumeLogsCollection
			backend.objAttrsProvider,
			backend.kubernetesManager,
			map[string]string{}, // nodeSelector (engine not node-pinned)
			nil,                 // tolerations
		)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred self-bootstrapping the logs aggregator")
		}
	}

	coll, err := logs_collector_functions.GetLogsCollector(ctx, backend.kubernetesManager)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred checking for an existing logs collector")
	}
	if coll == nil {
		logrus.Info("No logs collector found; self-bootstrapping it...")
		if _, _, err = logs_collector_functions.CreateLogsCollector(
			ctx,
			bootstrapCollectorTCPPort,
			bootstrapCollectorHTTPPort,
			fluentbit.NewFluentbitLogsCollector(),
			agg,
			nil, // filters
			nil, // parsers
			backend.kubernetesManager,
			backend.objAttrsProvider,
			nil, // tolerations
		); err != nil {
			return stacktrace.Propagate(err, "An error occurred self-bootstrapping the logs collector")
		}
	}
	return nil
}
