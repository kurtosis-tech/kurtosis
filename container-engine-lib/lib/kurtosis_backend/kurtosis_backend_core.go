package kurtosis_backend

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/kurtosis_backend/objects/engine"
	"github.com/kurtosis-tech/object-attributes-schema-lib/forever_constants"
	"github.com/sirupsen/logrus"
	"net"
)

// these constants are here because they are being used for both kubernetes and docker backends core
const (
	// The location where the engine data directory (on the Docker host machine) will be bind-mounted
	//  on the engine server
	EngineDataDirpathOnEngineServerContainer = "/engine-data"

	publicPortNumParsingBase     = 10
	publicPortNumParsingUintBits = 16

	shouldCleanRunningEngineContainers = false
)

var engineLabels = map[string]string{
	forever_constants.AppIDLabel:         forever_constants.AppIDValue,
	forever_constants.ContainerTypeLabel: forever_constants.ContainerType_EngineServer,
}

type KurtosisBackendCore interface {
	// Creates an engine with the given parameters
	CreateEngine(
		ctx context.Context,
		imageOrgAndRepo string,
		imageVersionTag string,
		logLevel logrus.Level,
		listenPortNum uint16,
		engineDataDirpathOnHostMachine string,
		envVars map[string]string,
	) (
		resultPublicIpAddr net.IP,
		resultPublicPortNum uint16,
		resultErr error,
	)

	// Gets engines using the given filters, returning a map of matched engines identified by their engine ID
	GetEngines(
		ctx context.Context,
		filters *engine.GetEnginesFilters,
	) (map[string]*engine.Engine, error)

	// Stops the engines with the given IDs
	StopEngines(
		ctx context.Context,
		ids map[string]bool,
	) (
		map[string]error, // Contains one engine ID key per engine we tried to stop, with the (potentially nil) error from attemping to do so
		error, // Represents an error before attempting to stop the engines
	)

	// Destroys the engines with the given IDs, regardless of if they're running or not
	DestroyEngines(ctx context.Context, ids map[string]bool) error
}
