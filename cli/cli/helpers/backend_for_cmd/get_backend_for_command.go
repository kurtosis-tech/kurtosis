package backend_for_cmd

import (
	"github.com/kurtosis-tech/container-engine-lib/lib"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/backend_creator"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	// TODO Remove this in favor of actual Kubernetes info in the config file
	storageClassName = "standard"
	volumeSizeInMegabytes = 100
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
	// TODO REFACTOR: we should get this backend from the config!!
	var apiContainerModeArgs *backend_creator.APIContainerModeArgs = nil  // Not an API container
	dockerBackend, err := backend_creator.GetLocalDockerKurtosisBackend(apiContainerModeArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get Docker backend, instead a non-nil error was returned")
	}

	return dockerBackend, err
}
