package centralized_logs

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"
)

const (
	testKurtosisBackendLogsClientEnclaveId = "test-enclave"
	testKurtosisBackendLogsClientUserService1Guid = "test-user-service-1"
	testKurtosisBackendLogsClientUserService2Guid = "test-user-service-2"
	testKurtosisBackendLogsClientUserService3Guid = "test-user-service-3"

	nonexistentKurtosisBackendLogsClientUserService4Guid = "test-user-service-4"
	nonexistentKurtosisBackendLogsClientUserService5Guid = "test-user-service-5"
	nonexistentKurtosisBackendLogsClientUserService6Guid = "test-user-service-6"

	expectedSuccessfulUserServiceLogLines= 3

	expectedFirstLogLineOnEachUserServiceRegex = "This is the first user service #[1-3] log line."
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
	require.Equal(t, len(userServiceGuids), len(userServiceLogsByUserServiceGuids))
	for userServiceGuid := range userServiceGuids {
		userServiceLogLines, found := userServiceLogsByUserServiceGuids[userServiceGuid]
		require.True(t, found, "Expected to receive the log lines for user service with GUID '%v' but it was not returned", userServiceGuid)
		require.Equal(t, expectedSuccessfulUserServiceLogLines, len(userServiceLogLines), "Expected to receive '%v' log lines from user service '%v', but '%v' log lines were returned", expectedSuccessfulUserServiceLogLines, userServiceGuid, len(userServiceLogLines))
		require.Regexp(t, regexp.MustCompile(expectedFirstLogLineOnEachUserServiceRegex), userServiceLogLines[0], "Expected the first log line '%v' from user service '%v' asserts the regex '%v', but it does not", userServiceLogLines[0], userServiceGuid, expectedFirstLogLineOnEachUserServiceRegex)
	}
}

func TestDoRequestWithKurtosisBackendLogsClientReturnsErrorResponse(t *testing.T) {
	ctx := context.Background()

	minimalKurtosisBackend := NewMinimalKurtosisBackendMock()
	kurtosisBackendLogsClient := NewKurtosisBackendLogClient(minimalKurtosisBackend)

	nonexistentUserServiceGuids := map[service.ServiceGUID]bool {
		nonexistentKurtosisBackendLogsClientUserService4Guid: true,
		nonexistentKurtosisBackendLogsClientUserService5Guid: true,
		nonexistentKurtosisBackendLogsClientUserService6Guid: true,
	}

	userServiceLogsByUserServiceGuids, err := kurtosisBackendLogsClient.GetUserServiceLogs(ctx, testKurtosisBackendLogsClientEnclaveId, nonexistentUserServiceGuids)
	require.Error(t, err, "Expected to receive an error when getting user service logs for service with GUIDs '%+v' but no error was returned", nonexistentUserServiceGuids)
	require.Nil(t, userServiceLogsByUserServiceGuids)
}
