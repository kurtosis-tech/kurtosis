package backend_interface

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
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

	/*
	CreateAPIContainer(
	) (
		*api_container.APIContainer,
		error,
	)

	 */

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

	// TODO GetUserServiceLogs

	// TODO StopUserServices

	// TODO GetShellOnUserService

	// TODO DestroyUserServices
}
