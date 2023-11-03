package stream_logs_strategy

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_consts"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_filesystem"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/logline"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"strings"
)

// This strategy pulls logs from filesytsem where there is a log file per enclave, per service
type PerFileStreamLogsStrategy struct {
}

func NewPerFileStreamLogsStrategy() *PerFileStreamLogsStrategy {
	return &PerFileStreamLogsStrategy{}
}

type JsonLog map[string]string

func (strategy *PerFileStreamLogsStrategy) StreamLogs(
	ctx context.Context,
	fs volume_filesystem.VolumeFilesystem,
	logsByKurtosisUserServiceUuidChan chan map[service.ServiceUUID][]logline.LogLine,
	streamErrChan chan error,
	enclaveUuid enclave.EnclaveUUID,
	serviceUuid service.ServiceUUID,
	conjunctiveLogLinesFiltersWithRegex []logline.LogLineFilterWithRegex,
	shouldFollowLogs bool,
	shouldReturnAllLogs bool,
	numLogLines uint32,
) {
	// logs are stored per enclave id, per service uuid, eg. <base path>/123440231421/54325342w2341.json
	logsFilepath := fmt.Sprintf(volume_consts.PerFileFmtStr, volume_consts.LogsStorageDirpath, string(enclaveUuid), string(serviceUuid), volume_consts.Filetype)
	logsFile, err := fs.Open(logsFilepath)
	if err != nil {
		streamErrChan <- stacktrace.Propagate(err, "An error occurred opening the logs file for service '%v' in enclave '%v' at the following path: %v.", serviceUuid, enclaveUuid, logsFilepath)
		return
	}
	logsReader := bufio.NewReader(logsFile)

	for {
		select {
		case <-ctx.Done():
			logrus.Debugf("Context was canceled, stopping streaming service logs for service '%v' in enclave '%v", serviceUuid, enclaveUuid)
			return
		default:
			var jsonLogStr string
			var readErr error
			var jsonLogNewStr string

			for {
				jsonLogNewStr, readErr = logsReader.ReadString(volume_consts.NewLineRune)
				jsonLogStr = jsonLogStr + jsonLogNewStr
				// check if it's an uncompleted Json line
				if jsonLogNewStr != "" && len(jsonLogNewStr) > 2 {
					jsonLogNewStrLastChars := jsonLogNewStr[len(jsonLogNewStr)-2:]
					if jsonLogNewStrLastChars != volume_consts.EndOfJsonLine+"\n" {
						// removes the newline char from the previous part of the json line
						jsonLogStr = strings.TrimSuffix(jsonLogStr, string(volume_consts.NewLineRune))
						continue
					}
				}
				if readErr != nil && errors.Is(readErr, io.EOF) {
					if shouldFollowLogs {
						continue
					}
					// exiting stream
					logrus.Debugf("EOF error returned when reading logs for service '%v' in enclave '%v'", serviceUuid, enclaveUuid)
					return
				}
				break
			}
			if readErr != nil {
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

			// Then we extract the actual log message using vectors log field
			logMsgStr, found := jsonLog[volume_consts.LogLabel]
			if !found {
				streamErrChan <- stacktrace.NewError("An error retrieving the log field from logs json file. This is a bug in Kurtosis.")
				return
			}

			// Extract the timestamp using vectors timestamp field
			logTimestamp, err := parseTimestampFromJsonLogLine(jsonLog)
			if err != nil {
				streamErrChan <- stacktrace.Propagate(err, "An error occurred parsing timestamp from json log line.")
				return
			}
			logLine := logline.NewLogLine(logMsgStr, *logTimestamp)

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
		}
	}
}
