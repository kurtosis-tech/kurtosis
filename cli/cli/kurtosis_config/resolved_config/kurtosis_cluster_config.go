package resolved_config

import (
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	v1 "github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/v1"
	"github.com/kurtosis-tech/stacktrace"
	"golang.org/x/image/colornames"
)

type kurtosisBackendSupplier func() (backend_interface.KurtosisBackend, error)

type KurtosisClusterConfig struct {
	kurtosisBackendSupplier kurtosisBackendSupplier
}

func NewKurtosisClusterConfigFromOverrides(overrides *v1.KurtosisClusterV1) (*KurtosisClusterConfig, error) {
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

	backendSupplier, err := getKurtosisBackendSupplier(clusterType, overrides.Config)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the Kurtosis backend supplier")
	}

	return &KurtosisClusterConfig{
		kurtosisBackendSupplier: backendSupplier,
	}, nil
}

func getKurtosisBackendSupplier(clusterType KurtosisClusterType, kubernetesConfig *v1.KubernetesClusterConfigV1) (
	kurtosisBackendSupplier,
	error,
) {
	var result kurtosisBackendSupplier
	switch clusterType {
	case KurtosisClusterType_Docker:
		if kubernetesConfig != nil {
			return nil, stacktrace.NewError("Cluster config must not be provided when cluster type is '%v'", clusterType.String())
		}
		result = func() (backend_interface.KurtosisBackend, error) {
			// TODO
			return nil, stacktrace.NewError("TODO IMPLEMENT THIS")
		}
	case KurtosisClusterType_Kubernetes:
		if kubernetesConfig == nil {
			return nil, stacktrace.NewError("Cluster config must be provided when cluster type is '%v'", clusterType.String())
		}
		if kubernetesConfig.KubernetesClusterName == nil {
			return nil, stacktrace.NewError(
				"Cluster type is '%v' but has no Kubernetes cluster name in its config map",
				clusterType,
			)
		}
		kubernetesClusterName := *kubernetesConfig.KubernetesClusterName

		if kubernetesConfig.StorageClass == nil {
			return nil, stacktrace.NewError(
				"Cluster type is '%v' but has no storage class in its config map",
				clusterType,
			)
		}
		storageClass := *kubernetesConfig.StorageClass

		if kubernetesConfig.EnclaveSizeInGigabytes == nil {
			return nil, stacktrace.NewError(
				"Cluster type is '%v' but has no enclave data volume size (in GB) specified in its config map",
				clusterType,
			)
		}
		enclaveDataVolSizeGb := *kubernetesConfig.EnclaveSizeInGigabytes

		result = func() (backend_interface.KurtosisBackend, error) {
			// TODO
			return nil, stacktrace.NewError(fmt.Sprintf("TODO IMPLEMENT THIS: %v %v %v", kubernetesClusterName, storageClass, enclaveDataVolSizeGb)
		}
	default:
		// This should never happen because we enforce this via unit tests
		return nil, stacktrace.NewError(
			"No Kurtosis backend supplier definition for cluster type '%v'; this is a bug in Kurtosis",
			clusterType.String(),
		)
	}
	return result, nil
}
