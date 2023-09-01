package stream_logs_strategy

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/consts"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_filesystem"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/logline"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"strings"
)

const (
	// This is how many weeks we promise to hold logs for
	// We use this to compute how far back in time we need pull logs
	defaultRetentionPeriodInWeeks = 4
)

// This strategy pulls logs from filesytsem where there is a log file per week, per enclave, per service
// Weeks are denoted 00-53, where 00 is the first week of the year
// eg.
// [.../28/d3e8832d671f/61830789f03a.json] is the file containing logs from service with uuid 61830789f03a, in enclave with uuid d3e8832d671f,
// in the 28th week of the current year
type PerWeekStreamLogsStrategy struct {
}

func (strategy *PerWeekStreamLogsStrategy) StreamLogs(
	ctx context.Context,
	fs volume_filesystem.VolumeFilesystem,
	logsByKurtosisUserServiceUuidChan chan map[service.ServiceUUID][]logline.LogLine,
	streamErrChan chan error,
	enclaveUuid enclave.EnclaveUUID,
	serviceUuid service.ServiceUUID,
	conjunctiveLogLinesFiltersWithRegex []logline.LogLineFilterWithRegex,
	shouldFollowLogs bool,
) {
	logPaths, err := getRetainedLogsFilepaths(fs, defaultRetentionPeriodInWeeks, 5, string(enclaveUuid), string(serviceUuid))
	if err != nil {
		streamErrChan <- stacktrace.Propagate(err, "An error retrieving the retained logs filepaths for service '%v' in enclave '%v'.", serviceUuid, enclaveUuid)
		return
	}

	fileReaders := []io.Reader{}
	for _, pathStr := range logPaths {
		logsFile, err := fs.Open(pathStr)
		if err != nil {
			streamErrChan <- stacktrace.Propagate(err, "An error occurred opening the logs file for service '%v' in enclave '%v' at the following path: %v.", serviceUuid, enclaveUuid, pathStr)
			return
		}
		fileReaders = append(fileReaders, logsFile)
	}

	combinedLogsReader := io.MultiReader(fileReaders...)

	logsReader := bufio.NewReader(combinedLogsReader)

	numLogsReturned := 0
	for shouldFollowLogs || numLogsReturned < consts.MaxNumLogsToReturn {
		select {
		case <-ctx.Done():
			logrus.Debugf("Context was canceled, stopping streaming service logs for service '%v' in enclave '%v", serviceUuid, enclaveUuid)
			return
		default:
			var jsonLogStr string
			var err error
			var readErr error
			var jsonLogNewStr string

			for {
				jsonLogNewStr, readErr = logsReader.ReadString(consts.NewLineRune)
				jsonLogStr = jsonLogStr + jsonLogNewStr
				// check if it's an uncompleted Json line
				if jsonLogNewStr != "" && len(jsonLogNewStr) > 2 {
					jsonLogNewStrLastChars := jsonLogNewStr[len(jsonLogNewStr)-2:]
					if jsonLogNewStrLastChars != consts.EndOfJsonLine {
						// removes the newline char from the previous part of the json line
						jsonLogStr = strings.TrimSuffix(jsonLogStr, string(consts.NewLineRune))
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
				streamErrChan <- stacktrace.Propagate(err, "An error occurred reading the logs file for service '%v' in enclave '%v'.", serviceUuid, enclaveUuid)
				return
			}

			// each logLineStr is of the following structure: {"enclave_uuid": "...", "service_uuid":"...", "log": "...",.. "timestamp":"..."}
			// eg. {"container_type":"api-container", "container_id":"8f8558ba", "container_name":"/kurtosis-api--ffd",
			// "log":"hi","timestamp":"2023-08-14T14:57:49Z"}

			// First, we decode the line
			var jsonLog JsonLog
			err = json.Unmarshal([]byte(jsonLogStr), &jsonLog)
			if err != nil {
				streamErrChan <- stacktrace.Propagate(err, "An error occurred parsing the json logs file for service '%v' in enclave '%v'.", serviceUuid, enclaveUuid)
				return
			}

			// Then we extract the actual log message using the "log" field
			logLineStr, found := jsonLog[consts.LogLabel]
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

// [getRetainedLogsFilepaths] returns a list of log filepaths containing logs for [serviceUuid] in [enclaveUuid]
// going back ([retentionPeriodInWeeks] + 1) back from [currentWeek] where these are 00-53(%U strftime specifier) values
// denoting the week of the year.
// Notes:
// - The +1 is because we retain an extra week of logs compared to what we promise to retain for safety.
// - The list of filepaths will be returned in order of most recent logs to the oldest logs e.g. [ 4/80124/1234.json, /3/801234/1234.json, ...]
// Returns error:
//   - if any of the filepaths don't exist in the underlying filesystem
//   - if there are less the ([retentionPeriodInWeeks] + 1) log filepaths found
//     This indicates logs were lost or manually removed.
func getRetainedLogsFilepaths(
	filesystem volume_filesystem.VolumeFilesystem,
	retentionPeriodInWeeks int,
	currentWeek int,
	enclaveUuid, serviceUuid string) ([]string, error) {
	return []string{}, nil
}
