package stream_logs_strategy

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/logs_clock"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_consts"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_filesystem"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/logline"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
	"io"
	"strconv"
	"strings"
	"time"
)

const (
	oneWeek = 7 * 24 * time.Hour
)

// This strategy pulls logs from filesytsem where there is a log file per year, per week, per enclave, per service
// Weeks are denoted 01-52
// eg.
// [.../28/d3e8832d671f/61830789f03a.json] is the file containing logs from service with uuid 61830789f03a, in enclave with uuid d3e8832d671f,
// in the 28th week of the current year
type PerWeekStreamLogsStrategy struct {
	time logs_clock.LogsClock
}

func NewPerWeekStreamLogsStrategy(time logs_clock.LogsClock) *PerWeekStreamLogsStrategy {
	return &PerWeekStreamLogsStrategy{
		time: time,
	}
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
	paths := strategy.getRetainedLogsFilePaths(fs, volume_consts.LogRetentionPeriodInWeeks, string(enclaveUuid), string(serviceUuid))
	if len(paths) == 0 {
		streamErrChan <- stacktrace.NewError(
			`No logs file paths for service '%v' in enclave '%v' were found. This means either:
					1) No logs for this service were detected/stored.
					2) Logs were manually removed.`,
			serviceUuid, enclaveUuid)
		return
	}
	if len(paths) > volume_consts.LogRetentionPeriodInWeeks+1 {
		logrus.Warnf(
			`We expected to retrieve logs going back '%v' weeks, but instead retrieved logs going back '%v' weeks. 
					This means logs past the retention period are being returned, likely a bug in Kurtosis.`,
			volume_consts.LogRetentionPeriodInWeeks+1, len(paths))
	}

	fileReaders := []io.Reader{}
	for _, pathStr := range paths {
		logsFile, err := fs.Open(pathStr)
		if err != nil {
			streamErrChan <- stacktrace.Propagate(err, "An error occurred opening the logs file for service '%v' in enclave '%v' at the following path: %v.", serviceUuid, enclaveUuid, pathStr)
			return
		}
		fileReaders = append(fileReaders, logsFile)
	}

	combinedLogsReader := io.MultiReader(fileReaders...)

	logsReader := bufio.NewReader(combinedLogsReader)

	for {
		select {
		case <-ctx.Done():
			logrus.Debugf("Context was canceled, stopping streaming service logs for service '%v' in enclave '%v", serviceUuid, enclaveUuid)
			return
		default:
			var jsonLogStr string
			var err error
			var readErr error
			var jsonLogNewStr string
			var shouldReturnAfterStreamingLastLine = false

			for {
				jsonLogNewStr, readErr = logsReader.ReadString(volume_consts.NewLineRune)
				jsonLogStr = jsonLogStr + jsonLogNewStr
				// check if it's an uncompleted Json line
				if jsonLogNewStr != "" && len(jsonLogNewStr) > 2 {
					jsonLogNewStrLastChars := jsonLogNewStr[len(jsonLogNewStr)-2:]
					if jsonLogNewStrLastChars != volume_consts.EndOfJsonLine {
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
					if jsonLogStr != "" {
						shouldReturnAfterStreamingLastLine = true
					} else {
						return
					}
				}
				break
			}
			if readErr != nil && !errors.Is(readErr, io.EOF) {
				streamErrChan <- stacktrace.Propagate(readErr, "An error occurred reading the logs file for service '%v' in enclave '%v'.", serviceUuid, enclaveUuid)
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
			logLineStr, found := jsonLog[volume_consts.LogLabel]
			if !found {
				streamErrChan <- stacktrace.NewError("An error retrieving the log field from logs json file. This is a bug in Kurtosis.")
				return
			}
			logLine := logline.NewLogLine(logLineStr)

			// Then we filter by checking if the log message is valid based on requested filtersr
			validLogLine, err := logLine.IsValidLogLineBaseOnFilters(conjunctiveLogLinesFiltersWithRegex)
			if err != nil {
				streamErrChan <- stacktrace.Propagate(err, "An error occurred filtering log line '%+v' using filters '%+v'", logLine, conjunctiveLogLinesFiltersWithRegex)
				break
			}
			// ensure this log line is within the retention period if it has a timestamp
			withinRetention, err := strategy.isWithinRetentionPeriod(jsonLog)
			if err != nil {
				streamErrChan <- stacktrace.Propagate(err, "An error occurred filtering log line '%+v' using filters '%+v'", logLine, conjunctiveLogLinesFiltersWithRegex)
				break
			}

			shouldReturnLogLine := validLogLine && withinRetention
			if !shouldReturnLogLine {
				break
			}

			// send the log line
			logLines := []logline.LogLine{*logLine}
			userServicesLogLinesMap := map[service.ServiceUUID][]logline.LogLine{
				serviceUuid: logLines,
			}
			logsByKurtosisUserServiceUuidChan <- userServicesLogLinesMap
			if shouldReturnAfterStreamingLastLine {
				return
			}
		}
	}
}

// [getRetainedLogsFilePaths] returns a list of log file paths containing logs for [serviceUuid] in [enclaveUuid]
// going back ([retentionPeriodInWeeks] + 1) back from [currentWeek].
// Notes:
// - File paths are of the format '/week/enclave uuid/service uuid.json' where 'week' is %W strftime specifier
// - The +1 is because we retain an extra week of logs compared to what we promise to retain for safety.
// - The list of file paths is returned in order of oldest logs to most recent logs e.g. [ 3/80124/1234.json, /4/801234/1234.json, ...]
// - If a file path does not exist, the function with exits and returns whatever file paths were found
func (strategy *PerWeekStreamLogsStrategy) getRetainedLogsFilePaths(
	filesystem volume_filesystem.VolumeFilesystem,
	retentionPeriodInWeeks int,
	enclaveUuid, serviceUuid string) []string {
	paths := []string{}

	// get log file paths as far back as they exist
	for i := 0; i < (retentionPeriodInWeeks + 1); i++ {
		year, week := strategy.time.Now().Add(time.Duration(-i) * oneWeek).ISOWeek()
		filePathStr := fmt.Sprintf(volume_consts.PerWeekFilePathFmtStr, volume_consts.LogsStorageDirpath, strconv.Itoa(year), strconv.Itoa(week), enclaveUuid, serviceUuid, volume_consts.Filetype)
		if _, err := filesystem.Stat(filePathStr); err != nil {
			break
		}
		paths = append(paths, filePathStr)
	}

	// reverse for oldest to most recent
	slices.Reverse(paths)

	return paths
}

// Returns true if no [logLine] has no timestamp
func (strategy *PerWeekStreamLogsStrategy) isWithinRetentionPeriod(logLine JsonLog) (bool, error) {
	retentionPeriod := strategy.time.Now().Add(time.Duration(-volume_consts.LogRetentionPeriodInWeeks-1) * oneWeek)
	timestampStr, found := logLine[volume_consts.TimestampLabel]
	if found {
		timestamp, err := time.Parse(time.RFC3339, timestampStr)
		if err != nil {
			return false, stacktrace.Propagate(err, "An error occurred retrieving the timestamp field from logs json log line. This is a bug in Kurtosis.")
		}
		return timestamp.After(retentionPeriod), nil
	}
	return true, nil
}
