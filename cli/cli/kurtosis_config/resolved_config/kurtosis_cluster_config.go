package resolved_config

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/backend_creator"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	v2 "github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/overrides_objects/v2"
	"github.com/kurtosis-tech/kurtosis-engine-server/launcher/engine_server_launcher"
	"github.com/kurtosis-tech/stacktrace"
	"strings"
)

const (
	defaultKubernetesEnclaveDataVolumeSizeInMegabytes = uint(1024)
)

// Nil because the CLI will never operate in API container mode
var dockerBackendApiContainerModeArgs *backend_creator.APIContainerModeArgs = nil

type kurtosisBackendSupplier func(ctx context.Context) (backend_interface.KurtosisBackend, error)

type KurtosisClusterConfig struct {
	kurtosisBackendSupplier     kurtosisBackendSupplier
	engineBackendConfigSupplier engine_server_launcher.KurtosisBackendConfigSupplier
	clusterType                 KurtosisClusterType
}

func NewKurtosisClusterConfigFromOverrides(clusterId string, overrides *v2.KurtosisClusterConfigV2) (*KurtosisClusterConfig, error) {
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

	return &KurtosisClusterConfig{
		kurtosisBackendSupplier:     backendSupplier,
		engineBackendConfigSupplier: engineBackendConfigSupplier,
		clusterType:                 clusterType,
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

// ====================================================================================================
//                                      Private Helpers
// ====================================================================================================
func getSuppliers(clusterId string, clusterType KurtosisClusterType, kubernetesConfig *v2.KubernetesClusterConfigV2) (
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
			backend, err := backend_creator.GetLocalDockerKurtosisBackend(dockerBackendApiContainerModeArgs)
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

		enclaveDataVolumeSizeInMb := defaultKubernetesEnclaveDataVolumeSizeInMegabytes
		if kubernetesConfig.EnclaveSizeInMegabytes != nil {
			enclaveDataVolumeSizeInMb = *kubernetesConfig.EnclaveSizeInMegabytes
		}

		backendSupplier = func(ctx context.Context) (backend_interface.KurtosisBackend, error) {
			backend, err := lib.GetCLIKubernetesKurtosisBackend(ctx)
			if err != nil {
				return nil, stacktrace.Propagate(
					err,
					"An error occurred getting a Kurtosis Kubernetes backend from cluster '%v' with storage class '%v' and enclave data volume size (in MB) '%v'",
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
