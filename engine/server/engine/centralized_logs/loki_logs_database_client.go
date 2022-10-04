package centralized_logs

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultHttpClientTimeOut = 1 * time.Minute

	httpProtocol              = "http://"
	baseLokiApiPath           = "/loki/api/v1"
	queryRangeEndpointSubpath = "/query_range"

	//the global retention period store logs for 30 days = 720h.
	maxRetentionPeriod = 720 * time.Hour

	//We use this header because we are using the Loki multi-tenancy feature to split logs by the EnclaveID
	//Read more about it here: https://grafana.com/docs/loki/latest/operations/multi-tenancy/
	//tenantID = enclaveID
	organizationIdHttpHeaderKey = "X-Scope-OrgID"

	startTimeQueryParamKey = "start"
	queryLogsQueryParamKey = "query"
	entriesLimitQueryParamKey = "limit"
	directionQueryParamKey = "direction"

	//We are establishing a big limit in order to get all the user-service's logs in one request
	//We will improve this in the future generating a pagination mechanism based on the limit number
	//and the unix epoch time (as the start time for the next request) returned by newest stream's log line
	defaultEntriesLimit = "4000"
	//The oldest item is first when using direction=forward
	defaultDirection = "forward"

	kurtosisContainerTypeLokiTagKey = "kurtosisContainerType"
	kurtosisGuidLokiTagKey          = "kurtosisGUID"

	orTagsOperator = "|"

	userServiceContainerType = "user-service"

	//A stream value should contain 2 items, the timestamp as the first one, and the log line
	//More here: https://grafana.com/docs/loki/latest/api/
	streamValueNumOfItems = 2

	streamValueLogLineIndex = 1

)

// A backoff schedule for when and how often to retry failed HTTP
// requests. The first element is the time to wait after the
// first failure, the second the time to wait after the second
// failure, etc. After reaching the last element, retries stop
// and the request is considered failed.
var httpRetriesBackoffSchedule = []time.Duration{
	20 * time.Millisecond,
	40 * time.Millisecond,
	1 * time.Second,
}

type LokiQueryRangeResponse struct {
	Status string `json:"status"`
	Data   *struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Stream struct {
				KurtosisContainerType string `json:"kurtosisContainerType"`
				KurtosisGUID          string `json:"kurtosisGUID"`
			} `json:"stream"`
			Values [][]string `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

type LogLine struct {
	Log string `json:"log"`
}

type LokiLogsDatabaseClient struct {
	logsDatabaseAddress string
	httpClient          httpClient
}

func NewLokiLogsDatabaseClient(logsDatabaseAddress string, httpClient httpClient) *LokiLogsDatabaseClient {
	return &LokiLogsDatabaseClient{logsDatabaseAddress: logsDatabaseAddress, httpClient: httpClient}
}

func NewLokiLogsDatabaseClientWithDefaultHttpClient(logsDatabaseAddress string) *LokiLogsDatabaseClient {
	httpClientObj := &http.Client{Timeout: defaultHttpClientTimeOut}
	return &LokiLogsDatabaseClient{logsDatabaseAddress: logsDatabaseAddress, httpClient: httpClientObj}
}

func (client *LokiLogsDatabaseClient) GetUserServiceLogs(
	ctx context.Context,
	enclaveID enclave.EnclaveID,
	userServiceGuids map[service.ServiceGUID]bool,
) (map[service.ServiceGUID][]string, error) {

	resultLogsByKurtosisUserServiceGuid := map[service.ServiceGUID][]string{}

	httpHeaderWithTenantID := http.Header{}
	httpHeaderWithTenantID.Add(organizationIdHttpHeaderKey, string(enclaveID))

	kurtosisGuid := []string{}
	for userServiceGuid := range userServiceGuids {
		kurtosisGuid = append(kurtosisGuid, string(userServiceGuid))
	}

	maxRetentionLogsTimeParamValue := getMaxRetentionLogsTimeParamValue()
	queryParamValue := getQueryParamValue(userServiceContainerType, kurtosisGuid)

	getLogsPath := baseLokiApiPath + queryRangeEndpointSubpath

	queryRangeEndpointUrlStr := fmt.Sprintf("%v%v%v", httpProtocol, client.logsDatabaseAddress, getLogsPath)

	queryRangeEndpointUrl, err := url.Parse(queryRangeEndpointUrlStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred parsing url string '%v'", queryRangeEndpointUrlStr)
	}

	queryRangeEndpointQuery := queryRangeEndpointUrl.Query()

	queryRangeEndpointQuery.Set(startTimeQueryParamKey, maxRetentionLogsTimeParamValue)
	queryRangeEndpointQuery.Set(queryLogsQueryParamKey, queryParamValue)
	queryRangeEndpointQuery.Set(entriesLimitQueryParamKey, defaultEntriesLimit)
	queryRangeEndpointQuery.Set(directionQueryParamKey, defaultDirection)

	queryRangeEndpointUrl.RawQuery = queryRangeEndpointQuery.Encode()

	httpRequest := &http.Request{
		Method: http.MethodGet,
		URL:    queryRangeEndpointUrl,
		Header: httpHeaderWithTenantID,
	}
	httpRequestWithContext := httpRequest.WithContext(ctx)

	httpResponseBodyBytes, err := client.doHttpRequestWithRetries(httpRequestWithContext)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred doing http request '%+v'", httpRequestWithContext)
	}

	lokiQueryRangeResponse := &LokiQueryRangeResponse{}
	if err = json.Unmarshal(httpResponseBodyBytes, lokiQueryRangeResponse); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred unmarshalling the loki query range response")
	}

	if lokiQueryRangeResponse == nil || lokiQueryRangeResponse.Data == nil {
		return nil, stacktrace.Propagate(err, "The response body's schema payload received '%+v' by calling the Loki's query range endpoint is not what was expected; this is a bug in Kurtosis", lokiQueryRangeResponse)
	}

	for _, queryRangeResult := range lokiQueryRangeResponse.Data.Result {
		resultKurtosisGuidStr := queryRangeResult.Stream.KurtosisGUID
		resultKurtosisGuid := service.ServiceGUID(resultKurtosisGuidStr)
		resultKurtosisGuidLogLinesStr := make([]string, len(queryRangeResult.Values))
		for queryRangeIndex, queryRangeValue := range queryRangeResult.Values {
			logLineStr, err := getLogLineStrFromStreamValue(queryRangeValue)
			if err != nil {
				return nil, stacktrace.Propagate(err, "An error occurred getting log line string from stream value '%+v'", queryRangeValue)
			}
			resultKurtosisGuidLogLinesStr[queryRangeIndex] = logLineStr
		}
		resultLogsByKurtosisUserServiceGuid[resultKurtosisGuid] = resultKurtosisGuidLogLinesStr
	}


	return resultLogsByKurtosisUserServiceGuid, nil
}

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================
func (client *LokiLogsDatabaseClient) doHttpRequestWithRetries(request *http.Request) ([]byte, error)  {

	var (
		httpResponse          *http.Response
		httpResponseBodyBytes []byte
		err                   error
	)

	for _, retryBackoff := range httpRetriesBackoffSchedule {

		httpResponse, httpResponseBodyBytes, err = client.doHttpRequest(request)
		if err != nil {
			logrus.Debugf("Doing http request '%+v' returned with the following error: %v", request, err.Error())
		}

		if httpResponse != nil {
			logrus.Debugf("Doing http request '%+v' returned with body '%v' and '%v' status code", request, string(httpResponseBodyBytes), httpResponse.StatusCode)

			if httpResponse.StatusCode == http.StatusOK {
				return httpResponseBodyBytes, nil
			}

			//Do not retry if the status code indicate problems in the client side
			if httpResponse.StatusCode > http.StatusBadRequest && httpResponse.StatusCode < http.StatusInternalServerError {
				return nil, stacktrace.NewError("Executing the http request '%+v' returned not valid status code '%v'", request, httpResponse.StatusCode)
			}
		}

		logrus.Debugf("Retrying request in '%v'", retryBackoff)
		time.Sleep(retryBackoff)
	}
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred doing http request '%+v' even after applying this retry backoff schedule '%+v'", request, httpRetriesBackoffSchedule)
	}

	return nil, stacktrace.NewError("The request '%+v' could not be executed successfully, even after applying this retry backoff schedule '%+v', the status code '%v' and body '%v' were the last one received", request, httpRetriesBackoffSchedule, httpResponse.StatusCode, string(httpResponseBodyBytes))
}


func (client *LokiLogsDatabaseClient) doHttpRequest(
	request *http.Request,
) (
	resultResponse *http.Response,
	resultResponseBodyBytes []byte,
	resultErr error,
)  {

	var (
		httpResponseBodyBytes []byte
		err error
	)

	httpResponse, err := client.httpClient.Do(request)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred doing http request '%+v'", request)
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode == http.StatusOK {
		httpResponseBodyBytes, err = io.ReadAll(httpResponse.Body)
		if err != nil {
			return nil, nil, stacktrace.Propagate(err,
				"An error occurred reading the response body from '%v'", request.URL)
		}
	}

	return httpResponse, httpResponseBodyBytes, nil
}

func getMaxRetentionLogsTimeParamValue() string {
	now := time.Now()
	maxRetentionLogsTime := now.Add(-maxRetentionPeriod)
	maxRetentionLogsTimeNano := maxRetentionLogsTime.UnixNano()
	maxRetentionLogsTimeNanoStr := fmt.Sprintf("%v", maxRetentionLogsTimeNano)
	return maxRetentionLogsTimeNanoStr
}

func getQueryParamValue(
	kurtosisContainerType string,
	kurtosisGuids []string,
) string {
	kurtosisGuidParaValues := getKurtosisGuidParaValues(kurtosisGuids)

	queryParamWithKurtosisTagsQueryValue := fmt.Sprintf(
		`{%v="%v",%v=~"%v"}`,
		kurtosisContainerTypeLokiTagKey,
		kurtosisContainerType,
		kurtosisGuidLokiTagKey,
		kurtosisGuidParaValues,
	)
	return queryParamWithKurtosisTagsQueryValue
}

func getKurtosisGuidParaValues(kurtosisGuids []string) string {
	kurtosisGuidsParamValues := strings.Join(kurtosisGuids, orTagsOperator)
	return kurtosisGuidsParamValues
}

func getLogLineStrFromStreamValue(streamValue []string) (string, error) {
	if len(streamValue) > streamValueNumOfItems {
		return "", stacktrace.NewError("The stream value '%+v' should contains only 2 items but '%v' items were found, this should never happen; this is a bug in Kurtosis", streamValue, len(streamValue))
	}

	logLineStr := streamValue[streamValueLogLineIndex]
	logLineBytes := []byte(logLineStr)
	logLineObj := &LogLine{}

	if err := json.Unmarshal(logLineBytes, logLineObj); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred unmarshalling logline '%+v'", logLineObj)
	}

	return logLineObj.Log, nil
}
