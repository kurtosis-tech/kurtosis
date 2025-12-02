package identifiers_test

import (
	"context"
	"testing"
	"time"

	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

const (
	testName = "identifiers-test"

	datastoreServiceName = "datastore"
	shortenedUuidLength  = 12

	invalidServiceName = "invalid-service"
	invalidEnclaveName = "invalid-enclave"

	secondsToSleepForK8SToDestroyEnclave = 10 * time.Second

	removalScript = `
def run(plan):
	plan.remove_service("datastore")
`
)

func TestIdentifiers(t *testing.T) {

	ctx := context.Background()

	// ------------------------------------- TEST SETUP ----------------------------------------------
	enclaveCtx, _, destroyEnclaveFunc, err := test_helpers.CreateEnclave(t, ctx, testName)
	require.NoError(t, err, "An error occurred creating an enclave")
	shouldDestroyEnclave := true
	defer func() {
		if !shouldDestroyEnclave {
			return
		}
		err = destroyEnclaveFunc()
		if err != nil {
			logrus.Errorf("An error occurred while destroying the enclave '%v'", testName)
		}
	}()

	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	require.NoError(t, err)

	enclaveUuid := enclaveCtx.GetEnclaveUuid()
	enclaveUuidStr := string(enclaveUuid)
	enclaveName := enclaveCtx.GetEnclaveName()
	shortenedEnclaveUuidStr := enclaveUuidStr[:shortenedUuidLength]

	datastoreServiceCtx, _, datastoreClientCloseFunc, err := test_helpers.AddDatastoreService(ctx, datastoreServiceName, enclaveCtx)
	require.NoError(t, err, "An error occurred adding the datastore service to the enclave")
	defer datastoreClientCloseFunc()

	serviceUuid := datastoreServiceCtx.GetServiceUUID()
	serviceUuidStr := string(serviceUuid)
	serviceNameStr := string(datastoreServiceCtx.GetServiceName())
	shortenedServiceUuidStr := serviceUuidStr[:shortenedUuidLength]

	// ------------------------------------- TEST RUN ----------------------------------------------

	// test with service and enclave alive and running

	enclaveIdentifiers, err := kurtosisCtx.GetExistingAndHistoricalEnclaveIdentifiers(ctx)
	require.NoError(t, err)

	enclaveUuidResult, err := enclaveIdentifiers.GetEnclaveUuidForIdentifier(shortenedEnclaveUuidStr)
	require.NoError(t, err)
	require.Equal(t, enclaveUuid, enclaveUuidResult)
	_, err = kurtosisCtx.GetEnclaveContext(ctx, shortenedEnclaveUuidStr)
	require.NoError(t, err)

	enclaveUuidResult, err = enclaveIdentifiers.GetEnclaveUuidForIdentifier(enclaveUuidStr)
	require.NoError(t, err)
	require.Equal(t, enclaveUuid, enclaveUuidResult)
	_, err = kurtosisCtx.GetEnclaveContext(ctx, enclaveUuidStr)
	require.NoError(t, err)

	enclaveUuidResult, err = enclaveIdentifiers.GetEnclaveUuidForIdentifier(enclaveName)
	require.NoError(t, err)
	require.Equal(t, enclaveUuid, enclaveUuidResult)
	_, err = kurtosisCtx.GetEnclaveContext(ctx, enclaveName)
	require.NoError(t, err)

	_, err = enclaveIdentifiers.GetEnclaveUuidForIdentifier(invalidEnclaveName)
	require.Error(t, err)

	serviceIdentifiers, err := enclaveCtx.GetExistingAndHistoricalServiceIdentifiers(ctx)
	require.Nil(t, err)

	serviceUuidResult, err := serviceIdentifiers.GetServiceUuidForIdentifier(serviceUuidStr)
	require.NoError(t, err)
	require.Equal(t, serviceUuid, serviceUuidResult)
	_, err = enclaveCtx.GetServiceContext(serviceUuidStr)
	require.NoError(t, err)

	serviceUuidResult, err = serviceIdentifiers.GetServiceUuidForIdentifier(shortenedServiceUuidStr)
	require.NoError(t, err)
	require.Equal(t, serviceUuid, serviceUuidResult)
	_, err = enclaveCtx.GetServiceContext(shortenedServiceUuidStr)
	require.NoError(t, err)

	serviceUuidResult, err = serviceIdentifiers.GetServiceUuidForIdentifier(serviceNameStr)
	require.NoError(t, err)
	require.Equal(t, serviceUuid, serviceUuidResult)
	_, err = enclaveCtx.GetServiceContext(serviceNameStr)
	require.NoError(t, err)

	_, err = serviceIdentifiers.GetServiceUuidForIdentifier(invalidServiceName)
	require.Error(t, err)

	// remove service and identifier look up should work but context lookup not

	result, err := test_helpers.RunScriptWithDefaultConfig(ctx, enclaveCtx, removalScript)
	require.NoError(t, err)
	require.Nil(t, result.ExecutionError)
	require.Nil(t, result.InterpretationError)
	require.Empty(t, result.ValidationErrors)

	serviceIdentifiersAfterDeletion, err := enclaveCtx.GetExistingAndHistoricalServiceIdentifiers(ctx)
	require.Nil(t, err)

	serviceUuidResult, err = serviceIdentifiersAfterDeletion.GetServiceUuidForIdentifier(serviceUuidStr)
	require.NoError(t, err)
	require.Equal(t, serviceUuid, serviceUuidResult)
	_, err = enclaveCtx.GetServiceContext(serviceUuidStr)
	require.Error(t, err)

	serviceUuidResult, err = serviceIdentifiersAfterDeletion.GetServiceUuidForIdentifier(shortenedServiceUuidStr)
	require.NoError(t, err)
	require.Equal(t, serviceUuid, serviceUuidResult)
	_, err = enclaveCtx.GetServiceContext(shortenedServiceUuidStr)
	require.Error(t, err)

	serviceUuidResult, err = serviceIdentifiersAfterDeletion.GetServiceUuidForIdentifier(serviceNameStr)
	require.NoError(t, err)
	require.Equal(t, serviceUuid, serviceUuidResult)
	_, err = enclaveCtx.GetServiceContext(serviceNameStr)
	require.Error(t, err)

	// now destroying the enclave
	err = destroyEnclaveFunc()
	require.NoError(t, err)
	shouldDestroyEnclave = false

	// we have to put this test to sleep as k8s doesn't do removals immediately
	time.Sleep(secondsToSleepForK8SToDestroyEnclave)

	// this should now fail as service identifiers are stored in memory & the enclave is destroyed
	_, err = enclaveCtx.GetExistingAndHistoricalServiceIdentifiers(ctx)
	require.Error(t, err)

	enclaveIdentifiersAfterStop, err := kurtosisCtx.GetExistingAndHistoricalEnclaveIdentifiers(ctx)
	require.NoError(t, err)

	enclaveUuidResult, err = enclaveIdentifiersAfterStop.GetEnclaveUuidForIdentifier(shortenedEnclaveUuidStr)
	require.NoError(t, err)
	require.Equal(t, enclaveUuid, enclaveUuidResult)
	_, err = kurtosisCtx.GetEnclaveContext(ctx, shortenedEnclaveUuidStr)
	require.Error(t, err)

	enclaveUuidResult, err = enclaveIdentifiersAfterStop.GetEnclaveUuidForIdentifier(enclaveUuidStr)
	require.NoError(t, err)
	require.Equal(t, enclaveUuid, enclaveUuidResult)
	_, err = kurtosisCtx.GetEnclaveContext(ctx, enclaveUuidStr)
	require.Error(t, err)

	enclaveUuidResult, err = enclaveIdentifiersAfterStop.GetEnclaveUuidForIdentifier(enclaveName)
	require.NoError(t, err)
	require.Equal(t, enclaveUuid, enclaveUuidResult)
	_, err = kurtosisCtx.GetEnclaveContext(ctx, enclaveName)
	require.Error(t, err)
}
