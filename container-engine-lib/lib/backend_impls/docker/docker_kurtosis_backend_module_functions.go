package docker

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/module"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"net"
)

func (backendCore *DockerKurtosisBackend) CreateModule(
	ctx context.Context,
	id string,
	containerImageName string,
	serializedParams string,
)(
	privateIp net.IP,
	privatePort *port_spec.PortSpec,
	publicIp net.IP,
	publicPort *port_spec.PortSpec,
	resultErr error,
) {
	panic("Implement me")
}

func (backendCore *DockerKurtosisBackend) GetModules(
	ctx context.Context,
	filters *module.ModuleFilters,
) (
	map[string]*module.Module,
	error,
) {
	panic("Implement me")
}

func (backendCore *DockerKurtosisBackend) DestroyModules(
	ctx context.Context,
	filters *module.ModuleFilters,
) (
	successfulModuleIds map[string]bool,
	erroredModuleIds map[string]error,
	resultErr error,
) {
	panic("Implement me")
}

