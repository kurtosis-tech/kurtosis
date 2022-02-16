package kurtosis_backend

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/kurtosis_backend_core"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
)

type KurtosisBackend struct {
	kurtosisBackendCore kurtosis_backend_core.KurtosisBackendCore
	log                 *logrus.Logger
}

func NewKurtosisBackend(log *logrus.Logger, kurtosisBackendCore kurtosis_backend_core.KurtosisBackendCore) *KurtosisBackend {
	return &KurtosisBackend{
		log:                 log,
		kurtosisBackendCore: kurtosisBackendCore,
	}
}

func (backend *KurtosisBackend) CreateEngine(
	ctx context.Context,
	imageVersionTag string,
	logLevel logrus.Level,
	listenPortNum uint16,
	engineDataDirpathOnHostMachine string,
	imageOrgAndRepo string,
	serializedEnvVars map[string]string,
) (
	resultPublicIpAddr net.IP,
	resultPublicPortNum uint16,
	resultErr error,
) {
	publicIpAddr, publicPortNum, err := backend.kurtosisBackendCore.CreateEngine(ctx, imageVersionTag, logLevel, listenPortNum, engineDataDirpathOnHostMachine, imageOrgAndRepo, serializedEnvVars)
	if err != nil {
		return nil, 0, stacktrace.Propagate(resultErr, "An error occurred while trying to create the kurtosis engine with publicIpAddr '%v' and publicPortNum '%v'", publicIpAddr, publicPortNum)
	}
	return publicIpAddr, publicPortNum, nil
}

func (backend *KurtosisBackend) StopEngine(ctx context.Context) error {
	err := backend.kurtosisBackendCore.StopEngine(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while trying to stop the kurtosis engine")
	}
	return nil
}

func (backend *KurtosisBackend) CleanStoppedEngines(ctx context.Context) ([]string, []error, error) {
	engineNames, engineErrors, err := backend.kurtosisBackendCore.CleanStoppedEngines(ctx)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred while trying to clean stopped Kurtosis engines")
	}
	return engineNames, engineErrors, nil
}

func (backend *KurtosisBackend) GetEngineStatus(
	ctx context.Context,
) (resultEngineStatus string, resultPublicIpAddr net.IP, resultPortNum uint16, resultErr error) {
	engineStatus, ipAddr, portNum, err := backend.kurtosisBackendCore.GetEngineStatus(ctx)
	if err != nil {
		return "", ipAddr, 0, stacktrace.Propagate(err, "An error occurred while trying to get the engine status")
	}
	return engineStatus, ipAddr, portNum, nil
}
