package logs_components

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
)

type LogsCollector interface {
	GetPrivateTcpPortSpec() (*port_spec.PortSpec, error)
	GetContainerArgs(
		containerName string,
		containerLabels map[string]string,
		networkId string,
	) (*docker_manager.CreateAndStartContainerArgs, error)
}
