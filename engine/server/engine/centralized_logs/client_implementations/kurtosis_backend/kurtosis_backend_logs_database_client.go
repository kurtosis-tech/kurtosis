package kurtosis_backend

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/logline"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"regexp"
	"strings"
	"sync"
)

const (
	oneSenderAdded = 1
	newlineRune    = '\n'
)

type kurtosisBackendLogsDatabaseClient struct {
	kurtosisBackend backend_interface.KurtosisBackend
}

func NewKurtosisBackendLogsDatabaseClient(kurtosisBackend backend_interface.KurtosisBackend) *kurtosisBackendLogsDatabaseClient {
	return &kurtosisBackendLogsDatabaseClient{
		kurtosisBackend: kurtosisBackend,
	}
}

func (client *kurtosisBackendLogsDatabaseClient) StreamUserServiceLogs(
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

	ctx, cancelCtxFunc := context.WithCancel(ctx)

	userServiceFilters := newUserServiceFilters(userServiceUuids)

	conjunctiveLogFiltersWithRegex, err := newConjunctiveLogFiltersWithRegex(conjunctiveLogLineFilters)
	if err != nil {
		cancelCtxFunc()
		return nil, nil, nil, stacktrace.Propagate(err, "An error occurred creating conjunctive log line filter with regex from filters '%+v'", conjunctiveLogLineFilters)
	}

	successfulUserServiceLogs, erroredUserServiceUuids, err := client.kurtosisBackend.GetUserServiceLogs(ctx, enclaveUuid, userServiceFilters, shouldFollowLogs)
	if err != nil {
		cancelCtxFunc()
		return nil, nil, nil, stacktrace.Propagate(
			err, "An error occurred getting user service logs using filters '%+v' on enclave with UUID '%v' "+
				"and with should follow logs value '%v'",
			userServiceFilters,
			enclaveUuid,
			shouldFollowLogs,
		)
	}

	if len(erroredUserServiceUuids) == len(userServiceUuids) && len(successfulUserServiceLogs) == 0 {
		cancelCtxFunc()
		var allServiceErrors []string
		for serviceUuid, serviceErr := range erroredUserServiceUuids {
			serviceErrStr := fmt.Sprintf("service UUID '%v' - error:%v", serviceUuid, serviceErr.Error())
			allServiceErrors = append(allServiceErrors, serviceErrStr)
		}
		allServiceErrorsStr := strings.Join(allServiceErrors, "\n")
		return nil, nil, nil, stacktrace.NewError("All the requested services with UUIDs '%+v' returned errors when calling for logs. Errors:\n%v", userServiceUuids, allServiceErrorsStr)
	}

	//this channel return an error if the stream fails at some point
	streamErrChan := make(chan error)

	for serviceUuid, serviceErr := range erroredUserServiceUuids {
		streamErrChan <- stacktrace.Propagate(serviceErr, "An error occurred getting user service logs for user service with UUID '%v'", serviceUuid)
	}

	wgSenders := &sync.WaitGroup{}

	//this channel will return the user service log lines by service UUID
	logsByKurtosisUserServiceUuidChan := make(chan map[service.ServiceUUID][]logline.LogLine)

	for serviceUuid, serviceReadCloser := range successfulUserServiceLogs {
		wgSenders.Add(oneSenderAdded)
		go streamServiceLogLines(
			ctx,
			wgSenders,
			logsByKurtosisUserServiceUuidChan,
			streamErrChan,
			serviceUuid,
			serviceReadCloser,
			conjunctiveLogFiltersWithRegex,
		)
	}

	//this go routine handles the stream cancellation
	go func() {
		//wait for all senders' end
		wgSenders.Wait()

		//close resources first
		for _, userServiceLogsReadCloser := range successfulUserServiceLogs {
			if err := userServiceLogsReadCloser.Close(); err != nil {
				logrus.Warnf("We tried to close the user service logs read-closer-objects after we're done using it, but doing so threw an error:\n%v", err)
			}
		}
		close(logsByKurtosisUserServiceUuidChan)
		close(streamErrChan)

		//then cancel the context
		cancelCtxFunc()
	}()

	return logsByKurtosisUserServiceUuidChan, streamErrChan, cancelCtxFunc, nil
}

func (client *kurtosisBackendLogsDatabaseClient) FilterExistingServiceUuids(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	userServiceUuids map[service.ServiceUUID]bool,
) (map[service.ServiceUUID]bool, error) {

	userServiceFilters := &service.ServiceFilters{
		Names:    nil,
		UUIDs:    userServiceUuids,
		Statuses: nil,
	}

	existingServicesByUuids, err := client.kurtosisBackend.GetUserServices(ctx, enclaveUuid, userServiceFilters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting user services for enclave with UUID '%v' and using filters '%+v'", enclaveUuid, userServiceFilters)
	}

	filteredServiceUuidsSet := map[service.ServiceUUID]bool{}
	for serviceUuid := range userServiceUuids {
		if _, found := existingServicesByUuids[serviceUuid]; found {
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
func newUserServiceFilters(userServiceGuids map[service.ServiceUUID]bool) *service.ServiceFilters {
	userServiceFilters := &service.ServiceFilters{
		Names:    nil,
		UUIDs:    userServiceGuids,
		Statuses: nil,
	}
	return userServiceFilters
}

func streamServiceLogLines(
	ctx context.Context,
	wgSenders *sync.WaitGroup,
	logsByKurtosisUserServiceUuidChan chan map[service.ServiceUUID][]logline.LogLine,
	streamErrChan chan error,
	serviceUuid service.ServiceUUID,
	userServiceReadCloserLog io.ReadCloser,
	conjunctiveLogLinesFiltersWithRegex []LogLineFilterWithRegex,
) {
	defer wgSenders.Done()

	logsReader := bufio.NewReader(userServiceReadCloserLog)

	for {
		select {
		//client cancel ctx case
		case <-ctx.Done():
			logrus.Debugf("Context was canceled, stopping streaming service logs for service '%v'", serviceUuid)
			return
		default:
			//getting a new single log line
			logLineStr, err := logsReader.ReadString(newlineRune)
			if err != nil && errors.Is(err, io.EOF) {
				//exiting stream
				logrus.Debugf("EOF error returned when reading logs for service '%v'", serviceUuid)
				return
			}
			if err != nil {
				streamErrChan <- stacktrace.Propagate(err, "An error occurred reading the user service read closer logs for service with UUID '%v'", serviceUuid)
				return
			}

			logLine := logline.NewLogLine(logLineStr)

			//filtering it
			shouldReturnLogLine, err := shouldReturnLogLineBaseOnFilters(logLine, conjunctiveLogLinesFiltersWithRegex)
			if err != nil {
				streamErrChan <- stacktrace.Propagate(err, "An error occurred filtering log line '%+v' using filters '%+v'", logLine, conjunctiveLogLinesFiltersWithRegex)
				break
			}
			if !shouldReturnLogLine {
				break
			}

			//send the log line
			logLines := []logline.LogLine{*logLine}
			userServicesLogLinesMap := map[service.ServiceUUID][]logline.LogLine{
				serviceUuid: logLines,
			}
			logsByKurtosisUserServiceUuidChan <- userServicesLogLinesMap
		}
	}
}

func shouldReturnLogLineBaseOnFilters(
	logLine *logline.LogLine,
	conjunctiveLogLinesFiltersWithRegex []LogLineFilterWithRegex,
) (bool, error) {

	shouldReturnIt := true

	for _, logLineFilter := range conjunctiveLogLinesFiltersWithRegex {
		operator := logLineFilter.GetOperator()

		logLineContent := logLine.GetContent()
		logLineContentLowerCase := strings.ToLower(logLineContent)
		textPatternLowerCase := strings.ToLower(logLineFilter.GetTextPattern())

		switch operator {
		case logline.LogLineOperator_DoesContainText:
			if !strings.Contains(logLineContentLowerCase, textPatternLowerCase) {
				shouldReturnIt = false
			}
		case logline.LogLineOperator_DoesNotContainText:
			if strings.Contains(logLineContentLowerCase, textPatternLowerCase) {
				shouldReturnIt = false
			}
		case logline.LogLineOperator_DoesContainMatchRegex:
			if !logLineFilter.compiledRegexPattern.MatchString(logLineContent) {
				shouldReturnIt = false
			}
		case logline.LogLineOperator_DoesNotContainMatchRegex:
			if logLineFilter.compiledRegexPattern.MatchString(logLineContent) {
				shouldReturnIt = false
			}
		default:
			return false, stacktrace.NewError("Unrecognized log line filter operator '%v' in filter '%v'; this is a bug in Kurtosis", operator, logLineFilter)
		}
		if !shouldReturnIt {
			break
		}
	}

	return shouldReturnIt, nil
}

func newConjunctiveLogFiltersWithRegex(conjunctiveLogLineFilters logline.ConjunctiveLogLineFilters) ([]LogLineFilterWithRegex, error) {
	conjunctiveLogFiltersWithRegex := []LogLineFilterWithRegex{}
	for _, logLineFilter := range conjunctiveLogLineFilters {
		logLineFilterWithRegex := newLogLineFilterWithRegex(logLineFilter, nil)

		if logLineFilter.IsRegexFilter() {
			filterRegexPattern := logLineFilter.GetTextPattern()
			logLineRegexPattern, err := regexp.Compile(filterRegexPattern)
			if err != nil {
				return nil, stacktrace.Propagate(err, "An error occurred compiling regex string '%v' for log line filter '%+v'", filterRegexPattern, logLineFilter)
			}
			logLineFilterWithRegex.compiledRegexPattern = logLineRegexPattern
		}
		conjunctiveLogFiltersWithRegex = append(conjunctiveLogFiltersWithRegex, *logLineFilterWithRegex)
	}

	return conjunctiveLogFiltersWithRegex, nil
}
