package resolved_config

import (
	"context"
	"strings"

	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/overrides_objects/v3"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/backend_creator"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/configs"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
	"github.com/kurtosis-tech/kurtosis/contexts-config-store/store"
	"github.com/kurtosis-tech/kurtosis/engine/launcher/engine_server_launcher"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	defaultKubernetesEnclaveDataVolumeSizeInMegabytes = uint(1024)
)

type kurtosisBackendSupplier func(ctx context.Context) (backend_interface.KurtosisBackend, error)

type KurtosisClusterConfig struct {
	kurtosisBackendSupplier     kurtosisBackendSupplier
	engineBackendConfigSupplier engine_server_launcher.KurtosisBackendConfigSupplier
	clusterType                 KurtosisClusterType
	logsAggregator              LogsAggregatorConfig
}

type LogsAggregatorConfig struct {
	Sinks logs_aggregator.Sinks
}

func NewKurtosisClusterConfigFromOverrides(clusterId string, overrides *v3.KurtosisClusterConfigV3) (*KurtosisClusterConfig, error) {
	if overrides.Type == nil {
		return nil, stacktrace.NewError("Kurtosis cluster must have a defined type")
	}

	clusterType, err := KurtosisClusterTypeString(*overrides.Type)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Cluster '%v' has unrecognized type '%v'; valid values are: %+v",
			clusterId,
			*overrides.Type,
			strings.Join(KurtosisClusterTypeStrings(), ", "),
		)
	}

	backendSupplier, engineBackendConfigSupplier, err := getSuppliers(clusterId, clusterType, overrides.Config)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the suppliers that cluster '%v' will use", clusterId)
	}

	logsAggregator := LogsAggregatorConfig{
		Sinks: nil,
	}

	if overrides.LogsAggregator != nil {
		if len(overrides.LogsAggregator.Sinks) > 0 {
			for sinkId := range overrides.LogsAggregator.Sinks {
				// We add a default file sink as the logs database for certain log commands (i.e. kurtosis service logs) to work, hence this validation
				// A potential improvement would be that all log-related commands are compatible with user-defined sinks
				if sinkId == logs_aggregator.DefaultSinkId {
					return nil, stacktrace.NewError("The LogsAggregator Sinks had a sink named %s which is reserved for Kurtosis default sink", logs_aggregator.DefaultSinkId)
				}
			}

			logsAggregator.Sinks = overrides.LogsAggregator.Sinks
		}
	}

	return &KurtosisClusterConfig{
		kurtosisBackendSupplier:     backendSupplier,
		engineBackendConfigSupplier: engineBackendConfigSupplier,
		clusterType:                 clusterType,
		logsAggregator:              logsAggregator,
	}, nil
}

func (clusterConfig *KurtosisClusterConfig) GetKurtosisBackend(ctx context.Context) (backend_interface.KurtosisBackend, error) {
	backend, err := clusterConfig.kurtosisBackendSupplier(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting a Kurtosis backend")
	}
	return backend, nil
}

func (clusterConfig *KurtosisClusterConfig) GetEngineBackendConfigSupplier() engine_server_launcher.KurtosisBackendConfigSupplier {
	return clusterConfig.engineBackendConfigSupplier
}

func (clusterConfig *KurtosisClusterConfig) GetClusterType() KurtosisClusterType {
	return clusterConfig.clusterType
}

func (clusterConfig *KurtosisClusterConfig) GetLogsAggregatorConfig() LogsAggregatorConfig {
	return clusterConfig.logsAggregator
}

// ====================================================================================================
//
//	Private Helpers
//
// ====================================================================================================
func getSuppliers(clusterId string, clusterType KurtosisClusterType, kubernetesConfig *v3.KubernetesClusterConfigV3) (
	kurtosisBackendSupplier,
	engine_server_launcher.KurtosisBackendConfigSupplier,
	error,
) {
	var backendSupplier kurtosisBackendSupplier
	var engineConfigSupplier engine_server_launcher.KurtosisBackendConfigSupplier
	switch clusterType {
	case KurtosisClusterType_Docker:
		if kubernetesConfig != nil {
			return nil, nil, stacktrace.NewError(
				"Cluster '%v' defines cluster config, but config must not be provided when cluster type is '%v'",
				clusterId,
				clusterType.String(),
			)
		}

		backendSupplier = func(_ context.Context) (backend_interface.KurtosisBackend, error) {
			var remoteBackendConfigMaybe *configs.KurtosisRemoteBackendConfig
			currentContext, err := store.GetContextsConfigStore().GetCurrentContext()
			if err != nil {
				return nil, stacktrace.Propagate(err, "An error occurred retrieving the current context")
			}
			if store.IsRemote(currentContext) {
				remoteBackendConfigMaybe = configs.NewRemoteBackendConfigFromRemoteContext(currentContext.GetRemoteContextV0())
			}
			// Get a local or remote docker backend based on the existence of the remote backend config.
			// We do not pass APIC mode args since we are dealing with the engine here.
			backend, err := backend_creator.GetDockerKurtosisBackend(backend_creator.NoAPIContainerModeArgs, remoteBackendConfigMaybe)
			if err != nil {
				return nil, stacktrace.Propagate(err, "An error occurred creating the Docker Kurtosis backend")
			}
			return backend, nil
		}

		engineConfigSupplier = engine_server_launcher.NewDockerKurtosisBackendConfigSupplier()
	case KurtosisClusterType_Kubernetes:
		if kubernetesConfig == nil {
			return nil, nil, stacktrace.NewError(
				"Cluster '%v' doesn't define cluster config, but config must be provided when cluster type is '%v'",
				clusterId,
				clusterType.String(),
			)
		}
		if kubernetesConfig.KubernetesClusterName == nil {
			return nil, nil, stacktrace.NewError(
				"Type of cluster '%v' is '%v' but has no Kubernetes cluster name in its config map",
				clusterId,
				clusterType,
			)
		}

		// TODO Use the Kubernetes cluster name when constructing the KubernetesBackend!
		_ = *kubernetesConfig.KubernetesClusterName

		if kubernetesConfig.StorageClass == nil {
			return nil, nil, stacktrace.NewError(
				"Type of cluster '%v' is '%v' but no storage class was defined in the config",
				clusterId,
				clusterType,
			)
		}
		storageClass := *kubernetesConfig.StorageClass
		engineNodeSelectors := map[string]string{} // TODO: add node selectors to kubernetes config

		enclaveDataVolumeSizeInMb := defaultKubernetesEnclaveDataVolumeSizeInMegabytes
		if kubernetesConfig.EnclaveSizeInMegabytes != nil {
			enclaveDataVolumeSizeInMb = *kubernetesConfig.EnclaveSizeInMegabytes
		}

		backendSupplier = func(ctx context.Context) (backend_interface.KurtosisBackend, error) {
			backend, err := kubernetes_kurtosis_backend.GetCLIBackend(ctx, *kubernetesConfig.StorageClass, engineNodeSelectors)
			if err != nil {
				return nil, stacktrace.Propagate(
					err,
					"An error occurred getting Kurtosis Kubernetes backend for CLI from cluster '%v' with storage class '%v' and enclave data volume size (in MB) '%v'",
					clusterId,
					storageClass,
					enclaveDataVolumeSizeInMb,
				)
			}
			return backend, nil
		}

		engineConfigSupplier = engine_server_launcher.NewKubernetesKurtosisBackendConfigSupplier(storageClass, enclaveDataVolumeSizeInMb)
	default:
		// This should never happen because we enforce this via unit tests
		return nil, nil, stacktrace.NewError(
			"Cluster '%v' has unrecognized type '%v'; this is a bug in Kurtosis",
			clusterId,
			clusterType.String(),
		)
	}
	return backendSupplier, engineConfigSupplier, nil
}
