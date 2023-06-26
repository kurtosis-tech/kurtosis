package metrics_reporting

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
	"io"
)

// TODO CALL THE METRICS LIBRARY EVENT-REGISTRATION FUNCTIONS HERE!!!!
type MetricsReportingKurtosisBackend struct {
	underlying backend_interface.KurtosisBackend
}

func NewMetricsReportingKurtosisBackend(underlying backend_interface.KurtosisBackend) *MetricsReportingKurtosisBackend {
	return &MetricsReportingKurtosisBackend{underlying: underlying}
}

func (backend *MetricsReportingKurtosisBackend) FetchImage(ctx context.Context, image string) error {
	if err := backend.underlying.FetchImage(ctx, image); err != nil {
		return stacktrace.Propagate(err, "An error occurred pulling image '%v'", image)
	}
	return nil
}

func (backend *MetricsReportingKurtosisBackend) CreateEngine(
	ctx context.Context,
	imageOrgAndRepo string,
	imageVersionTag string,
	grpcPortNum uint16,
	envVars map[string]string,
) (*engine.Engine, error) {
	result, err := backend.underlying.CreateEngine(
		ctx,
		imageOrgAndRepo,
		imageVersionTag,
		grpcPortNum,
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

func (backend *MetricsReportingKurtosisBackend) CreateEnclave(ctx context.Context, enclaveUuid enclave.EnclaveUUID, enclaveName string, isPartitioningEnabled bool) (*enclave.Enclave, error) {
	result, err := backend.underlying.CreateEnclave(ctx, enclaveUuid, enclaveName, isPartitioningEnabled)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating enclave with UUID '%v' and is-partitioning-enabled value '%v'", enclaveUuid, isPartitioningEnabled)
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

func (backend *MetricsReportingKurtosisBackend) RenameEnclave(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	newName string,
) error {

	if err := backend.underlying.RenameEnclave(ctx, enclaveUuid, newName); err != nil {
		return stacktrace.Propagate(err, "An error occurred renaming enclave with UUID '%v' to '%s'", enclaveUuid, newName)
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

func (backend *MetricsReportingKurtosisBackend) CreateNetworkingSidecar(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	serviceUuid service.ServiceUUID,
) (
	*networking_sidecar.NetworkingSidecar,
	error,
) {
	networkingSidecar, err := backend.underlying.CreateNetworkingSidecar(ctx, enclaveUuid, serviceUuid)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating networking sidecar for user service with UUID '%v' in enclave with UUID '%v'", serviceUuid, enclaveUuid)
	}
	return networkingSidecar, nil
}

func (backend *MetricsReportingKurtosisBackend) GetNetworkingSidecars(
	ctx context.Context,
	filters *networking_sidecar.NetworkingSidecarFilters,
) (
	map[service.ServiceUUID]*networking_sidecar.NetworkingSidecar,
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
	enclaveUuid enclave.EnclaveUUID,
	networkingSidecarsCommands map[service.ServiceUUID][]string,
) (
	map[service.ServiceUUID]*exec_result.ExecResult,
	map[service.ServiceUUID]error,
	error,
) {
	successfulNetworkingSidecarExecResults, erroredUserServiceUuids, err := backend.underlying.RunNetworkingSidecarExecCommands(ctx, enclaveUuid, networkingSidecarsCommands)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred running networking sidecar exec commands '%+v' in enclave with UUID '%v'", networkingSidecarsCommands, enclaveUuid)
	}
	return successfulNetworkingSidecarExecResults, erroredUserServiceUuids, nil
}

func (backend *MetricsReportingKurtosisBackend) StopNetworkingSidecars(
	ctx context.Context,
	filters *networking_sidecar.NetworkingSidecarFilters,
) (
	map[service.ServiceUUID]bool,
	map[service.ServiceUUID]error,
	error,
) {
	successfulUserServiceUuids, erroredUserServiceUuids, err := backend.underlying.StopNetworkingSidecars(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred stopping networking sidecars using filters '%+v'", filters)
	}
	return successfulUserServiceUuids, erroredUserServiceUuids, nil
}

func (backend *MetricsReportingKurtosisBackend) DestroyNetworkingSidecars(
	ctx context.Context,
	filters *networking_sidecar.NetworkingSidecarFilters,
) (
	map[service.ServiceUUID]bool,
	map[service.ServiceUUID]error,
	error,
) {
	successfulUserServiceUuids, erroredUserServiceUuids, err := backend.underlying.DestroyNetworkingSidecars(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred destroying networking sidecars using filters '%+v'", filters)
	}
	return successfulUserServiceUuids, erroredUserServiceUuids, nil
}

func (backend *MetricsReportingKurtosisBackend) CreateLogsDatabase(
	ctx context.Context,
	logsDatabaseHttpPortNumber uint16,
) (
	*logs_database.LogsDatabase,
	error,
) {

	logsDatabase, err := backend.underlying.CreateLogsDatabase(ctx, logsDatabaseHttpPortNumber)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the logs database with HTTP port number '%v'", logsDatabaseHttpPortNumber)
	}

	return logsDatabase, nil
}

// if nothing is found returns nil
func (backend *MetricsReportingKurtosisBackend) GetLogsDatabase(
	ctx context.Context,
) (
	resultMaybeLogsDatabase *logs_database.LogsDatabase,
	resultErr error,
) {
	maybeLogsDatabase, err := backend.underlying.GetLogsDatabase(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the logs database")
	}

	return maybeLogsDatabase, nil
}

func (backend *MetricsReportingKurtosisBackend) DestroyLogsDatabase(
	ctx context.Context,
) error {
	if err := backend.underlying.DestroyLogsDatabase(ctx); err != nil {
		return stacktrace.Propagate(err, "An error occurred destroying the logs database")
	}

	return nil
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

func (backend *MetricsReportingKurtosisBackend) DestroyDeprecatedCentralizedLogsResources(ctx context.Context) error {
	if err := backend.underlying.DestroyDeprecatedCentralizedLogsResources(ctx); err != nil {
		return stacktrace.Propagate(err, "An error occurred while destroying deprecated logs collector")
	}
	return nil
}
