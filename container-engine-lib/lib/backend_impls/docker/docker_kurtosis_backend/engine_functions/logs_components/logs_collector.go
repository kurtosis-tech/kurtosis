package logs_components

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
)

type LogsCollectorAddress *string
type LogsCollectorLabels []string

type LogsCollector interface {
	GetPrivateTcpPortSpec() (*port_spec.PortSpec, error)
	GetPrivateHttpPortSpec() (*port_spec.PortSpec, error)
	GetContainerArgs(
		containerName string,
		containerLabels map[string]string,
		volumeName string,
		networkId string,
		dockerManager *docker_manager.DockerManager,
	) (*docker_manager.CreateAndStartContainerArgs, error)
	WaitForAvailability() error
}


