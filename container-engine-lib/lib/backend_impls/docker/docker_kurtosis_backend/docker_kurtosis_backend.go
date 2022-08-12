package docker_kurtosis_backend

import (
	"context"
	"errors"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/engine_functions"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/user_services_functions"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_network_allocator"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifacts_expansion"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/free-ip-addr-tracker-lib/lib"
	"github.com/kurtosis-tech/stacktrace"
	"io"
	"net"
	"strings"
	"sync"
)

// Unfortunately, Docker doesn't have an enum for the protocols it supports, so we have to create this translation map
var portSpecProtosToDockerPortProtos = map[port_spec.PortProtocol]string{
	port_spec.PortProtocol_TCP:  "tcp",
	port_spec.PortProtocol_SCTP: "sctp",
	port_spec.PortProtocol_UDP:  "udp",
}

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
	enclaveFreeIpProviders map[enclave.EnclaveID]*lib.FreeIpAddrTracker

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
	enclaveFreeIpProviders map[enclave.EnclaveID]*lib.FreeIpAddrTracker,
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


func (backend *DockerKurtosisBackend) RegisterUserService(ctx context.Context, enclaveId enclave.EnclaveID, serviceId service.ServiceID) (*service.ServiceRegistration, error) {
	return user_service_functions.RegisterUserService(ctx, enclaveId, serviceId, backend.serviceRegistrations, backend.serviceRegistrationMutex, backend.enclaveFreeIpProviders)
}

// Registers a user service for each given serviceId, allocating each an IP and ServiceGUID
func (backend *DockerKurtosisBackend ) RegisterUserServices(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceIds map[service.ServiceID]bool,
	) (successfulUserServiceRegistrations map[service.ServiceID]*service.ServiceRegistration, erroredUserServiceIds map[service.ServiceID]error, resultErr error) {
	return user_service_functions.RegisterUserServices(ctx, enclaveId, serviceIds, backend.serviceRegistrations, backend.serviceRegistrationMutex, backend.enclaveFreeIpProviders)
}

func (backend *DockerKurtosisBackend) StartUserService(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceGuid service.ServiceGUID,
	containerImageName string,
	privatePorts map[string]*port_spec.PortSpec,
	publicPorts map[string]*port_spec.PortSpec, //TODO this is a huge hack to temporarily enable static ports for NEAR until we have a more productized solution
	entrypointArgs []string,
	cmdArgs []string,
	envVars map[string]string,
	filesArtifactsExpansion *files_artifacts_expansion.FilesArtifactsExpansion,
	cpuAllocationMillicpus uint64,
	memoryAllocationMegabytes uint64,
) (*service.Service, error) {
	return user_service_functions.StartUserService(
		ctx,
		enclaveId,
		serviceGuid,
		containerImageName,
		privatePorts,
		publicPorts,
		entrypointArgs,
		cmdArgs,
		envVars,
		filesArtifactsExpansion,
		cpuAllocationMillicpus,
		memoryAllocationMegabytes,
		backend.serviceRegistrations,
		backend.serviceRegistrationMutex,
		backend.enclaveFreeIpProviders,
		backend.dockerManager,
		backend.objAttrsProvider)
}

func (backend *DockerKurtosisBackend) StartUserServices(ctx context.Context, enclaveId enclave.EnclaveID, services map[service.ServiceGUID]*service.ServiceConfig) (map[service.ServiceGUID]service.Service, map[service.ServiceGUID]error, error){
	return user_service_functions.StartUserServices(
		ctx,
		enclaveId,
		services,
		backend.serviceRegistrations,
		backend.serviceRegistrationMutex,
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


// ====================================================================================================
//                       Private helper functions shared by multiple subfunctions files
// ====================================================================================================
// TODO MOVE THIS TO WHOMEVER CALLS KURTOSISBACKEND
// This is a helper function that will take multiple errors, each identified by an ID, and format them together
// If no errors are returned, this function returns nil
func buildCombinedError(errorsById map[string]error, titleStr string) error {
	allErrorStrs := []string{}
	for errorId, stopErr := range errorsById {
		errorFormatStr := ">>>>>>>>>>>>> %v %v <<<<<<<<<<<<<\n" +
			"%v\n" +
			">>>>>>>>>>>>> END %v %v <<<<<<<<<<<<<"
		errorStr := fmt.Sprintf(
			errorFormatStr,
			strings.ToUpper(titleStr),
			errorId,
			stopErr.Error(),
			strings.ToUpper(titleStr),
			errorId,
		)
		allErrorStrs = append(allErrorStrs, errorStr)
	}

	if len(allErrorStrs) > 0 {
		// NOTE: This is one of the VERY rare cases where we don't want to use stacktrace.Propagate, because
		// attaching stack information for this method (which simply combines errors) just isn't useful. The
		// expected behaviour is that the caller of this function will use stacktrace.Propagate
		return errors.New(strings.Join(
			allErrorStrs,
			"\n\n",
		))
	}

	return nil
}

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
	volumeSearchLabels :=  map[string]string{
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



