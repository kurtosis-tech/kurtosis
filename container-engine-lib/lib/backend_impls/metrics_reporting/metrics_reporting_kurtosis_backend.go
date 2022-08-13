package metrics_reporting

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/api_container"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifacts_expansion"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/module"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/networking_sidecar"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"io"
	"net"
	"strings"
)

// TODO CALL THE METRICS LIBRARY EVENT-REGISTRATION FUNCTIONS HERE!!!!
type MetricsReportingKurtosisBackend struct {
	underlying backend_interface.KurtosisBackend
}

func NewMetricsReportingKurtosisBackend(underlying backend_interface.KurtosisBackend) *MetricsReportingKurtosisBackend {
	return &MetricsReportingKurtosisBackend{underlying: underlying}
}

func (backend *MetricsReportingKurtosisBackend) PullImage(image string) error {
	if err := backend.underlying.PullImage(image); err != nil {
		return stacktrace.Propagate(err, "An error occurred pulling image '%v'", image)
	}
	return nil
}

func (backend *MetricsReportingKurtosisBackend) CreateEngine(ctx context.Context, imageOrgAndRepo string, imageVersionTag string, grpcPortNum uint16, grpcProxyPortNum uint16, envVars map[string]string) (*engine.Engine, error) {
	result, err := backend.underlying.CreateEngine(
		ctx,
		imageOrgAndRepo,
		imageVersionTag,
		grpcPortNum,
		grpcProxyPortNum,
		envVars,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the engine using image '%v' with tag '%v'", imageOrgAndRepo, imageVersionTag)
	}
	return result, nil
}

// Gets point-in-time data about engines matching the given filters
func (backend *MetricsReportingKurtosisBackend) GetEngines(ctx context.Context, filters *engine.EngineFilters) (map[engine.EngineGUID]*engine.Engine, error) {
	engines, err := backend.underlying.GetEngines(ctx, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting engines using filters: %+v", filters)
	}
	return engines, nil
}

func (backend *MetricsReportingKurtosisBackend) StopEngines(ctx context.Context, filters *engine.EngineFilters) (
	successfulIds map[engine.EngineGUID]bool,
	failedIds map[engine.EngineGUID]error,
	resultErr error,
) {
	successes, failures, err := backend.underlying.StopEngines(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred stopping engines using filters: %+v", filters)
	}
	return successes, failures, nil
}

func (backend *MetricsReportingKurtosisBackend) DestroyEngines(ctx context.Context, filters *engine.EngineFilters) (
	successfulIds map[engine.EngineGUID]bool,
	failedIds map[engine.EngineGUID]error,
	resultErr error,
) {
	successes, failures, err := backend.underlying.DestroyEngines(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred destroying engines using filters: %+v", filters)
	}
	return successes, failures, nil
}

func (backend *MetricsReportingKurtosisBackend) CreateEnclave(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	isPartitioningEnabled bool,
) (*enclave.Enclave, error) {
	result, err := backend.underlying.CreateEnclave(ctx, enclaveId, isPartitioningEnabled)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating enclave with ID '%v' and is-partitioning-enabled value '%v'", enclaveId, isPartitioningEnabled)
	}
	return result, nil
}

func (backend *MetricsReportingKurtosisBackend) GetEnclaves(
	ctx context.Context,
	filters *enclave.EnclaveFilters,
) (
	map[enclave.EnclaveID]*enclave.Enclave,
	error,
) {
	results, err := backend.underlying.GetEnclaves(ctx, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclaves using filters: %+v", filters)
	}
	return results, nil
}

func (backend *MetricsReportingKurtosisBackend) StopEnclaves(
	ctx context.Context,
	filters *enclave.EnclaveFilters,
) (
	successfulEnclaveIds map[enclave.EnclaveID]bool,
	erroredEnclaveIds map[enclave.EnclaveID]error,
	resultErr error,
) {
	successes, failures, err := backend.underlying.StopEnclaves(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred stopping enclaves using filters: %+v", filters)
	}
	return successes, failures, nil
}

func (backend *MetricsReportingKurtosisBackend) DumpEnclave(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	outputDirpath string,
) error {
	if err := backend.underlying.DumpEnclave(ctx, enclaveId, outputDirpath); err != nil {
		return stacktrace.Propagate(err, "An error occurred dumping enclave '%v' to path '%v'", enclaveId, outputDirpath)
	}
	return nil
}

func (backend *MetricsReportingKurtosisBackend) DestroyEnclaves(
	ctx context.Context,
	filters *enclave.EnclaveFilters,
) (
	successfulEnclaveIds map[enclave.EnclaveID]bool,
	erroredEnclaveIds map[enclave.EnclaveID]error,
	resultErr error,
) {
	successes, failures, err := backend.underlying.DestroyEnclaves(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred destroying enclaves using filters: %+v", filters)
	}
	return successes, failures, nil
}

func (backend *MetricsReportingKurtosisBackend) CreateAPIContainer(
	ctx context.Context,
	image string,
	enclaveId enclave.EnclaveID,
	grpcPortNum uint16,
	grpcProxyPortNum uint16,
	enclaveDataVolumeDirpath string,
	ownIpEnvVar string,
	customEnvVars map[string]string,
) (*api_container.APIContainer, error) {
	if _, found := customEnvVars[ownIpEnvVar]; found {
		return nil, stacktrace.NewError("Requested own IP environment variable '%v' conflicts with custom environment variable", ownIpEnvVar)
	}

	result, err := backend.underlying.CreateAPIContainer(
		ctx,
		image,
		enclaveId,
		grpcPortNum,
		grpcProxyPortNum,
		enclaveDataVolumeDirpath,
		ownIpEnvVar,
		customEnvVars,
	)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating an API container from image '%v' with envvars: %+v",
			image,
			customEnvVars,
		)
	}
	return result, nil
}

func (backend *MetricsReportingKurtosisBackend) GetAPIContainers(ctx context.Context, filters *api_container.APIContainerFilters) (map[enclave.EnclaveID]*api_container.APIContainer, error) {
	results, err := backend.underlying.GetAPIContainers(ctx, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting API containers matching filters: %+v", filters)
	}
	return results, nil
}

func (backend *MetricsReportingKurtosisBackend) StopAPIContainers(ctx context.Context, filters *api_container.APIContainerFilters) (successfulApiContainerIds map[enclave.EnclaveID]bool, erroredApiContainerIds map[enclave.EnclaveID]error, resultErr error) {
	successes, failures, err := backend.underlying.StopAPIContainers(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred stopping API containers using filters: %+v", filters)
	}
	return successes, failures, nil
}

func (backend *MetricsReportingKurtosisBackend) DestroyAPIContainers(ctx context.Context, filters *api_container.APIContainerFilters) (successfulApiContainerIds map[enclave.EnclaveID]bool, erroredApiContainerIds map[enclave.EnclaveID]error, resultErr error) {
	successes, failures, err := backend.underlying.DestroyAPIContainers(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred destroying API containers using filters: %+v", filters)
	}
	return successes, failures, nil
}

func (backend *MetricsReportingKurtosisBackend) CreateModule(
	ctx context.Context,
	image string,
	enclaveId enclave.EnclaveID,
	id module.ModuleID,
	grpcPortNum uint16,
	envVars map[string]string,
) (
	newModule *module.Module,
	resultErr error,
) {
	newModule, err := backend.underlying.CreateModule(
		ctx,
		image,
		enclaveId,
		id,
		grpcPortNum,
		envVars,
	)
	if err != nil {
		return nil,
			stacktrace.Propagate(
				err,
				"An error occurred creating module with ID '%v' in enclave '%v', and image '%v'",
				id,
				enclaveId,
				image,
			)
	}

	return newModule, nil
}

func (backend *MetricsReportingKurtosisBackend) GetModules(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *module.ModuleFilters,
) (
	map[module.ModuleGUID]*module.Module,
	error,
) {
	modules, err := backend.underlying.GetModules(ctx, enclaveId, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting modules in enclave '%v' using filters: %+v", enclaveId, filters)
	}
	return modules, nil
}

func (backend *MetricsReportingKurtosisBackend) GetModuleLogs(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *module.ModuleFilters,
	shouldFollowLogs bool,
) (
	map[module.ModuleGUID]io.ReadCloser,
	map[module.ModuleGUID]error,
	error,
) {
	moduleLogs, erroredModules, err := backend.underlying.GetModuleLogs(ctx, enclaveId, filters, shouldFollowLogs)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting module logs in enclave '%v' using filters '%+v'", enclaveId, filters)
	}
	return moduleLogs, erroredModules, nil
}

func (backend *MetricsReportingKurtosisBackend) StopModules(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *module.ModuleFilters,
) (
	successfulModuleIds map[module.ModuleGUID]bool,
	erroredModuleIds map[module.ModuleGUID]error,
	resultErr error,
) {
	successes, failures, err := backend.underlying.StopModules(ctx, enclaveId, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred stopping modules in enclave '%v' using filters: %+v", enclaveId, filters)
	}
	return successes, failures, nil
}

func (backend *MetricsReportingKurtosisBackend) DestroyModules(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *module.ModuleFilters,
) (
	successfulModuleIds map[module.ModuleGUID]bool,
	erroredModuleIds map[module.ModuleGUID]error,
	resultErr error,
) {
	successes, failures, err := backend.underlying.DestroyModules(ctx, enclaveId, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred destroying modules in enclave '%v' using filters: %+v", enclaveId, filters)
	}
	return successes, failures, nil
}

func (backend *MetricsReportingKurtosisBackend) RegisterUserService(ctx context.Context, enclaveId enclave.EnclaveID, serviceId service.ServiceID) (*service.ServiceRegistration, error) {
	serviceIdStr := string(serviceId)
	if len(strings.TrimSpace(serviceIdStr)) == 0 {
		return nil, stacktrace.NewError("Service ID cannot be whitespace or empty")
	}

	result, err := backend.underlying.RegisterUserService(ctx, enclaveId, serviceId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred registering user service '%v' in enclave '%v'", serviceId, enclaveId)
	}
	return result, nil
}

// Registers a user service for each given serviceId, allocating each an IP and ServiceGUID
func (backend *MetricsReportingKurtosisBackend) RegisterUserServices(ctx context.Context, enclaveId enclave.EnclaveID, serviceIds map[service.ServiceID]bool, ) (map[service.ServiceID]*service.ServiceRegistration, map[service.ServiceID]error, error){
	successes, failures, err := backend.underlying.RegisterUserServices(ctx, enclaveId, serviceIds)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred registering services in enclave '%v' with the following service ids: %+v", enclaveId, serviceIds)
	}
	return successes, failures, nil
}

func (backend *MetricsReportingKurtosisBackend) StartUserService(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	guid service.ServiceGUID,
	containerImageName string,
	privatePorts map[string]*port_spec.PortSpec,
	publicPorts map[string]*port_spec.PortSpec, //TODO this is a huge hack to temporarily enable static ports for NEAR until we have a more productized solution
	entrypointArgs []string,
	cmdArgs []string,
	envVars map[string]string,
	filesArtifactExpansion *files_artifacts_expansion.FilesArtifactsExpansion,
	cpuAllocationMillicpus uint64,
	memoryAllocationMegabytes uint64,
) (
	newUserService *service.Service,
	resultErr error,
) {
	userService, err := backend.underlying.StartUserService(
		ctx,
		enclaveId,
		guid,
		containerImageName,
		privatePorts,
		publicPorts,
		entrypointArgs,
		cmdArgs,
		envVars,
		filesArtifactExpansion,
		cpuAllocationMillicpus,
		memoryAllocationMegabytes,
	)
	if err != nil {
		filesArtifactMountDirpaths := []string{}
		if filesArtifactExpansion != nil {
			for _, mountpointOnUserService := range filesArtifactExpansion.ExpanderDirpathsToServiceDirpaths {
				filesArtifactMountDirpaths = append(filesArtifactMountDirpaths, mountpointOnUserService)
			}
		}
		return nil, stacktrace.Propagate(
			err,
			"An error occurred starting user service '%v' using image '%v' "+
				"with private ports '%+v' and entry point args '%+v', command args '%+v', environment "+
				"vars '%+v', and file artifacts mount dirpaths '%v'",
			guid,
			containerImageName,
			privatePorts,
			entrypointArgs,
			cmdArgs,
			envVars,
			strings.Join(filesArtifactMountDirpaths, ", "),
		)
	}
	return userService, nil
}

func (backend *MetricsReportingKurtosisBackend) StartUserServices(ctx context.Context, enclaveId enclave.EnclaveID, services map[service.ServiceGUID]*service.ServiceConfig) (map[service.ServiceGUID]*service.Service, map[service.ServiceGUID]error, error){
	successes, failures, err := backend.underlying.StartUserServices(ctx, enclaveId, services)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred starting services in enclave '%v' with the following service ids: %+v", enclaveId, services)
	}
	return successes, failures, nil
}

func (backend *MetricsReportingKurtosisBackend) GetUserServices(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *service.ServiceFilters,
) (
	map[service.ServiceGUID]*service.Service,
	error,
) {
	services, err := backend.underlying.GetUserServices(ctx, enclaveId, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting user services in enclave '%v' using filters '%+v'", enclaveId, filters)
	}
	return services, nil
}

func (backend *MetricsReportingKurtosisBackend) GetUserServiceLogs(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *service.ServiceFilters,
	shouldFollowLogs bool,
) (
	map[service.ServiceGUID]io.ReadCloser,
	map[service.ServiceGUID]error,
	error,
) {
	userServiceLogs, erroredUserServices, err := backend.underlying.GetUserServiceLogs(ctx, enclaveId, filters, shouldFollowLogs)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting user service logs in enclave '%v' using filters '%+v'", enclaveId, filters)
	}
	return userServiceLogs, erroredUserServices, nil
}

func (backend *MetricsReportingKurtosisBackend) PauseService(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceId service.ServiceGUID,
) error {
	err := backend.underlying.PauseService(ctx, enclaveId, serviceId)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to pause service '%v' in enclave '%v'", serviceId, enclaveId)
	}
	return nil
}

func (backend *MetricsReportingKurtosisBackend) UnpauseService(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceId service.ServiceGUID,
) error {
	err := backend.underlying.UnpauseService(ctx, enclaveId, serviceId)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to unpause service '%v' in enclave '%v'", serviceId, enclaveId)
	}
	return nil
}

func (backend *MetricsReportingKurtosisBackend) RunUserServiceExecCommands(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	userServiceCommands map[service.ServiceGUID][]string,
) (
	succesfulUserServiceExecResults map[service.ServiceGUID]*exec_result.ExecResult,
	erroredUserServiceGuids map[service.ServiceGUID]error,
	resultErr error,
) {
	succesfulUserServiceExecResults, erroredUserServiceGuids, err := backend.underlying.RunUserServiceExecCommands(ctx, enclaveId, userServiceCommands)
	if err != nil {
		return nil, nil, stacktrace.Propagate(
			err,
			"An error occurred running user service exec commands '%+v' on enclave '%v'",
			userServiceCommands,
			enclaveId,
		)
	}
	return succesfulUserServiceExecResults, erroredUserServiceGuids, nil
}

func (backend *MetricsReportingKurtosisBackend) GetConnectionWithUserService(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceGUID service.ServiceGUID,
) (
	resultConn net.Conn,
	resultErr error,
) {
	newConn, err := backend.underlying.GetConnectionWithUserService(ctx, enclaveId, serviceGUID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting connection with user service with GUID '%v'", serviceGUID)
	}
	return newConn, nil
}

func (backend *MetricsReportingKurtosisBackend) CopyFilesFromUserService(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceGuid service.ServiceGUID,
	srcPath string,
	output io.Writer,
) error {
	if err := backend.underlying.CopyFilesFromUserService(ctx, enclaveId, serviceGuid, srcPath, output); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred copying files from sourcepath '%v' in user service with GUID '%v' in enclave with ID '%v'",
			srcPath,
			serviceGuid,
			enclaveId,
		)
	}
	return nil
}

func (backend *MetricsReportingKurtosisBackend) StopUserServices(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *service.ServiceFilters,
) (
	successfulUserServiceGuids map[service.ServiceGUID]bool,
	erroredUserServiceGuids map[service.ServiceGUID]error,
	resultErr error,
) {
	successes, failures, err := backend.underlying.StopUserServices(ctx, enclaveId, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred stopping user services in enclave '%v' using filters: %+v", enclaveId, filters)
	}
	return successes, failures, nil
}

func (backend *MetricsReportingKurtosisBackend) DestroyUserServices(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *service.ServiceFilters,
) (
	successfulUserServiceGuids map[service.ServiceGUID]bool,
	erroredUserServiceGuids map[service.ServiceGUID]error,
	resultErr error,
) {
	successes, failures, err := backend.underlying.DestroyUserServices(ctx, enclaveId, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred destroying user services using filters: %+v", filters)
	}
	return successes, failures, nil
}

func (backend *MetricsReportingKurtosisBackend) CreateNetworkingSidecar(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceGuid service.ServiceGUID,
) (
	*networking_sidecar.NetworkingSidecar,
	error,
) {
	networkingSidecar, err := backend.underlying.CreateNetworkingSidecar(ctx, enclaveId, serviceGuid)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating networking sidecar for user service with GUID '%v' in enclave with ID '%v'", serviceGuid, enclaveId)
	}
	return networkingSidecar, nil
}

func (backend *MetricsReportingKurtosisBackend) GetNetworkingSidecars(
	ctx context.Context,
	filters *networking_sidecar.NetworkingSidecarFilters,
) (
	map[service.ServiceGUID]*networking_sidecar.NetworkingSidecar,
	error,
) {
	networkingSidecars, err := backend.underlying.GetNetworkingSidecars(ctx, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting networking sidecars using filters '%+v'", filters)
	}
	return networkingSidecars, nil
}

func (backend *MetricsReportingKurtosisBackend) RunNetworkingSidecarExecCommands(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	networkingSidecarsCommands map[service.ServiceGUID][]string,
) (
	map[service.ServiceGUID]*exec_result.ExecResult,
	map[service.ServiceGUID]error,
	error,
) {
	successfulNetworkingSidecarExecResults, erroredUserServiceGuids, err := backend.underlying.RunNetworkingSidecarExecCommands(ctx, enclaveId, networkingSidecarsCommands)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred running networking sidecar exec commands '%+v' in enclave with ID '%v'", networkingSidecarsCommands, enclaveId)
	}
	return successfulNetworkingSidecarExecResults, erroredUserServiceGuids, nil
}

func (backend *MetricsReportingKurtosisBackend) StopNetworkingSidecars(
	ctx context.Context,
	filters *networking_sidecar.NetworkingSidecarFilters,
) (
	map[service.ServiceGUID]bool,
	map[service.ServiceGUID]error,
	error,
) {
	successfulUserServiceGuids, erroredUserServiceGuids, err := backend.underlying.StopNetworkingSidecars(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred stopping networking sidecars using filters '%+v'", filters)
	}
	return successfulUserServiceGuids, erroredUserServiceGuids, nil
}

func (backend *MetricsReportingKurtosisBackend) DestroyNetworkingSidecars(
	ctx context.Context,
	filters *networking_sidecar.NetworkingSidecarFilters,
) (
	map[service.ServiceGUID]bool,
	map[service.ServiceGUID]error,
	error,
) {
	successfulUserServiceGuids, erroredUserServiceGuids, err := backend.underlying.DestroyNetworkingSidecars(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred destroying networking sidecars using filters '%+v'", filters)
	}
	return successfulUserServiceGuids, erroredUserServiceGuids, nil
}
