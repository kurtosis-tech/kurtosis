package backend_for_cmd

import (
	"github.com/kurtosis-tech/container-engine-lib/lib"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	// TODO Remove this in favor of actual Kubernetes info in the config file
	storageClassName = "standard"
	volumeSizeInMegabytes = 1
)

func GetBackendForCmd(useKubernetes bool) (backend_interface.KurtosisBackend, error) {
	if useKubernetes {
		kubernetesBackend, err := lib.GetLocalKubernetesKurtosisBackend(storageClassName, volumeSizeInMegabytes)
		if err != nil {
			return nil, stacktrace.Propagate(
				err,
				"Expected to be able to get a Kubernetes backend with storage class '%v' and " +
					"volume size of '%v' MB, instead a non-nil error was returned",
				storageClassName,
				volumeSizeInMegabytes,
			)
		}
		return kubernetesBackend, nil
	}
	dockerBackend, err := lib.GetLocalDockerKurtosisBackend()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get Docker backend, instead a non-nil error was returned")
	}

	return dockerBackend, err
}
