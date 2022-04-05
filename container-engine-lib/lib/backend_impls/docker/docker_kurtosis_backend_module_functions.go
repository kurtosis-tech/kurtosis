package docker

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/module"
)

func (backend *DockerKurtosisBackend) CreateModule(
	ctx context.Context,
	id module.ModuleID,
	guid module.ModuleGUID,
	containerImageName string,
	serializedParams string,
)(
	newModule *module.Module,
	resultErr error,
) {
	panic("Implement me")
}

func (backend *DockerKurtosisBackend) GetModules(
	ctx context.Context,
	filters *module.ModuleFilters,
) (
	map[string]*module.Module,
	error,
) {
	panic("Implement me")
}

func (backend *DockerKurtosisBackend) DestroyModules(
	ctx context.Context,
	filters *module.ModuleFilters,
) (
	successfulModuleIds map[module.ModuleGUID]bool,
	erroredModuleIds map[module.ModuleGUID]error,
	resultErr error,
) {
	panic("Implement me")
}

