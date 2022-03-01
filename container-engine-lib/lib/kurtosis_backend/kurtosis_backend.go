package kurtosis_backend

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/kurtosis_backend/objects/engine"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
)

// TODO CALL THE METRICS LIBRARY EVENT-REGISTRATION FUNCTIONS HERE!!!!
type KurtosisBackend struct {
	kurtosisBackendCore KurtosisBackendCore
	log                 *logrus.Logger
}

func NewKurtosisBackend(log *logrus.Logger, kurtosisBackendCore KurtosisBackendCore) *KurtosisBackend {
	return &KurtosisBackend{
		log:                 log,
		kurtosisBackendCore: kurtosisBackendCore,
	}
}

func (backend *KurtosisBackend) CreateEngine(
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
) {
	publicIpAddr, publicPortNum, err := backend.kurtosisBackendCore.CreateEngine(
		ctx,
		imageOrgAndRepo,
		imageVersionTag,
		logLevel,
		listenPortNum,
		engineDataDirpathOnHostMachine,
		envVars,
	)
	if err != nil {
		return nil, 0, stacktrace.Propagate(
			err,
			"An error occurred trying to create a Kurtosis engine using image '%v' with tag '%v'",
			imageOrgAndRepo,
			imageVersionTag,
		)
	}
	return publicIpAddr, publicPortNum, nil
}

// Gets point-in-time data about engines matching the given filters
func (backend *KurtosisBackend) GetEngines(ctx context.Context, filters *engine.GetEnginesFilters) (map[string]*engine.Engine, error) {
	engines, err := backend.kurtosisBackendCore.GetEngines(ctx, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting engines using filters: %+v", filters)
	}
	return engines, nil
}

func (backend *KurtosisBackend) StopEngines(ctx context.Context, ids map[string]bool) error {
	err := backend.kurtosisBackendCore.StopEngines(ctx, ids)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while trying to stop engines with IDs: %+v", ids)
	}
	return nil
}

func (backend *KurtosisBackend) DestroyEngines(ctx context.Context, ids map[string]bool) error {
	err := backend.kurtosisBackendCore.DestroyEngines(ctx, ids)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while trying to destroy engines with IDs: %+v", ids)
	}
	return nil
}

/*
func (backend *KurtosisBackend) CleanStoppedEngines(ctx context.Context) ([]string, []error, error) {
	engineNames, engineErrors, err := backend.kurtosisBackendCore.CleanStoppedEngines(ctx)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred while trying to clean stopped Kurtosis engines")
	}
	return engineNames, engineErrors, nil
}

 */

/*
func (backend *KurtosisBackend) GetEnginePublicIPAndPort(
	ctx context.Context,
) (
	resultPublicIpAddr net.IP,
	resultPublicPortNum uint16,
	resultIsEngineStopped bool,
	resultErr error,
) {
	publicIpAddr, publicPortNum, isEngineStopped, err := backend.kurtosisBackendCore.GetEnginePublicIPAndPort(ctx)
	if err != nil {
		return nil, 0, false, stacktrace.Propagate(err, "An error occurred while trying to get the engine status")
	}
	return publicIpAddr, publicPortNum, isEngineStopped, nil
}

 */