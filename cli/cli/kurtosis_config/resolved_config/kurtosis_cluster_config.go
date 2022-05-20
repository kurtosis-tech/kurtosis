package resolved_config

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/backend_creator"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/overrides_objects/v1"
	"github.com/kurtosis-tech/kurtosis-engine-server/launcher/engine_server_launcher"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	defaultKubernetesEnclaveDataVolumeSizeInMegabytes = uint(10)
)

// Nil because the CLI will never operate in API container mode
var dockerBackendApiContainerModeArgs *backend_creator.APIContainerModeArgs = nil

type kurtosisBackendSupplier func(ctx context.Context) (backend_interface.KurtosisBackend, error)

type KurtosisClusterConfig struct {
	kurtosisBackendSupplier kurtosisBackendSupplier
	engineBackendConfigSupplier engine_server_launcher.KurtosisBackendConfigSupplier
}

func NewKurtosisClusterConfigFromOverrides(overrides *v1.KurtosisClusterConfigV1) (*KurtosisClusterConfig, error) {
	if overrides.Type == nil {
		return nil, stacktrace.NewError("Kurtosis cluster must have a defined type")
	}

	clusterType, err := KurtosisClusterTypeString(*overrides.Type)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Unrecognized cluster type string '%v'; valid values are: %+v",
			*overrides.Type,
			KurtosisClusterTypeStrings(),
		)
	}

	backendSupplier, engineBackendConfigSupplier, err := getSuppliers(clusterType, overrides.Config)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the suppliers that the cluster config will use")
	}

	return &KurtosisClusterConfig{
		kurtosisBackendSupplier:     backendSupplier,
		engineBackendConfigSupplier: engineBackendConfigSupplier,
	}, nil
}

func (clusterConfig *KurtosisClusterConfig) GetKurtosisBackend(ctx context.Context) (backend_interface.KurtosisBackend, error) {
	backend, err := clusterConfig.kurtosisBackendSupplier(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting a Kurtosis backend")
	}
	return backend, nil
}

func (clusterConfig *KurtosisClusterConfig) GetEngineBackendConfigSupplier() (engine_server_launcher.KurtosisBackendConfigSupplier) {
	return clusterConfig.engineBackendConfigSupplier
}


// ====================================================================================================
//                                      Private Helpers
// ====================================================================================================
func getSuppliers(clusterType KurtosisClusterType, kubernetesConfig *v1.KubernetesClusterConfigV1) (
	kurtosisBackendSupplier,
	engine_server_launcher.KurtosisBackendConfigSupplier,
	error,
) {
	var backendSupplier kurtosisBackendSupplier
	var engineConfigSupplier engine_server_launcher.KurtosisBackendConfigSupplier
	switch clusterType {
	case KurtosisClusterType_Docker:
		if kubernetesConfig != nil {
			return nil, nil, stacktrace.NewError("Cluster config must not be provided when cluster type is '%v'", clusterType.String())
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
			return nil, nil, stacktrace.NewError("Cluster config must be provided when cluster type is '%v'", clusterType.String())
		}
		if kubernetesConfig.KubernetesClusterName == nil {
			return nil, nil, stacktrace.NewError(
				"Cluster type is '%v' but has no Kubernetes cluster name in its config map",
				clusterType,
			)
		}

		// TODO Use the Kubernetes cluster name when constructing the KubernetesBackend!
		_ = *kubernetesConfig.KubernetesClusterName

		if kubernetesConfig.StorageClass == nil {
			return nil, nil, stacktrace.NewError(
				"Cluster type is '%v' but has no storage class in its config map",
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
					"An error occurred getting a Kurtosis Kubernetes backend with storage class '%v' and enclave data volume size (in MB) '%v'",
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
			"No Kurtosis backend supplier definition for cluster type '%v'; this is a bug in Kurtosis",
			clusterType.String(),
		)
	}
	return backendSupplier, engineConfigSupplier, nil
}