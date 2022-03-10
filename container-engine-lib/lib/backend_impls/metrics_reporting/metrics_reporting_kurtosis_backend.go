package metrics_reporting

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/api_container"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/module"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"io"
	"net"
)

// TODO CALL THE METRICS LIBRARY EVENT-REGISTRATION FUNCTIONS HERE!!!!
type MetricsReportingKurtosisBackend struct {
	underlying backend_interface.KurtosisBackend
}

func NewMetricsReportingKurtosisBackend(underlying backend_interface.KurtosisBackend) *MetricsReportingKurtosisBackend {
	return &MetricsReportingKurtosisBackend{underlying: underlying}
}

func (backend *MetricsReportingKurtosisBackend) CreateEngine(ctx context.Context, imageOrgAndRepo string, imageVersionTag string, grpcPortNum uint16, grpcProxyPortNum uint16, engineDataDirpathOnHostMachine string, envVars map[string]string) (*engine.Engine, error) {
	result, err := backend.underlying.CreateEngine(ctx, imageOrgAndRepo, imageVersionTag, grpcPortNum, grpcProxyPortNum, engineDataDirpathOnHostMachine, envVars)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the engine using image '%v' with tag '%v'", imageOrgAndRepo, imageVersionTag)
	}
	return result, nil
}

// Gets point-in-time data about engines matching the given filters
func (backend *MetricsReportingKurtosisBackend) GetEngines(ctx context.Context, filters *engine.EngineFilters) (map[string]*engine.Engine, error) {
	engines, err := backend.underlying.GetEngines(ctx, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting engines using filters: %+v", filters)
	}
	return engines, nil
}

func (backend *MetricsReportingKurtosisBackend) StopEngines(ctx context.Context, filters *engine.EngineFilters) (
	successfulIds map[string]bool,
	failedIds map[string]error,
	resultErr error,
) {
	successes, failures, err := backend.underlying.StopEngines(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred stopping engines using filters: %+v", filters)
	}
	return successes, failures, nil
}

func (backend *MetricsReportingKurtosisBackend) DestroyEngines(ctx context.Context, filters *engine.EngineFilters) (
	successfulIds map[string]bool,
	failedIds map[string]error,
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
	enclaveId string,
	isPartitioningEnabled bool,
)(*enclave.Enclave, error) {
	result, err := backend.underlying.CreateEnclave(ctx, enclaveId, isPartitioningEnabled)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating enclave with ID '%v' and is-partitioning-enabled value '%v'", enclaveId, isPartitioningEnabled)
	}
	return result, nil
}

func (backend *MetricsReportingKurtosisBackend) GetEnclaves(ctx context.Context, filters *enclave.EnclaveFilters) (map[string]*enclave.Enclave, error) {
	results, err := backend.underlying.GetEnclaves(ctx, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclaves using filters: %+v", filters)
	}
	return results, nil
}

func (backend *MetricsReportingKurtosisBackend) StopEnclaves(ctx context.Context, filters *enclave.EnclaveFilters) (successfulEnclaveIds map[string]bool, erroredEnclaveIds map[string]error, resultErr error) {
	successes, failures, err := backend.underlying.StopEnclaves(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred stopping enclaves using filters: %+v", filters)
	}
	return successes, failures, nil
}

func (backend *MetricsReportingKurtosisBackend) DestroyEnclaves(ctx context.Context, filters *enclave.EnclaveFilters) (successfulEnclaveIds map[string]bool, erroredEnclaveIds map[string]error, resultErr error) {
	successes, failures, err := backend.underlying.DestroyEnclaves(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred destroying enclaves using filters: %+v", filters)
	}
	return successes, failures, nil
}

func (backend *MetricsReportingKurtosisBackend) RepartitionEnclave(
	ctx context.Context,
	enclaveId string,
	servicesConnections map[service.ServiceID]map[service.ServiceID]enclave.NetworkConnection,
)(
	resultErr error,
) {
	if err := backend.underlying.RepartitionEnclave(
		ctx,
		enclaveId,
		servicesConnections,
	); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred creating repartition with servicesConnections '%+v' for enclave with ID '%v'",
			servicesConnections,
			enclaveId,
		)
	}
	return nil
}

func (backend *MetricsReportingKurtosisBackend) CreateAPIContainer(
	ctx context.Context,
	image string,
	grpcPortId string,
	grpcPortSpec *port_spec.PortSpec,
	grpcProxyPortId string,
	grpcProxyPortSpec *port_spec.PortSpec,
	enclaveDataDirpathOnHostMachine string,
	envVars map[string]string,
) (*api_container.APIContainer, error) {
	result, err := backend.underlying.CreateAPIContainer(
		ctx,
		image,
		grpcPortId,
		grpcPortSpec,
		grpcProxyPortId,
		grpcProxyPortSpec,
		enclaveDataDirpathOnHostMachine,
		envVars,
	)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating an API container from image '%v' with envvars: %+v",
			image,
			envVars,
		)
	}
	return result, nil
}

func (backend *MetricsReportingKurtosisBackend) GetAPIContainers(ctx context.Context, filters *api_container.APIContainerFilters) (map[string]*api_container.APIContainer, error) {
	results, err := backend.underlying.GetAPIContainers(ctx, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting API containers matching filters: %+v", filters)
	}
	return results, nil
}

func (backend *MetricsReportingKurtosisBackend) StopAPIContainers(ctx context.Context, filters *enclave.EnclaveFilters) (successApiContainerIds map[string]bool, erroredApiContainerIds map[string]error, resultErr error) {
	successes, failures, err := backend.underlying.StopAPIContainers(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred stopping API containers using filters: %+v", filters)
	}
	return successes, failures, nil
}

func (backend *MetricsReportingKurtosisBackend) DestroyAPIContainers(ctx context.Context, filters *enclave.EnclaveFilters) (successApiContainerIds map[string]bool, erroredApiContainerIds map[string]error, resultErr error) {
	successes, failures, err := backend.underlying.DestroyAPIContainers(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred destroying API containers using filters: %+v", filters)
	}
	return successes, failures, nil
}


func (backend *MetricsReportingKurtosisBackend) CreateModule(
	ctx context.Context,
	id string,
	containerImageName string,
	serializedParams string,
)(
	resultPrivateIp net.IP,
	resultPrivatePort *port_spec.PortSpec,
	resultPublicIp net.IP,
	resultPublicPort *port_spec.PortSpec,
	resultErr error,
) {
	privateIp, privatePort, publicIp, publicPort, err := backend.underlying.CreateModule(
		ctx,
		id,
		containerImageName,
		serializedParams,
		)
	if err != nil {
		return nil, nil, nil, nil,
		stacktrace.Propagate(
			err,
			"An error occurred creating module with ID '%v', container image name '%v' and serialized params '%+v'",
			id,
			containerImageName,
			serializedParams)
	}

	return privateIp, privatePort, publicIp, publicPort, nil
}

func (backend *MetricsReportingKurtosisBackend) GetModules(
	ctx context.Context,
	filters *module.ModuleFilters,
)(
	map[string]*module.Module,
	error,
) {
	modules, err := backend.underlying.GetModules(ctx, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting modules using filters: %+v", filters)
	}
	return modules, nil
}

func (backend *MetricsReportingKurtosisBackend) DestroyModules(
	ctx context.Context,
	filters *module.ModuleFilters,
)(
	successfulModuleIds map[string]bool,
	erroredModuleIds map[string]error,
	resultErr error,
) {
	successes, failures, err := backend.underlying.DestroyModules(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred destroying modules using filters: %+v", filters)
	}
	return successes, failures, nil
}

func (backend *MetricsReportingKurtosisBackend) CreateUserService(
	ctx context.Context,
	id string,
	containerImageName string,
	privatePorts []*port_spec.PortSpec,
	entrypointArgs []string,
	cmdArgs []string,
	envVars map[string]string,
	enclaveDataDirMntDirpath string,
	filesArtifactMountDirpaths map[string]string,
)(
	maybePublicIpAddr net.IP,
	publicPorts map[string]*port_spec.PortSpec,
	resultErr error,
) {
	publicIpAddr, publicPort, err := backend.underlying.CreateUserService(
		ctx,
		id,
		containerImageName,
		privatePorts,
		entrypointArgs,
		cmdArgs,
		envVars,
		enclaveDataDirMntDirpath,
		filesArtifactMountDirpaths,
		)
	if err != nil {
		return nil, nil,
		stacktrace.Propagate(
			err,
			"An error occurred creating the user service with ID '%v' using image '%v' with private ports '%+v' with entry point args '%+v', command args '%+v', environment vars '%+v', enclave data mount dirpath '%v' and file artifacts mount dirpath '%v'",
			id,
			containerImageName,
			privatePorts,
			entrypointArgs,
			cmdArgs,
			envVars,
			enclaveDataDirMntDirpath,
			filesArtifactMountDirpaths,
			)
	}
	return publicIpAddr, publicPort, nil
}

func (backend *MetricsReportingKurtosisBackend) GetUserServices(
	ctx context.Context,
	filters *service.ServiceFilters,
)(
	map[string]*service.Service,
	error,
){
	services, err := backend.underlying.GetUserServices(ctx, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting user services using filters '%+v'", filters)
	}
	return services, nil
}

func (backend *MetricsReportingKurtosisBackend) GetUserServiceLogs(
	ctx context.Context,
	filters *service.ServiceFilters,
)(
	map[string]io.ReadCloser,
	error,
) {
	userServiceLogs, err := backend.underlying.GetUserServiceLogs(ctx, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting user service logs using filters '%+v'", filters)
	}
	return userServiceLogs, nil
}

func (backend *MetricsReportingKurtosisBackend) RunUserServiceExecCommand (
	ctx context.Context,
	serviceId string,
	commandArgs []string,
)(
	resultExitCode int32,
	resultOutput string,
	resultErr error,
) {
	exitCode, output, err := backend.underlying.RunUserServiceExecCommand(ctx, serviceId, commandArgs)
	if err != nil {
		return 0, "", stacktrace.Propagate(
			err,
			"An error occurred running user service exec command with user service ID '%v' and command args '%+v'",
			serviceId,
			commandArgs,
			)
	}
	return exitCode, output, nil
}

func (backend *MetricsReportingKurtosisBackend) WaitForUserServiceHttpEndpointAvailability(
	ctx context.Context,
	serviceId string,
	httpMethod string,
	port uint32,
	path string,
	requestBody string,
	bodyText string,
	initialDelayMilliseconds uint32,
	retries uint32,
	retriesDelayMilliseconds uint32,
)(
	resultErr error,
) {
	if err := backend.underlying.WaitForUserServiceHttpEndpointAvailability(
		ctx,
		serviceId,
		httpMethod,
		port,
		path,
		requestBody,
		bodyText,
		initialDelayMilliseconds,
		retries,
		retriesDelayMilliseconds,
		); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred waiting for http endpoint with path '%v', port '%v', request body '%v', body text '%v' from service with ID '%v' to become available after '%v' retries and '%v' milliseconds between retries,",
			path,
			port,
			requestBody,
			bodyText,
			serviceId,
			retries,
			retriesDelayMilliseconds,
		)
	}
	return nil
}

func (backend *MetricsReportingKurtosisBackend) GetShellOnUserService(
	ctx context.Context,
	userServiceId string,
)(
	resultErr error,
) {
	if err := backend.underlying.GetShellOnUserService(ctx, userServiceId); err != nil {
		return stacktrace.Propagate(err, "An error occurred getting shell on user service with ID '%v'", userServiceId)
	}
	return nil
}

func (backend *MetricsReportingKurtosisBackend) StopUserServices(
	ctx context.Context,
	filters *service.ServiceFilters,
)(
	successfulUserServiceIds map[string]bool,
	erroredUserServiceIds map[string]error,
	resultErr error,
) {
	successes, failures, err := backend.underlying.StopUserServices(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred stopping user services using filters: %+v", filters)
	}
	return successes, failures, nil
}

func (backend *MetricsReportingKurtosisBackend) DestroyUserServices(
	ctx context.Context,
	filters *service.ServiceFilters,
)(
	successfulUserServiceIds map[string]bool,
	erroredUserServiceIds map[string]error,
	resultErr error,
) {
	successes, failures, err := backend.underlying.DestroyUserServices(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred destroying user services using filters: %+v", filters)
	}
	return successes, failures, nil
}
