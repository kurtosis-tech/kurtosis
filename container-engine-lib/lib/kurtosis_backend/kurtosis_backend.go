package kurtosis_backend

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/kurtosis_backend_core"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
)

type KurtosisBackend struct {
	kurtosisBackendCore kurtosis_backend_core.KubernetesBackendCore
	log                 *logrus.Logger
}

func NewKurtosisBackend(log *logrus.Logger, kurtosisBackendCore kurtosis_backend_core.KubernetesBackendCore) *KurtosisBackend {
	return &KurtosisBackend{
		log:                 log,
		kurtosisBackendCore: kurtosisBackendCore,
	}
}

func (kb KurtosisBackend) CreateEngine(
	ctx context.Context,
	imageVersionTag string,
	logLevel logrus.Level,
	listenPortNum uint16,
	engineDataDirpathOnHostMachine string,
	containerImage string,
) (
	resultPublicIpAddr net.IP,
	resultPublicPortNum uint16,
	resultErr error,
) {
	resultPublicIpAddr, resultPublicPortNum, resultErr = kb.kurtosisBackendCore.CreateEngine(ctx, imageVersionTag, logLevel, listenPortNum, engineDataDirpathOnHostMachine, containerImage)
	if resultErr != nil {
		return nil, 0, stacktrace.Propagate(resultErr, " an error ocurred while trying to create the kurtosis engine")
	}
	return resultPublicIpAddr, resultPublicPortNum, resultErr
}

func (kb KurtosisBackend) StopEngine(ctx context.Context) error {
	err := kb.kurtosisBackendCore.StopEngine(ctx)
	if err != nil {
		return stacktrace.Propagate(err, " an error ocurred while trying to stop the kurtosis engine")
	}
	return nil
}

func (kb KurtosisBackend) CleanStoppedEngines(ctx context.Context) ([]string, []error, error) {
	engineNames, engineErrors, err := kb.kurtosisBackendCore.CleanStoppedEngines(ctx)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, " an error ocurred while trying to clean the kurtosis engines")
	}
	return engineNames, engineErrors, nil
}

func (kb KurtosisBackend) GetEngineStatus(
	ctx context.Context,
) (engineStatus string, ipAddr net.IP, portNum uint16, err error) {
	engineStatus, ipAddr, portNum, err = kb.kurtosisBackendCore.GetEngineStatus(ctx)
	if err != nil {
		return "", ipAddr, 0, stacktrace.Propagate(err, " an error ocurred while trying to get the engine status")
	}
	return engineStatus, ipAddr, portNum, nil
}
