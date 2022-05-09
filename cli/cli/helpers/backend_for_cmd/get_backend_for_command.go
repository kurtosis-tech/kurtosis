package backend_for_cmd

import (
	"github.com/kurtosis-tech/container-engine-lib/lib"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
)

const (
	withKubernetesFlagName = "with-kubernetes"
)

func GetBackendForCmd(useKubernetes bool) (backend_interface.KurtosisBackend, error) {
	if useKubernetes {
		return lib.GetLocalKubernetesKurtosisBackend()
	}
	return lib.GetLocalDockerKurtosisBackend()
}
