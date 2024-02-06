package backend_interface

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_build_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_registry_spec"
	"io"
	"time"

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
)

// TODO This mega-backend should really have its individual functionalities split up into
//  the appropriate places it's used - e.g. APIContainerBackend, EngineBackend, etc.
//  The reason we don't do this right now is because the CLI uses some KurtosisBackend methods
//  (e.g. GetUserServices to power 'enclave inspect', GetUserServiceLogs) because the Kurtosis
//  APIs don't yet support them. Once the Kurtosis APIs support everything, we'll have the CLI
//  use purely the Kurtosis SDK (as it should)

// KurtosisBackend abstracts a Kurtosis backend, which will be a container engine (Docker or Kubernetes).
// The heuristic for "do I need a method in KurtosisBackend?" here is "will I make one or more calls to
// the underlying container engine?"
// mockery -r --name=KurtosisBackend --filename=mock_kurtosis_backend.go --structname=MockKurtosisBackend --with-expecter --inpackage
type KurtosisBackend interface {
	// FetchImage always attempts to retrieve the latest image.
	// If retrieving the latest [dockerImage] fails, the local image will be used.
	// Returns True is it was retrieved from cloud or False if it's a local image
	// Returns a string that represents the architecture of the image
	FetchImage(ctx context.Context, image string, registrySpec *image_registry_spec.ImageRegistrySpec, downloadMode image_download_mode.ImageDownloadMode) (bool, string, error)

	PruneUnusedImages(ctx context.Context) ([]string, error)

	// Creates an engine with the given parameters
	CreateEngine(
		ctx context.Context,
		imageOrgAndRepo string,
		imageVersionTag string,
		grpcPortNum uint16,
		envVars map[string]string,
		shouldStartInDebugMode bool,
	) (
		*engine.Engine,
		error,
	)

	// Gets engines using the given filters, returning a map of matched engines identified by their engine GUID
	GetEngines(ctx context.Context, filters *engine.EngineFilters) (map[engine.EngineGUID]*engine.Engine, error)

	// Stops the engines matching the given filters
	StopEngines(
		ctx context.Context,
		filters *engine.EngineFilters,
	) (
		successfulEngineGuids map[engine.EngineGUID]bool, // "set" of engine GUIDs that were successfully stopped
		erroredEngineGuids map[engine.EngineGUID]error, // "set" of engine GUIDs that errored when stopping, with the error
		resultErr error, // Represents an error with the function itself, rather than the engines
	)

	// Destroys the engines matching the given filters, regardless of if they're running or not
	DestroyEngines(
		ctx context.Context,
		filters *engine.EngineFilters,
	) (
		successfulEngineGuids map[engine.EngineGUID]bool, // "set" of engine GUIDs that were successfully destroyed
		erroredEngineGuids map[engine.EngineGUID]error, // "set" of engine GUIDs that errored when destroying, with the error
		resultErr error, // Represents an error with the function itself, rather than the engines
	)

	// Gets logs of all engines
	GetEngineLogs(ctx context.Context, outputDirpath string) error

	// Dumps all of Kurtosis (engines + all enclaves)
	DumpKurtosis(ctx context.Context, outputDirpath string) error

	// Creates an enclave with the given enclave UUID
	CreateEnclave(ctx context.Context, enclaveUuid enclave.EnclaveUUID, enclaveName string) (*enclave.Enclave, error)

	// Update an enclave by UUID, it's only possible to udpate the name and creation time so far
	// The newCreationTime param is optional, it won't be updated if the value is nil
	UpdateEnclave(
		ctx context.Context,
		enclaveUuid enclave.EnclaveUUID,
		newName string,
		creationTime *time.Time,
	) error

	// Gets enclaves matching the given filters
	GetEnclaves(
		ctx context.Context,
		filters *enclave.EnclaveFilters,
	) (
		map[enclave.EnclaveUUID]*enclave.Enclave,
		error,
	)

	// Stops enclaves matching the given filters
	StopEnclaves(
		ctx context.Context,
		filters *enclave.EnclaveFilters,
	) (
		successfulEnclaveIds map[enclave.EnclaveUUID]bool,
		erroredEnclaveIds map[enclave.EnclaveUUID]error,
		resultErr error,
	)

	// Dumps the contents of the given enclave to the given directory
	// TODO add this to K8S
	DumpEnclave(
		ctx context.Context,
		enclaveUuid enclave.EnclaveUUID,
		outputDirpath string,
	) error

	// Destroys enclaves matching the given filters
	DestroyEnclaves(
		ctx context.Context,
		filters *enclave.EnclaveFilters,
	) (
		successfulEnclaveIds map[enclave.EnclaveUUID]bool,
		erroredEnclaveIds map[enclave.EnclaveUUID]error,
		resultErr error,
	)

	CreateAPIContainer(
		ctx context.Context,
		image string,
		enclaveUuid enclave.EnclaveUUID,
		grpcPortNum uint16,
		enclaveDataVolumeDirpath string,
		// The environment variable that the user is requesting to populate with the container's own IP address
		// Must not conflict with the custom environment variables
		ownIpAddressEnvVar string,
		customEnvVars map[string]string,
		shouldStartInDebugMode bool,
	) (
		*api_container.APIContainer,
		error,
	)

	GetAPIContainers(
		ctx context.Context,
		filters *api_container.APIContainerFilters,
	) (
		// Matching API containers, keyed by their enclave ID
		map[enclave.EnclaveUUID]*api_container.APIContainer,
		error,
	)

	// Stops API containers matching the given filters
	StopAPIContainers(
		ctx context.Context,
		filters *api_container.APIContainerFilters,
	) (
		// Successful & errored API containers are keyed by their enclave ID
		successfulApiContainerIds map[enclave.EnclaveUUID]bool,
		erroredApiContainerIds map[enclave.EnclaveUUID]error,
		resultErr error,
	)

	// Stops API containers matching the given filters
	DestroyAPIContainers(
		ctx context.Context,
		filters *api_container.APIContainerFilters,
	) (
		// Successful & errored API containers are keyed by their enclave ID
		successfulApiContainerIds map[enclave.EnclaveUUID]bool,
		erroredApiContainerIds map[enclave.EnclaveUUID]error,
		resultErr error,
	)

	/*
		KURTOSIS SERVICE STATE DIAGRAM

			                                |
			                        RegisterUserServices
			                                |
			                                V
			                            REGISTERED
			                                |
			                    StartRegisteredUserServices
			                                |
			                                V
			            .--------------- STARTED
			            |                   |
			            |           StopUserService
			            |                   |
			            |                   V
			    DestroyUserServices      STOPPED
			            |                   |
			            |           DestroyUserServices
			            |                   |
			            |                   V
			            '-------------> DESTROYED

			- Note the above state diagram doesn't account for PauseService or UnpauseService
	*/

	// RegisterUserServices registers the services allocating them an IP address and a UUID. The service is not started!
	RegisterUserServices(
		ctx context.Context,
		enclaveUuid enclave.EnclaveUUID,
		services map[service.ServiceName]bool,
	) (
		map[service.ServiceName]*service.ServiceRegistration, // "set" of user service Names that were successfully registered
		map[service.ServiceName]error, // "set" of user service Names that errored when being registered, with the error
		error,
	)

	// UnregisterUserServices unregisters a set of services. If a service isn't registered, it no-ops for this service
	UnregisterUserServices(
		ctx context.Context,
		enclaveUuid enclave.EnclaveUUID,
		services map[service.ServiceUUID]bool,
	) (
		map[service.ServiceUUID]bool, // "set" of user service UUIDs that were successfully unregistered
		map[service.ServiceUUID]error, // "set" of user service UUIDs that errored when being unregistered, with the error
		error,
	)

	// StartRegisteredUserServices consumes service registrations to create auser container for each registration, given each service config
	StartRegisteredUserServices(
		ctx context.Context,
		enclaveUuid enclave.EnclaveUUID,
		services map[service.ServiceUUID]*service.ServiceConfig,
	) (
		map[service.ServiceUUID]*service.Service, // "set" of user UUIDs that were successfully started
		map[service.ServiceUUID]error, // "set" of user service UUIDs that errored when attempting to start, with the error
		error, // represents an error with the function itself, rather than the user services
	)

	// RemoveRegisteredUserServiceProcesses removes the running user service process but keeps the service registration
	// TODO: As we don't persist user service logs anywhere, removing the container/pod will also remove all its
	//  logs. We need a persistent log storage to address this issue.
	RemoveRegisteredUserServiceProcesses(
		ctx context.Context,
		enclaveUuid enclave.EnclaveUUID,
		services map[service.ServiceUUID]bool,
	) (
		map[service.ServiceUUID]bool, // user service UUIDs that were successfully removed
		map[service.ServiceUUID]error, // user service UUIDs that failed to be removed, with the error
		error, // represents an error with the function itself, rather than the user services
	)

	// Gets user services using the given filters, returning a map of matched user services identified by their UUID
	GetUserServices(
		ctx context.Context,
		enclaveUuid enclave.EnclaveUUID,
		filters *service.ServiceFilters,
	) (
		map[service.ServiceUUID]*service.Service,
		error,
	)

	// Get user service logs using the given filters, returning a map of matched user services identified by their GUID and a readCloser object for each one
	// User is responsible for closing the 'ReadCloser' object returned in the successfulUserServiceLogs map
	GetUserServiceLogs(
		ctx context.Context,
		enclaveUuid enclave.EnclaveUUID,
		filters *service.ServiceFilters,
		shouldFollowLogs bool,
	) (
		successfulUserServiceLogs map[service.ServiceUUID]io.ReadCloser,
		erroredUserServiceUuids map[service.ServiceUUID]error,
		resultError error,
	)

	// Executes a shell command inside an user service instance indenfified by its ID
	RunUserServiceExecCommands(
		ctx context.Context,
		enclaveUuid enclave.EnclaveUUID,
		userServiceCommands map[service.ServiceUUID][]string,
	) (
		succesfulUserServiceExecResults map[service.ServiceUUID]*exec_result.ExecResult,
		erroredUserServiceUuids map[service.ServiceUUID]error,
		resultErr error,
	)

	// Executes a shell command inside an user service, where output and final response are streamed back to user
	RunUserServiceExecCommandWithStreamedOutput(
		ctx context.Context,
		enclaveUuid enclave.EnclaveUUID,
		serviceUuid service.ServiceUUID,
		cmd []string,
	) (execOutputChan chan string, finalExecResultChan chan *exec_result.ExecResult, resultErr error)

	// Get a connection with user service to execute commands in
	GetShellOnUserService(ctx context.Context, enclaveUuid enclave.EnclaveUUID, serviceUuid service.ServiceUUID) (resultErr error)

	// Copy files, packaged as a TAR, from the given user service and writes the bytes to the given output writer
	CopyFilesFromUserService(
		ctx context.Context,
		enclaveUuid enclave.EnclaveUUID,
		serviceUuid service.ServiceUUID,
		srcPathOnService string,
		output io.Writer,
	) error

	// StopUserServices stops the user containers for the services matching the given filters
	StopUserServices(
		ctx context.Context,
		enclaveUuid enclave.EnclaveUUID,
		filters *service.ServiceFilters,
	) (
		successfulUserServiceUuids map[service.ServiceUUID]bool, // "set" of user service UUIDs that were successfully stopped
		erroredUserServiceUuids map[service.ServiceUUID]error, // "set" of user service UUIDs that errored when stopping, with the error
		resultErr error, // Represents an error with the function itself, rather than the user services
	)

	// DestroyUserServices destroys user services matching the given filters, removing all resources associated with it
	DestroyUserServices(
		ctx context.Context,
		enclaveUuid enclave.EnclaveUUID,
		filters *service.ServiceFilters,
	) (
		successfulUserServiceUuids map[service.ServiceUUID]bool, // "set" of user service UUIDs that were successfully destroyed
		erroredUserServiceUuids map[service.ServiceUUID]error, // "set" of user service UUIDs that errored when destroying, with the error
		resultErr error, // Represents an error with the function itself, rather than the user services
	)

	CreateLogsAggregator(ctx context.Context) (*logs_aggregator.LogsAggregator, error)

	// Returns nil if logs aggregator was not found
	GetLogsAggregator(ctx context.Context) (*logs_aggregator.LogsAggregator, error)

	DestroyLogsAggregator(ctx context.Context) error

	// Create a new Logs Collector for sending container's logs to the logs aggregator server
	CreateLogsCollectorForEnclave(ctx context.Context, enclaveUuid enclave.EnclaveUUID, logsCollectorHttpPortNumber uint16, logsCollectorTcpPortNumber uint16) (*logs_collector.LogsCollector, error)

	// Gets the logs collector, if nothing is found returns nil
	GetLogsCollectorForEnclave(ctx context.Context, enclaveUuid enclave.EnclaveUUID) (*logs_collector.LogsCollector, error)

	// Destroy the logs collector for enclave with UUID
	DestroyLogsCollectorForEnclave(ctx context.Context, enclaveUuid enclave.EnclaveUUID) error

	CreateReverseProxy(ctx context.Context, engineGuid engine.EngineGUID) (*reverse_proxy.ReverseProxy, error)

	// Returns nil if logs aggregator was not found
	GetReverseProxy(ctx context.Context) (*reverse_proxy.ReverseProxy, error)

	DestroyReverseProxy(ctx context.Context) error

	// GetAvailableCPUAndMemory - gets available memory in megabytes and cpu in millicores, the boolean indicates whether the information is complete
	GetAvailableCPUAndMemory(ctx context.Context) (compute_resources.MemoryInMegaBytes, compute_resources.CpuMilliCores, bool, error)

	// BuildImage builds a container image based on the [imageBuildSpec] with [imageName]
	// Returns image architecture and if error occurred
	BuildImage(ctx context.Context, imageName string, imageBuildSpec *image_build_spec.ImageBuildSpec) (string, error)
}
