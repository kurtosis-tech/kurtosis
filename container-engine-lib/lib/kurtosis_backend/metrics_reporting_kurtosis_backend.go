package kurtosis_backend

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/kurtosis_backend/objects/engine"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

// TODO CALL THE METRICS LIBRARY EVENT-REGISTRATION FUNCTIONS HERE!!!!
type MetricsReportingKurtosisBackend struct {
	kurtosisBackendCore KurtosisBackendCore
}

func NewMetricsReportingKurtosisBackend(log *logrus.Logger, kurtosisBackendCore KurtosisBackendCore) *MetricsReportingKurtosisBackend {
	return &MetricsReportingKurtosisBackend{
		kurtosisBackendCore: kurtosisBackendCore,
	}
}

func (backend *MetricsReportingKurtosisBackend) CreateEngine(ctx context.Context, imageOrgAndRepo string, imageVersionTag string, grpcPortNum uint16, grpcProxyPortNum uint16, engineDataDirpathOnHostMachine string, envVars map[string]string) (*engine.Engine, error) {
	result, err := backend.kurtosisBackendCore.CreateEngine(ctx, imageOrgAndRepo, imageVersionTag, grpcPortNum, grpcProxyPortNum, engineDataDirpathOnHostMachine, envVars)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the engine using image '%v' with tag '%v'", imageOrgAndRepo, imageVersionTag)
	}
	return result, nil
}

// Gets point-in-time data about engines matching the given filters
func (backend *MetricsReportingKurtosisBackend) GetEngines(ctx context.Context, filters *engine.GetEnginesFilters) (map[string]*engine.Engine, error) {
	engines, err := backend.kurtosisBackendCore.GetEngines(ctx, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting engines using filters: %+v", filters)
	}
	return engines, nil
}

func (backend *MetricsReportingKurtosisBackend) StopEngines(ctx context.Context, filters *engine.GetEnginesFilters) (
	successfulIds map[string]bool,
	failedIds map[string]error,
	resultErr error,
) {
	successes, failures, err := backend.kurtosisBackendCore.StopEngines(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred stopping engines using filters: %+v", filters)
	}
	return successes, failures, nil
}

func (backend *MetricsReportingKurtosisBackend) DestroyEngines(ctx context.Context, filters *engine.GetEnginesFilters) (
	successfulIds map[string]bool,
	failedIds map[string]error,
	resultErr error,
) {
	successes, failures, err := backend.kurtosisBackendCore.DestroyEngines(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred destroying engines using filters: %+v", filters)
	}
	return successes, failures, nil
}