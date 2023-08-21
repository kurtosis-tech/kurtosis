package persistent_volume

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/logline"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"sync"
)

const (
	oneSenderAdded = 1

	// Location of logs on the filesystem of the engine
	logsStorageDirpath = "/var/log/kurtosis/"
	filetype           = ".json"

	newlineRune = '\n'

	logLabel = "log"

	maxNumLogsToReturn = 200
)

type JsonLog map[string]string

// persistentVolumeLogsDatabaseClient pulls logs from a Docker volume the engine is mounted to
type persistentVolumeLogsDatabaseClient struct {
	kurtosisBackend backend_interface.KurtosisBackend

	filesystem VolumeFilesystem
}

func NewPersistentVolumeLogsDatabaseClient(kurtosisBackend backend_interface.KurtosisBackend, filesystem VolumeFilesystem) *persistentVolumeLogsDatabaseClient {
	return &persistentVolumeLogsDatabaseClient{
		kurtosisBackend: kurtosisBackend,
		filesystem:      filesystem,
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
	ctx, cancelCtxFunc := context.WithCancel(ctx)

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
	for serviceUuid := range userServiceUuids {
		wgSenders.Add(oneSenderAdded)
		go streamServiceLogLines(
			ctx,
			client.filesystem,
			wgSenders,
			logsByKurtosisUserServiceUuidChan,
			streamErrChan,
			enclaveUuid,
			serviceUuid,
			conjunctiveLogFiltersWithRegex,
			shouldFollowLogs,
		)
	}

	// this go routine handles the stream cancellation
	go func() {
		//wait for stream go routine to end
		wgSenders.Wait()

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
	fs VolumeFilesystem,
	wgSenders *sync.WaitGroup,
	logsByKurtosisUserServiceUuidChan chan map[service.ServiceUUID][]logline.LogLine,
	streamErrChan chan error,
	enclaveUuid enclave.EnclaveUUID,
	serviceUuid service.ServiceUUID,
	conjunctiveLogLinesFiltersWithRegex []logline.LogLineFilterWithRegex,
	shouldFollowLogs bool,
) {
	defer wgSenders.Done()

	// logs are stored per enclave id, per service uuid, eg. <base path>/123440231421/54325342w2341.json
	logsFilepath := fmt.Sprintf("%s%s/%s%s", logsStorageDirpath, string(enclaveUuid), string(serviceUuid), filetype)
	logsFile, err := fs.Open(logsFilepath)
	if err != nil {
		streamErrChan <- stacktrace.Propagate(err, "An error occurred opening the logs file for service '%v' in enclave '%v' at the following path: %v.", serviceUuid, enclaveUuid, logsFilepath)
		return
	}
	logsReader := bufio.NewReader(logsFile)

	numLogsReturned := 0
	for shouldFollowLogs || numLogsReturned < maxNumLogsToReturn {
		select {
		case <-ctx.Done():
			logrus.Debugf("Context was canceled, stopping streaming service logs for service '%v' in enclave '%v", serviceUuid, enclaveUuid)
			return
		default:
			jsonLogStr, err := logsReader.ReadString(newlineRune)
			if err != nil && errors.Is(err, io.EOF) {
				// exiting stream
				logrus.Debugf("EOF error returned when reading logs for service '%v' in enclave '%v'", serviceUuid, enclaveUuid)
				return
			}
			if err != nil {
				streamErrChan <- stacktrace.Propagate(err, "An error occurred reading the logs file for service '%v' in enclave '%v' at the following path: %v.", serviceUuid, enclaveUuid, logsFilepath)
				return
			}

			// each logLineStr is of the following structure: {"enclave_uuid": "...", "service_uuid":"...", "log": "...",.. "timestamp":"..."}
			// eg. {"container_type":"api-container", "container_id":"8f8558ba", "container_name":"/kurtosis-api--ffd",
			// "log":"hi","timestamp":"2023-08-14T14:57:49Z"}

			// First, we decode the line
			var jsonLog JsonLog
			err = json.Unmarshal([]byte(jsonLogStr), &jsonLog)
			if err != nil {
				streamErrChan <- stacktrace.Propagate(err, "An error occurred parsing the json logs file for service '%v' in enclave '%v' at the following path: %v.", serviceUuid, enclaveUuid, logsFilepath)
				return
			}

			// Then we extract the actual log message using the "log" field
			logLineStr, found := jsonLog[logLabel]
			if !found {
				streamErrChan <- stacktrace.NewError("An error retrieving the log field from logs json file. This is a bug in Kurtosis.")
				return
			}
			logLine := logline.NewLogLine(logLineStr)

			// Then we filter by checking if the log message is valid based on requested filters
			shouldReturnLogLine, err := logLine.IsValidLogLineBaseOnFilters(conjunctiveLogLinesFiltersWithRegex)
			if err != nil {
				streamErrChan <- stacktrace.Propagate(err, "An error occurred filtering log line '%+v' using filters '%+v'", logLine, conjunctiveLogLinesFiltersWithRegex)
				break
			}
			if !shouldReturnLogLine {
				break
			}

			// send the log line
			logLines := []logline.LogLine{*logLine}
			userServicesLogLinesMap := map[service.ServiceUUID][]logline.LogLine{
				serviceUuid: logLines,
			}
			logsByKurtosisUserServiceUuidChan <- userServicesLogLinesMap
			numLogsReturned++
		}
	}
}
