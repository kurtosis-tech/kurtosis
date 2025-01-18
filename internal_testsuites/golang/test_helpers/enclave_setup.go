package test_helpers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kurtosis-tech/kurtosis/api/golang/v2/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/v2/util"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

const (
	testsuiteNameEnclaveIDFragment = "go-test"

	destroyEnclaveRetries                  = 3
	destroyEnclaveRetriesDelayMilliseconds = 1000
)

func CreateEnclave(t *testing.T, ctx context.Context, testName string) (resultEnclaveCtx *enclaves.EnclaveContext, resultStopEnclaveFunc func(), resultDestroyEnclaveFunc func() error, resultErr error) {
	enclaveName := fmt.Sprintf(
		"%v-%v-%v",
		testsuiteNameEnclaveIDFragment,
		testName,
		time.Now().Unix(),
	)
	enclaveCtx, stopEnclaveFunc, destroyEnclaveFunc, err := util.CreateEmptyEnclave(ctx, enclaveName)
	require.NoError(t, err, "An error occurred creating enclave '%v'", enclaveName)
	stopEnclaveFuncWrapped := func() {
		if err := stopEnclaveFunc(); err != nil {
			logrus.Errorf("An error occurred stopping enclave '%v' that we created for this test:\n%v", enclaveName, err)
			logrus.Errorf("ACTION REQUIRED: You'll need to stop enclave '%v' manually!!!!", enclaveName)
		}

	}
	destroyEnclaveFuncWrapped := func() error {
		for i := 0; i < destroyEnclaveRetries; i++ {
			if err := destroyEnclaveFunc(); err != nil {
				logrus.Warnf("An error occurred destroying enclave '%v' that we created for this test:\n%v", enclaveName, err)
				if i == destroyEnclaveRetries-1 {
					logrus.Errorf("An error occurred destroying enclave '%v' that we created for this test:\n%v", enclaveName, err)
					logrus.Errorf("ACTION REQUIRED: You'll need to destroy enclave '%v' manually!!!!", enclaveName)
					return stacktrace.NewError("An error occurred after trying to destroy the enclave '%v' %d times", enclaveName, destroyEnclaveRetries)
				}
				logrus.Warnf("Retrying %d more time(s)", destroyEnclaveRetries-i-1)
				time.Sleep(time.Duration(destroyEnclaveRetriesDelayMilliseconds) * time.Millisecond)
			} else {
				break
			}
		}
		return nil
	}

	return enclaveCtx, stopEnclaveFuncWrapped, destroyEnclaveFuncWrapped, nil
}
