package common

import (
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/kurtosis_config_getter"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/resolved_config"
	"github.com/kurtosis-tech/stacktrace"
)

func ValidateEnclavePoolSizeFlag(enclavePoolSize uint8) error {
	if enclavePoolSize > 0 {
		isEnclavePoolAvailableForCurrentCluster, err := isEnclavePoolAvailableForCurrentClusterType()
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred validating if the enclave pool feature is valid for the current cluster type")
		}
		if !isEnclavePoolAvailableForCurrentCluster {
			return stacktrace.NewError("The enclave pool feature is not available for the current cluster type, please switch to the 'Kubernetes' Kurtosis backend type if you want to use it.")
		}
	}
	return nil
}

func isEnclavePoolAvailableForCurrentClusterType() (bool, error) {
	clusterConfig, err := kurtosis_config_getter.GetKurtosisClusterConfig()
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred getting the Kurtosis cluster config")
	}

	clusterType := clusterConfig.GetClusterType()

	if clusterType == resolved_config.KurtosisClusterType_Kubernetes {
		return true, nil
	}
	return false, nil
}
