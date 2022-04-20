package test_helpers

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang/lib/kurtosis_context"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const (
	testsuiteNameEnclaveIDFragment = "go-testsuite"

	millisInNanos = 1000
)

func CreateEnclave(t *testing.T, ctx context.Context, testName string, isPartitioningEnabled bool) (resultEnclaveCtx *enclaves.EnclaveContext, resultStopEnclaveFunc func(), restultKurtosisCtx *kurtosis_context.KurtosisContext, resultErr error) {
	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	require.NoError(t, err, "An error occurred connecting to the Kurtosis engine for running test '%v'", testName)
	enclaveId := enclaves.EnclaveID(fmt.Sprintf(
		"%v.%v.%v",
		testsuiteNameEnclaveIDFragment,
		testName,
		time.Now().UnixNano() / millisInNanos,
	))
	enclaveCtx, err := kurtosisCtx.CreateEnclave(ctx, enclaveId, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating enclave '%v'", enclaveId)
	stopEnclaveFunc := func() {
		if err := kurtosisCtx.StopEnclave(ctx, enclaveId); err != nil {
			logrus.Errorf("An error occurred stopping enclave '%v' that we created for this test:\n%v", enclaveId, err)
			logrus.Errorf("ACTION REQUIRED: You'll need to stop enclave '%v' manually!!!!", enclaveId)
		}
	}

	return enclaveCtx, stopEnclaveFunc, kurtosisCtx, nil
}
