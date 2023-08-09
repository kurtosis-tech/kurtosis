package docker_kurtosis_backend

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/engine_functions"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/logs_aggregator_functions"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/logs_aggregator_functions/implementations/vector"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/logs_collector_functions"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/logs_collector_functions/implementations/fluentbit"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/user_services_functions"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_network_allocator"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/compute_resources"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db/free_ip_addr_tracker"
	"github.com/kurtosis-tech/stacktrace"
	"io"
	"sync"
)

const (
	isResourceInformationComplete = true
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

func (backend *DockerKurtosisBackend) GetEngineLogs(ctx context.Context, outputDirpath string) error {
	return engine_functions.GetEngineLogs(ctx, outputDirpath, backend.dockerManager)
}

func (backend *DockerKurtosisBackend) DumpKurtosis(ctx context.Context, outputDirpath string) error {
	return engine_functions.DumpKurtosis(ctx, outputDirpath, backend)
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

	logsCollector, err := backend.GetLogsCollectorForEnclave(ctx, enclaveUuid)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting the logs collector")
	}
	if logsCollector == nil || logsCollector.GetStatus() != container_status.ContainerStatus_Running {
		return nil, nil, stacktrace.NewError("The user services can't be started because no logs collector is running for sending the logs to")
	}

	logsCollectorIpAddressInEnclaveNetwork := logsCollector.GetEnclaveNetworkIpAddress()
	if logsCollectorIpAddressInEnclaveNetwork == nil {
		return nil, nil, stacktrace.NewError("Expected the logs collector has ip address in enclave network but this is nil")
	}

	logsCollectorAvailabilityChecker := fluentbit.NewFluentbitAvailabilityChecker(logsCollectorIpAddressInEnclaveNetwork, logsCollector.GetPrivateHttpPort().GetNumber())

	successfullyStartedService, failedService, err := user_service_functions.StartRegisteredUserServices(
		ctx,
		enclaveUuid,
		services,
		serviceRegistrationsForEnclave,
		logsCollector,
		logsCollectorAvailabilityChecker,
		backend.objAttrsProvider,
		freeIpAddrProviderForEnclave,
		backend.dockerManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "Unexpected error while starting user service")
	}
	return successfullyStartedService, failedService, nil
}

func (backend *DockerKurtosisBackend) RemoveRegisteredUserServiceProcesses(ctx context.Context, enclaveUuid enclave.EnclaveUUID, services map[service.ServiceUUID]bool) (map[service.ServiceUUID]bool, map[service.ServiceUUID]error, error) {
	serviceRegistrationsForEnclave, found := backend.serviceRegistrations[enclaveUuid]
	if !found {
		return nil, nil, stacktrace.NewError(
			"No service registrations are being tracked for enclave '%v'; this likely means that the registration "+
				"request is being called where it shouldn't be (i.e. outside the API container)",
			enclaveUuid,
		)
	}

	successfullyStartedService, failedService, err := user_service_functions.RemoveRegisteredUserServiceProcesses(
		ctx,
		enclaveUuid,
		services,
		serviceRegistrationsForEnclave,
		backend.dockerManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "Unexpected error while updating user services")
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

func (backend *DockerKurtosisBackend) GetShellOnUserService(ctx context.Context, enclaveUuid enclave.EnclaveUUID, serviceUuid service.ServiceUUID) error {
	return user_service_functions.GetShellOnUserService(ctx, enclaveUuid, serviceUuid, backend.dockerManager)
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
	serviceRegistrationsForEnclave, found := backend.serviceRegistrations[enclaveUuid]
	if !found {
		return nil, nil, stacktrace.NewError(
			"No service registrations are being tracked for enclave '%v'",
			enclaveUuid,
		)
	}
	return user_service_functions.StopUserServices(ctx, enclaveUuid, filters, serviceRegistrationsForEnclave, backend.dockerManager)
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

func (backend *DockerKurtosisBackend) CreateLogsAggregator(ctx context.Context, logsAggregatorPortNum uint16) (*logs_aggregator.LogsAggregator, error) {
	logsAggregatorContainer := vector.NewVectorLogsAggregatorContainer() //Declaring the implementation

	logsAggregator, err := logs_aggregator_functions.CreateLogsAggregator(
		ctx,
		logsAggregatorPortNum,
		logsAggregatorContainer,
		backend.dockerManager,
		backend.objAttrsProvider,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the logs aggregator using the logs aggregator container '%+v' and the port number '%v'", logsAggregatorContainer, logsAggregatorPortNum)
	}
	return logsAggregator, nil
}

func (backend *DockerKurtosisBackend) GetLogsAggregator(ctx context.Context) (*logs_aggregator.LogsAggregator, error) {
	maybeLogsAggregator, err := logs_aggregator_functions.GetLogsAggregator(
		ctx,
		backend.dockerManager,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the logs aggregator")
	}

	return maybeLogsAggregator, nil
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
	logsAggregator, err := backend.GetLogsAggregator(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the logs aggregator; the logs collector cannot be run without a logs aggregator")
	}

	if logsAggregator == nil || logsAggregator.GetStatus() != container_status.ContainerStatus_Running {
		return nil, stacktrace.NewError("The logs aggregator is not running; the logs collector cannot be run without a running logs aggregator")
	}

	//Declaring the implementation
	logsCollectorContainer := fluentbit.NewFluentbitLogsCollectorContainer()

	logsCollector, err := logs_collector_functions.CreateLogsCollectorForEnclave(
		ctx,
		enclaveUuid,
		logsCollectorTcpPortNumber,
		logsCollectorHttpPortNumber,
		logsCollectorContainer,
		logsAggregator,
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

func (backend *DockerKurtosisBackend) GetAvailableCPUAndMemory(ctx context.Context) (compute_resources.MemoryInMegaBytes, compute_resources.CpuMilliCores, bool, error) {
	availableMemory, availableCpu, err := backend.dockerManager.GetAvailableCPUAndMemory(ctx)
	if err != nil {
		return 0, 0, false, stacktrace.Propagate(err, "an error occurred fetching resource information from the docker backend")
	}
	return availableMemory, availableCpu, isResourceInformationComplete, nil
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
