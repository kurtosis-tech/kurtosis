package lib

import (
	"context"
	engine2 "github.com/kurtosis-tech/container-engine-lib/lib/objects/engine"
)

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
		*engine2.Engine,
		error,
	)

	// Gets engines using the given filters, returning a map of matched engines identified by their engine ID
	GetEngines(ctx context.Context, filters *engine2.GetEnginesFilters) (map[string]*engine2.Engine, error)

	// Stops the engines with the given IDs
	StopEngines(
		ctx context.Context,
		filters *engine2.GetEnginesFilters,
	) (
		successfulEngineIds map[string]bool, // "set" of engine IDs that were successfully stopped
		erroredEngineIds map[string]error, // "set" of engine IDs that errored when stopping, with the error
		resultErr error, // Represents an error with the function itself, rather than the engines
	)

	// Destroys the engines with the given IDs, regardless of if they're running or not
	DestroyEngines(
		ctx context.Context,
		filters *engine2.GetEnginesFilters,
	) (
		successfulEngineIds map[string]bool, // "set" of engine IDs that were successfully destroyed
		erroredEngineIds map[string]error, // "set" of engine IDs that errored when destroying, with the error
		resultErr error, // Represents an error with the function itself, rather than the engines
	)
}
