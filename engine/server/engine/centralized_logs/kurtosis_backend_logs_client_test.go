package centralized_logs

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testKurtosisBackendLogsClientEnclaveId = "test-enclave"
	testKurtosisBackendLogsClientUserService1Guid = "test-user-service-1"
	testKurtosisBackendLogsClientUserService2Guid = "test-user-service-2"
	testKurtosisBackendLogsClientUserService3Guid = "test-user-service-3"
)

func TestDoRequestWithKurtosisBackendLogsClientReturnsValidResponse(t *testing.T) {
	ctx := context.Background()

	minimalKurtosisBackend := NewMinimalKurtosisBackendMock()
	kurtosisBackendLogsClient := NewKurtosisBackendLogClient(minimalKurtosisBackend)

	userServiceGuids := map[service.ServiceGUID]bool {
		testKurtosisBackendLogsClientUserService1Guid: true,
		testKurtosisBackendLogsClientUserService2Guid: true,
		testKurtosisBackendLogsClientUserService3Guid: true,
	}

	userServiceLogsByUserServiceGuids, err := kurtosisBackendLogsClient.GetUserServiceLogs(ctx, testKurtosisBackendLogsClientEnclaveId, userServiceGuids)
	require.NoError(t, err, "Expected to receive a successful response after calling GetUserServiceLogs but this returned an error")

	require.NotNil(t, userServiceLogsByUserServiceGuids)

	//TODO compare the response values

}
