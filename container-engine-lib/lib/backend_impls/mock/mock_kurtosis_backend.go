package mock

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/api_container"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/module"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/networking_sidecar"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"io"
	"net"
)

type MockKurtosisBackend struct {}

func NewMockKurtosisBackend() *MockKurtosisBackend {
	return &MockKurtosisBackend{}
}

func (backend *MockKurtosisBackend) CreateEngine(ctx context.Context, imageOrgAndRepo string, imageVersionTag string, grpcPortNum uint16, grpcProxyPortNum uint16, engineDataDirpathOnHostMachine string, envVars map[string]string) (*engine.Engine, error) {
	panic("implement me")
}

func (backend *MockKurtosisBackend) GetEngines(ctx context.Context, filters *engine.EngineFilters) (map[string]*engine.Engine, error) {
	panic("implement me")
}

func (backend *MockKurtosisBackend) StopEngines(ctx context.Context, filters *engine.EngineFilters) (successfulEngineIds map[string]bool, erroredEngineIds map[string]error, resultErr error) {
	panic("implement me")
}

func (backend *MockKurtosisBackend) DestroyEngines(ctx context.Context, filters *engine.EngineFilters) (successfulEngineIds map[string]bool, erroredEngineIds map[string]error, resultErr error) {
	panic("implement me")
}

func (backend *MockKurtosisBackend) CreateEnclave(ctx context.Context, enclaveId enclave.EnclaveID, isPartitioningEnabled bool) (*enclave.Enclave, error) {
	panic("implement me")
}

func (backend *MockKurtosisBackend) GetEnclaves(ctx context.Context, filters *enclave.EnclaveFilters) (map[enclave.EnclaveID]*enclave.Enclave, error) {
	panic("implement me")
}

func (backend *MockKurtosisBackend) StopEnclaves(ctx context.Context, filters *enclave.EnclaveFilters) (successfulEnclaveIds map[enclave.EnclaveID]bool, erroredEnclaveIds map[enclave.EnclaveID]error, resultErr error) {
	panic("implement me")
}

func (backend *MockKurtosisBackend) DestroyEnclaves(ctx context.Context, filters *enclave.EnclaveFilters) (successfulEnclaveIds map[enclave.EnclaveID]bool, erroredEnclaveIds map[enclave.EnclaveID]error, resultErr error) {
	panic("implement me")
}

func (backend *MockKurtosisBackend) CreateAPIContainer(ctx context.Context, image string, grpcPortId string, grpcPortSpec *port_spec.PortSpec, grpcProxyPortId string, grpcProxyPortSpec *port_spec.PortSpec, enclaveDataDirpathOnHostMachine string, envVars map[string]string) (*api_container.APIContainer, error) {
	panic("implement me")
}

func (backend *MockKurtosisBackend) GetAPIContainers(ctx context.Context, filters *api_container.APIContainerFilters) (map[string]*api_container.APIContainer, error) {
	panic("implement me")
}

func (backend *MockKurtosisBackend) StopAPIContainers(ctx context.Context, filters *enclave.EnclaveFilters) (successApiContainerIds map[string]bool, erroredApiContainerIds map[string]error, resultErr error) {
	panic("implement me")
}

func (backend *MockKurtosisBackend) DestroyAPIContainers(ctx context.Context, filters *enclave.EnclaveFilters) (successApiContainerIds map[string]bool, erroredApiContainerIds map[string]error, resultErr error) {
	panic("implement me")
}

func (backend *MockKurtosisBackend) CreateModule(ctx context.Context, id module.ModuleID, guid module.ModuleGUID, containerImageName string, serializedParams string) (newModule *module.Module, resultErr error) {
	panic("implement me")
}

func (backend *MockKurtosisBackend) GetModules(ctx context.Context, filters *module.ModuleFilters) (map[string]*module.Module, error) {
	panic("implement me")
}

func (backend *MockKurtosisBackend) DestroyModules(ctx context.Context, filters *module.ModuleFilters) (successfulModuleIds map[module.ModuleGUID]bool, erroredModuleIds map[module.ModuleGUID]error, resultErr error) {
	panic("implement me")
}

func (backend *MockKurtosisBackend) CreateUserService(ctx context.Context, id service.ServiceID, guid service.ServiceGUID, containerImageName string, privatePorts []*port_spec.PortSpec, entrypointArgs []string, cmdArgs []string, envVars map[string]string, enclaveDataDirMntDirpath string, filesArtifactMountDirpaths map[string]string) (newUserService *service.Service, resultErr error) {
	panic("implement me")
}

func (backend *MockKurtosisBackend) GetUserServices(ctx context.Context, filters *service.ServiceFilters) (map[service.ServiceGUID]*service.Service, error) {
	panic("implement me")
}

func (backend *MockKurtosisBackend) GetUserServiceLogs(ctx context.Context, filters *service.ServiceFilters) (map[service.ServiceGUID]io.ReadCloser, error) {
	panic("implement me")
}

func (backend *MockKurtosisBackend) RunUserServiceExecCommand(ctx context.Context, serviceGUID service.ServiceGUID, commandArgs []string) (exitCode int32, output string, resultErr error) {
	panic("implement me")
}

func (backend *MockKurtosisBackend) WaitForUserServiceHttpEndpointAvailability(ctx context.Context, serviceGUID service.ServiceGUID, httpMethod string, port uint32, path string, requestBody string, bodyText string, initialDelayMilliseconds uint32, retries uint32, retriesDelayMilliseconds uint32) (resultErr error) {
	panic("implement me")
}

func (backend *MockKurtosisBackend) GetShellOnUserService(ctx context.Context, serviceGUID service.ServiceGUID) (resultErr error) {
	panic("implement me")
}

func (backend *MockKurtosisBackend) StopUserServices(ctx context.Context, filters *service.ServiceFilters) (successfulUserServiceIds map[service.ServiceGUID]bool, erroredUserServiceIds map[service.ServiceGUID]error, resultErr error) {
	panic("implement me")
}

func (backend *MockKurtosisBackend) DestroyUserServices(ctx context.Context, filters *service.ServiceFilters) (successfulUserServiceIds map[service.ServiceGUID]bool, erroredUserServiceIds map[service.ServiceGUID]error, resultErr error) {
	panic("implement me")
}

func (backend *MockKurtosisBackend) CreateNetworkingSidecar(ctx context.Context, enclaveId enclave.EnclaveID, serviceGuid service.ServiceGUID, ipAddr net.IP) (*networking_sidecar.NetworkingSidecar, error) {
	panic("implement me")
}

func (backend *MockKurtosisBackend) GetNetworkingSidecars(ctx context.Context, filters *networking_sidecar.NetworkingSidecarFilters) (map[networking_sidecar.NetworkingSidecarGUID]*networking_sidecar.NetworkingSidecar, error) {
	panic("implement me")
}

func (backend *MockKurtosisBackend) RunNetworkingSidecarsExecCommand(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	networkingSidecarsCommands map[networking_sidecar.NetworkingSidecarGUID][]string,
)(
	map[networking_sidecar.NetworkingSidecarGUID]bool,
	map[networking_sidecar.NetworkingSidecarGUID]error,
	error,
){

	successfulSidecarGuids := map[networking_sidecar.NetworkingSidecarGUID]bool{}
	erroredSidecarGuids := map[networking_sidecar.NetworkingSidecarGUID]error{}

	for networkingSidecarGuid := range networkingSidecarsCommands {
		successfulSidecarGuids[networkingSidecarGuid] = true
	}

	return successfulSidecarGuids, erroredSidecarGuids, nil
}

func (backend *MockKurtosisBackend) StopNetworkingSidecars(ctx context.Context, filters *networking_sidecar.NetworkingSidecarFilters) (successfulSidecarGuids map[networking_sidecar.NetworkingSidecarGUID]bool, erroredSidecarGuids map[networking_sidecar.NetworkingSidecarGUID]error, resultErr error) {
	panic("implement me")
}

func (backend *MockKurtosisBackend) DestroyNetworkingSidecars(ctx context.Context, filters *networking_sidecar.NetworkingSidecarFilters) (successfulSidecarGuids map[networking_sidecar.NetworkingSidecarGUID]bool, erroredSidecarGuids map[networking_sidecar.NetworkingSidecarGUID]error, resultErr error) {
	panic("implement me")
}

