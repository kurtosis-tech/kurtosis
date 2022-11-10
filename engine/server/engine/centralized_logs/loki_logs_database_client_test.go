package centralized_logs

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/stretchr/testify/require"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"
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

	userServiceContainerType = "user-service"

	testTimeOut = 30 * time.Second

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

	userServiceLogsByGuidChan, errChan, closeStreamFunc, err := logsDatabaseClient.GetUserServiceLogs(ctx, enclaveId, userServiceGuids)
	defer closeStreamFunc()

	require.NoError(t, err, "An error occurred getting user service logs for GUIDs '%+v' in enclave '%v'", userServiceGuids, enclaveId)
	require.NotNil(t, userServiceLogsByGuidChan, "Received a nil user service logs channel, but a non-nil value was expected")
	require.Nil(t, errChan, "Received a not nil error channel, but a nil value was expected")

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

	userServiceLogsByGuidChan, errChan, closeStreamFunc, err := logsDatabaseClient.GetUserServiceLogs(ctx, enclaveId, userServiceGuids)
	defer closeStreamFunc()

	require.NoError(t, err, "An error occurred getting user service logs for GUIDs '%+v' in enclave '%v'", userServiceGuids, enclaveId)
	require.NotNil(t, userServiceLogsByGuidChan, "Received a nil user service logs channel, but a non-nil value was expected")
	require.Nil(t, errChan, "Received a not nil error channel, but a nil value was expected")

	var testEvaluationErr error

	shouldReceiveStream := true
	for shouldReceiveStream {
		select {
		case <-time.Tick(testTimeOut):
			testEvaluationErr = stacktrace.NewError("Receiving stream logs in the test has reached the '%v' time out", testTimeOut)
			shouldReceiveStream = false
			break
		case userServiceLogsByGuid, isChanOpen := <-userServiceLogsByGuidChan:
			if !isChanOpen {
				shouldReceiveStream = false
				break
			}

			require.Equal(t, len(userServiceGuids), len(userServiceLogsByGuid))

			for userServiceGuid := range userServiceGuids {
				logLines, found := userServiceLogsByGuid[userServiceGuid]
				require.True(t, found)

				expectedAmountLogLines, found := expectedUserServiceAmountLogLinesByUserServiceGuid[userServiceGuid]
				require.True(t, found)

				require.Equal(t, expectedAmountLogLines, len(logLines))

				require.Equal(t, expectedFirstLogLineOnEachService, logLines[0].GetContent())
			}

			shouldReceiveStream = false
			break
		}
	}

	require.NoError(t, testEvaluationErr)

}

func TestNewUserServiceLogLinesByUserServiceGuidFromLokiStreamsReturnSuccessfullyForLogTailJsonResponseBody(t *testing.T) {

	expectedLogLines := []string{"kurtosis", "test", "running", "successfully"}
	userServiceGuidStr := "stream-logs-test-service-1666785469"
	userServiceGuid := service.ServiceGUID(userServiceGuidStr)

	expectedValuesInStream1 := [][]string{
		{"1666785473000000000", "{\"container_id\":\"b0735bc50a76a0476928607aca13a4c73c814036bdbf8b989c2f3b458cc21eab\",\"container_name\":\"/ts-testsuite.stream-logs-test.1666785464--user-service--stream-logs-test-service-1666785469\",\"source\":\"stdout\",\"log\":\"kurtosis\",\"comKurtosistechGuid\":\"stream-logs-test-service-1666785469\",\"comKurtosistechContainerType\":\"user-service\",\"com.kurtosistech.enclave-id\":\"ts-testsuite.stream-logs-test.1666785464\"}"},
	}

	expectedValuesInStream2 := [][]string{
		{"1666785473000000000", "{\"comKurtosistechGuid\":\"stream-logs-test-service-1666785469\",\"container_id\":\"b0735bc50a76a0476928607aca13a4c73c814036bdbf8b989c2f3b458cc21eab\",\"container_name\":\"/ts-testsuite.stream-logs-test.1666785464--user-service--stream-logs-test-service-1666785469\",\"source\":\"stdout\",\"log\":\"test\",\"comKurtosistechContainerType\":\"user-service\",\"com.kurtosistech.enclave-id\":\"ts-testsuite.stream-logs-test.1666785464\"}"},
	}

	expectedValuesInStream3 := [][]string{
		{"1666785473000000000", "{\"comKurtosistechContainerType\":\"user-service\",\"com.kurtosistech.enclave-id\":\"ts-testsuite.stream-logs-test.1666785464\",\"comKurtosistechGuid\":\"stream-logs-test-service-1666785469\",\"container_id\":\"b0735bc50a76a0476928607aca13a4c73c814036bdbf8b989c2f3b458cc21eab\",\"container_name\":\"/ts-testsuite.stream-logs-test.1666785464--user-service--stream-logs-test-service-1666785469\",\"source\":\"stdout\",\"log\":\"running\"}"},
	}

	expectedValuesInStream4 := [][]string{
		{"1666785473000000000", "{\"container_name\":\"/ts-testsuite.stream-logs-test.1666785464--user-service--stream-logs-test-service-1666785469\",\"source\":\"stdout\",\"log\":\"successfully\",\"comKurtosistechGuid\":\"stream-logs-test-service-1666785469\",\"comKurtosistechContainerType\":\"user-service\",\"com.kurtosistech.enclave-id\":\"ts-testsuite.stream-logs-test.1666785464\",\"container_id\":\"b0735bc50a76a0476928607aca13a4c73c814036bdbf8b989c2f3b458cc21eab\"}"},
	}

	lokiStreams1 := newLokiStreamValueForTest(userServiceGuid, expectedValuesInStream1)
	lokiStreams2 := newLokiStreamValueForTest(userServiceGuid, expectedValuesInStream2)
	lokiStreams3 := newLokiStreamValueForTest(userServiceGuid, expectedValuesInStream3)
	lokiStreams4 := newLokiStreamValueForTest(userServiceGuid, expectedValuesInStream4)

	lokiStreams := []lokiStreamValue{
		lokiStreams1,
		lokiStreams2,
		lokiStreams3,
		lokiStreams4,
	}

	resultLogsByKurtosisUserServiceGuid, err := newUserServiceLogLinesByUserServiceGuidFromLokiStreams(lokiStreams)
	require.NoError(t, err)
	require.NotNil(t, resultLogsByKurtosisUserServiceGuid)
	require.Equal(t, len(lokiStreams), len(resultLogsByKurtosisUserServiceGuid[userServiceGuid]))
	for expectedLogLineIndex, expectedLogLine := range expectedLogLines {
		actualLogLine := resultLogsByKurtosisUserServiceGuid[userServiceGuid][expectedLogLineIndex].GetContent()
		require.Equal(t, expectedLogLine, actualLogLine)
	}

}

// ====================================================================================================
//
//	Private Helper Functions
//
// ====================================================================================================
func newLokiStreamValueForTest(userServiceGuid service.ServiceGUID, expectedValues [][]string) lokiStreamValue {
	newLokiStreamValue := lokiStreamValue{
		Stream: struct {
			KurtosisContainerType string `json:"comKurtosistechContainerType"`
			KurtosisGUID          string `json:"comKurtosistechGuid"`
		}(struct {
			KurtosisContainerType string
			KurtosisGUID          string
		}{KurtosisContainerType: userServiceContainerType, KurtosisGUID: string(userServiceGuid)}),
		Values: expectedValues,
	}
	return newLokiStreamValue
}
