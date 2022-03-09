package backend_interface

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/api_container"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/module"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/partition"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"io"
	"net"
)

// KurtosisBackend abstracts a Kurtosis backend, which will be a container engine (Docker or Kubernetes).
// The heuristic for "do I need a method in KurtosisBackend?" here is "will I make one or more calls to
// the underlying container engine?"
type KurtosisBackend interface {
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
		enclaveId string,
	) (
		*enclave.Enclave,
		error,
	)

	// Gets enclaves matching the given filters
	GetEnclaves(
		ctx context.Context,
		filters *enclave.EnclaveFilters,
	) (
		map[string]*enclave.Enclave,
		error,
	)

	// TODO MAYYYYYYYBE DumpEnclaves?

	// Stops enclaves matching the given filters
	StopEnclaves(
		ctx context.Context,
		filters *enclave.EnclaveFilters,
	) (
		successfulEnclaveIds map[string]bool,
		erroredEnclaveIds map[string]error,
		resultErr error,
	)

	// Destroys enclaves matching the given filters
	DestroyEnclaves(
		ctx context.Context,
		filters *enclave.EnclaveFilters,
	) (
		successfulEnclaveIds map[string]bool,
		erroredEnclaveIds map[string]error,
		resultErr error,
	)

	// Register some file artifacts which can be used for the enclave's services and modules
	RegisterFileArtifacts(
		ctx context.Context,
		fileArtifactsUrls map[service.FilesArtifactID]string,
	)(
		resultErr error,
	)

	// Repartition the Enclave network defining which services will be on each part
	CreateRepartition(
		ctx context.Context,
		partitions []*partition.Partition,
		newPartitionConnections map[partition.PartitionConnectionID]partition.PartitionConnection,
		newDefaultConnection partition.PartitionConnection,
	)(
		resultErr error,
	)

	CreateAPIContainer(
		ctx context.Context,
		image string,
		grpcPortId string,
		grpcPortSpec *port_spec.PortSpec,
		grpcProxyPortId string,
		grpcProxyPortSpec *port_spec.PortSpec,
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
		map[string]*api_container.APIContainer,
		error,
	)

	// Stops API containers matching the given filters
	StopAPIContainers(
		ctx context.Context,
		filters *enclave.EnclaveFilters,
	) (
	// Successful & errored API containers are keyed by their enclave ID
		successApiContainerIds map[string]bool,
		erroredApiContainerIds map[string]error,
		resultErr error,
	)

	// Stops API containers matching the given filters
	DestroyAPIContainers(
		ctx context.Context,
		filters *enclave.EnclaveFilters,
	) (
	// Successful & errored API containers are keyed by their enclave ID
		successApiContainerIds map[string]bool,
		erroredApiContainerIds map[string]error,
		resultErr error,
	)

	// Create a module from a container image with serialized params
	CreateModule(
		ctx context.Context,
		id string,
		containerImageName string,
		serializedParams string,
	)(
		privateIp net.IP,
		privatePort *port_spec.PortSpec,
		publicIp net.IP,
		publicPort *port_spec.PortSpec,
		resultErr error,
	)

	// Gets modules using the given filters, returning a map of matched modules identified by their module ID
	GetModules(ctx context.Context, filters *module.ModuleFilters) (map[string]*module.Module, error)

	// Destroys the modules with the given filters, regardless of if they're running or not
	DestroyModules(
		ctx context.Context,
		filters *module.ModuleFilters,
	) (
		successfulModuleIds map[string]bool, // "set" of module IDs that were successfully destroyed
		erroredModuleIds map[string]error, // "set" of module IDs that errored when destroying, with the error
		resultErr error, // Represents an error with the function itself, rather than the modules
	)

	// Creates a user service inside an enclave with the given configuration
	CreateUserService(
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
		maybePublicIpAddr net.IP, // The ip exposed in the host machine. Will be nil if the service doesn't declare any private ports
		publicPorts map[string]*port_spec.PortSpec, //Mapping of port-used-by-service -> port-on-the-host-machine where the user can make requests to the port to access the port. If a used port doesn't have a host port bound, then the value will be nil.
		resultErr error,
	)

	// Gets user services using the given filters, returning a map of matched user services identified by their ID
	GetUserServices(ctx context.Context, filters *service.ServiceFilters) (map[string]*service.Service, error)

	// Get user service logs using the given filters, returning a map of matched user services identified by their ID and a readCloser object for each one
	GetUserServiceLogs(ctx context.Context, filters *service.ServiceFilters) (map[string]io.ReadCloser, error)

	// Executes a shell command inside an user service instance indenfified by its ID
	RunUserServiceExecCommand (
		ctx context.Context,
		serviceId string,
		commandArgs []string,
	)(
		exitCode int32,
		output string,
		resultErr error,
	)

	// Wait for succesful http endpoint response which can be used to check if the service is available
	WaitForUserServiceHttpEndpointAvailability(
		ctx context.Context,
		serviceId string,
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

	// Stop user services using the given filters,
	StopUserServices(
		ctx context.Context,
		filters *service.ServiceFilters,
	)(
		successfulUserServiceIds map[string]bool, // "set" of user service IDs that were successfully stopped
		erroredUserServiceIds map[string]error, // "set" of user service IDs that errored when stopping, with the error
		resultErr error, // Represents an error with the function itself, rather than the user services
	)

	// Destroy user services using the given filters,
	DestroyUserServices(
		ctx context.Context,
		filters *service.ServiceFilters,
	)(
		successfulUserServiceIds map[string]bool, // "set" of user service IDs that were successfully destroyed
		erroredUserServiceIds map[string]error, // "set" of user service IDs that errored when destroying, with the error
		resultErr error, // Represents an error with the function itself, rather than the user services
	)

	// Get an interactive shell to execute commands in an user service
	GetShellOnUserService(
		ctx context.Context,
		userServiceId string,
	)(
		resultErr error,
	)

	// TODO CreateRepl

	// TODO AttachToRepl

	// TODO GetRepls

	// TODO StopRepl

	// TODO DestroyRepl

	// TODO RunReplExecCommand

}
