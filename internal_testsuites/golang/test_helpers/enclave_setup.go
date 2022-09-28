package test_helpers

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

const (
	testsuiteNameEnclaveIDFragment = "go-testsuite"

	millisInNanos = 1000

	coreAndEngineVersion = "CORE_ENGINE_VERSION_TAG"
)

func CreateEnclave(t *testing.T, ctx context.Context, testName string, isPartitioningEnabled bool) (resultEnclaveCtx *enclaves.EnclaveContext, resultStopEnclaveFunc func(), resultDestroyEnclaveFunc func() error, resultErr error) {
	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	require.NoError(t, err, "An error occurred connecting to the Kurtosis engine for running test '%v'", testName)
	enclaveId := enclaves.EnclaveID(fmt.Sprintf(
		"%v.%v.%v",
		testsuiteNameEnclaveIDFragment,
		testName,
		time.Now().UnixNano()/millisInNanos,
	))

	apiContainerVersion := getAPIContainerVersionFromEnvironmentOrDefault()

	enclaveCtx, err := kurtosisCtx.CreateEnclaveWithCustomAPIContainerVersion(ctx, enclaveId, isPartitioningEnabled, apiContainerVersion)
	require.NoError(t, err, "An error occurred creating enclave '%v'", enclaveId)
	stopEnclaveFunc := func() {

		if err := kurtosisCtx.StopEnclave(ctx, enclaveId); err != nil {
			logrus.Errorf("An error occurred stopping enclave '%v' that we created for this test:\n%v", enclaveId, err)
			logrus.Errorf("ACTION REQUIRED: You'll need to stop enclave '%v' manually!!!!", enclaveId)
		}

	}
	destroyEnclaveFunc := func() error {
		if err := kurtosisCtx.DestroyEnclave(ctx, enclaveId); err != nil {
			logrus.Errorf("An error occurred destroying enclave '%v' that we created for this test:\n%v", enclaveId, err)
			logrus.Errorf("ACTION REQUIRED: You'll need to destroy enclave '%v' manually!!!!", enclaveId)
			return err
		}
		return nil
	}

	return enclaveCtx, stopEnclaveFunc, destroyEnclaveFunc, nil
}

func getAPIContainerVersionFromEnvironmentOrDefault() string {
	apiContainerVersion := os.Getenv(coreAndEngineVersion)
	if apiContainerVersion == "" {
		logrus.Debugf("Environment variable '%s' not set, proceeding with the launchers default.", coreAndEngineVersion)
	}
	return apiContainerVersion
}
