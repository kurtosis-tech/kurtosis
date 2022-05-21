package backend_interface

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/api_container"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifact_expansion"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/module"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/networking_sidecar"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"io"
	"net"
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
		envVars map[string]string,
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
		grpcPortNum uint16,
		grpcProxyPortNum uint16,
		enclaveDataVolumeDirpath string,
		// The environment variable that the user is requesting to populate with the container's own IP address
		// Must not conflict with the custom environment variables
		ownIpAddressEnvVar string,
		customEnvVars map[string]string,
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
		grpcPortNum uint16,
		envVars map[string]string,
	) (
		newModule *module.Module,
		resultErr error,
	)

	// Gets modules using the given filters, returning a map of matched modules identified by their module ID
	GetModules(
		ctx context.Context,
		enclaveId enclave.EnclaveID,
		filters *module.ModuleFilters,
	) (
		map[module.ModuleGUID]*module.Module,
		error,
	)

	// Get module logs using the given filters, returning a map of matched modules identified by their GUID and a readCloser object for each one
	// User is responsible for closing the 'ReadCloser' object returned in the successfulModuleLogs map
	GetModuleLogs(
		ctx context.Context,
		enclaveId enclave.EnclaveID,
		filters *module.ModuleFilters,
		shouldFollowLogs bool,
	) (
		successfulModuleLogs map[module.ModuleGUID]io.ReadCloser,
		erroredModuleGuids map[module.ModuleGUID]error,
		resultError error,
	)

	// Stops the modules matching the given filters
	StopModules(
		ctx context.Context,
		enclaveId enclave.EnclaveID,
		filters *module.ModuleFilters,
	) (
		successfulModuleIds map[module.ModuleGUID]bool, // "set" of module IDs that were successfully stopped
		erroredModuleIds map[module.ModuleGUID]error, // "set" of module IDs that errored when being stopped, with the error
		resultErr error, // Represents an error with the function itself, rather than the modules
	)

	// Destroys the modules with the given filters, regardless of if they're running or not
	DestroyModules(
		ctx context.Context,
		enclaveId enclave.EnclaveID,
		filters *module.ModuleFilters,
	) (
		successfulModuleIds map[module.ModuleGUID]bool, // "set" of module IDs that were successfully destroyed
		erroredModuleIds map[module.ModuleGUID]error, // "set" of module IDs that errored when destroying, with the error
		resultErr error, // Represents an error with the function itself, rather than the modules
	)





	/*
	                           KURTOSIS SERVICE STATE DIAGRAM

                                .-----------------DestroyServices--------------------.
                               /                                                      \
	  RegisterService--> REGISTERED ---StopServices---> STOPPED ---DestroyServices---> DESTROYED
	                           \                          /                           /
	                      StartService              StopServices                     /
	                             \                      /                           /
	                              '---------------> RUNNING ---DestroyServices-----'

	Considerations:
	- We have REGISTERED as a state separate from RUNNING because some user containers need to know their own
		IP address when they start, which means we need to know the IP address of the container BEFORE it starts.
	- As of 2022-05-15, Kurtosis services can never be restarted once stopped.
	*/


	// Registers a user service, allocating it an IP and ServiceGUID
	RegisterUserService(
		ctx context.Context,
		enclaveId enclave.EnclaveID,
		serviceId service.ServiceID,
	) (*service.ServiceRegistration, error, )

	// StartUserService consumes a service registration to create a user container with the given parameters
	StartUserService(
		ctx context.Context,
		enclaveId enclave.EnclaveID,
		serviceGuid service.ServiceGUID,
		containerImageName string,
		privatePorts map[string]*port_spec.PortSpec,
		entrypointArgs []string,
		cmdArgs []string,
		envVars map[string]string,
		filesArtifactVolumeMountDirpaths map[files_artifact_expansion.FilesArtifactExpansionGUID]string,
	) (
		*service.Service,
		error,
	)

	// Gets user services using the given filters, returning a map of matched user services identified by their GUID
	GetUserServices(
		ctx context.Context,
		enclaveId enclave.EnclaveID,
		filters *service.ServiceFilters,
	) (
		map[service.ServiceGUID]*service.Service,
		error,
	)

	// Get user service logs using the given filters, returning a map of matched user services identified by their GUID and a readCloser object for each one
	// User is responsible for closing the 'ReadCloser' object returned in the successfulUserServiceLogs map
	GetUserServiceLogs(
		ctx context.Context,
		enclaveId enclave.EnclaveID,
		filters *service.ServiceFilters,
		shouldFollowLogs bool,
	) (
		successfulUserServiceLogs map[service.ServiceGUID]io.ReadCloser,
		erroredUserServiceGuids map[service.ServiceGUID]error,
		resultError error,
	)

	// Pauses execution of all processes on a service, but does not shut down the service (memory state is preserved)
	PauseService(
		ctx context.Context,
		enclaveId enclave.EnclaveID,
		serviceId service.ServiceGUID,
	) (
		resultErr error,
	)

	// Unpauses a service, resuming execution of all processes on the service that were previously paused.
	UnpauseService(
		ctx context.Context,
		enclaveId enclave.EnclaveID,
		serviceId service.ServiceGUID,
	) (
		resultErr error,
	)

	// Executes a shell command inside an user service instance indenfified by its ID
	RunUserServiceExecCommands(
		ctx context.Context,
		enclaveId enclave.EnclaveID,
		userServiceCommands map[service.ServiceGUID][]string,
	) (
		succesfulUserServiceExecResults map[service.ServiceGUID]*exec_result.ExecResult,
		erroredUserServiceGuids map[service.ServiceGUID]error,
		resultErr error,
	)

	// Get a connection with user service to execute commands in
	GetConnectionWithUserService(
		ctx context.Context,
		enclaveId enclave.EnclaveID,
		serviceGuid service.ServiceGUID,
	) (
		resultConn net.Conn,
		resultErr error,
	)

	// Copy files, packaged as a TAR, from the given user service and writes the bytes to the given output writer
	CopyFilesFromUserService(
		ctx context.Context,
		enclaveId enclave.EnclaveID,
		serviceGuid service.ServiceGUID,
		srcPathOnService string,
		output io.Writer,
	)(
		error,
	)

	// StopUserServices stops the user containers for the services matching the given filters
	// A stopped service cannot be activated again as of 2022-05-14
	StopUserServices(
		ctx context.Context,
		enclaveId enclave.EnclaveID,
		filters *service.ServiceFilters,
	) (
		successfulUserServiceGuids map[service.ServiceGUID]bool, // "set" of user service GUIDs that were successfully stopped
		erroredUserServiceGuids map[service.ServiceGUID]error, // "set" of user service GUIDs that errored when stopping, with the error
		resultErr error, // Represents an error with the function itself, rather than the user services
	)

	// DestroyUserServices destroys user services matching the given filters, removing all resources associated with it
	DestroyUserServices(
		ctx context.Context,
		enclaveId enclave.EnclaveID,
		filters *service.ServiceFilters,
	) (
		successfulUserServiceGuids map[service.ServiceGUID]bool, // "set" of user service GUIDs that were successfully destroyed
		erroredUserServiceGuids map[service.ServiceGUID]error, // "set" of user service GUIDs that errored when destroying, with the error
		resultErr error, // Represents an error with the function itself, rather than the user services
	)

	// TODO Move this logic inside the user service, so that we have tighter controls on what can happen and what can't
	//Create a user service's  networking sidecar inside enclave
	CreateNetworkingSidecar(
		ctx context.Context,
		enclaveId enclave.EnclaveID,
		serviceGuid service.ServiceGUID,
	) (
		*networking_sidecar.NetworkingSidecar,
		error,
	)

	// Gets networking sidecars using the given filters, returning a map of matched networking sidecars identified by their GUID
	GetNetworkingSidecars(
		ctx context.Context,
		filters *networking_sidecar.NetworkingSidecarFilters,
	) (
		map[service.ServiceGUID]*networking_sidecar.NetworkingSidecar,
		error,
	)

	//Executes many shell commands inside multiple networking sidecar instances indenfified by User Service GUIDs
	RunNetworkingSidecarExecCommands(
		ctx context.Context,
		enclaveId enclave.EnclaveID,
		networkingSidecarsCommands map[service.ServiceGUID][]string,
	) (
		successfulNetworkingSidecarExecResults map[service.ServiceGUID]*exec_result.ExecResult,
		erroredUserServiceGuids map[service.ServiceGUID]error,
		resultErr error,
	)

	// Stop networking sidecars using the given filters,
	StopNetworkingSidecars(
		ctx context.Context,
		filters *networking_sidecar.NetworkingSidecarFilters,
	) (
		successfulUserServiceGuids map[service.ServiceGUID]bool,
		erroredUserServiceGuids map[service.ServiceGUID]error,
		resultErr error,
	)

	// Destroy networking sidecars using the given filters,
	DestroyNetworkingSidecars(
		ctx context.Context,
		filters *networking_sidecar.NetworkingSidecarFilters,
	) (
		successfulUserServiceGuids map[service.ServiceGUID]bool,
		erroredUserServiceGuids map[service.ServiceGUID]error,
		resultErr error,
	)

	//Create and execute a files artifact expansion for given file's FilesArtifactID
	CreateFilesArtifactExpansion(
		ctx context.Context,
		enclaveId enclave.EnclaveID,
		serviceGuid service.ServiceGUID,
		filesArtifactId service.FilesArtifactID,
		filesArtifactFilepathRelativeToEnclaveDatadirRoot string, ) (
		*files_artifact_expansion.FilesArtifactExpansion,
		error,
	)

	//Destroy files artifact expansion using the given filters
	DestroyFilesArtifactExpansions(
		ctx context.Context,
		enclaveId enclave.EnclaveID,
		filters *files_artifact_expansion.FilesArtifactExpansionFilters,
	) (
		resultSuccessfulFileArtifactExpansionGUIDs map[files_artifact_expansion.FilesArtifactExpansionGUID]bool,
		resultErroredFileArtifactExpansionGUIDs map[files_artifact_expansion.FilesArtifactExpansionGUID]error,
		resultErr error,
	)
}
