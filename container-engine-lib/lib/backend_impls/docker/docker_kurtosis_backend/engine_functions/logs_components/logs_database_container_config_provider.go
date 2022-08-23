package logs_components

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
)

type LogsDatabaseContainerConfigProvider interface {
	GetPrivateHttpPortSpec() (*port_spec.PortSpec, error)
	GetContainerArgs(
		containerName string,
		containerLabels map[string]string,
		volumeName string,
		networkId string,
	) (*docker_manager.CreateAndStartContainerArgs, error)
}
