package backend_for_cmd

import (
	"github.com/kurtosis-tech/container-engine-lib/lib"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/stacktrace"
)

// TODO Remove this in favor of actual Kubernetes info in the config file
func GetBackendForCmd(useKubernetes bool) (backend_interface.KurtosisBackend, error) {
	if useKubernetes {
		kubernetesBackend, err := lib.GetLocalKubernetesKurtosisBackend()
		if err != nil {
			return nil, stacktrace.Propagate(err, "Expected to be able to get a Kubernetes backend, instead a non-nil error was returned")
		}
		return kubernetesBackend, nil
	}
	dockerBackend, err := lib.GetLocalDockerKurtosisBackend()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get Docker backend, instead a non-nil error was returned")
	}

	return dockerBackend, err
}
