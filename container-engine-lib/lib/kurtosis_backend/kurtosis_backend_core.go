package kurtosis_backend

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/kurtosis_backend/objects/engine"
)

type KurtosisBackendCore interface {
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
	GetEngines(ctx context.Context, filters *engine.GetEnginesFilters) (map[string]*engine.Engine, error)

	// Stops the engines with the given IDs
	StopEngines(
		ctx context.Context,
		filters *engine.GetEnginesFilters,
	) (
		map[string]error, // Contains one engine ID key per engine we tried to stop, with the (potentially nil) error from attemping to do so
		error, // Represents an error before attempting to stop the engines
	)

	// Destroys the engines with the given IDs, regardless of if they're running or not
	DestroyEngines(
		ctx context.Context,
		filters *engine.GetEnginesFilters,
	) (
		map[string]error, // Contains one engine ID key per engine we tried to destroy, with the (potentially nil) error from attemping to do so
		error, // Represents an error before attempting to destroy the engines
	)
}
