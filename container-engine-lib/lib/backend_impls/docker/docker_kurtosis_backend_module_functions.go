package docker

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/module"
)

func (backendCore *DockerKurtosisBackend) CreateModule(
	ctx context.Context,
	id string,
	containerImageName string,
	serializedParams string,
)(
	newModule *module.Module,
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

