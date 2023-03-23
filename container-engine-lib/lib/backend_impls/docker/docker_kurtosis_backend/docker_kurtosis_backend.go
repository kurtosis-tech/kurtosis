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
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db/free_ip_addr_tracker"
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
	enclaveFreeIpProviders map[enclave.EnclaveUUID]*free_ip_addr_tracker.FreeIpAddrTracker

	// TODO Migrate this to an on-disk database, so that the API container can be shut down & restarted!
	// Canonical store of the registrations being tracked by this *DockerKurtosisBackend instance
	// NOTE: Unlike Kubernetes, Docker doesn't have a concrete object representing a service registration/IP address
	//  allocation. We use this in-memory store to accomplish the same thing.
	serviceRegistrations map[enclave.EnclaveUUID]map[service.ServiceUUID]*service.ServiceRegistration

	// Control concurrent access to serviceRegistrations
	serviceRegistrationMutex *sync.Mutex
}

func NewDockerKurtosisBackend(
	dockerManager *docker_manager.DockerManager,
	enclaveFreeIpProviders map[enclave.EnclaveUUID]*free_ip_addr_tracker.FreeIpAddrTracker,
) *DockerKurtosisBackend {
	dockerNetworkAllocator := docker_network_allocator.NewDockerNetworkAllocator(dockerManager)
	serviceRegistrations := map[enclave.EnclaveUUID]map[service.ServiceUUID]*service.ServiceRegistration{}
	for enclaveUuid := range enclaveFreeIpProviders {
		serviceRegistrations[enclaveUuid] = map[service.ServiceUUID]*service.ServiceRegistration{}
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

func (backend *DockerKurtosisBackend) FetchImage(ctx context.Context, image string) error {
	err := backend.dockerManager.FetchImage(ctx, image)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred fetching image from kurtosis backend")
	}
	return nil
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
	resultSuccessfulEngineUuids map[engine.EngineGUID]bool,
	resultErroredEngineUuids map[engine.EngineGUID]error,
	resultErr error,
) {
	return engine_functions.StopEngines(ctx, filters, backend.dockerManager)
}

func (backend *DockerKurtosisBackend) DestroyEngines(
	ctx context.Context,
	filters *engine.EngineFilters,
) (
	resultSuccessfulEngineUuids map[engine.EngineGUID]bool,
	resultErroredEngineUuids map[engine.EngineGUID]error,
	resultErr error,
) {
	return engine_functions.DestroyEngines(ctx, filters, backend.dockerManager)
}

func (backend *DockerKurtosisBackend) EngineLogs(ctx context.Context, outputDirpath string) error {
	return engine_functions.EngineLogs(ctx, outputDirpath, backend.dockerManager)
}

func (backend *DockerKurtosisBackend) RegisterUserServices(_ context.Context, enclaveUuid enclave.EnclaveUUID, services map[service.ServiceName]bool) (map[service.ServiceName]*service.ServiceRegistration, map[service.ServiceName]error, error) {
	serviceRegistrationsForEnclave, found := backend.serviceRegistrations[enclaveUuid]
	if !found {
		return nil, nil, stacktrace.NewError(
			"No service registrations are being tracked for enclave '%v'; this likely means that the registration "+
				"request is being called where it shouldn't be (i.e. outside the API container)",
			enclaveUuid,
		)
	}

	freeIpAddrProviderForEnclave, found := backend.enclaveFreeIpProviders[enclaveUuid]
	if !found {
		return nil, nil, stacktrace.NewError(
			"Received a request to register services in enclave '%v', but no free IP address provider was "+
				"defined for this enclave; this likely means that the start request is being called where it shouldn't "+
				"be (i.e. outside the API container)",
			enclaveUuid,
		)
	}

	registeredService, failedServices, err := user_service_functions.RegisterUserServices(enclaveUuid, services, serviceRegistrationsForEnclave, freeIpAddrProviderForEnclave, backend.serviceRegistrationMutex)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "Unexpected error registering services to enclave '%s'", enclaveUuid)
	}
	return registeredService, failedServices, nil
}

func (backend *DockerKurtosisBackend) UnregisterUserServices(_ context.Context, enclaveUuid enclave.EnclaveUUID, services map[service.ServiceUUID]bool) (map[service.ServiceUUID]bool, map[service.ServiceUUID]error, error) {
	serviceRegistrationsForEnclave, found := backend.serviceRegistrations[enclaveUuid]
	if !found {
		return nil, nil, stacktrace.NewError(
			"No service registrations are being tracked for enclave '%v'; this likely means that the registration "+
				"request is being called where it shouldn't be (i.e. outside the API container)",
			enclaveUuid,
		)
	}

	freeIpAddrProviderForEnclave, found := backend.enclaveFreeIpProviders[enclaveUuid]
	if !found {
		return nil, nil, stacktrace.NewError(
			"Received a request to unregister services in enclave '%v', but no free IP address provider was "+
				"defined for this enclave; this likely means that the start request is being called where it shouldn't "+
				"be (i.e. outside the API container)",
			enclaveUuid,
		)
	}

	servicesSuccessfullyUnregistered, failedServices := user_service_functions.UnregisterUserServices(services, serviceRegistrationsForEnclave, freeIpAddrProviderForEnclave, backend.serviceRegistrationMutex)
	return servicesSuccessfullyUnregistered, failedServices, nil
}

func (backend *DockerKurtosisBackend) StartRegisteredUserServices(ctx context.Context, enclaveUuid enclave.EnclaveUUID, services map[service.ServiceUUID]*service.ServiceConfig) (map[service.ServiceUUID]*service.Service, map[service.ServiceUUID]error, error) {
	serviceRegistrationsForEnclave, found := backend.serviceRegistrations[enclaveUuid]
	if !found {
		return nil, nil, stacktrace.NewError(
			"No service registrations are being tracked for enclave '%v'; this likely means that the registration "+
				"request is being called where it shouldn't be (i.e. outside the API container)",
			enclaveUuid,
		)
	}

	freeIpAddrProviderForEnclave, found := backend.enclaveFreeIpProviders[enclaveUuid]
	if !found {
		return nil, nil, stacktrace.NewError(
			"Received a request to start services in enclave '%v', but no free IP address provider was "+
				"defined for this enclave; this likely means that the start request is being called where it shouldn't "+
				"be (i.e. outside the API container)",
			enclaveUuid,
		)
	}

	successfullyStartedService, failedService, err := user_service_functions.StartUserServices(
		ctx,
		enclaveUuid,
		services,
		serviceRegistrationsForEnclave,
		backend.objAttrsProvider,
		freeIpAddrProviderForEnclave,
		backend.dockerManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "Unexpected error while starting user service")
	}
	return successfullyStartedService, failedService, nil
}

func (backend *DockerKurtosisBackend) GetUserServices(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	filters *service.ServiceFilters,
) (
	map[service.ServiceUUID]*service.Service,
	error,
) {
	return user_service_functions.GetUserServices(ctx, enclaveUuid, filters, backend.dockerManager)
}

func (backend *DockerKurtosisBackend) GetUserServiceLogs(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	filters *service.ServiceFilters,
	shouldFollowLogs bool,
) (
	map[service.ServiceUUID]io.ReadCloser,
	map[service.ServiceUUID]error,
	error,
) {
	return user_service_functions.GetUserServiceLogs(ctx, enclaveUuid, filters, shouldFollowLogs, backend.dockerManager)
}

func (backend *DockerKurtosisBackend) PauseService(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	serviceUuid service.ServiceUUID,
) error {
	return user_service_functions.PauseService(ctx, enclaveUuid, serviceUuid, backend.dockerManager)
}

func (backend *DockerKurtosisBackend) UnpauseService(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	serviceUuid service.ServiceUUID,
) error {
	return user_service_functions.UnpauseService(ctx, enclaveUuid, serviceUuid, backend.dockerManager)
}

// TODO Switch these to streaming so that huge command outputs don't blow up the API container memory
// NOTE: This function will block while the exec is ongoing; if we need more perf we can make it async
func (backend *DockerKurtosisBackend) RunUserServiceExecCommands(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	userServiceCommands map[service.ServiceUUID][]string,
) (
	map[service.ServiceUUID]*exec_result.ExecResult,
	map[service.ServiceUUID]error,
	error,
) {
	return user_service_functions.RunUserServiceExecCommands(ctx, enclaveUuid, userServiceCommands, backend.dockerManager)
}

func (backend *DockerKurtosisBackend) GetConnectionWithUserService(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	serviceUuid service.ServiceUUID,
) (
	net.Conn,
	error,
) {
	return user_service_functions.GetConnectionWithUserService(ctx, enclaveUuid, serviceUuid, backend.dockerManager)
}

// It returns io.ReadCloser which is a tar stream. It's up to the caller to close the reader.
func (backend *DockerKurtosisBackend) CopyFilesFromUserService(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	serviceUuid service.ServiceUUID,
	srcPathOnContainer string,
	output io.Writer,
) error {
	return user_service_functions.CopyFilesFromUserService(ctx, enclaveUuid, serviceUuid, srcPathOnContainer, output, backend.dockerManager)
}

func (backend *DockerKurtosisBackend) StopUserServices(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	filters *service.ServiceFilters,
) (
	resultSuccessfulServiceUUIDs map[service.ServiceUUID]bool,
	resultErroredServiceUUIDs map[service.ServiceUUID]error,
	resultErr error,
) {
	return user_service_functions.StopUserServices(ctx, enclaveUuid, filters, backend.dockerManager)
}

func (backend *DockerKurtosisBackend) DestroyUserServices(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	filters *service.ServiceFilters,
) (
	resultSuccessfulUuids map[service.ServiceUUID]bool,
	resultErroredUuids map[service.ServiceUUID]error,
	resultErr error,
) {
	serviceRegistrationsForEnclave, found := backend.serviceRegistrations[enclaveUuid]
	if !found {
		return nil, nil, stacktrace.NewError(
			"No service registrations are being tracked for enclave '%v'; this likely means that the registration "+
				"request is being called where it shouldn't be (i.e. outside the API container)",
			enclaveUuid,
		)
	}

	freeIpAddrProviderForEnclave, found := backend.enclaveFreeIpProviders[enclaveUuid]
	if !found {
		return nil, nil, stacktrace.NewError(
			"Received a request to start services in enclave '%v', but no free IP address provider was "+
				"defined for this enclave; this likely means that the start request is being called where it shouldn't "+
				"be (i.e. outside the API container)",
			enclaveUuid,
		)
	}
	successfullyDestroyedServices, failedServices, err := user_service_functions.DestroyUserServices(
		ctx,
		enclaveUuid,
		filters,
		serviceRegistrationsForEnclave,
		backend.serviceRegistrationMutex,
		freeIpAddrProviderForEnclave,
		backend.dockerManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "Unexpected error destroying services in enclave '%s'", enclaveUuid)
	}
	return successfullyDestroyedServices, failedServices, nil
}

func (backend *DockerKurtosisBackend) CreateLogsDatabase(
	ctx context.Context,
	logsDatabaseHttpPortNumber uint16,
) (
	*logs_database.LogsDatabase,
	error,
) {

	//Declaring the implementation
	logsDatabaseContainer := loki.NewLokiLogDatabaseContainer()

	logsDatabase, err := logs_database_functions.CreateLogsDatabase(
		ctx,
		logsDatabaseHttpPortNumber,
		logsDatabaseContainer,
		backend.dockerManager,
		backend.objAttrsProvider,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the logs database using the logs database container '%+v' and the HTTP port number '%v'", logsDatabaseContainer, logsDatabaseHttpPortNumber)
	}
	return logsDatabase, nil
}

// If nothing is found returns nil
func (backend *DockerKurtosisBackend) GetLogsDatabase(
	ctx context.Context,
) (
	resultMaybeLogsDatabase *logs_database.LogsDatabase,
	resultErr error,
) {
	maybeLogsDatabase, err := logs_database_functions.GetLogsDatabase(
		ctx,
		backend.dockerManager,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the logs database")
	}

	return maybeLogsDatabase, nil
}

func (backend *DockerKurtosisBackend) DestroyLogsDatabase(
	ctx context.Context,
) error {

	if err := logs_database_functions.DestroyLogsDatabase(
		ctx,
		backend.dockerManager,
	); err != nil {
		return stacktrace.Propagate(err, "An error occurred destroying the logs database")
	}

	return nil
}

func (backend *DockerKurtosisBackend) CreateLogsCollectorForEnclave(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	logsCollectorTcpPortNumber uint16,
	logsCollectorHttpPortNumber uint16,
) (
	*logs_collector.LogsCollector,
	error,
) {

	//TODO we we'd have to replace this part if we ever wanted to send to an external source
	logsDatabase, err := backend.GetLogsDatabase(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the logs database; the logs collector cannot be run without a logs database")
	}

	if logsDatabase == nil || logsDatabase.GetStatus() != container_status.ContainerStatus_Running {
		return nil, stacktrace.NewError("The logs database is not running; the logs collector cannot be run without a running logs database")
	}

	//Declaring the implementation
	logsCollectorContainer := fluentbit.NewFluentbitLogsCollectorContainer()

	logsCollector, err := logs_collector_functions.CreateLogsCollectorForEnclave(
		ctx,
		enclaveUuid,
		logsCollectorTcpPortNumber,
		logsCollectorHttpPortNumber,
		logsCollectorContainer,
		logsDatabase,
		backend.dockerManager,
		backend.objAttrsProvider,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the logs collector using the '%v' TCP port number, the '%v' HTTP port number and the los collector container '%+v'", logsCollectorTcpPortNumber, logsCollectorHttpPortNumber, logsCollectorContainer)
	}

	return logsCollector, nil
}

// If nothing is found returns nil
func (backend *DockerKurtosisBackend) GetLogsCollectorForEnclave(ctx context.Context, enclaveUuid enclave.EnclaveUUID) (resultMaybeLogsCollector *logs_collector.LogsCollector, resultErr error) {
	maybeLogsCollector, err := logs_collector_functions.GetLogsCollectorForEnclave(
		ctx,
		enclaveUuid,
		backend.dockerManager,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the logs collector")
	}

	return maybeLogsCollector, nil
}

func (backend *DockerKurtosisBackend) DestroyLogsCollectorForEnclave(ctx context.Context, enclaveUuid enclave.EnclaveUUID) error {

	if err := logs_collector_functions.DestroyLogsCollector(ctx, enclaveUuid, backend.dockerManager); err != nil {
		return stacktrace.Propagate(err, "An error occurred destroying the logs collector")
	}

	return nil
}

// DestroyDeprecatedCentralizedLogsResources Destroy the deprecated centralized logs resources (containers and volumes)
// It doesn't complain if it couldn't find the centralized logs resources
// TODO(centralized-logs-resources-deprecation) remove this once we know people are on > 0.68.0
func (backend *DockerKurtosisBackend) DestroyDeprecatedCentralizedLogsResources(ctx context.Context) error {

	if err := logs_database_functions.DestroyDeprecatedCentralizedLogsDatabase(ctx, backend.dockerManager); err != nil {
		return stacktrace.Propagate(err, "An error occurred while destroying the deprecated centralized logs database")
	}

	if err := logs_collector_functions.DestroyDeprecatedCentralizedLogsCollectors(ctx, backend.dockerManager); err != nil {
		return stacktrace.Propagate(err, "An error occurred while destroying the deprecated centralized logs collector")
	}
	return nil
}

// ====================================================================================================
//
//	Private helper functions shared by multiple subfunctions files
//
// ====================================================================================================
func (backend *DockerKurtosisBackend) getEnclaveNetworkByEnclaveUuid(ctx context.Context, enclaveUuid enclave.EnclaveUUID) (*types.Network, error) {
	networkSearchLabels := map[string]string{
		label_key_consts.AppIDDockerLabelKey.GetString():       label_value_consts.AppIDDockerLabelValue.GetString(),
		label_key_consts.EnclaveUUIDDockerLabelKey.GetString(): string(enclaveUuid),
	}

	enclaveNetworksFound, err := backend.dockerManager.GetNetworksByLabels(ctx, networkSearchLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Docker networks by enclave ID '%v'", enclaveUuid)
	}
	numMatchingNetworks := len(enclaveNetworksFound)
	if numMatchingNetworks == 0 {
		return nil, stacktrace.NewError("No network was found for enclave with ID '%v'", enclaveUuid)
	}
	if numMatchingNetworks > 1 {
		return nil, stacktrace.NewError(
			"Expected exactly one network matching enclave ID '%v', but got %v",
			enclaveUuid,
			numMatchingNetworks,
		)
	}
	return enclaveNetworksFound[0], nil
}

// Guaranteed to either return an enclave data volume name or throw an error
func (backend *DockerKurtosisBackend) getEnclaveDataVolumeByEnclaveUuid(ctx context.Context, enclaveUuid enclave.EnclaveUUID) (string, error) {
	volumeSearchLabels := map[string]string{
		label_key_consts.AppIDDockerLabelKey.GetString():       label_value_consts.AppIDDockerLabelValue.GetString(),
		label_key_consts.EnclaveUUIDDockerLabelKey.GetString(): string(enclaveUuid),
		label_key_consts.VolumeTypeDockerLabelKey.GetString():  label_value_consts.EnclaveDataVolumeTypeDockerLabelValue.GetString(),
	}
	foundVolumes, err := backend.dockerManager.GetVolumesByLabels(ctx, volumeSearchLabels)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting enclave data volumes matching labels '%+v'", volumeSearchLabels)
	}
	if len(foundVolumes) > 1 {
		return "", stacktrace.NewError("Found multiple enclave data volumes matching enclave ID '%v'; this should never happen", enclaveUuid)
	}
	if len(foundVolumes) == 0 {
		return "", stacktrace.NewError("No enclave data volume found for enclave '%v'", enclaveUuid)
	}
	volume := foundVolumes[0]
	return volume.Name, nil
}
