package backend_interface

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/api_container"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/module"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"io"
	"net"
)

// KurtosisBackend abstracts a Kurtosis backend, which will be a container engine (Docker or Kubernetes).
// The heuristic for "do I need a method in KurtosisBackend?" here is "will I make one or more calls to
// the underlying container engine?"
type KurtosisBackend interface {
	// Attempts to pull the image from remote to locally, overwriting the local if it exists
	PullImage(image string) error

	// Creates an engine with the given parameters
	CreateEngine(
		ctx context.Context,
		imageOrgAndRepo string,
		imageVersionTag string,
		grpcPortNum uint16,
		grpcProxyPortNum uint16,
		engineDataDirpathOnHostMachine string,
		envVars map[string]string,
	) (
		*engine.Engine,
		error,
	)

	// Gets engines using the given filters, returning a map of matched engines identified by their engine ID
	GetEngines(ctx context.Context, filters *engine.EngineFilters) (map[string]*engine.Engine, error)

	// Stops the engines matching the given filters
	StopEngines(
		ctx context.Context,
		filters *engine.EngineFilters,
	) (
		successfulEngineIds map[string]bool, // "set" of engine IDs that were successfully stopped
		erroredEngineIds map[string]error, // "set" of engine IDs that errored when stopping, with the error
		resultErr error, // Represents an error with the function itself, rather than the engines
	)

	// Destroys the engines matching the given filters, regardless of if they're running or not
	DestroyEngines(
		ctx context.Context,
		filters *engine.EngineFilters,
	) (
		successfulEngineIds map[string]bool, // "set" of engine IDs that were successfully destroyed
		erroredEngineIds map[string]error, // "set" of engine IDs that errored when destroying, with the error
		resultErr error, // Represents an error with the function itself, rather than the engines
	)

	// Creates an enclave with the given enclave ID
	CreateEnclave(
		ctx context.Context,
		enclaveId enclave.EnclaveID,
		isPartitioningEnabled bool,
	) (
		*enclave.Enclave,
		error,
	)

	// Gets enclaves matching the given filters
	GetEnclaves(
		ctx context.Context,
		filters *enclave.EnclaveFilters,
	) (
		map[enclave.EnclaveID]*enclave.Enclave,
		error,
	)

	// TODO MAYYYYYYYBE DumpEnclaves?

	// Stops enclaves matching the given filters
	StopEnclaves(
		ctx context.Context,
		filters *enclave.EnclaveFilters,
	) (
		successfulEnclaveIds map[enclave.EnclaveID]bool,
		erroredEnclaveIds map[enclave.EnclaveID]error,
		resultErr error,
	)

	// Dumps the contents of the given enclave to the given directory
	DumpEnclave(
		ctx context.Context,
		enclaveId enclave.EnclaveID,
		outputDirpath string,
	) error

	// Destroys enclaves matching the given filters
	DestroyEnclaves(
		ctx context.Context,
		filters *enclave.EnclaveFilters,
	) (
		successfulEnclaveIds map[enclave.EnclaveID]bool,
		erroredEnclaveIds map[enclave.EnclaveID]error,
		resultErr error,
	)


	CreateAPIContainer(
		ctx context.Context,
		image string,
		enclaveId enclave.EnclaveID,
		ipAddr net.IP, // TODO REMOVE THIS ONCE WE FIX THE STATIC IP PROBLEM!!
		grpcPortNum uint16,
		grpcProxyPortNum uint16,
		enclaveDataDirpathOnHostMachine string,	// TODO DELETE WHEN WE HAVE AN ENCLAVE DATA VOLUME!
		envVars map[string]string,
	) (
		*api_container.APIContainer,
		error,
	)

	GetAPIContainers(
		ctx context.Context,
		filters *api_container.APIContainerFilters,
	) (
		// Matching API containers, keyed by their enclave ID
		map[enclave.EnclaveID]*api_container.APIContainer,
		error,
	)

	// Stops API containers matching the given filters
	StopAPIContainers(
		ctx context.Context,
		filters *api_container.APIContainerFilters,
	) (
		// Successful & errored API containers are keyed by their enclave ID
		successfulApiContainerIds map[enclave.EnclaveID]bool,
		erroredApiContainerIds map[enclave.EnclaveID]error,
		resultErr error,
	)

	// Stops API containers matching the given filters
	DestroyAPIContainers(
		ctx context.Context,
		filters *api_container.APIContainerFilters,
	) (
		// Successful & errored API containers are keyed by their enclave ID
		successfulApiContainerIds map[enclave.EnclaveID]bool,
		erroredApiContainerIds map[enclave.EnclaveID]error,
		resultErr error,
	)

	// Create a module from a container image with serialized params
	CreateModule(
		ctx context.Context,
		image string,
		enclaveId enclave.EnclaveID,
		id module.ModuleID,
		guid module.ModuleGUID,
		ipAddr net.IP, // TODO REMOVE THIS ONCE WE FIX THE STATIC IP PROBLEM!!
		grpcPortNum uint16,
		enclaveDataDirpathOnHostMachine string,
		envVars map[string]string,
	)(
		newModule *module.Module,
		resultErr error,
	)

	// Gets modules using the given filters, returning a map of matched modules identified by their module ID
	GetModules(
		ctx context.Context,
		filters *module.ModuleFilters,
	) (
		map[module.ModuleGUID]*module.Module,
		error,
	)

	// Destroys the modules with the given filters, regardless of if they're running or not
	DestroyModules(
		ctx context.Context,
		filters *module.ModuleFilters,
	) (
		successfulModuleIds map[module.ModuleGUID]bool, // "set" of module IDs that were successfully destroyed
		erroredModuleIds map[module.ModuleGUID]error, // "set" of module IDs that errored when destroying, with the error
		resultErr error, // Represents an error with the function itself, rather than the modules
	)

	// Creates a user service inside an enclave with the given configuration
	CreateUserService(
		ctx context.Context,
		id service.ServiceID,
		guid service.ServiceGUID,
		containerImageName string,
		privatePorts []*port_spec.PortSpec,
		entrypointArgs []string,
		cmdArgs []string,
		envVars map[string]string,
		enclaveDataDirMntDirpath string,
		filesArtifactMountDirpaths map[string]string,
    )(
		newUserService *service.Service,
		resultErr error,
	)

	// Gets user services using the given filters, returning a map of matched user services identified by their ID
	GetUserServices(ctx context.Context, filters *service.ServiceFilters) (map[service.ServiceGUID]*service.Service, error)

	// Get user service logs using the given filters, returning a map of matched user services identified by their ID and a readCloser object for each one
	GetUserServiceLogs(ctx context.Context, filters *service.ServiceFilters) (map[service.ServiceGUID]io.ReadCloser, error)

	// Executes a shell command inside an user service instance indenfified by its ID
	RunUserServiceExecCommand (
		ctx context.Context,
		serviceGUID service.ServiceGUID,
		commandArgs []string,
	)(
		exitCode int32,
		output string,
		resultErr error,
	)

	// Wait for succesful http endpoint response which can be used to check if the service is available
	WaitForUserServiceHttpEndpointAvailability(
		ctx context.Context,
		serviceGUID service.ServiceGUID,
		httpMethod string, //The httpMethod used to execute the request. Valid values: GET and POST
		port uint32, //The port of the service to check. For instance 8080
		path string, //The path of the service to check. It mustn't start with the first slash. For instance `service/health`
		requestBody string, //The content of the request body. Only valid when the httpMethod is POST
		bodyText string, //If the endpoint returns this value, the service will be marked as available (e.g. Hello World).
		initialDelayMilliseconds uint32, //The number of milliseconds to wait until executing the first HTTP call
		retries uint32, //Max number of HTTP call attempts that this will execute until giving up and returning an error
		retriesDelayMilliseconds uint32, //Number of milliseconds to wait between retries
	)(
		resultErr error,
	)

	// Get an interactive shell to execute commands in an user service
	GetShellOnUserService(
		ctx context.Context,
		serviceGUID service.ServiceGUID,
	)(
		resultErr error,
	)

	// Stop user services using the given filters,
	StopUserServices(
		ctx context.Context,
		filters *service.ServiceFilters,
	)(
		successfulUserServiceIds map[service.ServiceGUID]bool, // "set" of user service IDs that were successfully stopped
		erroredUserServiceIds map[service.ServiceGUID]error, // "set" of user service IDs that errored when stopping, with the error
		resultErr error, // Represents an error with the function itself, rather than the user services
	)

	// Destroy user services using the given filters,
	DestroyUserServices(
		ctx context.Context,
		filters *service.ServiceFilters,
	)(
		successfulUserServiceIds map[service.ServiceGUID]bool, // "set" of user service IDs that were successfully destroyed
		erroredUserServiceIds map[service.ServiceGUID]error, // "set" of user service IDs that errored when destroying, with the error
		resultErr error, // Represents an error with the function itself, rather than the user services
	)

	// TODO CreateRepl

	// TODO AttachToRepl

	// TODO GetRepls

	// TODO StopRepl

	// TODO DestroyRepl

	// TODO RunReplExecCommand

}
