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
	"os"
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
	shouldReturnAllLogs bool,
	numLogLines uint32,
) {
	paths, err := strategy.getLogFilePaths(fs, volume_consts.LogRetentionPeriodInWeeks, string(enclaveUuid), string(serviceUuid))
	if err != nil {
		streamErrChan <- stacktrace.Propagate(err, "An error occurred retrieving log file paths for service '%v' in enclave '%v'.", serviceUuid, enclaveUuid)
		return
	}
	if len(paths) == 0 {
		streamErrChan <- stacktrace.NewError(
			`No logs file paths for service '%v' in enclave '%v' were found. This means either:
					1) No logs for this service were detected/stored.
					2) Logs were manually removed.`,
			serviceUuid, enclaveUuid)
		return
	}
	if len(paths) > volume_consts.LogRetentionPeriodInWeeks {
		logrus.Warnf(
			`We expected to retrieve logs going back '%v' weeks, but instead retrieved logs going back '%v' weeks. 
					This means logs past the retention period are being returned, likely a bug in Kurtosis.`,
			volume_consts.LogRetentionPeriodInWeeks, len(paths))
	}

	logsReader, files, err := getLogsReader(fs, paths)
	if err != nil {
		streamErrChan <- stacktrace.Propagate(err, "An error occurred creating a logs reader for service '%v' in enclave '%v'", serviceUuid, enclaveUuid)
		return
	}
	defer func() {
		for _, file := range files {
			_ = file.Close()
		}
	}()

	if shouldReturnAllLogs {
		if err := strategy.streamAllLogs(ctx, logsReader, logsByKurtosisUserServiceUuidChan, serviceUuid, conjunctiveLogLinesFiltersWithRegex); err != nil {
			streamErrChan <- stacktrace.Propagate(err, "An error occurred streaming all logs for service '%v' in enclave '%v'", serviceUuid, enclaveUuid)
			return
		}
	} else {
		if err := strategy.streamTailLogs(ctx, logsReader, numLogLines, logsByKurtosisUserServiceUuidChan, serviceUuid, conjunctiveLogLinesFiltersWithRegex); err != nil {
			streamErrChan <- stacktrace.Propagate(err, "An error occurred streaming '%v' logs for service '%v' in enclave '%v'", numLogLines, serviceUuid, enclaveUuid)
			return
		}
	}

	if shouldFollowLogs {
		latestLogFile := paths[len(paths)-1]
		if err := strategy.followLogs(latestLogFile, logsByKurtosisUserServiceUuidChan, serviceUuid, conjunctiveLogLinesFiltersWithRegex); err != nil {
			streamErrChan <- stacktrace.Propagate(err, "An error occurred creating following logs for service '%v' in enclave '%v'", serviceUuid, enclaveUuid)
			return
		}
		logrus.Debugf("Following logs...")
	}
}

// [getLogFilePaths] returns a list of log file paths containing logs for [serviceUuid] in [enclaveUuid]
// going [retentionPeriodInWeeks] back from the [currentWeek].
// Notes:
// - File paths are of the format '/week/enclave uuid/service uuid.json' where 'week' is %W strftime specifier
// - The list of file paths is returned in order of oldest logs to most recent logs e.g. [ 3/80124/1234.json, /4/801234/1234.json, ...]
// - If a file path does not exist, the function with exits and returns whatever file paths were found
func (strategy *PerWeekStreamLogsStrategy) getLogFilePaths(filesystem volume_filesystem.VolumeFilesystem, retentionPeriodInWeeks int, enclaveUuid, serviceUuid string) ([]string, error) {
	var paths []string
	currentTime := strategy.time.Now()

	// scan for first existing log file
	firstWeekWithLogs := 0
	for i := 0; i < retentionPeriodInWeeks; i++ {
		year, week := currentTime.Add(time.Duration(-i) * oneWeek).ISOWeek()
		filePathStr := fmt.Sprintf(volume_consts.PerWeekFilePathFmtStr, volume_consts.LogsStorageDirpath, strconv.Itoa(year), strconv.Itoa(week), enclaveUuid, serviceUuid, volume_consts.Filetype)
		if _, err := filesystem.Stat(filePathStr); err == nil {
			paths = append(paths, filePathStr)
			firstWeekWithLogs = i
			break
		} else {
			// return if error is not due to nonexistent file path
			if !os.IsNotExist(err) {
				return paths, err
			}
		}
	}

	// scan for remaining files as far back as they exist
	for i := firstWeekWithLogs + 1; i < retentionPeriodInWeeks; i++ {
		year, week := currentTime.Add(time.Duration(-i) * oneWeek).ISOWeek()
		filePathStr := fmt.Sprintf(volume_consts.PerWeekFilePathFmtStr, volume_consts.LogsStorageDirpath, strconv.Itoa(year), strconv.Itoa(week), enclaveUuid, serviceUuid, volume_consts.Filetype)
		if _, err := filesystem.Stat(filePathStr); err != nil {
			break
		}
		paths = append(paths, filePathStr)
	}

	// reverse for oldest to most recent
	slices.Reverse(paths)

	return paths, nil
}

// Returns a Reader over all logs in [logFilePaths] and the open file descriptors of the associated [logFilePaths]
func getLogsReader(filesystem volume_filesystem.VolumeFilesystem, logFilePaths []string) (*bufio.Reader, []volume_filesystem.VolumeFile, error) {
	var fileReaders []io.Reader
	var files []volume_filesystem.VolumeFile

	// get a reader for each log file
	for _, pathStr := range logFilePaths {
		logsFile, err := filesystem.Open(pathStr)
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred opening the logs file at the following path: %v", pathStr)
		}
		fileReaders = append(fileReaders, logsFile)
		files = append(files, logsFile)
	}

	// combine log file readers into a single reader
	combinedLogsReader := io.MultiReader(fileReaders...)

	return bufio.NewReader(combinedLogsReader), files, nil
}

func (strategy *PerWeekStreamLogsStrategy) streamAllLogs(
	ctx context.Context,
	logsReader *bufio.Reader,
	logsByKurtosisUserServiceUuidChan chan map[service.ServiceUUID][]logline.LogLine,
	serviceUuid service.ServiceUUID,
	conjunctiveLogLinesFiltersWithRegex []logline.LogLineFilterWithRegex) error {
	for {
		select {
		case <-ctx.Done():
			logrus.Debugf("Context was canceled, stopping streaming service logs for service '%v'", serviceUuid)
			return nil
		default:
			jsonLogStr, err := getCompleteJsonLogString(logsReader)
			if isValidJsonEnding(jsonLogStr) {
				if err = strategy.sendJsonLogLine(jsonLogStr, logsByKurtosisUserServiceUuidChan, serviceUuid, conjunctiveLogLinesFiltersWithRegex); err != nil {
					return err
				}
			}

			if err != nil {
				// if we've reached end of logs, return success, otherwise return the error
				if errors.Is(err, io.EOF) {
					return nil
				} else {
					return err
				}
			}
		}
	}
}

// tail -n X
func (strategy *PerWeekStreamLogsStrategy) streamTailLogs(
	ctx context.Context,
	logsReader *bufio.Reader,
	numLogLines uint32,
	logsByKurtosisUserServiceUuidChan chan map[service.ServiceUUID][]logline.LogLine,
	serviceUuid service.ServiceUUID,
	conjunctiveLogLinesFiltersWithRegex []logline.LogLineFilterWithRegex) error {
	tailLogLines := make([]string, 0, numLogLines)

	for {
		select {
		case <-ctx.Done():
			logrus.Debugf("Context was canceled, stopping streaming service logs for service '%v'", serviceUuid)
			return nil
		default:
			jsonLogStr, err := getCompleteJsonLogString(logsReader)
			if isValidJsonEnding(jsonLogStr) {
				// collect all log lines in tail log lines
				tailLogLines = append(tailLogLines, jsonLogStr)
				if len(tailLogLines) > int(numLogLines) {
					tailLogLines = tailLogLines[1:]
				}
				continue
			}

			if err != nil {
				// stop reading if end of logs reached, otherwise return the error
				if errors.Is(err, io.EOF) {
					break
				} else {
					return err
				}
			}
		}
		break
	}

	for _, jsonLogStr := range tailLogLines {
		if err := strategy.sendJsonLogLine(jsonLogStr, logsByKurtosisUserServiceUuidChan, serviceUuid, conjunctiveLogLinesFiltersWithRegex); err != nil {
			return err
		}
	}

	return nil
}

// Returns the next complete json log string from [logsReader], unless err is reached in which case an incomplete json log line could be returned
func getCompleteJsonLogString(logsReader *bufio.Reader) (string, error) {
	var completeJsonLogStr string
	for {
		jsonLogStr, isJsonEnding, err := getJsonLogString(logsReader)
		completeJsonLogStr = completeJsonLogStr + jsonLogStr
		// err could be EOF or something else so return the string and err for caller to handle
		if err != nil {
			return completeJsonLogStr, err
		}

		// if appended string is valid json ending, can now return a complete json log
		if isJsonEnding {
			return completeJsonLogStr, nil
		}
	}
}

// Return the next json log string from [logsReader] and whether the string ends in valid json or not
func getJsonLogString(logsReader *bufio.Reader) (string, bool, error) {
	jsonLogStr, err := logsReader.ReadString(volume_consts.NewLineRune)

	jsonLogStr = strings.TrimSuffix(jsonLogStr, string(volume_consts.NewLineRune))

	return jsonLogStr, isValidJsonEnding(jsonLogStr), err
}

func isValidJsonEnding(line string) bool {
	if len(line) < len(volume_consts.EndOfJsonLine) {
		return false
	}
	endOfLine := line[len(line)-len(volume_consts.EndOfJsonLine):]
	return endOfLine == volume_consts.EndOfJsonLine
}

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
		logrus.Warnf("An error occurred parsing the json log string: %v. Skipping sending this log line.", jsonLogLineStr)
		return nil
	}

	// Then extract the actual log message using the vectors log field
	logMsgStr, found := jsonLog[volume_consts.LogLabel]
	if !found {
		return stacktrace.NewError("An error retrieving the log field '%v' from json log string: %v\n", volume_consts.LogLabel, jsonLogLineStr)
	}

	// Extract the timestamp using vectors timestamp field
	logTimestamp, err := parseTimestampFromJsonLogLine(jsonLog)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing timestamp from json log line.")
	}
	logLine := logline.NewLogLine(logMsgStr, *logTimestamp)

	// Then filter by checking if the log message is valid based on requested filters
	validLogLine, err := logLine.IsValidLogLineBaseOnFilters(conjunctiveLogLinesFiltersWithRegex)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred filtering log line '%+v' using filters '%+v'", logLine, conjunctiveLogLinesFiltersWithRegex)
	}
	if !validLogLine {
		return nil
	}

	// ensure this log line is within the retention period if it has a timestamp
	withinRetentionPeriod, err := strategy.isWithinRetentionPeriod(logLine)
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
func (strategy *PerWeekStreamLogsStrategy) isWithinRetentionPeriod(logLine *logline.LogLine) (bool, error) {
	retentionPeriod := strategy.time.Now().Add(time.Duration(-volume_consts.LogRetentionPeriodInWeeks) * oneWeek)
	timestamp := logLine.GetTimestamp()
	return timestamp.After(retentionPeriod), nil
}

// Continue streaming log lines as they are written to log file (tail -f [filepath])
func (strategy *PerWeekStreamLogsStrategy) followLogs(
	filepath string,
	logsByKurtosisUserServiceUuidChan chan map[service.ServiceUUID][]logline.LogLine,
	serviceUuid service.ServiceUUID,
	conjunctiveLogLinesFiltersWithRegex []logline.LogLineFilterWithRegex,
) error {
	logTail, err := tail.TailFile(filepath, tail.Config{
		Location: &tail.SeekInfo{
			Offset: 0,
			Whence: io.SeekEnd, // start tailing from end of log file
		},
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

// Converts a string in UTC format to a time.Time, returns error if no time is found or time is incorrectly formatted
func parseTimestampFromJsonLogLine(logLine JsonLog) (*time.Time, error) {
	timestampStr, found := logLine[volume_consts.TimestampLabel]
	if !found {
		return nil, stacktrace.NewError("An error occurred retrieving the timestamp field '%v' from json: %v", volume_consts.TimestampLabel, logLine)
	}
	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred parsing the timestamp string '%v' from UTC to a time.Time object.", timestampStr)
	}
	return &timestamp, nil
}
