package test_helpers

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const (
	testsuiteNameEnclaveIDFragment = "go-testsuite"

	millisInNanos = 1000
)

func CreateEnclave(t *testing.T, ctx context.Context, testName string) (resultEnclaveCtx *enclaves.EnclaveContext, resultStopEnclaveFunc func(), resultDestroyEnclaveFunc func() error, resultErr error) {
	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	require.NoError(t, err, "An error occurred connecting to the Kurtosis engine for running test '%v'", testName)
	enclaveName := fmt.Sprintf(
		"%v.%v.%v",
		testsuiteNameEnclaveIDFragment,
		testName,
		time.Now().UnixNano()/millisInNanos,
	)
	enclaveCtx, err := kurtosisCtx.CreateEnclave(ctx, enclaveName)
	require.NoError(t, err, "An error occurred creating enclave '%v'", enclaveName)
	stopEnclaveFunc := func() {

		if err := kurtosisCtx.StopEnclave(ctx, enclaveName); err != nil {
			logrus.Errorf("An error occurred stopping enclave '%v' that we created for this test:\n%v", enclaveName, err)
			logrus.Errorf("ACTION REQUIRED: You'll need to stop enclave '%v' manually!!!!", enclaveName)
		}

	}
	destroyEnclaveFunc := func() error {
		if err := kurtosisCtx.DestroyEnclave(ctx, enclaveName); err != nil {
			logrus.Errorf("An error occurred destroying enclave '%v' that we created for this test:\n%v", enclaveName, err)
			logrus.Errorf("ACTION REQUIRED: You'll need to destroy enclave '%v' manually!!!!", enclaveName)
			return err
		}
		return nil
	}

	return enclaveCtx, stopEnclaveFunc, destroyEnclaveFunc, nil
}
