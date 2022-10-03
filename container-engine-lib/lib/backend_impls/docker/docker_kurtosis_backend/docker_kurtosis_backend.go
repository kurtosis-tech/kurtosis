package docker_kurtosis_backend

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/engine_functions"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/logs_collector_functions"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/logs_collector_functions/implementations/fluentbit"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/logs_database_functions"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/logs_database_functions/implementations/loki"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/user_services_functions"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_network_allocator"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_database"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/free_ip_addr_tracker"
	"github.com/kurtosis-tech/stacktrace"
	"io"
	"net"
	"sync"
)

type DockerKurtosisBackend struct {
	dockerManager *docker_manager.DockerManager

	dockerNetworkAllocator *docker_network_allocator.DockerNetworkAllocator

	objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider

	// TODO This is ONLY relevant to internal-to-enclave functions, meaning that we now have a *DockerKurtosisBackend
	//  which takes in some values which are only useful for certain functions. What we should really do is split all
	//  KurtosisBackend functions into API container, engine server, and CLI functions, and move the functionality there
	//  (in essence creating APIContainerKurtosisBackend, EngineKurtosisBackend, CLIKurtosisBackend). That way, everything
	//  will be cleaner. HOWEVER, the reason it's not done this way as of 2022-05-12 is because the CLI still uses some
	//  KurtosisBackend functionality that it shouldn't (e.g. GetUserServiceLogs). This should all flow through the API
	//  container API instaed.
	// This map is set exactly once, upon creation of the DockerKubernetesBackend, and never modified afterwards. Therefore, it doesn't need to be protected with a mutex (because the FreeIPProviders are themselves threadsafe)
	enclaveFreeIpProviders map[enclave.EnclaveID]*free_ip_addr_tracker.FreeIpAddrTracker

	// TODO Migrate this to an on-disk database, so that the API container can be shut down & restarted!
	// Canonical store of the registrations being tracked by this *DockerKurtosisBackend instance
	// NOTE: Unlike Kubernetes, Docker doesn't have a concrete object representing a service registration/IP address
	//  allocation. We use this in-memory store to accomplish the same thing.
	serviceRegistrations map[enclave.EnclaveID]map[service.ServiceGUID]*service.ServiceRegistration

	// Control concurrent access to serviceRegistrations
	serviceRegistrationMutex *sync.Mutex
}

func NewDockerKurtosisBackend(
	dockerManager *docker_manager.DockerManager,
	enclaveFreeIpProviders map[enclave.EnclaveID]*free_ip_addr_tracker.FreeIpAddrTracker,
) *DockerKurtosisBackend {
	dockerNetworkAllocator := docker_network_allocator.NewDockerNetworkAllocator(dockerManager)
	serviceRegistrations := map[enclave.EnclaveID]map[service.ServiceGUID]*service.ServiceRegistration{}
	for enclaveId := range enclaveFreeIpProviders {
		serviceRegistrations[enclaveId] = map[service.ServiceGUID]*service.ServiceRegistration{}
	}
	return &DockerKurtosisBackend{
		dockerManager:            dockerManager,
		dockerNetworkAllocator:   dockerNetworkAllocator,
		objAttrsProvider:         object_attributes_provider.GetDockerObjectAttributesProvider(),
		enclaveFreeIpProviders:   enclaveFreeIpProviders,
		serviceRegistrations:     serviceRegistrations,
		serviceRegistrationMutex: &sync.Mutex{},
	}
}

func (backend *DockerKurtosisBackend) PullImage(image string) error {
	return stacktrace.NewError("PullImage isn't implemented for Docker yet")
}

func (backend *DockerKurtosisBackend) CreateEngine(
	ctx context.Context,
	imageOrgAndRepo string,
	imageVersionTag string,
	grpcPortNum uint16,
	grpcProxyPortNum uint16,
	envVars map[string]string,
) (
	*engine.Engine,
	error,
) {
	return engine_functions.CreateEngine(
		ctx,
		imageOrgAndRepo,
		imageVersionTag,
		grpcPortNum,
		grpcProxyPortNum,
		envVars,
		backend.dockerManager,
		backend.objAttrsProvider,
	)
}

func (backend *DockerKurtosisBackend) GetEngines(
	ctx context.Context,
	filters *engine.EngineFilters,
) (
	map[engine.EngineGUID]*engine.Engine,
	error,
) {
	return engine_functions.GetEngines(ctx, filters, backend.dockerManager)
}

func (backend *DockerKurtosisBackend) StopEngines(
	ctx context.Context,
	filters *engine.EngineFilters,
) (
	resultSuccessfulEngineGuids map[engine.EngineGUID]bool,
	resultErroredEngineGuids map[engine.EngineGUID]error,
	resultErr error,
) {
	return engine_functions.StopEngines(ctx, filters, backend.dockerManager)
}

func (backend *DockerKurtosisBackend) DestroyEngines(
	ctx context.Context,
	filters *engine.EngineFilters,
) (
	resultSuccessfulEngineGuids map[engine.EngineGUID]bool,
	resultErroredEngineGuids map[engine.EngineGUID]error,
	resultErr error,
) {
	return engine_functions.DestroyEngines(ctx, filters, backend.dockerManager)
}

func (backend *DockerKurtosisBackend) StartUserServices(ctx context.Context, enclaveId enclave.EnclaveID, services map[service.ServiceID]*service.ServiceConfig) (map[service.ServiceID]*service.Service, map[service.ServiceID]error, error) {

	logsCollectorFilters := &logs_collector.LogsCollectorFilters{
		Status: container_status.ContainerStatus_Running,
	}

	logsCollector, err := backend.GetLogsCollector(ctx, logsCollectorFilters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting the logs collector")
	}
	if logsCollector == nil || logsCollector.GetStatus() != container_status.ContainerStatus_Running{
		return nil, nil, stacktrace.NewError("The user services can't be started because there is not logs collector running for sending the logs")
	}

	return user_service_functions.StartUserServices(
		ctx,
		enclaveId,
		services,
		backend.serviceRegistrations,
		backend.serviceRegistrationMutex,
		logsCollector,
		backend.objAttrsProvider,
		backend.enclaveFreeIpProviders,
		backend.dockerManager)
}

func (backend *DockerKurtosisBackend) GetUserServices(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *service.ServiceFilters,
) (
	map[service.ServiceGUID]*service.Service,
	error,
) {
	return user_service_functions.GetUserServices(ctx, enclaveId, filters, backend.dockerManager)
}

func (backend *DockerKurtosisBackend) GetUserServiceLogs(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *service.ServiceFilters,
	shouldFollowLogs bool,
) (
	map[service.ServiceGUID]io.ReadCloser,
	map[service.ServiceGUID]error,
	error,
) {
	return user_service_functions.GetUserServiceLogs(ctx, enclaveId, filters, shouldFollowLogs, backend.dockerManager)
}

func (backend *DockerKurtosisBackend) PauseService(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceGuid service.ServiceGUID,
) error {
	return user_service_functions.PauseService(ctx, enclaveId, serviceGuid, backend.dockerManager)
}

func (backend *DockerKurtosisBackend) UnpauseService(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceGuid service.ServiceGUID,
) error {
	return user_service_functions.UnpauseService(ctx, enclaveId, serviceGuid, backend.dockerManager)
}

// TODO Switch these to streaming so that huge command outputs don't blow up the API container memory
// NOTE: This function will block while the exec is ongoing; if we need more perf we can make it async
func (backend *DockerKurtosisBackend) RunUserServiceExecCommands(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	userServiceCommands map[service.ServiceGUID][]string,
) (
	map[service.ServiceGUID]*exec_result.ExecResult,
	map[service.ServiceGUID]error,
	error,
) {
	return user_service_functions.RunUserServiceExecCommands(ctx, enclaveId, userServiceCommands, backend.dockerManager)
}

func (backend *DockerKurtosisBackend) GetConnectionWithUserService(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceGuid service.ServiceGUID,
) (
	net.Conn,
	error,
) {
	return user_service_functions.GetConnectionWithUserService(ctx, enclaveId, serviceGuid, backend.dockerManager)
}

// It returns io.ReadCloser which is a tar stream. It's up to the caller to close the reader.
func (backend *DockerKurtosisBackend) CopyFilesFromUserService(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceGuid service.ServiceGUID,
	srcPathOnContainer string,
	output io.Writer,
) error {
	return user_service_functions.CopyFilesFromUserService(ctx, enclaveId, serviceGuid, srcPathOnContainer, output, backend.dockerManager)
}

func (backend *DockerKurtosisBackend) StopUserServices(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *service.ServiceFilters,
) (
	resultSuccessfulServiceGUIDs map[service.ServiceGUID]bool,
	resultErroredServiceGUIDs map[service.ServiceGUID]error,
	resultErr error,
) {
	return user_service_functions.StopUserServices(ctx, enclaveId, filters, backend.dockerManager)
}

func (backend *DockerKurtosisBackend) DestroyUserServices(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *service.ServiceFilters,
) (
	resultSuccessfulGuids map[service.ServiceGUID]bool,
	resultErroredGuids map[service.ServiceGUID]error,
	resultErr error,
) {
	return user_service_functions.DestroyUserServices(
		ctx,
		enclaveId,
		filters,
		backend.serviceRegistrations,
		backend.serviceRegistrationMutex,
		backend.enclaveFreeIpProviders,
		backend.dockerManager)
}

func (backend *DockerKurtosisBackend) CreateLogsDatabase(
	ctx context.Context,
) (
	*logs_database.LogsDatabase,
	error,
){

	//Declaring the implementation
	logsDatabaseContainer := loki.NewLokiLogDatabaseContainer()

	return logs_database_functions.CreateLogsDatabase(
		ctx,
		logsDatabaseContainer,
		backend.dockerManager,
		backend.objAttrsProvider,
	)
}

func (backend *DockerKurtosisBackend) GetLogsDatabase(
	ctx context.Context,
	filters *logs_database.LogsDatabaseFilters,
) (
	*logs_database.LogsDatabase,
	error,
) {
	return logs_database_functions.GetLogsDatabase(
		ctx,
		filters,
		backend.dockerManager,
	)
}

func (backend *DockerKurtosisBackend) StopLogsDatabase(
	ctx context.Context,
	filters *logs_database.LogsDatabaseFilters,
) (
	error,
) {

	logsCollectorFilters := &logs_collector.LogsCollectorFilters{
		Status: container_status.ContainerStatus_Running,
	}

	logsCollector, err := backend.GetLogsCollector(ctx, logsCollectorFilters)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the logs collector")
	}

	if logsCollector != nil && logsCollector.GetStatus() == container_status.ContainerStatus_Running {
		return stacktrace.NewError("The logs database can't be stopped due the logs collector is running")
	}

	return logs_database_functions.StopLogsDatabase(
		ctx,
		filters,
		backend.dockerManager,
	)
}

func (backend *DockerKurtosisBackend) DestroyLogsDatabase(
	ctx context.Context,
	filters *logs_database.LogsDatabaseFilters,
) (
	error,
) {

	logsCollectorFilters := &logs_collector.LogsCollectorFilters{
		Status: container_status.ContainerStatus_Running,
	}

	logsCollector, err := backend.GetLogsCollector(ctx, logsCollectorFilters)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the logs collector")
	}

	if logsCollector != nil && logsCollector.GetStatus() == container_status.ContainerStatus_Running {
		return stacktrace.NewError("The logs database can't be destroyed due the logs collector is running")
	}

	return logs_database_functions.DestroyLogsDatabase(
		ctx,
		filters,
		backend.dockerManager,
	)
}

func (backend *DockerKurtosisBackend) CreateLogsCollector(
	ctx context.Context,
	logsCollectorHttpPortNumber uint16,
) (
	*logs_collector.LogsCollector,
	error,
) {

	logsDatabaseFilters := &logs_database.LogsDatabaseFilters{
		Status: container_status.ContainerStatus_Running,
	}

	//TODO we we'd have to replace this part if we ever wanted to send to an external source
	logsDatabase, err := backend.GetLogsDatabase(ctx, logsDatabaseFilters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the logs database, it's not possible to run the logs collector without a logs database")
	}

	if logsDatabase == nil || logsDatabase.GetStatus() != container_status.ContainerStatus_Running {
		return nil,stacktrace.NewError("The logs database is not running, it's not possible to run the logs collector without a running logs database")
	}

	//Declaring the implementation
	logsCollectorContainer := fluentbit.NewFluentbitLogsCollectorContainer()

	return logs_collector_functions.CreateLogsCollector(
		ctx,
		logsCollectorHttpPortNumber,
		logsCollectorContainer,
		logsDatabase,
		backend.dockerManager,
		backend.objAttrsProvider,
	)
}

func (backend *DockerKurtosisBackend) GetLogsCollector(
	ctx context.Context,
	filters *logs_collector.LogsCollectorFilters,
) (
	*logs_collector.LogsCollector,
	error,
) {
	return logs_collector_functions.GetLogsCollector(
		ctx,
		filters,
		backend.dockerManager,
	)
}

func (backend *DockerKurtosisBackend) StopLogsCollector(
	ctx context.Context,
	filters *logs_collector.LogsCollectorFilters,
) (
	error,
) {

	return logs_collector_functions.StopLogsCollector(
		ctx,
		filters,
		backend.dockerManager,
	)
}

func (backend *DockerKurtosisBackend) DestroyLogsCollector(
	ctx context.Context,
	filters *logs_collector.LogsCollectorFilters,
) (
	error,
) {

	return logs_collector_functions.DestroyLogsCollector(
		ctx,
		filters,
		backend.dockerManager,
	)
}

// ====================================================================================================
//
//	Private helper functions shared by multiple subfunctions files
//
// ====================================================================================================
func (backend *DockerKurtosisBackend) getEnclaveNetworkByEnclaveId(ctx context.Context, enclaveId enclave.EnclaveID) (*types.Network, error) {
	networkSearchLabels := map[string]string{
		label_key_consts.AppIDDockerLabelKey.GetString():     label_value_consts.AppIDDockerLabelValue.GetString(),
		label_key_consts.EnclaveIDDockerLabelKey.GetString(): string(enclaveId),
	}

	enclaveNetworksFound, err := backend.dockerManager.GetNetworksByLabels(ctx, networkSearchLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Docker networks by enclave ID '%v'", enclaveId)
	}
	numMatchingNetworks := len(enclaveNetworksFound)
	if numMatchingNetworks == 0 {
		return nil, stacktrace.NewError("No network was found for enclave with ID '%v'", enclaveId)
	}
	if numMatchingNetworks > 1 {
		return nil, stacktrace.NewError(
			"Expected exactly one network matching enclave ID '%v', but got %v",
			enclaveId,
			numMatchingNetworks,
		)
	}
	return enclaveNetworksFound[0], nil
}

// Guaranteed to either return an enclave data volume name or throw an error
func (backend *DockerKurtosisBackend) getEnclaveDataVolumeByEnclaveId(ctx context.Context, enclaveId enclave.EnclaveID) (string, error) {
	volumeSearchLabels := map[string]string{
		label_key_consts.AppIDDockerLabelKey.GetString():      label_value_consts.AppIDDockerLabelValue.GetString(),
		label_key_consts.EnclaveIDDockerLabelKey.GetString():  string(enclaveId),
		label_key_consts.VolumeTypeDockerLabelKey.GetString(): label_value_consts.EnclaveDataVolumeTypeDockerLabelValue.GetString(),
	}
	foundVolumes, err := backend.dockerManager.GetVolumesByLabels(ctx, volumeSearchLabels)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting enclave data volumes matching labels '%+v'", volumeSearchLabels)
	}
	if len(foundVolumes) > 1 {
		return "", stacktrace.NewError("Found multiple enclave data volumes matching enclave ID '%v'; this should never happen", enclaveId)
	}
	if len(foundVolumes) == 0 {
		return "", stacktrace.NewError("No enclave data volume found for enclave '%v'", enclaveId)
	}
	volume := foundVolumes[0]
	return volume.Name, nil
}
