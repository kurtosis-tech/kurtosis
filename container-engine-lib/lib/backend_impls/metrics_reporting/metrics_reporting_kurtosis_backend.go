package metrics_reporting

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_build_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_registry_spec"
	"io"
	"time"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/api_container"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/compute_resources"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_download_mode"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/reverse_proxy"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
)

// TODO CALL THE METRICS LIBRARY EVENT-REGISTRATION FUNCTIONS HERE!!!!
type MetricsReportingKurtosisBackend struct {
	underlying backend_interface.KurtosisBackend
}

func NewMetricsReportingKurtosisBackend(underlying backend_interface.KurtosisBackend) *MetricsReportingKurtosisBackend {
	return &MetricsReportingKurtosisBackend{underlying: underlying}
}

func (backend *MetricsReportingKurtosisBackend) FetchImage(ctx context.Context, image string, registrySpec *image_registry_spec.ImageRegistrySpec, downloadMode image_download_mode.ImageDownloadMode) (bool, string, error) {
	pulledFromRemote, architecture, err := backend.underlying.FetchImage(ctx, image, registrySpec, downloadMode)
	if err != nil {
		return false, "", stacktrace.Propagate(err, "An error occurred pulling image '%v'", image)
	}
	return pulledFromRemote, architecture, nil
}

func (backend *MetricsReportingKurtosisBackend) PruneUnusedImages(ctx context.Context) ([]string, error) {
	prunedImages, err := backend.underlying.PruneUnusedImages(ctx)
	if err != nil {
		return prunedImages, stacktrace.Propagate(err, "An error occurred pruning unused images")
	}
	return prunedImages, nil
}

func (backend *MetricsReportingKurtosisBackend) CreateEngine(
	ctx context.Context,
	imageOrgAndRepo string,
	imageVersionTag string,
	grpcPortNum uint16,
	envVars map[string]string,
	shouldStartInDebugMode bool,
) (*engine.Engine, error) {
	result, err := backend.underlying.CreateEngine(
		ctx,
		imageOrgAndRepo,
		imageVersionTag,
		grpcPortNum,
		envVars,
		shouldStartInDebugMode,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the engine using image '%v' with tag '%v' and debug mode '%v'", imageOrgAndRepo, imageVersionTag, shouldStartInDebugMode)
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

func (backend *MetricsReportingKurtosisBackend) GetEngineLogs(ctx context.Context, outputDirpath string) error {
	if err := backend.underlying.GetEngineLogs(ctx, outputDirpath); err != nil {
		return stacktrace.Propagate(err, "An error occurred while dumping engine logs to dir '%v'", outputDirpath)
	}
	return nil
}

func (backend *MetricsReportingKurtosisBackend) DumpKurtosis(ctx context.Context, outputDirpath string) error {
	if err := backend.underlying.DumpKurtosis(ctx, outputDirpath); err != nil {
		return stacktrace.Propagate(err, "An error occurred while dumping the state of Kurtosis to dir '%v'", outputDirpath)
	}
	return nil
}

func (backend *MetricsReportingKurtosisBackend) CreateEnclave(ctx context.Context, enclaveUuid enclave.EnclaveUUID, enclaveName string) (*enclave.Enclave, error) {
	result, err := backend.underlying.CreateEnclave(ctx, enclaveUuid, enclaveName)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating enclave with UUID '%v'", enclaveUuid)
	}
	return result, nil
}

func (backend *MetricsReportingKurtosisBackend) GetEnclaves(
	ctx context.Context,
	filters *enclave.EnclaveFilters,
) (
	map[enclave.EnclaveUUID]*enclave.Enclave,
	error,
) {
	results, err := backend.underlying.GetEnclaves(ctx, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclaves using filters: %+v", filters)
	}
	return results, nil
}

func (backend *MetricsReportingKurtosisBackend) UpdateEnclave(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	newName string,
	newCreationTime *time.Time,
) error {

	if err := backend.underlying.UpdateEnclave(ctx, enclaveUuid, newName, newCreationTime); err != nil {
		return stacktrace.Propagate(err, "An error occurred updating enclave with UUID '%v', updating name to '%s' and creation time to '%v'", enclaveUuid, newName, newCreationTime)
	}

	return nil
}

func (backend *MetricsReportingKurtosisBackend) StopEnclaves(
	ctx context.Context,
	filters *enclave.EnclaveFilters,
) (
	successfulEnclaveIds map[enclave.EnclaveUUID]bool,
	erroredEnclaveIds map[enclave.EnclaveUUID]error,
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
	enclaveUuid enclave.EnclaveUUID,
	outputDirpath string,
) error {
	if err := backend.underlying.DumpEnclave(ctx, enclaveUuid, outputDirpath); err != nil {
		return stacktrace.Propagate(err, "An error occurred dumping enclave '%v' to path '%v'", enclaveUuid, outputDirpath)
	}
	return nil
}

func (backend *MetricsReportingKurtosisBackend) DestroyEnclaves(
	ctx context.Context,
	filters *enclave.EnclaveFilters,
) (
	successfulEnclaveIds map[enclave.EnclaveUUID]bool,
	erroredEnclaveIds map[enclave.EnclaveUUID]error,
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
	enclaveUuid enclave.EnclaveUUID,
	grpcPortNum uint16,
	enclaveDataVolumeDirpath string,
	ownIpEnvVar string,
	customEnvVars map[string]string,
	shouldStartInDebugMode bool,
) (*api_container.APIContainer, error) {
	if _, found := customEnvVars[ownIpEnvVar]; found {
		return nil, stacktrace.NewError("Requested own IP environment variable '%v' conflicts with custom environment variable", ownIpEnvVar)
	}

	result, err := backend.underlying.CreateAPIContainer(
		ctx,
		image,
		enclaveUuid,
		grpcPortNum,
		enclaveDataVolumeDirpath,
		ownIpEnvVar,
		customEnvVars,
		shouldStartInDebugMode,
	)
	if err != nil {
		// WARNING: remember not to print 'customEnvVars' because it could end up creating a secret info leak
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating an API container from image '%v'",
			image,
		)
	}
	return result, nil
}

func (backend *MetricsReportingKurtosisBackend) GetAPIContainers(ctx context.Context, filters *api_container.APIContainerFilters) (map[enclave.EnclaveUUID]*api_container.APIContainer, error) {
	results, err := backend.underlying.GetAPIContainers(ctx, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting API containers matching filters: %+v", filters)
	}
	return results, nil
}

func (backend *MetricsReportingKurtosisBackend) StopAPIContainers(ctx context.Context, filters *api_container.APIContainerFilters) (successfulApiContainerIds map[enclave.EnclaveUUID]bool, erroredApiContainerIds map[enclave.EnclaveUUID]error, resultErr error) {
	successes, failures, err := backend.underlying.StopAPIContainers(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred stopping API containers using filters: %+v", filters)
	}
	return successes, failures, nil
}

func (backend *MetricsReportingKurtosisBackend) DestroyAPIContainers(ctx context.Context, filters *api_container.APIContainerFilters) (successfulApiContainerIds map[enclave.EnclaveUUID]bool, erroredApiContainerIds map[enclave.EnclaveUUID]error, resultErr error) {
	successes, failures, err := backend.underlying.DestroyAPIContainers(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred destroying API containers using filters: %+v", filters)
	}
	return successes, failures, nil
}

func (backend *MetricsReportingKurtosisBackend) RegisterUserServices(ctx context.Context, enclaveUuid enclave.EnclaveUUID, services map[service.ServiceName]bool) (map[service.ServiceName]*service.ServiceRegistration, map[service.ServiceName]error, error) {
	successes, failures, err := backend.underlying.RegisterUserServices(ctx, enclaveUuid, services)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred registering services to enclave '%v' with the following service ids: %+v", enclaveUuid, services)
	}
	return successes, failures, nil
}

func (backend *MetricsReportingKurtosisBackend) UnregisterUserServices(ctx context.Context, enclaveUuid enclave.EnclaveUUID, services map[service.ServiceUUID]bool) (map[service.ServiceUUID]bool, map[service.ServiceUUID]error, error) {
	successes, failures, err := backend.underlying.UnregisterUserServices(ctx, enclaveUuid, services)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred unregistering services from enclave '%v' with the following service uuids: %+v", enclaveUuid, services)
	}
	return successes, failures, nil
}

func (backend *MetricsReportingKurtosisBackend) StartRegisteredUserServices(ctx context.Context, enclaveUuid enclave.EnclaveUUID, services map[service.ServiceUUID]*service.ServiceConfig) (map[service.ServiceUUID]*service.Service, map[service.ServiceUUID]error, error) {
	successes, failures, err := backend.underlying.StartRegisteredUserServices(ctx, enclaveUuid, services)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred starting services in enclave '%v' with the following service ids: %+v", enclaveUuid, services)
	}
	return successes, failures, nil
}

func (backend *MetricsReportingKurtosisBackend) RemoveRegisteredUserServiceProcesses(ctx context.Context, enclaveUuid enclave.EnclaveUUID, services map[service.ServiceUUID]bool) (map[service.ServiceUUID]bool, map[service.ServiceUUID]error, error) {
	successes, failures, err := backend.underlying.RemoveRegisteredUserServiceProcesses(ctx, enclaveUuid, services)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred removing service processes in enclave '%v' with the following service ids: %+v", enclaveUuid, services)
	}
	return successes, failures, nil
}

func (backend *MetricsReportingKurtosisBackend) GetUserServices(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	filters *service.ServiceFilters,
) (
	map[service.ServiceUUID]*service.Service,
	error,
) {
	services, err := backend.underlying.GetUserServices(ctx, enclaveUuid, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting user services in enclave '%v' using filters '%+v'", enclaveUuid, filters)
	}
	return services, nil
}

func (backend *MetricsReportingKurtosisBackend) GetUserServiceLogs(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	filters *service.ServiceFilters,
	shouldFollowLogs bool,
) (
	map[service.ServiceUUID]io.ReadCloser,
	map[service.ServiceUUID]error,
	error,
) {
	userServiceLogs, erroredUserServices, err := backend.underlying.GetUserServiceLogs(ctx, enclaveUuid, filters, shouldFollowLogs)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting user service logs in enclave '%v' using filters '%+v'", enclaveUuid, filters)
	}
	return userServiceLogs, erroredUserServices, nil
}

func (backend *MetricsReportingKurtosisBackend) RunUserServiceExecCommands(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	userServiceCommands map[service.ServiceUUID][]string,
) (
	succesfulUserServiceExecResults map[service.ServiceUUID]*exec_result.ExecResult,
	erroredUserServiceUuids map[service.ServiceUUID]error,
	resultErr error,
) {
	succesfulUserServiceExecResults, erroredUserServiceUuids, err := backend.underlying.RunUserServiceExecCommands(ctx, enclaveUuid, userServiceCommands)
	if err != nil {
		return nil, nil, stacktrace.Propagate(
			err,
			"An error occurred running user service exec commands '%+v' on enclave '%v'",
			userServiceCommands,
			enclaveUuid,
		)
	}
	return succesfulUserServiceExecResults, erroredUserServiceUuids, nil
}

func (backend *MetricsReportingKurtosisBackend) RunUserServiceExecCommandWithStreamedOutput(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	serviceUuid service.ServiceUUID,
	cmd []string,
) (chan string, chan *exec_result.ExecResult, error) {
	return backend.underlying.RunUserServiceExecCommandWithStreamedOutput(ctx, enclaveUuid, serviceUuid, cmd)
}

func (backend *MetricsReportingKurtosisBackend) GetShellOnUserService(ctx context.Context, enclaveUuid enclave.EnclaveUUID, serviceUuid service.ServiceUUID) (resultErr error) {
	err := backend.underlying.GetShellOnUserService(ctx, enclaveUuid, serviceUuid)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting connection with user service with UUID '%v'", serviceUuid)
	}
	return nil
}

func (backend *MetricsReportingKurtosisBackend) CopyFilesFromUserService(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	serviceUuid service.ServiceUUID,
	srcPath string,
	output io.Writer,
) error {
	if err := backend.underlying.CopyFilesFromUserService(ctx, enclaveUuid, serviceUuid, srcPath, output); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred copying files from sourcepath '%v' in user service with UUID '%v' in enclave with UUID '%v'",
			srcPath,
			serviceUuid,
			enclaveUuid,
		)
	}
	return nil
}

func (backend *MetricsReportingKurtosisBackend) StopUserServices(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	filters *service.ServiceFilters,
) (
	successfulUserServiceUuids map[service.ServiceUUID]bool,
	erroredUserServiceUuids map[service.ServiceUUID]error,
	resultErr error,
) {
	successes, failures, err := backend.underlying.StopUserServices(ctx, enclaveUuid, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred stopping user services in enclave '%v' using filters: %+v", enclaveUuid, filters)
	}
	return successes, failures, nil
}

func (backend *MetricsReportingKurtosisBackend) DestroyUserServices(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	filters *service.ServiceFilters,
) (
	successfulUserServiceUuids map[service.ServiceUUID]bool,
	erroredUserServiceUuids map[service.ServiceUUID]error,
	resultErr error,
) {
	successes, failures, err := backend.underlying.DestroyUserServices(ctx, enclaveUuid, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred destroying user services using filters: %+v", filters)
	}
	return successes, failures, nil
}

func (backend *MetricsReportingKurtosisBackend) CreateLogsAggregator(ctx context.Context) (*logs_aggregator.LogsAggregator, error) {
	return backend.underlying.CreateLogsAggregator(ctx)
}

func (backend *MetricsReportingKurtosisBackend) GetLogsAggregator(ctx context.Context) (*logs_aggregator.LogsAggregator, error) {
	return backend.underlying.GetLogsAggregator(ctx)
}

func (backend *MetricsReportingKurtosisBackend) DestroyLogsAggregator(ctx context.Context) error {
	return backend.underlying.DestroyLogsAggregator(ctx)
}

func (backend *MetricsReportingKurtosisBackend) CreateLogsCollectorForEnclave(ctx context.Context, enclaveUuid enclave.EnclaveUUID, logsCollectorHttpPortNumber uint16, logsCollectorTcpPortNumber uint16) (*logs_collector.LogsCollector, error) {

	logsCollector, err := backend.underlying.CreateLogsCollectorForEnclave(ctx, enclaveUuid, logsCollectorHttpPortNumber, logsCollectorTcpPortNumber)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the logs collector with TCP port number '%v' and HTTP port number '%v'", logsCollectorTcpPortNumber, logsCollectorHttpPortNumber)
	}

	return logsCollector, nil
}

// if nothing is found returns nil
func (backend *MetricsReportingKurtosisBackend) GetLogsCollectorForEnclave(ctx context.Context, enclaveUuid enclave.EnclaveUUID) (resultMaybeLogsCollector *logs_collector.LogsCollector, resultErr error) {
	maybeLogsCollector, err := backend.underlying.GetLogsCollectorForEnclave(ctx, enclaveUuid)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the logs collector")
	}

	return maybeLogsCollector, nil
}

func (backend *MetricsReportingKurtosisBackend) DestroyLogsCollectorForEnclave(ctx context.Context, enclaveUuid enclave.EnclaveUUID) error {

	if err := backend.underlying.DestroyLogsCollectorForEnclave(ctx, enclaveUuid); err != nil {
		return stacktrace.Propagate(err, "An error occurred destroying the logs collector")
	}

	return nil
}

func (backend *MetricsReportingKurtosisBackend) CreateReverseProxy(ctx context.Context, engineGuid engine.EngineGUID) (*reverse_proxy.ReverseProxy, error) {
	return backend.underlying.CreateReverseProxy(ctx, engineGuid)
}

func (backend *MetricsReportingKurtosisBackend) GetReverseProxy(ctx context.Context) (*reverse_proxy.ReverseProxy, error) {
	return backend.underlying.GetReverseProxy(ctx)
}

func (backend *MetricsReportingKurtosisBackend) DestroyReverseProxy(ctx context.Context) error {
	return backend.underlying.DestroyReverseProxy(ctx)
}

func (backend *MetricsReportingKurtosisBackend) GetAvailableCPUAndMemory(ctx context.Context) (compute_resources.MemoryInMegaBytes, compute_resources.CpuMilliCores, bool, error) {
	availableMemory, availableCpu, isResourceInformationComplete, err := backend.underlying.GetAvailableCPUAndMemory(ctx)
	if err != nil {
		return 0, 0, false, stacktrace.Propagate(err, "An error occurred while fetching cpu & memory information from the underlying backend")
	}
	return availableMemory, availableCpu, isResourceInformationComplete, nil
}

func (backend *MetricsReportingKurtosisBackend) BuildImage(ctx context.Context, imageName string, imageBuildSpec *image_build_spec.ImageBuildSpec) (string, error) {
	return backend.underlying.BuildImage(ctx, imageName, imageBuildSpec)
}
