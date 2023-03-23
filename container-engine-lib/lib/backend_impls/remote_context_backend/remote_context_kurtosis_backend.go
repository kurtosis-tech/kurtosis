package remote_context_backend

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/api_container"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_database"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/networking_sidecar"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"golang.org/x/sync/errgroup"
	"io"
	"net"
)

// RemoteContextKurtosisBackend is a dual context holding a reference to a local backend running on Docker (k8s is
// currently not supported to work with a remote Kurtosis backend) and a reference to a remote Kurtosis backend (which
// currently is a Docker backend, we do not support anything else).
// Basically, requests are dispatched among those 2 backends depending on which area of Kurtosis they touch.
// This might change, but right now the engine runs locally, therefore engine interacting endpoint goes to the local
// backend, and all the rest goes to the remote backend, as all other containers (including API container) should be
// running in the remote backend
type RemoteContextKurtosisBackend struct {
	localKurtosisBackend  backend_interface.KurtosisBackend
	remoteKurtosisBackend backend_interface.KurtosisBackend
}

// newRemoteContextKurtosisBackend instantiate a new RemoteContextKurtosisBackend but is private. Use the corresponding
// backend_creator to build such a backend
func newRemoteContextKurtosisBackend(
	localBackend backend_interface.KurtosisBackend,
	contextBackend backend_interface.KurtosisBackend,
) *RemoteContextKurtosisBackend {
	return &RemoteContextKurtosisBackend{
		localKurtosisBackend:  localBackend,
		remoteKurtosisBackend: contextBackend,
	}
}

func (backend *RemoteContextKurtosisBackend) FetchImage(ctx context.Context, image string) error {
	// Without knowing which backend will need the image, it's tricky to filter which one should fetch it.
	// As said in the above comment, in practice, this is not the end of the world because this dual-context backend
	// is effectively used only by the engine and the CLI. Therefore, everything it downloads is the Engine and APIC
	// images.
	// If we start using this dual-context backend in other places, we might want to optimize this behaviour.
	errorGroup := errgroup.Group{}
	errorGroup.Go(func() error {
		return backend.localKurtosisBackend.FetchImage(ctx, image)
	})
	errorGroup.Go(func() error {
		return backend.remoteKurtosisBackend.FetchImage(ctx, image)
	})
	if err := errorGroup.Wait(); err != nil {
		return stacktrace.Propagate(err, "Error fetching the image '%s' in one of the backends", image)
	}
	return nil
}

func (backend *RemoteContextKurtosisBackend) CreateEngine(ctx context.Context, imageOrgAndRepo string, imageVersionTag string, grpcPortNum uint16, grpcProxyPortNum uint16, envVars map[string]string) (*engine.Engine, error) {
	return backend.localKurtosisBackend.CreateEngine(ctx, imageOrgAndRepo, imageVersionTag, grpcPortNum, grpcProxyPortNum, envVars)
}

func (backend *RemoteContextKurtosisBackend) GetEngines(ctx context.Context, filters *engine.EngineFilters) (map[engine.EngineGUID]*engine.Engine, error) {
	return backend.localKurtosisBackend.GetEngines(ctx, filters)
}

func (backend *RemoteContextKurtosisBackend) StopEngines(ctx context.Context, filters *engine.EngineFilters) (successfulEngineGuids map[engine.EngineGUID]bool, erroredEngineGuids map[engine.EngineGUID]error, resultErr error) {
	return backend.localKurtosisBackend.StopEngines(ctx, filters)
}

func (backend *RemoteContextKurtosisBackend) DestroyEngines(ctx context.Context, filters *engine.EngineFilters) (successfulEngineGuids map[engine.EngineGUID]bool, erroredEngineGuids map[engine.EngineGUID]error, resultErr error) {
	return backend.localKurtosisBackend.DestroyEngines(ctx, filters)
}

func (backend *RemoteContextKurtosisBackend) EngineLogs(ctx context.Context, outputDirpath string) error {
	return backend.localKurtosisBackend.EngineLogs(ctx, outputDirpath)
}

func (backend *RemoteContextKurtosisBackend) CreateEnclave(ctx context.Context, enclaveUuid enclave.EnclaveUUID, enclaveName string, isPartitioningEnabled bool) (*enclave.Enclave, error) {
	return backend.remoteKurtosisBackend.CreateEnclave(ctx, enclaveUuid, enclaveName, isPartitioningEnabled)
}

func (backend *RemoteContextKurtosisBackend) GetEnclaves(ctx context.Context, filters *enclave.EnclaveFilters) (map[enclave.EnclaveUUID]*enclave.Enclave, error) {
	return backend.remoteKurtosisBackend.GetEnclaves(ctx, filters)
}

func (backend *RemoteContextKurtosisBackend) StopEnclaves(ctx context.Context, filters *enclave.EnclaveFilters) (successfulEnclaveIds map[enclave.EnclaveUUID]bool, erroredEnclaveIds map[enclave.EnclaveUUID]error, resultErr error) {
	return backend.remoteKurtosisBackend.StopEnclaves(ctx, filters)
}

func (backend *RemoteContextKurtosisBackend) DumpEnclave(ctx context.Context, enclaveUuid enclave.EnclaveUUID, outputDirpath string) error {
	return backend.remoteKurtosisBackend.DumpEnclave(ctx, enclaveUuid, outputDirpath)
}

func (backend *RemoteContextKurtosisBackend) DestroyEnclaves(ctx context.Context, filters *enclave.EnclaveFilters) (successfulEnclaveIds map[enclave.EnclaveUUID]bool, erroredEnclaveIds map[enclave.EnclaveUUID]error, resultErr error) {
	return backend.remoteKurtosisBackend.DestroyEnclaves(ctx, filters)
}

func (backend *RemoteContextKurtosisBackend) CreateAPIContainer(ctx context.Context, image string, enclaveUuid enclave.EnclaveUUID, grpcPortNum uint16, grpcProxyPortNum uint16, enclaveDataVolumeDirpath string, ownIpAddressEnvVar string, customEnvVars map[string]string) (*api_container.APIContainer, error) {
	return backend.remoteKurtosisBackend.CreateAPIContainer(ctx, image, enclaveUuid, grpcPortNum, grpcProxyPortNum, enclaveDataVolumeDirpath, ownIpAddressEnvVar, customEnvVars)
}

func (backend *RemoteContextKurtosisBackend) GetAPIContainers(ctx context.Context, filters *api_container.APIContainerFilters) (map[enclave.EnclaveUUID]*api_container.APIContainer, error) {
	return backend.remoteKurtosisBackend.GetAPIContainers(ctx, filters)
}

func (backend *RemoteContextKurtosisBackend) StopAPIContainers(ctx context.Context, filters *api_container.APIContainerFilters) (successfulApiContainerIds map[enclave.EnclaveUUID]bool, erroredApiContainerIds map[enclave.EnclaveUUID]error, resultErr error) {
	return backend.remoteKurtosisBackend.StopAPIContainers(ctx, filters)
}

func (backend *RemoteContextKurtosisBackend) DestroyAPIContainers(ctx context.Context, filters *api_container.APIContainerFilters) (successfulApiContainerIds map[enclave.EnclaveUUID]bool, erroredApiContainerIds map[enclave.EnclaveUUID]error, resultErr error) {
	return backend.remoteKurtosisBackend.DestroyAPIContainers(ctx, filters)
}

func (backend *RemoteContextKurtosisBackend) RegisterUserServices(ctx context.Context, enclaveUuid enclave.EnclaveUUID, services map[service.ServiceName]bool) (map[service.ServiceName]*service.ServiceRegistration, map[service.ServiceName]error, error) {
	return backend.remoteKurtosisBackend.RegisterUserServices(ctx, enclaveUuid, services)
}

func (backend *RemoteContextKurtosisBackend) UnregisterUserServices(ctx context.Context, enclaveUuid enclave.EnclaveUUID, services map[service.ServiceUUID]bool) (map[service.ServiceUUID]bool, map[service.ServiceUUID]error, error) {
	return backend.remoteKurtosisBackend.UnregisterUserServices(ctx, enclaveUuid, services)
}

func (backend *RemoteContextKurtosisBackend) StartRegisteredUserServices(ctx context.Context, enclaveUuid enclave.EnclaveUUID, services map[service.ServiceUUID]*service.ServiceConfig) (map[service.ServiceUUID]*service.Service, map[service.ServiceUUID]error, error) {
	return backend.remoteKurtosisBackend.StartRegisteredUserServices(ctx, enclaveUuid, services)
}

func (backend *RemoteContextKurtosisBackend) GetUserServices(ctx context.Context, enclaveUuid enclave.EnclaveUUID, filters *service.ServiceFilters) (map[service.ServiceUUID]*service.Service, error) {
	return backend.remoteKurtosisBackend.GetUserServices(ctx, enclaveUuid, filters)
}

func (backend *RemoteContextKurtosisBackend) GetUserServiceLogs(ctx context.Context, enclaveUuid enclave.EnclaveUUID, filters *service.ServiceFilters, shouldFollowLogs bool) (successfulUserServiceLogs map[service.ServiceUUID]io.ReadCloser, erroredUserServiceUuids map[service.ServiceUUID]error, resultError error) {
	return backend.remoteKurtosisBackend.GetUserServiceLogs(ctx, enclaveUuid, filters, shouldFollowLogs)
}

func (backend *RemoteContextKurtosisBackend) PauseService(ctx context.Context, enclaveUuid enclave.EnclaveUUID, serviceUUID service.ServiceUUID) (resultErr error) {
	return backend.remoteKurtosisBackend.PauseService(ctx, enclaveUuid, serviceUUID)
}

func (backend *RemoteContextKurtosisBackend) UnpauseService(ctx context.Context, enclaveUuid enclave.EnclaveUUID, serviceUUID service.ServiceUUID) (resultErr error) {
	return backend.remoteKurtosisBackend.UnpauseService(ctx, enclaveUuid, serviceUUID)
}

func (backend *RemoteContextKurtosisBackend) RunUserServiceExecCommands(ctx context.Context, enclaveUuid enclave.EnclaveUUID, userServiceCommands map[service.ServiceUUID][]string) (succesfulUserServiceExecResults map[service.ServiceUUID]*exec_result.ExecResult, erroredUserServiceUuids map[service.ServiceUUID]error, resultErr error) {
	return backend.remoteKurtosisBackend.RunUserServiceExecCommands(ctx, enclaveUuid, userServiceCommands)
}

func (backend *RemoteContextKurtosisBackend) GetConnectionWithUserService(ctx context.Context, enclaveUuid enclave.EnclaveUUID, serviceUuid service.ServiceUUID) (resultConn net.Conn, resultErr error) {
	return backend.remoteKurtosisBackend.GetConnectionWithUserService(ctx, enclaveUuid, serviceUuid)
}

func (backend *RemoteContextKurtosisBackend) CopyFilesFromUserService(ctx context.Context, enclaveUuid enclave.EnclaveUUID, serviceUuid service.ServiceUUID, srcPathOnService string, output io.Writer) error {
	return backend.remoteKurtosisBackend.CopyFilesFromUserService(ctx, enclaveUuid, serviceUuid, srcPathOnService, output)
}

func (backend *RemoteContextKurtosisBackend) StopUserServices(ctx context.Context, enclaveUuid enclave.EnclaveUUID, filters *service.ServiceFilters) (successfulUserServiceUuids map[service.ServiceUUID]bool, erroredUserServiceUuids map[service.ServiceUUID]error, resultErr error) {
	return backend.remoteKurtosisBackend.StopUserServices(ctx, enclaveUuid, filters)
}

func (backend *RemoteContextKurtosisBackend) DestroyUserServices(ctx context.Context, enclaveUuid enclave.EnclaveUUID, filters *service.ServiceFilters) (successfulUserServiceUuids map[service.ServiceUUID]bool, erroredUserServiceUuids map[service.ServiceUUID]error, resultErr error) {
	return backend.remoteKurtosisBackend.DestroyUserServices(ctx, enclaveUuid, filters)

}

func (backend *RemoteContextKurtosisBackend) CreateNetworkingSidecar(ctx context.Context, enclaveUuid enclave.EnclaveUUID, serviceUuid service.ServiceUUID) (*networking_sidecar.NetworkingSidecar, error) {
	return backend.remoteKurtosisBackend.CreateNetworkingSidecar(ctx, enclaveUuid, serviceUuid)
}

func (backend *RemoteContextKurtosisBackend) GetNetworkingSidecars(ctx context.Context, filters *networking_sidecar.NetworkingSidecarFilters) (map[service.ServiceUUID]*networking_sidecar.NetworkingSidecar, error) {
	return backend.remoteKurtosisBackend.GetNetworkingSidecars(ctx, filters)
}

func (backend *RemoteContextKurtosisBackend) RunNetworkingSidecarExecCommands(ctx context.Context, enclaveUuid enclave.EnclaveUUID, networkingSidecarsCommands map[service.ServiceUUID][]string) (successfulNetworkingSidecarExecResults map[service.ServiceUUID]*exec_result.ExecResult, erroredUserServiceUuids map[service.ServiceUUID]error, resultErr error) {
	return backend.remoteKurtosisBackend.RunNetworkingSidecarExecCommands(ctx, enclaveUuid, networkingSidecarsCommands)
}

func (backend *RemoteContextKurtosisBackend) StopNetworkingSidecars(ctx context.Context, filters *networking_sidecar.NetworkingSidecarFilters) (successfulUserServiceUuids map[service.ServiceUUID]bool, erroredUserServiceUuids map[service.ServiceUUID]error, resultErr error) {
	return backend.remoteKurtosisBackend.StopNetworkingSidecars(ctx, filters)
}

func (backend *RemoteContextKurtosisBackend) DestroyNetworkingSidecars(ctx context.Context, filters *networking_sidecar.NetworkingSidecarFilters) (successfulUserServiceUuids map[service.ServiceUUID]bool, erroredUserServiceUuids map[service.ServiceUUID]error, resultErr error) {
	return backend.remoteKurtosisBackend.DestroyNetworkingSidecars(ctx, filters)
}

func (backend *RemoteContextKurtosisBackend) CreateLogsDatabase(ctx context.Context, logsDatabaseHttpPortNumber uint16) (*logs_database.LogsDatabase, error) {
	return backend.remoteKurtosisBackend.CreateLogsDatabase(ctx, logsDatabaseHttpPortNumber)
}

func (backend *RemoteContextKurtosisBackend) GetLogsDatabase(ctx context.Context) (*logs_database.LogsDatabase, error) {
	return backend.remoteKurtosisBackend.GetLogsDatabase(ctx)
}

func (backend *RemoteContextKurtosisBackend) DestroyLogsDatabase(ctx context.Context) error {
	return backend.remoteKurtosisBackend.DestroyLogsDatabase(ctx)
}

func (backend *RemoteContextKurtosisBackend) CreateLogsCollectorForEnclave(ctx context.Context, enclaveUuid enclave.EnclaveUUID, logsCollectorHttpPortNumber uint16, logsCollectorTcpPortNumber uint16) (*logs_collector.LogsCollector, error) {
	return backend.remoteKurtosisBackend.CreateLogsCollectorForEnclave(ctx, enclaveUuid, logsCollectorHttpPortNumber, logsCollectorHttpPortNumber)

}

func (backend *RemoteContextKurtosisBackend) GetLogsCollectorForEnclave(ctx context.Context, enclaveUuid enclave.EnclaveUUID) (*logs_collector.LogsCollector, error) {
	return backend.remoteKurtosisBackend.GetLogsCollectorForEnclave(ctx, enclaveUuid)
}

func (backend *RemoteContextKurtosisBackend) DestroyLogsCollectorForEnclave(ctx context.Context, enclaveUuid enclave.EnclaveUUID) error {
	return backend.remoteKurtosisBackend.DestroyLogsCollectorForEnclave(ctx, enclaveUuid)
}

func (backend *RemoteContextKurtosisBackend) DestroyDeprecatedCentralizedLogsResources(ctx context.Context) error {
	return backend.remoteKurtosisBackend.DestroyDeprecatedCentralizedLogsResources(ctx)
}
