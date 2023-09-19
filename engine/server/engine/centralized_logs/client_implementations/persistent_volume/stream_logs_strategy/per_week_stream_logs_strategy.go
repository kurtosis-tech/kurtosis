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
	"github.com/nxadm/tail"
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

// PerWeekStreamLogsStrategy pulls logs from filesystem where there is a log file per year, per week, per enclave, per service
// Weeks are denoted 01-52
// e.g.
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
	latestLogFile := paths[len(paths)-1]

	var fileReaders []io.Reader
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
			var jsonLogNewStr string
			var readErr error
			var err error
			var isLastLogLine = false

			// get a complete log line
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
					// exiting stream
					logrus.Debugf("EOF error returned when reading logs for service '%v' in enclave '%v'", serviceUuid, enclaveUuid)
					if jsonLogStr != "" {
						isLastLogLine = true
					} else {
						if shouldFollowLogs {
							if err = strategy.tailLogs(latestLogFile, logsByKurtosisUserServiceUuidChan, serviceUuid, conjunctiveLogLinesFiltersWithRegex); err != nil {
								streamErrChan <- stacktrace.Propagate(err, "An error occurred following logs for service '%v' in enclave '%v'.", serviceUuid, enclaveUuid)
								return
							}
						} else {
							return
						}
					}
				}
				break
			}
			if readErr != nil && !errors.Is(readErr, io.EOF) {
				streamErrChan <- stacktrace.Propagate(readErr, "An error occurred reading the logs file for service '%v' in enclave '%v'.", serviceUuid, enclaveUuid)
				return
			}

			if err = strategy.sendJsonLogLine(jsonLogStr, logsByKurtosisUserServiceUuidChan, serviceUuid, conjunctiveLogLinesFiltersWithRegex); err != nil {
				streamErrChan <- stacktrace.Propagate(err, "An error occurred sending log line for service '%v' in enclave '%v'.", serviceUuid, enclaveUuid)
				return
			}

			if isLastLogLine {
				if shouldFollowLogs {
					if err = strategy.tailLogs(latestLogFile, logsByKurtosisUserServiceUuidChan, serviceUuid, conjunctiveLogLinesFiltersWithRegex); err != nil {
						streamErrChan <- stacktrace.Propagate(err, "An error occurred following logs for service '%v' in enclave '%v'.", serviceUuid, enclaveUuid)
						return
					}
				} else {
					return
				}
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
	var paths []string
	currentTime := strategy.time.Now()

	// scan for first existing log file
	firstWeekWithLogs := 0
	for i := 0; i < (retentionPeriodInWeeks + 1); i++ {
		year, week := currentTime.Add(time.Duration(-i) * oneWeek).ISOWeek()
		filePathStr := fmt.Sprintf(volume_consts.PerWeekFilePathFmtStr, volume_consts.LogsStorageDirpath, strconv.Itoa(year), strconv.Itoa(week), enclaveUuid, serviceUuid, volume_consts.Filetype)
		if _, err := filesystem.Stat(filePathStr); err == nil {
			paths = append(paths, filePathStr)
			firstWeekWithLogs = i
			break
		}
	}

	// scan for remaining files as far back as they exist
	for i := firstWeekWithLogs + 1; i < (retentionPeriodInWeeks + 1); i++ {
		year, week := currentTime.Add(time.Duration(-i) * oneWeek).ISOWeek()
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

// tail -f [filepath]
func (strategy *PerWeekStreamLogsStrategy) tailLogs(
	filepath string,
	logsByKurtosisUserServiceUuidChan chan map[service.ServiceUUID][]logline.LogLine,
	serviceUuid service.ServiceUUID,
	conjunctiveLogLinesFiltersWithRegex []logline.LogLineFilterWithRegex,
) error {
	logTail, err := tail.TailFile(filepath, tail.Config{
		Location:    nil,
		ReOpen:      false,
		MustExist:   true,
		Poll:        false,
		Pipe:        false,
		Follow:      true,
		MaxLineSize: 0,
		RateLimiter: nil,
		Logger:      logrus.StandardLogger()})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while attempting to tail the log file.")
	}

	for logLine := range logTail.Lines {
		err = strategy.sendJsonLogLine(logLine.Text, logsByKurtosisUserServiceUuidChan, serviceUuid, conjunctiveLogLinesFiltersWithRegex)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred sending json log line '%v'.", logLine.Text)
		}
	}
	return nil
}

// Returns error if [jsonLogLineStr] is not a valid log line
func (strategy *PerWeekStreamLogsStrategy) sendJsonLogLine(
	jsonLogLineStr string,
	logsByKurtosisUserServiceUuidChan chan map[service.ServiceUUID][]logline.LogLine,
	serviceUuid service.ServiceUUID,
	conjunctiveLogLinesFiltersWithRegex []logline.LogLineFilterWithRegex) error {
	// each logLineStr is of the following structure: {"enclave_uuid": "...", "service_uuid":"...", "log": "...",.. "timestamp":"..."}
	// eg. {"container_type":"api-container", "container_id":"8f8558ba", "container_name":"/kurtosis-api--ffd",
	// "log":"hi","timestamp":"2023-08-14T14:57:49Z"}

	// First decode the line
	var jsonLog JsonLog
	if err := json.Unmarshal([]byte(jsonLogLineStr), &jsonLog); err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing the json log string: %v\n", jsonLogLineStr)
	}

	// Then extract the actual log message using the "log" field
	logLineStr, found := jsonLog[volume_consts.LogLabel]
	if !found {
		return stacktrace.NewError("An error retrieving the log field from json log string: %v\n", jsonLogLineStr)
	}
	logLine := logline.NewLogLine(logLineStr)

	// Then filter by checking if the log message is valid based on requested filters
	validLogLine, err := logLine.IsValidLogLineBaseOnFilters(conjunctiveLogLinesFiltersWithRegex)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred filtering log line '%+v' using filters '%+v'", logLine, conjunctiveLogLinesFiltersWithRegex)
	}
	if !validLogLine {
		return nil
	}

	// ensure this log line is within the retention period if it has a timestamp
	withinRetentionPeriod, err := strategy.isWithinRetentionPeriod(jsonLog)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred filtering log line '%+v' using filters '%+v'", logLine, conjunctiveLogLinesFiltersWithRegex)
	}
	if !withinRetentionPeriod {
		return nil
	}

	// send the log line
	logLines := []logline.LogLine{*logLine}
	userServicesLogLinesMap := map[service.ServiceUUID][]logline.LogLine{
		serviceUuid: logLines,
	}
	logsByKurtosisUserServiceUuidChan <- userServicesLogLinesMap
	return nil
}

// Returns true if [logLine] has no timestamp
func (strategy *PerWeekStreamLogsStrategy) isWithinRetentionPeriod(logLine JsonLog) (bool, error) {
	retentionPeriod := strategy.time.Now().Add(time.Duration(-volume_consts.LogRetentionPeriodInWeeks-1) * oneWeek)
	timestampStr, found := logLine[volume_consts.TimestampLabel]
	if !found {
		return true, nil
	}

	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred retrieving the timestamp field from logs json log line. This is a bug in Kurtosis.")
	}
	return timestamp.After(retentionPeriod), nil
}
