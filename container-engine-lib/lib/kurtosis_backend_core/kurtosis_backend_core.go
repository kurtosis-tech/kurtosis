package kurtosis_backend_core

import (
	"context"
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

	EngineStatus_Stopped                                 = "STOPPED"
)

var engineLabels = map[string]string{
	forever_constants.AppIDLabel: forever_constants.AppIDValue,
	forever_constants.ContainerTypeLabel: forever_constants.ContainerType_EngineServer,
}

type KurtosisBackendCore interface {
	CreateEngine(
		ctx context.Context,
		imageVersionTag string,
		logLevel logrus.Level,
		listenPortNum uint16,
		engineDataDirpathOnHostMachine string,
		containerImage string,
		envVars map[string]string,
	) (
		resultPublicIpAddr net.IP,
		resultPublicPortNum uint16,
		resultErr error,
	)
	StopEngine(ctx context.Context) error
	CleanStoppedEngines(ctx context.Context) ([]string, []error, error)
	GetEngineStatus(
		ctx context.Context,
	) (engineStatus string, ipAddr net.IP, portNum uint16, err error)
}
