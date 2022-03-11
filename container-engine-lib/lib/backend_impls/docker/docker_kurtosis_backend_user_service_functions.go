package docker

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"io"
)

func (backendCore *DockerKurtosisBackend) CreateUserService(
	ctx context.Context,
	id string,
	containerImageName string,
	privatePorts []*port_spec.PortSpec,
	entrypointArgs []string,
	cmdArgs []string,
	envVars map[string]string,
	enclaveDataDirMntDirpath string,
	filesArtifactMountDirpaths map[string]string,
)(
	newUserService *service.Service,
	resultErr error,
){
	panic("Implement me")
}

func (backendCore *DockerKurtosisBackend) GetUserServices(
	ctx context.Context,
	filters *service.ServiceFilters,
)(
	map[string]*service.Service,
	error,
){
	panic("Implement me")
}

func (backendCore *DockerKurtosisBackend) GetUserServiceLogs(
	ctx context.Context,
	filters *service.ServiceFilters,
)(
	map[string]io.ReadCloser,
	error,
){
	panic("Implement me")
}

func (backendCore *DockerKurtosisBackend) RunUserServiceExecCommand (
	ctx context.Context,
	serviceId string,
	commandArgs []string,
)(
	exitCode int32,
	output string,
	resultErr error,
){
	panic("Implement me")
}

func (backendCore *DockerKurtosisBackend) WaitForUserServiceHttpEndpointAvailability(
	ctx context.Context,
	serviceId string,
	httpMethod string,
	port uint32,
	path string,
	requestBody string,
	bodyText string,
	initialDelayMilliseconds uint32,
	retries uint32,
	retriesDelayMilliseconds uint32,
)(
	resultErr error,
) {
	panic("Implement me")
}

func (backendCore *DockerKurtosisBackend) GetShellOnUserService(
	ctx context.Context,
	userServiceId string,
)(
	resultErr error,
) {
	panic("Implement me")
}

func (backendCore *DockerKurtosisBackend) StopUserServices(
	ctx context.Context,
	filters *service.ServiceFilters,
)(
	successfulUserServiceIds map[string]bool,
	erroredUserServiceIds map[string]error,
	resultErr error,
) {
	panic("Implement me")
}

func (backendCore *DockerKurtosisBackend) DestroyUserServices(
	ctx context.Context,
	filters *service.ServiceFilters,
)(
	successfulUserServiceIds map[string]bool,
	erroredUserServiceIds map[string]error,
	resultErr error,
) {
	panic("Implement me")
}
