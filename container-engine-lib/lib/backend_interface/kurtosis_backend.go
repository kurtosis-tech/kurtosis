package backend_interface

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/api_container"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/engine"
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

	// Stops the engines with the given IDs
	StopEngines(
		ctx context.Context,
		filters *engine.EngineFilters,
	) (
		successfulEngineIds map[string]bool, // "set" of engine IDs that were successfully stopped
		erroredEngineIds map[string]error, // "set" of engine IDs that errored when stopping, with the error
		resultErr error, // Represents an error with the function itself, rather than the engines
	)

	// Destroys the engines with the given IDs, regardless of if they're running or not
	DestroyEngines(
		ctx context.Context,
		filters *engine.EngineFilters,
	) (
		successfulEngineIds map[string]bool, // "set" of engine IDs that were successfully destroyed
		erroredEngineIds map[string]error, // "set" of engine IDs that errored when destroying, with the error
		resultErr error, // Represents an error with the function itself, rather than the engines
	)

	CreateAPIContainer(
	) (
		*api_container.APIContainer,
		error,
	)




	// TODO CreateEnclave

	// TODO GetEnclaves

	// TODO StopEnclaves

	// TODO DestroyEnclaves

	// TODO MAYYYYYYYBE DumpEnclaves?

	// TODO CreateRepl

	// TODO AttachToRepl

	// TODO GetRepls

	// TODO StopRepl

	// TODO DestroyRepl

	// TODO RunReplExecCommand

	// TODO CreateModule

	// TODO DestroyModule

	// TODO CreateUserService

	// TODO GetUserServices

	// TODO StopUserServices

	// TODO DestroyUserServices

	// TODO GetUserServiceLogs

	// TODO GetShellOnUserService
}
