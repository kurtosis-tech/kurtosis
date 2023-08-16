package persistent_volume

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/logline"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"sync"
)

const (
	oneSenderAdded = 1

	// Location of logs on the filesystem of the engine
	logsFilepath = "/var/log/kurtosis/logs.json"

	newlineRune = '\n'

	serviceUUIDLogLabel = "container_id"
	enclaveUUIDLogLabel = "com.kurtosistech.enclave-id"
	logLabel            = "log"

	maxNumLogsToReturn = 200
)

type JsonLog map[string]string

// persistentVolumeLogsDatabaseClient pulls logs from a Docker volume the engine is mounted to
type persistentVolumeLogsDatabaseClient struct {
	kurtosisBackend backend_interface.KurtosisBackend
}

func NewPersistentVolumeLogsDatabaseClient(kurtosisBackend backend_interface.KurtosisBackend) *persistentVolumeLogsDatabaseClient {
	return &persistentVolumeLogsDatabaseClient{
		kurtosisBackend: kurtosisBackend,
	}
}

func (client *persistentVolumeLogsDatabaseClient) StreamUserServiceLogs(
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
	logrus.Debugf("ENCLAVE UUID: %v", enclaveUuid)
	logrus.Debugf("USER SERVICE UUIDS: %v", userServiceUuids)

	ctx, cancelCtxFunc := context.WithCancel(ctx)

	logsFile, err := os.Open(logsFilepath)
	if err != nil {
		cancelCtxFunc()
		return nil, nil, nil, stacktrace.Propagate(err, "An error occurred opening logs file.")
	}

	conjunctiveLogFiltersWithRegex, err := logline.NewConjunctiveLogFiltersWithRegex(conjunctiveLogLineFilters)
	if err != nil {
		cancelCtxFunc()
		return nil, nil, nil, stacktrace.Propagate(err, "An error occurred creating conjunctive log line filter with regex from filters '%+v'", conjunctiveLogLineFilters)
	}

	// this channel return an error if the stream fails at some point
	streamErrChan := make(chan error)

	// this channel will return the user service log lines by service UUID
	logsByKurtosisUserServiceUuidChan := make(chan map[service.ServiceUUID][]logline.LogLine)

	wgSenders := &sync.WaitGroup{}
	wgSenders.Add(oneSenderAdded)
	go streamServiceLogLines(
		ctx,
		wgSenders,
		logsByKurtosisUserServiceUuidChan,
		streamErrChan,
		enclaveUuid,
		userServiceUuids,
		logsFile,
		conjunctiveLogFiltersWithRegex,
		shouldFollowLogs,
	)

	// this go routine handles the stream cancellation
	go func() {
		//wait for stream go routine to end
		wgSenders.Wait()

		//close resources first
		if err := logsFile.Close(); err != nil {
			logrus.Warnf("An error occurred attempting to close the user service logs file after streaming:\n%v", err)
		}
		close(logsByKurtosisUserServiceUuidChan)
		close(streamErrChan)

		//then cancel the context
		cancelCtxFunc()
	}()

	return logsByKurtosisUserServiceUuidChan, streamErrChan, cancelCtxFunc, nil
}

func (client *persistentVolumeLogsDatabaseClient) FilterExistingServiceUuids(
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
func streamServiceLogLines(
	ctx context.Context,
	wgSenders *sync.WaitGroup,
	logsByKurtosisUserServiceUuidChan chan map[service.ServiceUUID][]logline.LogLine,
	streamErrChan chan error,
	enclaveUuid enclave.EnclaveUUID,
	userServiceUuids map[service.ServiceUUID]bool,
	logsFile io.Reader,
	conjunctiveLogLinesFiltersWithRegex []logline.LogLineFilterWithRegex,
	shouldFollowLogs bool,
) {
	defer wgSenders.Done()

	logsReader := bufio.NewReader(logsFile)

	numLogsReturned := 0
	for shouldFollowLogs || numLogsReturned < maxNumLogsToReturn {
		select {
		case <-ctx.Done():
			logrus.Debugf("Context was canceled, stopping streaming service logs for services '%v'", userServiceUuids)
			return
		default:
			jsonLogStr, err := logsReader.ReadString(newlineRune)
			if err != nil && errors.Is(err, io.EOF) {
				// exiting stream
				logrus.Debugf("EOF error returned when reading logs for services '%v'", userServiceUuids)
				return
			}
			if err != nil {
				streamErrChan <- stacktrace.Propagate(err, "An error occurred reading the logs file '%v'", userServiceUuids)
				return
			}

			// each logLineStr is of the following structure: {"label1": "...", "label2":"...", "log": "...", "timestamp", ... }
			// eg. {"container_type":"api-container","enclave_id":"ffd1c0ba29e1a464c","container_id":"8f8558ba",
			//	"container_name":"/kurtosis-api--ffd", "log":"hi","timestamp":"2023-08-14T14:57:49Z"}

			// First, we decode the line
			var jsonLog JsonLog
			err = json.Unmarshal([]byte(jsonLogStr), &jsonLog)
			if err != nil {
				streamErrChan <- stacktrace.Propagate(err, "An error occurred reading the logs file '%v'", userServiceUuids)
				return
			}

			// Then we extract the actual log message using the "log" field
			logLineStr, found := jsonLog[logLabel]
			if !found {
				streamErrChan <- stacktrace.NewError("An error retrieving the log field from logs json file. This is a bug in Kurtosis.")
				return
			}
			logLine := logline.NewLogLine(logLineStr)

			// We also extract enclave uuid and service uuid
			logEnclaveUuidStr, found := jsonLog[enclaveUUIDLogLabel]
			if !found {
				streamErrChan <- stacktrace.NewError("An error retrieving the enclave uuid field from logs json file. This is a bug in Kurtosis.")
				return
			}
			logServiceUuidStr, found := jsonLog[serviceUUIDLogLabel]
			if !found {
				streamErrChan <- stacktrace.NewError("An error retrieving the enclave uuid field from logs json file. This is a bug in Kurtosis.")
				return
			}

			logEnclaveUuid := enclave.EnclaveUUID(logEnclaveUuidStr)
			logServiceUuid := service.ServiceUUID(logServiceUuidStr)

			// Then we filter by checking:
			// 1. if the log message is valid based on requested filters
			// 2. if the log is associated with the requested enclave and one of the requested services
			//	  we check this bc currently all logs are in one file
			isValidBasedOnFilters, err := logLine.IsValidLogLineBaseOnFilters(conjunctiveLogLinesFiltersWithRegex)
			if err != nil {
				streamErrChan <- stacktrace.Propagate(err, "An error occurred filtering log line '%+v' using filters '%+v'", logLine, conjunctiveLogLinesFiltersWithRegex)
				break
			}
			isValidBasedOnEnclaveAndServiceUuid := (enclaveUuid == logEnclaveUuid) && (userServiceUuids[logServiceUuid])
			shouldReturnLogLine := isValidBasedOnFilters && isValidBasedOnEnclaveAndServiceUuid
			if !shouldReturnLogLine {
				break
			}

			// send the log line
			logLines := []logline.LogLine{*logLine}
			userServicesLogLinesMap := map[service.ServiceUUID][]logline.LogLine{
				logServiceUuid: logLines,
			}
			logsByKurtosisUserServiceUuidChan <- userServicesLogLinesMap
			numLogsReturned++
		}
	}
}
