package docker_kurtosis_backend

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_build_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_registry_spec"
	"io"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/engine_functions"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/logs_aggregator_functions"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/logs_aggregator_functions/implementations/vector"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/logs_collector_functions"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/logs_collector_functions/implementations/fluentbit"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/reverse_proxy_functions"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/reverse_proxy_functions/implementations/traefik"
	user_service_functions "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/user_services_functions"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_network_allocator"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/compute_resources"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_download_mode"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/reverse_proxy"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db/free_ip_addr_tracker"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db/service_registration"
	"github.com/kurtosis-tech/stacktrace"
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
	//  container API instead.
	// This map is set exactly once, upon creation of the DockerKubernetesBackend, and never modified afterwards. Therefore, it doesn't need to be protected with a mutex (because the FreeIPProviders are themselves threadsafe)
	enclaveFreeIpProviders map[enclave.EnclaveUUID]*free_ip_addr_tracker.FreeIpAddrTracker

	serviceRegistrationRepository *service_registration.ServiceRegistrationRepository

	productionMode bool

	// Control concurrent access to serviceRegistrations
	serviceRegistrationMutex *sync.Mutex
}

func NewDockerKurtosisBackend(
	dockerManager *docker_manager.DockerManager,
	enclaveFreeIpProviders map[enclave.EnclaveUUID]*free_ip_addr_tracker.FreeIpAddrTracker,
	serviceRegistrationRepository *service_registration.ServiceRegistrationRepository,
	productionMode bool,
) *DockerKurtosisBackend {
	dockerNetworkAllocator := docker_network_allocator.NewDockerNetworkAllocator(dockerManager)
	return &DockerKurtosisBackend{
		dockerManager:                 dockerManager,
		dockerNetworkAllocator:        dockerNetworkAllocator,
		objAttrsProvider:              object_attributes_provider.GetDockerObjectAttributesProvider(),
		enclaveFreeIpProviders:        enclaveFreeIpProviders,
		serviceRegistrationRepository: serviceRegistrationRepository,
		productionMode:                productionMode,
		serviceRegistrationMutex:      &sync.Mutex{},
	}
}

func (backend *DockerKurtosisBackend) FetchImage(ctx context.Context, image string, registrySpec *image_registry_spec.ImageRegistrySpec, downloadMode image_download_mode.ImageDownloadMode) (bool, string, error) {
	return backend.dockerManager.FetchImage(ctx, image, registrySpec, downloadMode)
}

func (backend *DockerKurtosisBackend) PruneUnusedImages(ctx context.Context) ([]string, error) {
	prunedImages, err := backend.dockerManager.PruneUnusedImages(ctx)
	prunedImageNames := []string{}
	for _, prunedImage := range prunedImages {
		if lenPrunedImageTags := len(prunedImage.RepoTags); lenPrunedImageTags != 1 {
			return nil, stacktrace.NewError("Expected exactly one repo tag, but found %d (%v). This is a bug in Kurtosis.", lenPrunedImageTags, prunedImage.RepoTags)
		}
		prunedImageNames = append(prunedImageNames, prunedImage.RepoTags[0])
	}
	if err != nil {
		return prunedImageNames, stacktrace.Propagate(err, "An error occurred pruning image from kurtosis backend")
	}
	return prunedImageNames, nil
}

func (backend *DockerKurtosisBackend) CreateEngine(
	ctx context.Context,
	imageOrgAndRepo string,
	imageVersionTag string,
	grpcPortNum uint16,
	envVars map[string]string,
	shouldStartInDebugMode bool,
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
		shouldStartInDebugMode,
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
	freeIpAddrProviderForEnclave, found := backend.enclaveFreeIpProviders[enclaveUuid]
	if !found {
		return nil, nil, stacktrace.NewError(
			"Received a request to register services in enclave '%v', but no free IP address provider was "+
				"defined for this enclave; this likely means that the start request is being called where it shouldn't "+
				"be (i.e. outside the API container)",
			enclaveUuid,
		)
	}

	registeredService, failedServices, err := user_service_functions.RegisterUserServices(enclaveUuid, services, backend.serviceRegistrationRepository, freeIpAddrProviderForEnclave, backend.serviceRegistrationMutex)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "Unexpected error registering services to enclave '%s'", enclaveUuid)
	}
	return registeredService, failedServices, nil
}

func (backend *DockerKurtosisBackend) UnregisterUserServices(_ context.Context, enclaveUuid enclave.EnclaveUUID, services map[service.ServiceUUID]bool) (map[service.ServiceUUID]bool, map[service.ServiceUUID]error, error) {
	freeIpAddrProviderForEnclave, found := backend.enclaveFreeIpProviders[enclaveUuid]
	if !found {
		return nil, nil, stacktrace.NewError(
			"Received a request to unregister services in enclave '%v', but no free IP address provider was "+
				"defined for this enclave; this likely means that the start request is being called where it shouldn't "+
				"be (i.e. outside the API container)",
			enclaveUuid,
		)
	}

	servicesSuccessfullyUnregistered, failedServices, err := user_service_functions.UnregisterUserServices(enclaveUuid, services, backend.serviceRegistrationRepository, freeIpAddrProviderForEnclave, backend.serviceRegistrationMutex)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred unregistering services '%+v' for enclave with UUID '%v'", services, enclaveUuid)
	}

	return servicesSuccessfullyUnregistered, failedServices, nil
}

func (backend *DockerKurtosisBackend) StartRegisteredUserServices(ctx context.Context, enclaveUuid enclave.EnclaveUUID, services map[service.ServiceUUID]*service.ServiceConfig) (map[service.ServiceUUID]*service.Service, map[service.ServiceUUID]error, error) {
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
	if logsCollector == nil || logsCollector.GetStatus() != container.ContainerStatus_Running {
		return nil, nil, stacktrace.NewError("The user services can't be started because no logs collector is running for it to send logs to.")
	}

	logsCollectorIpAddressInEnclaveNetwork := logsCollector.GetEnclaveNetworkIpAddress()
	if logsCollectorIpAddressInEnclaveNetwork == nil {
		return nil, nil, stacktrace.NewError("Expected the logs collector to have an ip address in the enclave network but it does not.")
	}

	logsCollectorAvailabilityChecker := fluentbit.NewFluentbitAvailabilityChecker(logsCollectorIpAddressInEnclaveNetwork, logsCollector.GetPrivateHttpPort().GetNumber())

	var restartPolicy docker_manager.RestartPolicy = docker_manager.NoRestart
	if backend.productionMode {
		restartPolicy = docker_manager.RestartAlways
	}

	successfullyStartedService, failedService, err := user_service_functions.StartRegisteredUserServices(
		ctx,
		enclaveUuid,
		services,
		backend.serviceRegistrationRepository,
		logsCollector,
		logsCollectorAvailabilityChecker,
		backend.objAttrsProvider,
		freeIpAddrProviderForEnclave,
		backend.dockerManager,
		restartPolicy)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "Unexpected error while starting user service")
	}
	return successfullyStartedService, failedService, nil
}

func (backend *DockerKurtosisBackend) RemoveRegisteredUserServiceProcesses(ctx context.Context, enclaveUuid enclave.EnclaveUUID, services map[service.ServiceUUID]bool) (map[service.ServiceUUID]bool, map[service.ServiceUUID]error, error) {
	successfullyStartedService, failedService, err := user_service_functions.RemoveRegisteredUserServiceProcesses(
		ctx,
		enclaveUuid,
		services,
		backend.serviceRegistrationRepository,
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

func (backend *DockerKurtosisBackend) RunUserServiceExecCommandWithStreamedOutput(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	serviceUuid service.ServiceUUID,
	cmd []string,
) (chan string, chan *exec_result.ExecResult, error) {
	return user_service_functions.RunUserServiceExecCommandWithStreamedOutput(ctx, enclaveUuid, serviceUuid, cmd, backend.dockerManager)
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
	return user_service_functions.StopUserServices(ctx, enclaveUuid, filters, backend.serviceRegistrationRepository, backend.dockerManager)
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
		backend.serviceRegistrationRepository,
		backend.serviceRegistrationMutex,
		freeIpAddrProviderForEnclave,
		backend.dockerManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "Unexpected error destroying services in enclave '%s'", enclaveUuid)
	}
	return successfullyDestroyedServices, failedServices, nil
}

func (backend *DockerKurtosisBackend) CreateLogsAggregator(ctx context.Context) (*logs_aggregator.LogsAggregator, error) {
	logsAggregatorContainer := vector.NewVectorLogsAggregatorContainer() //Declaring the implementation

	logsAggregator, _, err := logs_aggregator_functions.CreateLogsAggregator(
		ctx,
		logsAggregatorContainer,
		backend.dockerManager,
		backend.objAttrsProvider,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the logs aggregator using the logs aggregator container '%+v'.", logsAggregatorContainer)
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

func (backend *DockerKurtosisBackend) DestroyLogsAggregator(ctx context.Context) error {
	if err := logs_aggregator_functions.DestroyLogsAggregator(ctx, backend.dockerManager); err != nil {
		return stacktrace.Propagate(err, "An error occurred destroying the logs aggregator")
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
	var logsAggregator *logs_aggregator.LogsAggregator
	maybeLogsAggregator, err := logs_aggregator_functions.GetLogsAggregator(ctx, backend.dockerManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the logs aggregator. The logs collector cannot be run without a logs aggregator.")
	}
	if maybeLogsAggregator == nil {
		logrus.Warnf("Logs aggregator container does not exist. This is unexpected as docker should have restarted the container automatically.")
		logrus.Warnf("This can be fixed by restarting the engine using `kurtosis engine restart` and attempting to create the enclave again.")
		return nil, stacktrace.NewError("No logs aggregator container exists. The logs collector cannot be run without a logs aggregator.")
	}
	if maybeLogsAggregator.GetStatus() != container.ContainerStatus_Running {
		logrus.Warnf("Logs aggregator exists but is not running. Instead container status is '%v'. This is unexpected as docker should have restarted the container automatically.",
			maybeLogsAggregator.GetStatus())
		logrus.Warnf("This can be fixed by restarting the engine using `kurtosis engine restart` and attempting to create the enclave again.")
		return nil, stacktrace.NewError(
			"The logs aggregator container exists but is not running. Instead logs aggregator container status is '%v'. The logs collector cannot be run without a logs aggregator.",
			maybeLogsAggregator.GetStatus(),
		)
	}
	logsAggregator = maybeLogsAggregator

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

func (backend *DockerKurtosisBackend) CreateReverseProxy(ctx context.Context, engineGuid engine.EngineGUID) (*reverse_proxy.ReverseProxy, error) {
	reverseProxyContainer := traefik.NewTraefikReverseProxyContainer()

	reverseProxy, _, err := reverse_proxy_functions.CreateReverseProxy(
		ctx,
		engineGuid,
		reverseProxyContainer,
		backend.dockerManager,
		backend.objAttrsProvider,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the reverse proxy using the reverse proxy container '%+v'.", reverseProxyContainer)
	}
	return reverseProxy, nil
}

func (backend *DockerKurtosisBackend) GetReverseProxy(ctx context.Context) (*reverse_proxy.ReverseProxy, error) {
	maybeReverseProxy, err := reverse_proxy_functions.GetReverseProxy(
		ctx,
		backend.dockerManager,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the reverse proxy")
	}

	return maybeReverseProxy, nil
}

func (backend *DockerKurtosisBackend) DestroyReverseProxy(ctx context.Context) error {
	if err := reverse_proxy_functions.DestroyReverseProxy(ctx, backend.dockerManager); err != nil {
		return stacktrace.Propagate(err, "An error occurred destroying the reverse proxy")
	}

	return nil
}

func (backend *DockerKurtosisBackend) ConnectReverseProxyToNetwork(ctx context.Context, networkId string) error {
	if err := reverse_proxy_functions.ConnectReverseProxyToNetwork(ctx, backend.dockerManager, networkId); err != nil {
		return stacktrace.Propagate(err, "An error occurred connecting the reverse proxy to the network with ID '%v'", networkId)
	}

	return nil
}

func (backend *DockerKurtosisBackend) DisconnectReverseProxyFromNetwork(ctx context.Context, networkId string) error {
	if err := reverse_proxy_functions.DisconnectReverseProxyFromNetwork(ctx, backend.dockerManager, networkId); err != nil {
		return stacktrace.Propagate(err, "An error occurred disconnecting the reverse proxy from the network with ID '%v'", networkId)
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

func (backend *DockerKurtosisBackend) BuildImage(ctx context.Context, imageName string, imageBuildSpec *image_build_spec.ImageBuildSpec) (string, error) {
	return backend.dockerManager.BuildImage(ctx, imageName, imageBuildSpec)
}

// ====================================================================================================
//
//	Private helper functions shared by multiple subfunctions files
//
// ====================================================================================================
func (backend *DockerKurtosisBackend) getEnclaveNetworkByEnclaveUuid(ctx context.Context, enclaveUuid enclave.EnclaveUUID) (*types.Network, error) {
	networkSearchLabels := map[string]string{
		docker_label_key.AppIDDockerLabelKey.GetString():       label_value_consts.AppIDDockerLabelValue.GetString(),
		docker_label_key.EnclaveUUIDDockerLabelKey.GetString(): string(enclaveUuid),
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
		docker_label_key.AppIDDockerLabelKey.GetString():       label_value_consts.AppIDDockerLabelValue.GetString(),
		docker_label_key.EnclaveUUIDDockerLabelKey.GetString(): string(enclaveUuid),
		docker_label_key.VolumeTypeDockerLabelKey.GetString():  label_value_consts.EnclaveDataVolumeTypeDockerLabelValue.GetString(),
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
