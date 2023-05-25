package loki

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
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/logline"
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

	httpScheme                              = "http"
	websocketScheme                         = "ws"
	baseLokiApiPath                         = "/loki/api/v1"
	queryRangeEndpointSubpath               = "/query_range"
	queryListLabelValuesWithinRangeEndpoint = "/label/%s/values"
	tailEndpointSubpath                     = "/tail"

	lokiSuccessStatusInResponse = "success"
	kurtosisGuidLokiTagKey      = "comKurtosistechGuid"

	//the global retention period store logs for 30 days = 720h.
	maxRetentionPeriodHours = loki.LimitsRetentionPeriodHours * time.Hour

	//We use this header because we are using the Loki multi-tenancy feature to split logs by the EnclaveUUID
	//Read more about it here: https://grafana.com/docs/loki/latest/operations/multi-tenancy/
	//tenantID = enclaveUUID
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

	//It' for using in it the "start" query param
	maxRetentionPeriodHoursStartTime = -maxRetentionPeriodHours

	logsByKurtosisUserServiceUuidChanBuffSize = 5
	errorChanBuffSize                         = 2

	lokiEqualOperator        = "="
	lokiRegexMatchesOperator = "=~"
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

// fields are public because it's needed for JSON decoding
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

type lokiLabelValuesResponse struct {
	Status string   `json:"status"`
	Data   []string `json:"data"`
}

type LokiLogLine struct {
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
	httpClientObj := &http.Client{
		Transport:     nil,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       defaultHttpClientTimeOut,
	}
	return &lokiLogsDatabaseClient{logsDatabaseAddress: logsDatabaseAddress, httpClient: httpClientObj}
}

func (client *lokiLogsDatabaseClient) StreamUserServiceLogs(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	userServiceUuids map[service.ServiceUUID]bool,
	conjunctiveLogLineFilters logline.ConjunctiveLogLineFilters,
	shouldFollowLogs bool,
) (
	chan map[service.ServiceUUID][]logline.LogLine,
	chan error,
	context.CancelFunc,
	error,
) {

	var (
		serviceLogsByServiceUuidChan chan map[service.ServiceUUID][]logline.LogLine
		errChan                      chan error
		cancelCtxFunc                func()
		err                          error
	)

	lokiFilterLogsPipeline, err := newLokiLogFiltersPipelineFromConjunctiveLogLineFilters(conjunctiveLogLineFilters)
	if err != nil {
		return nil, nil, nil, stacktrace.Propagate(err, "An error occurred creating a new Loki filters log pipeline from these conjunctive log lines filters '%+v'", conjunctiveLogLineFilters)
	}

	if shouldFollowLogs {
		serviceLogsByServiceUuidChan, errChan, cancelCtxFunc, err = client.streamUserServiceLogs(ctx, enclaveUuid, userServiceUuids, lokiFilterLogsPipeline)
		if err != nil {
			return nil, nil, nil, stacktrace.Propagate(err, "An error occurred streaming service logs for UUIDs '%+v' in enclave with ID '%v'", userServiceUuids, enclaveUuid)
		}
	} else {
		serviceLogsByServiceUuidChan, cancelCtxFunc, err = client.getUserServiceLogs(ctx, enclaveUuid, userServiceUuids, lokiFilterLogsPipeline)
		if err != nil {
			return nil, nil, nil, stacktrace.Propagate(err, "An error occurred streaming service logs for UUIDs '%+v' in enclave with ID '%v'", userServiceUuids, enclaveUuid)
		}
	}

	return serviceLogsByServiceUuidChan, errChan, cancelCtxFunc, nil
}

func (client *lokiLogsDatabaseClient) FilterExistingServiceUuids(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	userServiceUuids map[service.ServiceUUID]bool,
) (map[service.ServiceUUID]bool, error) {
	httpHeaderWithTenantID := http.Header{}
	httpHeaderWithTenantID.Add(organizationIdHttpHeaderKey, string(enclaveUuid))

	getLogsPath := fmt.Sprintf(baseLokiApiPath+queryListLabelValuesWithinRangeEndpoint, kurtosisGuidLokiTagKey)

	queryRangeEndpointUrl := createUrl(httpScheme, client.logsDatabaseAddress, getLogsPath)

	queryRangeEndpointQuery := queryRangeEndpointUrl.Query()

	startTimeQueryParamValue := getStartTimeForFilteringExistingServiceUuidsParamValue()

	queryRangeEndpointQuery.Set(startTimeQueryParamKey, startTimeQueryParamValue)

	queryRangeEndpointUrl.RawQuery = queryRangeEndpointQuery.Encode()

	httpRequest := &http.Request{
		Method:           http.MethodGet,
		URL:              queryRangeEndpointUrl,
		Proto:            "",
		ProtoMajor:       0,
		ProtoMinor:       0,
		Header:           httpHeaderWithTenantID,
		Body:             nil,
		GetBody:          nil,
		ContentLength:    0,
		TransferEncoding: nil,
		Close:            false,
		Host:             "",
		Form:             nil,
		PostForm:         nil,
		MultipartForm:    nil,
		Trailer:          nil,
		RemoteAddr:       "",
		RequestURI:       "",
		TLS:              nil,
		Cancel:           nil,
		Response:         nil,
	}
	httpRequestWithContext := httpRequest.WithContext(ctx)

	httpResponseBodyBytes, err := client.doHttpRequestWithRetries(httpRequestWithContext)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred doing HTTP request '%+v'", httpRequestWithContext)
	}

	lokiLabelValuesResponseObj := &lokiLabelValuesResponse{
		Status: "",
		Data:   nil,
	}
	err = json.Unmarshal(httpResponseBodyBytes, lokiLabelValuesResponseObj)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred reading the existing service UUIDs from the logs database")
	}
	if lokiLabelValuesResponseObj.Status != lokiSuccessStatusInResponse {
		return nil, stacktrace.NewError("The logs database returns an error status when fetching the existing service UUIDs. Response was: \n%v", lokiLabelValuesResponseObj)
	}

	existingServiceUuidsSet := make(map[service.ServiceUUID]bool, len(lokiLabelValuesResponseObj.Data))
	for _, serviceUuid := range lokiLabelValuesResponseObj.Data {
		existingServiceUuidsSet[service.ServiceUUID(serviceUuid)] = true
	}
	filteredServiceUuidsSet := make(map[service.ServiceUUID]bool, len(lokiLabelValuesResponseObj.Data))
	for serviceUuid := range userServiceUuids {
		if _, found := existingServiceUuidsSet[serviceUuid]; found {
			filteredServiceUuidsSet[serviceUuid] = true
		}
	}
	return filteredServiceUuidsSet, nil
}

// ====================================================================================================
//
//	Private helper functions
//
// ====================================================================================================
func (client *lokiLogsDatabaseClient) getUserServiceLogs(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	userServiceUuids map[service.ServiceUUID]bool,
	lokiFilterLogsPipeline *lokiLogPipeline,
) (
	chan map[service.ServiceUUID][]logline.LogLine,
	context.CancelFunc,
	error,
) {

	httpHeaderWithTenantID := http.Header{}
	httpHeaderWithTenantID.Add(organizationIdHttpHeaderKey, string(enclaveUuid))

	kurtosisUuids := []string{}
	for userServiceUuid := range userServiceUuids {
		kurtosisUuids = append(kurtosisUuids, string(userServiceUuid))
	}

	maxRetentionLogsTimeParamValue := getMaxRetentionLogsTimeParamValue()

	userServiceContainerTypeDockerValue := label_value_consts.UserServiceContainerTypeDockerLabelValue.GetString()

	queryParamValue := getQueryParamValue(userServiceContainerTypeDockerValue, kurtosisUuids, lokiFilterLogsPipeline)

	getLogsPath := baseLokiApiPath + queryRangeEndpointSubpath

	queryRangeEndpointUrl := createUrl(httpScheme, client.logsDatabaseAddress, getLogsPath)

	queryRangeEndpointQuery := queryRangeEndpointUrl.Query()

	queryRangeEndpointQuery.Set(startTimeQueryParamKey, maxRetentionLogsTimeParamValue)
	queryRangeEndpointQuery.Set(queryLogsQueryParamKey, queryParamValue)
	queryRangeEndpointQuery.Set(entriesLimitQueryParamKey, defaultEntriesLimit)
	queryRangeEndpointQuery.Set(directionQueryParamKey, defaultDirection)

	queryRangeEndpointUrl.RawQuery = queryRangeEndpointQuery.Encode()

	httpRequest := &http.Request{
		Method:           http.MethodGet,
		URL:              queryRangeEndpointUrl,
		Proto:            "",
		ProtoMajor:       0,
		ProtoMinor:       0,
		Header:           httpHeaderWithTenantID,
		Body:             nil,
		GetBody:          nil,
		ContentLength:    0,
		TransferEncoding: nil,
		Close:            false,
		Host:             "",
		Form:             nil,
		PostForm:         nil,
		MultipartForm:    nil,
		Trailer:          nil,
		RemoteAddr:       "",
		RequestURI:       "",
		TLS:              nil,
		Cancel:           nil,
		Response:         nil,
	}

	ctxWithCancel, cancelCtxFunc := context.WithCancel(ctx)
	defer cancelCtxFunc()

	httpRequestWithContext := httpRequest.WithContext(ctxWithCancel)

	httpResponseBodyBytes, err := client.doHttpRequestWithRetries(httpRequestWithContext)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred doing HTTP request '%+v'", httpRequestWithContext)
	}

	lokiQueryRangeResponseObj := &lokiQueryRangeResponse{
		Status: "",
		Data:   nil,
	}
	if err = json.Unmarshal(httpResponseBodyBytes, lokiQueryRangeResponseObj); err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred unmarshalling the Loki query range response")
	}

	if lokiQueryRangeResponseObj.Status != lokiSuccessStatusInResponse {
		return nil, nil, stacktrace.NewError("The logs database return an error status when getting user service logs for service UUIDs '%+v'. Response was: \n%v", kurtosisUuids, lokiQueryRangeResponseObj)
	}

	if lokiQueryRangeResponseObj == nil || lokiQueryRangeResponseObj.Data == nil {
		return nil, nil, stacktrace.Propagate(err, "The response body's schema payload received '%+v' by calling the Loki's query range endpoint is not what was expected; this is a bug in Kurtosis", lokiQueryRangeResponseObj)
	}

	lokiStreams := lokiQueryRangeResponseObj.Data.Result

	//this channel will return the user service log lines by service GUI
	logsByKurtosisUserServiceUuidChan := make(chan map[service.ServiceUUID][]logline.LogLine, logsByKurtosisUserServiceUuidChanBuffSize)
	defer close(logsByKurtosisUserServiceUuidChan)

	resultLogsByKurtosisUserServiceUuid, err := newUserServiceLogLinesByUserServiceUuidFromLokiStreams(lokiStreams)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting user service log lines from loki streams '%+v'", lokiStreams)
	}

	logsByKurtosisUserServiceUuidChan <- resultLogsByKurtosisUserServiceUuid

	return logsByKurtosisUserServiceUuidChan, cancelCtxFunc, nil
}

func (client *lokiLogsDatabaseClient) streamUserServiceLogs(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	userServiceUuids map[service.ServiceUUID]bool,
	lokiFilterLogsPipeline *lokiLogPipeline,
) (
	chan map[service.ServiceUUID][]logline.LogLine,
	chan error,
	context.CancelFunc,
	error,
) {

	websocketDeadlineTime := getWebsocketDeadlineTime()

	ctxWithDeadline, cancelCtxFunc := context.WithDeadline(ctx, websocketDeadlineTime)
	shouldCancelCtx := true
	defer func() {
		if shouldCancelCtx {
			cancelCtxFunc()
		}
	}()

	tailLogsEndpointURL, httpHeaderWithTenantID := client.getTailLogEndpointURLAndHeader(enclaveUuid, userServiceUuids, lokiFilterLogsPipeline)

	//this channel will return the user service log lines by service UUID
	logsByKurtosisUserServiceUuidChan := make(chan map[service.ServiceUUID][]logline.LogLine, logsByKurtosisUserServiceUuidChanBuffSize)

	//this channel return an error if the stream fails at some point
	streamErrChan := make(chan error, errorChanBuffSize)

	tailLogsWebsocketConn, tailLogsHttpResponse, err := websocket.DefaultDialer.DialContext(ctxWithDeadline, tailLogsEndpointURL.String(), httpHeaderWithTenantID)
	if err != nil {
		errMsg := fmt.Sprintf("An error occurred calling the logs-database-tail-logs-endpoint with URL '%v' using headers '%+v'", tailLogsEndpointURL.String(), httpHeaderWithTenantID)
		if tailLogsHttpResponse != nil {
			errMsg = fmt.Sprintf("%v; the '%v' status code was received from the server", errMsg, tailLogsHttpResponse.StatusCode)
		}
		return nil, nil, nil, stacktrace.Propagate(err, errMsg)
	}

	go runReadStreamResponseAndAddUserServiceLogLinesToUserServiceLogsChannel(
		ctx,
		cancelCtxFunc,
		tailLogsWebsocketConn,
		logsByKurtosisUserServiceUuidChan,
		streamErrChan,
	)

	//We need to cancel the websocket connection only if something fails because we need the connection open after returning
	shouldCancelCtx = false
	return logsByKurtosisUserServiceUuidChan, streamErrChan, cancelCtxFunc, nil
}

func runReadStreamResponseAndAddUserServiceLogLinesToUserServiceLogsChannel(
	ctx context.Context,
	cancelCtxFunc context.CancelFunc,
	tailLogsWebsocketConn *websocket.Conn,
	logsByKurtosisUserServiceUuidChan chan map[service.ServiceUUID][]logline.LogLine,
	errChan chan error,
) {

	//Closing all the open resources at the end
	defer func() {
		if err := tailLogsWebsocketConn.Close(); err != nil {
			logrus.Warnf("We tried to close the tail logs websocket connection, but doing so threw an error:\n%v", err)
		}
		close(logsByKurtosisUserServiceUuidChan)
		close(errChan)
		cancelCtxFunc()
	}()

	for {

		streamResponse := &lokiStreamLogsResponse{
			Streams: nil,
		}

		if readingStreamResponseErr := tailLogsWebsocketConn.ReadJSON(streamResponse); readingStreamResponseErr != nil {

			logrus.Debugf("Reading the tail logs streams response has returned the following error:\n'%v'", readingStreamResponseErr)

			if websocket.IsCloseError(readingStreamResponseErr) {
				logrus.Debug("Reading the tail logs streams has been closed")
				return
			}

			if ctxErr := ctx.Err(); ctxErr != nil {
				switch ctx.Err() {
				case context.Canceled:
					logrus.Debug("The tail logs streams context has been canceled")
					return
				case context.DeadlineExceeded:
					logrus.Debug("The tail logs streams context deadline has been exceeded")
					deadlineErrMsg := "Reading the tail logs streams has exceeded the deadline time"
					ctxDeadlineTime, ok := ctx.Deadline()
					if ok {
						deadlineErrMsg = fmt.Sprintf("%v with value '%v'", deadlineErrMsg, ctxDeadlineTime)
					}
					errChan <- stacktrace.NewError(deadlineErrMsg)
					return
				default:
					logrus.Debugf("The tail logs streams context contains this error '%v' ", ctxErr)
				}
			}

			errChan <- stacktrace.Propagate(readingStreamResponseErr, "An error occurred reading the Loki's tail log endpoint")
			return
		}

		//Does the reading
		resultLogsByKurtosisUserServiceUuid, err := newUserServiceLogLinesByUserServiceUuidFromLokiStreams(streamResponse.Streams)
		if err != nil {
			errChan <- stacktrace.Propagate(err, "An error occurred getting user service log lines from loki streams '%+v'", streamResponse.Streams)
			return
		}
		logsByKurtosisUserServiceUuidChan <- resultLogsByKurtosisUserServiceUuid
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

func (client *lokiLogsDatabaseClient) getTailLogEndpointURLAndHeader(
	enclaveUuid enclave.EnclaveUUID,
	userServiceUuids map[service.ServiceUUID]bool,
	lokiFilterLogsPipeline *lokiLogPipeline,
) (url.URL, http.Header) {

	kurtosisUuids := []string{}
	for userServiceUuid := range userServiceUuids {
		kurtosisUuids = append(kurtosisUuids, string(userServiceUuid))
	}

	maxRetentionLogsTimeForTailingLogsParamValue := getStartTimeForStreamingLogsParamValue()

	userServiceContainerTypeDockerValue := label_value_consts.UserServiceContainerTypeDockerLabelValue.GetString()

	queryParamValue := getQueryParamValue(userServiceContainerTypeDockerValue, kurtosisUuids, lokiFilterLogsPipeline)

	tailLogsPath := baseLokiApiPath + tailEndpointSubpath

	tailLogsEndpointUrl := *createUrl(websocketScheme, client.logsDatabaseAddress, tailLogsPath)

	tailLogsEndpointQuery := tailLogsEndpointUrl.Query()

	tailLogsEndpointQuery.Set(queryLogsQueryParamKey, queryParamValue)
	tailLogsEndpointQuery.Set(delayForQueryParamKey, defaultDelayForSeconds)
	tailLogsEndpointQuery.Set(entriesLimitQueryParamKey, defaultEntriesLimitForTailingLogs)
	tailLogsEndpointQuery.Set(startTimeQueryParamKey, maxRetentionLogsTimeForTailingLogsParamValue)

	tailLogsEndpointUrl.RawQuery = tailLogsEndpointQuery.Encode()

	httpHeaderWithTenantID := http.Header{}
	httpHeaderWithTenantID.Add(organizationIdHttpHeaderKey, string(enclaveUuid))

	return tailLogsEndpointUrl, httpHeaderWithTenantID
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
	lokiFilterLogsPipeline *lokiLogPipeline,
) string {
	kurtosisGuidParaValues := getKurtosisGuidParamValues(kurtosisGuids)

	allLogsDatabaseKurtosisTrackedValidLokiTagsByDockerLabelKey := docker_kurtosis_backend.GetAllLogsDatabaseKurtosisTrackedValidLokiTagsByDockerLabelKey()

	kurtosisContainerTypeDockerLabelKey := label_key_consts.ContainerTypeDockerLabelKey

	kurtosisContainerTypeLokiTagKey := allLogsDatabaseKurtosisTrackedValidLokiTagsByDockerLabelKey[kurtosisContainerTypeDockerLabelKey]

	kurtosisGuidDockerLabelKey := label_key_consts.GUIDDockerLabelKey

	kurtosisGuidLokiTagKey := allLogsDatabaseKurtosisTrackedValidLokiTagsByDockerLabelKey[kurtosisGuidDockerLabelKey]

	streamSelectorInQuery := getLokiStreamSelectorStrWithContainerTypeAndGuidsTags(
		kurtosisContainerTypeLokiTagKey,
		kurtosisContainerType,
		kurtosisGuidLokiTagKey,
		kurtosisGuidParaValues,
	)

	var queryParamValue string
	if lokiFilterLogsPipeline != nil {
		queryParamValue = fmt.Sprintf("%s %s", streamSelectorInQuery, lokiFilterLogsPipeline.GetConjunctiveLogLineFiltersString())
	} else {
		queryParamValue = streamSelectorInQuery
	}

	return queryParamValue
}

// The stream selector determines which log streams to include in a query’s results.
// This stream selector will return log streams which contains the declared "container-type" in "kurtosisContainerType"
// and one of the "guids" defined in "kurtosisGuidParaValues"
func getLokiStreamSelectorStrWithContainerTypeAndGuidsTags(
	kurtosisContainerTypeLokiTagKey,
	kurtosisContainerType,
	kurtosisGuidLokiTagKey,
	kurtosisGuidParaValues string,
) string {
	streamSelectorInQuery := fmt.Sprintf(
		`{%v%v"%v",%v%v"%v"}`,
		kurtosisContainerTypeLokiTagKey,
		lokiEqualOperator,
		kurtosisContainerType,
		kurtosisGuidLokiTagKey,
		lokiRegexMatchesOperator,
		kurtosisGuidParaValues,
	)

	return streamSelectorInQuery
}

func getKurtosisGuidParamValues(kurtosisGuids []string) string {
	kurtosisGuidsParamValues := strings.Join(kurtosisGuids, disjunctionTagOperator)
	return kurtosisGuidsParamValues
}

func createUrl(scheme string, host string, path string) *url.URL {
	return &url.URL{ //nolint:exhaustruct
		Scheme:      scheme,
		Opaque:      "",
		User:        nil,
		Host:        host,
		Path:        path,
		RawPath:     "",
		ForceQuery:  false,
		RawQuery:    "",
		Fragment:    "",
		RawFragment: "",
	}
}

func getStartTimeForStreamingLogsParamValue() string {
	now := time.Now()
	startTime := now.Add(oneHourLess)
	startTimeNanoStr := getTimeInNanoString(startTime)
	return startTimeNanoStr
}

// Because the logs can be consumed from several days before the current date
// we have to set the start time from the max retention period time
func getStartTimeForFilteringExistingServiceUuidsParamValue() string {
	now := time.Now()
	startTime := now.Add(maxRetentionPeriodHoursStartTime)
	startTimeNanoStr := getTimeInNanoString(startTime)
	return startTimeNanoStr
}

func getTimeInNanoString(timeObj time.Time) string {
	timeNano := timeObj.UnixNano()
	timeNanoStr := fmt.Sprintf("%v", timeNano)
	return timeNanoStr
}

func newUserServiceLogLinesByUserServiceUuidFromLokiStreams(lokiStreams []lokiStreamValue) (map[service.ServiceUUID][]logline.LogLine, error) {

	resultLogsByKurtosisUserServiceUuid := map[service.ServiceUUID][]logline.LogLine{}

	for _, queryRangeResult := range lokiStreams {
		resultKurtosisUuidStr := queryRangeResult.Stream.KurtosisGUID
		resultKurtosisUuid := service.ServiceUUID(resultKurtosisUuidStr)
		resultKurtosisUuidLogLines := make([]logline.LogLine, len(queryRangeResult.Values))
		for queryRangeIndex, queryRangeValue := range queryRangeResult.Values {
			logLineObj, err := newLogLineFromStreamValue(queryRangeValue)
			if err != nil {
				return nil, stacktrace.Propagate(err, "An error occurred getting log line string from stream value '%+v'", queryRangeValue)
			}
			resultKurtosisUuidLogLines[queryRangeIndex] = *logLineObj
		}

		userServiceLogLines, found := resultLogsByKurtosisUserServiceUuid[resultKurtosisUuid]
		if found {
			userServiceLogLines = append(userServiceLogLines, resultKurtosisUuidLogLines...)
		} else {
			userServiceLogLines = resultKurtosisUuidLogLines
		}

		resultLogsByKurtosisUserServiceUuid[resultKurtosisUuid] = userServiceLogLines
	}

	return resultLogsByKurtosisUserServiceUuid, nil
}

func newLogLineFromStreamValue(streamValue []string) (*logline.LogLine, error) {
	if len(streamValue) > streamValueNumOfItems {
		return nil, stacktrace.NewError("The stream value '%+v' should contains only 2 items but '%v' items were found, this should never happen; this is a bug in Kurtosis", streamValue, len(streamValue))
	}

	lokiLogLineStr := streamValue[streamValueLogLineIndex]
	lokiLogLineBytes := []byte(lokiLogLineStr)
	lokiLogLine := &LokiLogLine{
		Log: "",
	}

	if err := json.Unmarshal(lokiLogLineBytes, lokiLogLine); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred unmarshalling Loki log line '%+v'", lokiLogLine)
	}

	newLogLineObj := logline.NewLogLine(lokiLogLine.Log)

	return newLogLineObj, nil
}

func getWebsocketDeadlineTime() time.Time {
	now := time.Now()
	deadlineTime := now.Add(maxAllowedWebsocketConnectionDurationOnServerSide)
	return deadlineTime
}

func newLokiLogFiltersPipelineFromConjunctiveLogLineFilters(
	conjunctiveLogLineFilters logline.ConjunctiveLogLineFilters,
) (*lokiLogPipeline, error) {

	var lokiLogLineFilters []LokiLineFilter

	for _, logLineFilter := range conjunctiveLogLineFilters {
		var lokiLogLineFilter *LokiLineFilter
		operator := logLineFilter.GetOperator()
		filterTextPattern := logLineFilter.GetTextPattern()
		switch operator {
		case logline.LogLineOperator_DoesContainText:
			lokiLogLineFilter = NewDoesContainTextLokiLineFilter(filterTextPattern)
		case logline.LogLineOperator_DoesNotContainText:
			lokiLogLineFilter = NewDoesNotContainTextLokiLineFilter(filterTextPattern)
		case logline.LogLineOperator_DoesContainMatchRegex:
			lokiLogLineFilter = NewDoesContainMatchRegexLokiLineFilter(filterTextPattern)
		case logline.LogLineOperator_DoesNotContainMatchRegex:
			lokiLogLineFilter = NewDoesNotContainMatchRegexLokiLineFilter(filterTextPattern)
		default:
			return nil, stacktrace.NewError("Unrecognized log line filter operator '%v' in filter '%v'; this is a bug in Kurtosis", operator, logLineFilter)
		}
		lokiLogLineFilters = append(lokiLogLineFilters, *lokiLogLineFilter)
	}

	logPipeline := NewLokiLogPipeline(lokiLogLineFilters)

	return logPipeline, nil
}
