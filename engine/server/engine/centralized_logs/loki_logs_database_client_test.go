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
	expectedKurtosisContainerTypeLokiTagKey = "comKurtosistechContainerType"
	expectedKurtosisGuidLokiTagKey          = "comKurtosistechGuid"
	expectedURLScheme                       = "http"
	expectedQueryRangeURLPath               = "/loki/api/v1/query_range"
	expectedQueryLogsQueryParamValueRegex   = `{comKurtosistechContainerType="user-service",comKurtosistechGuid=~"test-user-service-[1-3]\|test-user-service-[1-3]\|test-user-service-[1-3]"}`
	expectedEntriesLimitQueryParamValue     = "4000"
	expectedDirectionQueryParamValue        = "forward"
	expectedAmountQueryParams               = 4
)

func TestIfHttpRequestIsValidWhenCallingGetUserServiceLogs(t *testing.T) {
	enclaveId := enclave.EnclaveID(testEnclaveId)
	userServiceGuids := map[service.ServiceGUID]bool{
		testUserService1Guid: true,
		testUserService2Guid: true,
		testUserService3Guid: true,
	}
	httpClientObj := NewMockedHttpClient()
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

	require.Equal(t, expectedAmountQueryParams, len(request.URL.Query()), "Expected to request contains '%v' query params, but '%v' query params were found", expectedAmountQueryParams, len(request.URL.Query()))

	found = request.URL.Query().Has(expectedStartTimeQueryParamKey)
	require.True(t, found, "Expected to find query param with key '%v' in request form values '%+v', but it was not found", expectedStartTimeQueryParamKey, request.Form)

	found = request.URL.Query().Has(expectedQueryLogsQueryParamKey)
	require.True(t, found, "Expected to find query param with key '%v' in request form values '%+v', but it was not found", expectedStartTimeQueryParamKey, request.Form)

	queryLogsQueryParams := request.URL.Query().Get(expectedQueryLogsQueryParamKey)
	require.Regexp(t, regexp.MustCompile(expectedQueryLogsQueryParamValueRegex), queryLogsQueryParams)

	var (
		foundExpectedKurtosisContainerTypeLokiTagKey bool
		foundExpectedKurtosisGuidLokiTagKey          bool
	)

	foundKurtosisContainerTypeLokiTagKey := strings.Contains(queryLogsQueryParams, expectedKurtosisContainerTypeLokiTagKey)
	if foundKurtosisContainerTypeLokiTagKey {
		foundExpectedKurtosisContainerTypeLokiTagKey = true
	}
	foundKurtosisGuidLokiTagKey := strings.Contains(queryLogsQueryParams, expectedKurtosisGuidLokiTagKey)
	if foundKurtosisGuidLokiTagKey {
		foundExpectedKurtosisGuidLokiTagKey = true
	}

	require.True(t, foundExpectedKurtosisContainerTypeLokiTagKey, "Expected to find Loki's tag key key '%v' in request query params '%+v', but it was not found", expectedKurtosisContainerTypeLokiTagKey, queryLogsQueryParams)
	require.True(t, foundExpectedKurtosisGuidLokiTagKey, "Expected to find Loki's tag key key '%v' in request query params '%+v', but it was not found", expectedKurtosisGuidLokiTagKey, queryLogsQueryParams)

	found = request.URL.Query().Has(expectedEntriesLimitQueryParamKey)
	require.True(t, found, "Expected to find query param with key '%v' in request form values '%+v', but it was not found", expectedEntriesLimitQueryParamKey, request.Form)
	limitQueryParam := request.URL.Query().Get(expectedEntriesLimitQueryParamKey)
	require.Equal(t, expectedEntriesLimitQueryParamValue, limitQueryParam)

	found = request.URL.Query().Has(expectedDirectionQueryParamKey)
	require.True(t, found, "Expected to find query param with key '%v' in request form values '%+v', but it was not found", expectedDirectionQueryParamKey, request.Form)
	directionQueryParam := request.URL.Query().Get(expectedDirectionQueryParamKey)
	require.Equal(t, expectedDirectionQueryParamValue, directionQueryParam)

}

func TestDoRequestWithLokiLogsDatabaseClientReturnsValidResponse(t *testing.T) {
	enclaveId := enclave.EnclaveID(testEnclaveId)
	userServiceGuids := map[service.ServiceGUID]bool{
		testUserService1Guid: true,
		testUserService2Guid: true,
		testUserService3Guid: true,
	}
	httpClientObj := NewMockedHttpClient()
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

func TestNewUserServiceLogLinesByUserServiceGuidFromLokiStreamsReturnSuccessfullyForLogTailJsonResponseBody(t *testing.T) {

	expectedLogLines := []string{"kurtosis", "test", "running", "successfully"}
	userServiceGuidStr := "stream-logs-test-service-1666785469"
	userServiceGuid := service.ServiceGUID(userServiceGuidStr)

	lokiStreams := []lokiStreamValue{
		 {
			 Stream: struct {
				 KurtosisContainerType string `json:"comKurtosistechContainerType"`
				 KurtosisGUID          string `json:"comKurtosistechGuid"`
			 }(struct {
				 KurtosisContainerType string
				 KurtosisGUID          string
			 }{KurtosisContainerType: "user-service", KurtosisGUID: userServiceGuidStr}),
			 Values: [][]string{
				 {"1666785473000000000", "{\"container_id\":\"b0735bc50a76a0476928607aca13a4c73c814036bdbf8b989c2f3b458cc21eab\",\"container_name\":\"/ts-testsuite.stream-logs-test.1666785464--user-service--stream-logs-test-service-1666785469\",\"source\":\"stdout\",\"log\":\"kurtosis\",\"comKurtosistechGuid\":\"stream-logs-test-service-1666785469\",\"comKurtosistechContainerType\":\"user-service\",\"com.kurtosistech.enclave-id\":\"ts-testsuite.stream-logs-test.1666785464\"}"},
			 },

		},
		{
			Stream: struct {
				KurtosisContainerType string `json:"comKurtosistechContainerType"`
				KurtosisGUID          string `json:"comKurtosistechGuid"`
			}(struct {
				KurtosisContainerType string
				KurtosisGUID          string
			}{KurtosisContainerType: "user-service", KurtosisGUID: userServiceGuidStr}),
			Values: [][]string{
				{"1666785473000000000", "{\"comKurtosistechGuid\":\"stream-logs-test-service-1666785469\",\"container_id\":\"b0735bc50a76a0476928607aca13a4c73c814036bdbf8b989c2f3b458cc21eab\",\"container_name\":\"/ts-testsuite.stream-logs-test.1666785464--user-service--stream-logs-test-service-1666785469\",\"source\":\"stdout\",\"log\":\"test\",\"comKurtosistechContainerType\":\"user-service\",\"com.kurtosistech.enclave-id\":\"ts-testsuite.stream-logs-test.1666785464\"}"},
			},

		},
		{
			Stream: struct {
				KurtosisContainerType string `json:"comKurtosistechContainerType"`
				KurtosisGUID          string `json:"comKurtosistechGuid"`
			}(struct {
				KurtosisContainerType string
				KurtosisGUID          string
			}{KurtosisContainerType: "user-service", KurtosisGUID: userServiceGuidStr}),
			Values: [][]string{
				{"1666785473000000000", "{\"comKurtosistechContainerType\":\"user-service\",\"com.kurtosistech.enclave-id\":\"ts-testsuite.stream-logs-test.1666785464\",\"comKurtosistechGuid\":\"stream-logs-test-service-1666785469\",\"container_id\":\"b0735bc50a76a0476928607aca13a4c73c814036bdbf8b989c2f3b458cc21eab\",\"container_name\":\"/ts-testsuite.stream-logs-test.1666785464--user-service--stream-logs-test-service-1666785469\",\"source\":\"stdout\",\"log\":\"running\"}"},
			},

		},
		{
			Stream: struct {
				KurtosisContainerType string `json:"comKurtosistechContainerType"`
				KurtosisGUID          string `json:"comKurtosistechGuid"`
			}(struct {
				KurtosisContainerType string
				KurtosisGUID          string
			}{KurtosisContainerType: "user-service", KurtosisGUID: userServiceGuidStr}),
			Values: [][]string{
				{"1666785473000000000", "{\"container_name\":\"/ts-testsuite.stream-logs-test.1666785464--user-service--stream-logs-test-service-1666785469\",\"source\":\"stdout\",\"log\":\"successfully\",\"comKurtosistechGuid\":\"stream-logs-test-service-1666785469\",\"comKurtosistechContainerType\":\"user-service\",\"com.kurtosistech.enclave-id\":\"ts-testsuite.stream-logs-test.1666785464\",\"container_id\":\"b0735bc50a76a0476928607aca13a4c73c814036bdbf8b989c2f3b458cc21eab\"}"},
			},

		},
	}

	resultLogsByKurtosisUserServiceGuid, err := newUserServiceLogLinesByUserServiceGuidFromLokiStreams(lokiStreams)
	require.NoError(t, err)
	require.NotNil(t, resultLogsByKurtosisUserServiceGuid)
	require.Equal(t, len(lokiStreams), len(resultLogsByKurtosisUserServiceGuid[userServiceGuid]))
	require.Equal(t, expectedLogLines, resultLogsByKurtosisUserServiceGuid[userServiceGuid])
}
