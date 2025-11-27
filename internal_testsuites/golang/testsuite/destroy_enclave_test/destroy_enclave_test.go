package destroy_enclave_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/dzobbe/PoTE-kurtosis/api/golang/core/lib/services"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testName = "destroy-enclave"

	datastoreServiceName services.ServiceName = "datastore-service"
)

func TestDestroyEnclave(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, stopEnclaveFunc, destroyEnclaveFunc, err := test_helpers.CreateEnclave(t, ctx, testName)
	require.NoError(t, err, "An error occurred creating an enclave")
	shouldStopEnclaveAtTheEnd := true
	defer func() {
		if shouldStopEnclaveAtTheEnd {
			stopEnclaveFunc()
		}
	}()

	// ------------------------------------- TEST SETUP ----------------------------------------------
	_, _, _, err = test_helpers.AddDatastoreService(ctx, datastoreServiceName, enclaveCtx)
	require.NoError(t, err, "An error occurred adding the file server service")

	err = destroyEnclaveFunc()
	require.NoErrorf(t, err, "An error occurred destroying enclave with UUID '%v'", enclaveCtx.GetEnclaveUuid())
	shouldStopEnclaveAtTheEnd = false
}
