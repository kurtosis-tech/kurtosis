package test_helpers

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/util"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const (
	testsuiteNameEnclaveIDFragment = "go-test"
)

func CreateEnclave(t *testing.T, ctx context.Context, testName string) (resultEnclaveCtx *enclaves.EnclaveContext, resultStopEnclaveFunc func(), resultDestroyEnclaveFunc func() error, resultErr error) {
	enclaveName := fmt.Sprintf(
		"%v-%v-%v",
		testsuiteNameEnclaveIDFragment,
		testName,
		time.Now().UnixNano(),
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
		if err := destroyEnclaveFunc(); err != nil {
			logrus.Errorf("An error occurred destroying enclave '%v' that we created for this test:\n%v", enclaveName, err)
			logrus.Errorf("ACTION REQUIRED: You'll need to destroy enclave '%v' manually!!!!", enclaveName)
			return err
		}
		return nil
	}

	return enclaveCtx, stopEnclaveFuncWrapped, destroyEnclaveFuncWrapped, nil
}
