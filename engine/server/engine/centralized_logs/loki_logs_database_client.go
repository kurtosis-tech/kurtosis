package centralized_logs

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/logs_database_functions/implementations/loki"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/logs_database_functions/implementations/loki/tags"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultHttpClientTimeOut = 1 * time.Minute

	httpScheme                = "http"
	websocketScheme           = "ws"
	baseLokiApiPath           = "/loki/api/v1"
	queryRangeEndpointSubpath = "/query_range"
	tailEndpointSubpath       = "/tail"

	//the global retention period store logs for 30 days = 720h.
	maxRetentionPeriodHours = loki.LimitsRetentionPeriodHours * time.Hour


	//We use this header because we are using the Loki multi-tenancy feature to split logs by the EnclaveID
	//Read more about it here: https://grafana.com/docs/loki/latest/operations/multi-tenancy/
	//tenantID = enclaveID
	organizationIdHttpHeaderKey = "X-Scope-OrgID"

	startTimeQueryParamKey    = "start"
	queryLogsQueryParamKey    = "query"
	entriesLimitQueryParamKey = "limit"
	directionQueryParamKey    = "direction"
	delayForQueryParamKey     = "delay_for"

	//We are establishing a big limit in order to get all the user-service's logs in one request
	//We will improve this in the future generating a pagination mechanism based on the limit number
	//and the unix epoch time (as the start time for the next request) returned by newest stream's log line
	//We chose 4k because 4000 x 1kb lines = 4 MB, which is just under the Protobuf limit of 5MB
	defaultEntriesLimit = "4000"
	//For tailing logs a lower entries limit is established because the websocket endpoint will be constantly streaming the new lines
	defaultEntriesLimitForTailingLogs = "100"
	//The oldest item is first when using direction=forward
	defaultDirection = "forward"

	//The number of seconds to delay retrieving logs to let slow loggers catch up. Defaults to 0 and cannot be larger than 5.
	defaultDelayForSeconds = "0"

	disjunctionTagOperator = "|"

	//A stream value should contain 2 items, the timestamp as the first one, and the log line
	//More here: https://grafana.com/docs/loki/latest/api/
	streamValueNumOfItems = 2

	streamValueLogLineIndex = 1

	//Left the connection open from the server-side for 4 days
	maxAllowedWebsocketConnectionDurationOnServerSide = loki.TailMaxDurationHours * time.Hour

	oneHourLess = -time.Hour

	logsByKurtosisUserServiceGuidChanBufferSize = 2
	errorChanBufferSize = 2
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

//fields are public because it's needed for JSON decoding
type lokiQueryRangeResponse struct {
	Status string `json:"status"`
	Data   *struct {
		ResultType string            `json:"resultType"`
		Result     []lokiStreamValue `json:"result"`
	} `json:"data"`
}

type lokiStreamLogsResponse struct {
	Streams []lokiStreamValue `json:"streams"`
}

type logLine struct {
	Log string `json:"log"`
}

type lokiLogsDatabaseClient struct {
	logsDatabaseAddress string
	httpClient          httpClient
}
type lokiStreamValue struct {
	Stream struct {
		KurtosisContainerType string `json:"comKurtosistechContainerType"`
		KurtosisGUID          string `json:"comKurtosistechGuid"`
	} `json:"stream"`
	Values [][]string `json:"values"`
}

func NewLokiLogsDatabaseClient(logsDatabaseAddress string, httpClient httpClient) *lokiLogsDatabaseClient {
	return &lokiLogsDatabaseClient{logsDatabaseAddress: logsDatabaseAddress, httpClient: httpClient}
}

func NewLokiLogsDatabaseClientWithDefaultHttpClient(logsDatabaseAddress string) *lokiLogsDatabaseClient {
	httpClientObj := &http.Client{Timeout: defaultHttpClientTimeOut}
	return &lokiLogsDatabaseClient{logsDatabaseAddress: logsDatabaseAddress, httpClient: httpClientObj}
}

func (client *lokiLogsDatabaseClient) GetUserServiceLogs(
	ctx context.Context,
	enclaveID enclave.EnclaveID,
	userServiceGUIDs map[service.ServiceGUID]bool,
) (
	map[service.ServiceGUID][]string,
	error,
) {

	httpHeaderWithTenantID := http.Header{}
	httpHeaderWithTenantID.Add(organizationIdHttpHeaderKey, string(enclaveID))

	kurtosisGuids := []string{}
	for userServiceGuid := range userServiceGUIDs {
		kurtosisGuids = append(kurtosisGuids, string(userServiceGuid))
	}

	maxRetentionLogsTimeParamValue := getMaxRetentionLogsTimeParamValue()

	userServiceContainerTypeDockerValue := label_value_consts.UserServiceContainerTypeDockerLabelValue.GetString()

	queryParamValue := getQueryParamValue(userServiceContainerTypeDockerValue, kurtosisGuids)

	getLogsPath := baseLokiApiPath + queryRangeEndpointSubpath

	queryRangeEndpointUrl := &url.URL{Scheme: httpScheme, Host: client.logsDatabaseAddress, Path: getLogsPath}

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
		return nil, stacktrace.Propagate(err, "An error occurred doing HTTP request '%+v'", httpRequestWithContext)
	}

	lokiQueryRangeResponseObj := &lokiQueryRangeResponse{}
	if err = json.Unmarshal(httpResponseBodyBytes, lokiQueryRangeResponseObj); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred unmarshalling the Loki query range response")
	}

	if lokiQueryRangeResponseObj == nil || lokiQueryRangeResponseObj.Data == nil {
		return nil, stacktrace.Propagate(err, "The response body's schema payload received '%+v' by calling the Loki's query range endpoint is not what was expected; this is a bug in Kurtosis", lokiQueryRangeResponseObj)
	}

	lokiStreams := lokiQueryRangeResponseObj.Data.Result

	resultLogsByKurtosisUserServiceGuid, err := newUserServiceLogLinesByUserServiceGuidFromLokiStreams(lokiStreams)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting user service log lines from loki streams '%+v'", lokiStreams)
	}

	return resultLogsByKurtosisUserServiceGuid, nil
}

func (client *lokiLogsDatabaseClient) StreamUserServiceLogs(
	ctx context.Context,
	enclaveID enclave.EnclaveID,
	userServiceGUIDs map[service.ServiceGUID]bool,
) (
	chan map[service.ServiceGUID][]string,
	chan error,
	error,
) {

	tailLogsEndpointURL, httpHeaderWithTenantID := client.getTailLogEndpointURLAndHeader(enclaveID, userServiceGUIDs)

	websocketDeadlineTime := getWebsocketDeadlineTime()

	websocketCallCtxWithDeadline, websocketCallCtxCancelFunc := context.WithDeadline(ctx, websocketDeadlineTime)
	//We need to cancel the websocket connection only if something fails because we need the connection open after returning
	shouldCancelWebsocketCall := false
	defer func() {
		if shouldCancelWebsocketCall {
			websocketCallCtxCancelFunc()
		}
	}()

	logsByKurtosisUserServiceGuidChan := make(chan map[service.ServiceGUID][]string, logsByKurtosisUserServiceGuidChanBufferSize)
	errChan := make(chan error, errorChanBufferSize)

	tailLogsWebsocketConn, tailLogsHttpResponse, err := websocket.DefaultDialer.DialContext(websocketCallCtxWithDeadline, tailLogsEndpointURL.String(), httpHeaderWithTenantID)
	if err != nil {
		shouldCancelWebsocketCall = true
		errMsg := fmt.Sprintf("An error occurred calling the logs-database-tail-logs-endpoint with URL '%v' using headers '%+v'", tailLogsEndpointURL.String(), httpHeaderWithTenantID)
		if tailLogsHttpResponse != nil {
			errMsg = fmt.Sprintf("%v; the '%v' status code was received from the server", errMsg, tailLogsHttpResponse.StatusCode)
		}
		return nil, nil, stacktrace.Propagate(err, errMsg)
	}

	if err = tailLogsWebsocketConn.SetReadDeadline(websocketDeadlineTime); err != nil {
		shouldCancelWebsocketCall = true
		return nil, nil, stacktrace.Propagate(err, "An error occurred setting the websocket read deadline time '%v'", websocketDeadlineTime.String())
	}

	go readStreamResponseAndAddUserServiceLogLinesToUserServiceLogsChannel(
		websocketCallCtxWithDeadline,
		websocketCallCtxCancelFunc,
		tailLogsWebsocketConn,
		logsByKurtosisUserServiceGuidChan,
		errChan,
	)

	return logsByKurtosisUserServiceGuidChan, errChan, nil
}

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================
func readStreamResponseAndAddUserServiceLogLinesToUserServiceLogsChannel(
	websocketCallCtxWithDeadline context.Context,
	websocketCallCtxCancelFunc context.CancelFunc,
	tailLogsWebsocketConn *websocket.Conn,
	logsByKurtosisUserServiceGuidChan chan map[service.ServiceGUID][]string,
	errChan chan error,
) {

	defer func() {
		//Closing open resources at the end
		websocketCallCtxCancelFunc()
		if err := tailLogsWebsocketConn.Close(); err != nil {
			logrus.Warnf("We tried to close the tail logs websocket connection, but doing so threw an error:\n%v", err)
		}
		close(logsByKurtosisUserServiceGuidChan)
		close(errChan)
	}()

	for {
		select {

		case <-websocketCallCtxWithDeadline.Done():
			errChan <- stacktrace.Propagate(websocketCallCtxWithDeadline.Err(), "An error occurred streaming user service logs from the Loki logs database client, the stream context has done")
			return

		default:
			streamResponse := &lokiStreamLogsResponse{}

			if err := tailLogsWebsocketConn.ReadJSON(streamResponse); err != nil {
				logrus.Debugf("Reading the tail logs streams response has retunerd the following error:\n'%v'", err)
				readTailLogsStreamsJsonErr := stacktrace.Propagate(err, "An error occurred reading the websocket endpoint")
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					logrus.Debugf("Reading the tail logs streams has reached the time out, error description:\n'%v'", err)
				}
				errChan <- readTailLogsStreamsJsonErr
				return
			}

			resultLogsByKurtosisUserServiceGuid, err := newUserServiceLogLinesByUserServiceGuidFromLokiStreams(streamResponse.Streams)
			if err != nil {
				errChan <-  stacktrace.Propagate(err, "An error occurred getting user service log lines from loki streams '%+v'", streamResponse.Streams)
				return
			}

			logsByKurtosisUserServiceGuidChan <- resultLogsByKurtosisUserServiceGuid
		}
	}
}

func (client *lokiLogsDatabaseClient) doHttpRequestWithRetries(request *http.Request) ([]byte, error) {

	var (
		httpResponse          *http.Response
		httpResponseBodyBytes []byte
		err                   error
	)

	for _, retryBackoff := range httpRetriesBackoffSchedule {

		httpResponse, httpResponseBodyBytes, err = client.doHttpRequest(request)
		if err != nil {
			logrus.Debugf("Doing HTTP request '%+v' returned with the following error: %v", request, err.Error())
		}

		if httpResponse != nil {
			logrus.Debugf("Doing HTTP request '%+v' returned with body '%v' and '%v' status code", request, string(httpResponseBodyBytes), httpResponse.StatusCode)

			if httpResponse.StatusCode == http.StatusOK {
				return httpResponseBodyBytes, nil
			}

			//Do not retry if the status code indicate problems in the client side
			if httpResponse.StatusCode > http.StatusBadRequest && httpResponse.StatusCode < http.StatusInternalServerError {
				return nil, stacktrace.NewError("Executing the HTTP request '%+v' returned not valid status code '%v'", request, httpResponse.StatusCode)
			}
		}

		logrus.Debugf("Retrying request in '%v'", retryBackoff)
		time.Sleep(retryBackoff)
	}
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred doing HTTP request '%+v' even after applying this retry backoff schedule '%+v'", request, httpRetriesBackoffSchedule)
	}

	return nil, stacktrace.NewError("The request '%+v' could not be executed successfully, even after applying this retry backoff schedule '%+v', the status code '%v' and body '%v' were the last one received", request, httpRetriesBackoffSchedule, httpResponse.StatusCode, string(httpResponseBodyBytes))
}

func (client *lokiLogsDatabaseClient) doHttpRequest(
	request *http.Request,
) (
	resultResponse *http.Response,
	resultResponseBodyBytes []byte,
	resultErr error,
) {

	var (
		httpResponseBodyBytes []byte
		err                   error
	)

	httpResponse, err := client.httpClient.Do(request)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred doing HTTP request '%+v'", request)
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
	maxRetentionLogsTime := now.Add(-maxRetentionPeriodHours)
	maxRetentionLogsTimeNano := maxRetentionLogsTime.UnixNano()
	maxRetentionLogsTimeNanoStr := fmt.Sprintf("%v", maxRetentionLogsTimeNano)
	return maxRetentionLogsTimeNanoStr
}

func getQueryParamValue(
	kurtosisContainerType string,
	kurtosisGuids []string,
) string {
	kurtosisGuidParaValues := getKurtosisGuidParamValues(kurtosisGuids)

	allLogsDatabaseKurtosisTrackedValidLokiTagsByDockerLabelKey := docker_kurtosis_backend.GetAllLogsDatabaseKurtosisTrackedValidLokiTagsByDockerLabelKey()

	kurtosisContainerTypeDockerLabelKey := label_key_consts.ContainerTypeDockerLabelKey

	kurtosisContainerTypeLokiTagKey := allLogsDatabaseKurtosisTrackedValidLokiTagsByDockerLabelKey[kurtosisContainerTypeDockerLabelKey]

	kurtosisGuidDockerLabelKey := label_key_consts.GUIDDockerLabelKey

	kurtosisGuidLokiTagKey := allLogsDatabaseKurtosisTrackedValidLokiTagsByDockerLabelKey[kurtosisGuidDockerLabelKey]

	queryParamWithKurtosisTagsQueryValue := fmt.Sprintf(
		`{%v="%v",%v=~"%v"}`,
		kurtosisContainerTypeLokiTagKey,
		kurtosisContainerType,
		kurtosisGuidLokiTagKey,
		kurtosisGuidParaValues,
	)
	return queryParamWithKurtosisTagsQueryValue
}

func getKurtosisGuidParamValues(kurtosisGuids []string) string {
	kurtosisGuidsParamValues := strings.Join(kurtosisGuids, disjunctionTagOperator)
	return kurtosisGuidsParamValues
}

func (client *lokiLogsDatabaseClient) getTailLogEndpointURLAndHeader(
	enclaveID enclave.EnclaveID,
	userServiceGuids map[service.ServiceGUID]bool,
) (url.URL, http.Header) {

	kurtosisGuids := []string{}
	for userServiceGuid := range userServiceGuids {
		kurtosisGuids = append(kurtosisGuids, string(userServiceGuid))
	}

	maxRetentionLogsTimeForTailingLogsParamValue := getStartTimeForStreamingLogsParaValue()

	userServiceContainerTypeDockerValue := label_value_consts.UserServiceContainerTypeDockerLabelValue.GetString()

	queryParamValue := getQueryParamValue(userServiceContainerTypeDockerValue, kurtosisGuids)

	tailLogsPath := baseLokiApiPath + tailEndpointSubpath

	tailLogsEndpointUrl := url.URL{Scheme: websocketScheme, Host: client.logsDatabaseAddress, Path: tailLogsPath}

	tailLogsEndpointQuery := tailLogsEndpointUrl.Query()

	tailLogsEndpointQuery.Set(queryLogsQueryParamKey, queryParamValue)
	tailLogsEndpointQuery.Set(delayForQueryParamKey, defaultDelayForSeconds)
	tailLogsEndpointQuery.Set(entriesLimitQueryParamKey, defaultEntriesLimitForTailingLogs)
	tailLogsEndpointQuery.Set(startTimeQueryParamKey, maxRetentionLogsTimeForTailingLogsParamValue)

	tailLogsEndpointUrl.RawQuery = tailLogsEndpointQuery.Encode()

	httpHeaderWithTenantID := http.Header{}
	httpHeaderWithTenantID.Add(organizationIdHttpHeaderKey, string(enclaveID))

	return tailLogsEndpointUrl, httpHeaderWithTenantID
}

func getStartTimeForStreamingLogsParaValue() string {
	now := time.Now()
	startTime := now.Add(oneHourLess)
	startTimeNano := startTime.UnixNano()
	startTimeNanoStr := fmt.Sprintf("%v", startTimeNano)
	return startTimeNanoStr
}

func newUserServiceLogLinesByUserServiceGuidFromLokiStreams(lokiStreams []lokiStreamValue) (map[service.ServiceGUID][]string, error) {

	resultLogsByKurtosisUserServiceGuid := map[service.ServiceGUID][]string{}

	for _, queryRangeResult := range lokiStreams {
		resultKurtosisGuidStr := queryRangeResult.Stream.KurtosisGUID
		resultKurtosisGuid := service.ServiceGUID(resultKurtosisGuidStr)
		resultKurtosisGuidLogLinesStr := make([]string, len(queryRangeResult.Values))
		for queryRangeIndex, queryRangeValue := range queryRangeResult.Values {
			logLineStr, err := newLogLineStrFromStreamValue(queryRangeValue)
			if err != nil {
				return nil, stacktrace.Propagate(err, "An error occurred getting log line string from stream value '%+v'", queryRangeValue)
			}
			resultKurtosisGuidLogLinesStr[queryRangeIndex] = logLineStr
		}
		resultLogsByKurtosisUserServiceGuid[resultKurtosisGuid] = resultKurtosisGuidLogLinesStr
	}

	return resultLogsByKurtosisUserServiceGuid, nil
}

func newLogLineStrFromStreamValue(streamValue []string) (string, error) {
	if len(streamValue) > streamValueNumOfItems {
		return "", stacktrace.NewError("The stream value '%+v' should contains only 2 items but '%v' items were found, this should never happen; this is a bug in Kurtosis", streamValue, len(streamValue))
	}

	logLineStr := streamValue[streamValueLogLineIndex]
	logLineBytes := []byte(logLineStr)
	logLineObj := &logLine{}

	if err := json.Unmarshal(logLineBytes, logLineObj); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred unmarshalling logline '%+v'", logLineObj)
	}

	return logLineObj.Log, nil
}

func getWebsocketDeadlineTime() time.Time {
	now := time.Now()
	deadlineTime := now.Add(maxAllowedWebsocketConnectionDurationOnServerSide)
	return deadlineTime
}