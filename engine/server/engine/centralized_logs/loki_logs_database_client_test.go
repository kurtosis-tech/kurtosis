package centralized_logs

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/stretchr/testify/require"
	"net/http"
	"regexp"
	"strings"
	"testing"
)

const (
	fakeLogsDatabaseAddress = "1.2.3.4:8080"

	testEnclaveId        = "test-enclave"
	testUserService1Guid = "test-user-service-1"
	testUserService2Guid = "test-user-service-2"
	testUserService3Guid = "test-user-service-3"

	//Expected values
	expectedFirstLogLineOnEachService       = "This is the first log line."
	expectedOrganizationIdHttpHeaderKey     = "X-Scope-Orgid"
	expectedStartTimeQueryParamKey          = "start"
	expectedQueryLogsQueryParamKey          = "query"
	expectedEntriesLimitQueryParamKey       = "limit"
	expectedDirectionQueryParamKey          = "direction"
	expectedKurtosisContainerTypeLokiTagKey = "kurtosisContainerType"
	expectedKurtosisGuidLokiTagKey          = "kurtosisGUID"
	expectedURLScheme                       = "http"
	expectedQueryRangeURLPath               = "/loki/api/v1/query_range"
	expectedQueryLogsQueryParamValueRegex   = `{kurtosisContainerType="user-service",kurtosisGUID=~"test-user-service-[1-3]\|test-user-service-[1-3]\|test-user-service-[1-3]"}`
	expectedEntriesLimitQueryParamValue     = "4000"
	expectedDirectionQueryParamValue        = "forward"
)


func TestIfHttpRequestIsValidWhenCallingGetUserServiceLogs(t *testing.T) {
	enclaveId        := enclave.EnclaveID(testEnclaveId)
	userServiceGuids := map[service.ServiceGUID]bool{
		testUserService1Guid: true,
		testUserService2Guid: true,
		testUserService3Guid: true,
	}
	httpClientObj      := NewMockedHttpClient()
	logsDatabaseClient := NewLokiLogsDatabaseClient(fakeLogsDatabaseAddress, httpClientObj)

	ctx := context.Background()

	userServiceLogsByUserServiceGuids, err := logsDatabaseClient.GetUserServiceLogs(ctx, enclaveId, userServiceGuids)
	require.NoError(t, err, "An error occurred getting user service logs for GUIDs '%+v' in enclave '%v'", userServiceGuids, enclaveId)
	require.NotNil(t, userServiceLogsByUserServiceGuids, "Received a nil user service logs, but a non-nil value was expected")

	mockedHttpClientObj, ok := logsDatabaseClient.httpClient.(*mockedHttpClient)
	require.True(t, ok)

	request := mockedHttpClientObj.GetRequest()

	require.Equal(t, expectedURLScheme, request.URL.Scheme)
	require.Equal(t, fakeLogsDatabaseAddress, request.URL.Host)
	require.Equal(t, expectedQueryRangeURLPath, request.URL.Path)
	require.Equal(t, http.MethodGet, request.Method)

	organizationIds, found := request.Header[expectedOrganizationIdHttpHeaderKey]
	require.True(t, found, "Expected to find header key '%v' in request header '%+v', but it was not found", expectedOrganizationIdHttpHeaderKey, request.Header)

	expectedEnclaveId := enclaveId
	var foundExpectedEnclaveId bool
	for _, organizationId := range organizationIds {
		enclaveIdObj := enclave.EnclaveID(organizationId)
		if enclaveIdObj == expectedEnclaveId {
			foundExpectedEnclaveId = true
		}
	}
	require.True(t, foundExpectedEnclaveId, "Expected to find enclave ID '%v' in request header values '%+v' for header with key '%v', but it was not found", expectedEnclaveId, organizationIds, expectedOrganizationIdHttpHeaderKey)

	_, found = request.Form[expectedStartTimeQueryParamKey]
	require.True(t, found, "Expected to find query param with key '%v' in request form values '%+v', but it was not found", expectedStartTimeQueryParamKey, request.Form)

	queryLogsQueryParams, found := request.Form[expectedQueryLogsQueryParamKey]
	require.True(t, found, "Expected to find query param with key '%v' in request form values '%+v', but it was not found", expectedStartTimeQueryParamKey, request.Form)

	require.Regexp(t, regexp.MustCompile(expectedQueryLogsQueryParamValueRegex), queryLogsQueryParams)

	var (
		foundExpectedKurtosisContainerTypeLokiTagKey bool
		foundExpectedKurtosisGuidLokiTagKey          bool
	)

	for _, queryLogParam := range queryLogsQueryParams {
		foundKurtosisContainerTypeLokiTagKey := strings.Contains(queryLogParam, expectedKurtosisContainerTypeLokiTagKey)
		if foundKurtosisContainerTypeLokiTagKey {
			foundExpectedKurtosisContainerTypeLokiTagKey = true
		}
		foundKurtosisGuidLokiTagKey := strings.Contains(queryLogParam, expectedKurtosisGuidLokiTagKey)
		if foundKurtosisGuidLokiTagKey {
			foundExpectedKurtosisGuidLokiTagKey = true
		}
	}
	require.True(t, foundExpectedKurtosisContainerTypeLokiTagKey, "Expected to find Loki's tag key key '%v' in request query params '%+v', but it was not found", expectedKurtosisContainerTypeLokiTagKey, queryLogsQueryParams)
	require.True(t, foundExpectedKurtosisGuidLokiTagKey, "Expected to find Loki's tag key key '%v' in request query params '%+v', but it was not found", expectedKurtosisGuidLokiTagKey, queryLogsQueryParams)


	limitQueryParams, found := request.Form[expectedEntriesLimitQueryParamKey]
	require.True(t, found, "Expected to find query param with key '%v' in request form values '%+v', but it was not found", expectedEntriesLimitQueryParamKey, request.Form)
	require.Equal(t, expectedEntriesLimitQueryParamValue, limitQueryParams[0])

	directionQueryParams, found := request.Form[expectedDirectionQueryParamKey]
	require.True(t, found, "Expected to find query param with key '%v' in request form values '%+v', but it was not found", expectedDirectionQueryParamKey, request.Form)
	require.Equal(t, expectedDirectionQueryParamValue, directionQueryParams[0])

}

func TestDoRequestWithLokiLogsDatabaseClientReturnsValidResponse(t *testing.T) {
	enclaveId        := enclave.EnclaveID(testEnclaveId)
	userServiceGuids := map[service.ServiceGUID]bool{
		testUserService1Guid: true,
		testUserService2Guid: true,
		testUserService3Guid: true,
	}
	httpClientObj      := NewMockedHttpClient()
	logsDatabaseClient := NewLokiLogsDatabaseClient(fakeLogsDatabaseAddress, httpClientObj)

	ctx := context.Background()

	expectedUserServiceAmountLogLinesByUserServiceGuid := map[service.ServiceGUID]int{
		testUserService1Guid: 3,
		testUserService2Guid: 4,
		testUserService3Guid: 2,
	}

	userServiceLogsByUserServiceGuids, err := logsDatabaseClient.GetUserServiceLogs(ctx, enclaveId, userServiceGuids)
	require.NoError(t, err, "An error occurred getting user service logs for GUIDs '%+v' in enclave '%v'", userServiceGuids, enclaveId)
	require.NotNil(t, userServiceLogsByUserServiceGuids, "Received a nil user service logs, but a non-nil value was expected")

	require.Equal(t, len(userServiceGuids), len(userServiceLogsByUserServiceGuids))

	for userServiceGuid := range userServiceGuids {
		logLines, found := userServiceLogsByUserServiceGuids[userServiceGuid]
		require.True(t, found)

		expectedAmountLogLines, found := expectedUserServiceAmountLogLinesByUserServiceGuid[userServiceGuid]
		require.True(t, found)

		require.Equal(t, expectedAmountLogLines, len(logLines))

		require.Equal(t, expectedFirstLogLineOnEachService, logLines[0])
	}
}
