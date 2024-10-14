package file_layout

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/logs_clock"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_consts"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_filesystem"
	"golang.org/x/exp/slices"
	"math"
	"os"
	"strconv"
	"time"
)

const (
	oneWeekInHours  = 7 * 24
	oneWeekDuration = oneWeekInHours * time.Hour

	// basepath year/week
	perWeekDirPathFmtStr = "%s%s/%s/"

	// ... enclave uuid/service uuid <filetype>
	perWeekFilePathFmtStr = perWeekDirPathFmtStr + "%s/%s%s"
)

type PerWeekFileLayout struct {
	time logs_clock.LogsClock
}

func NewPerWeekFileLayout(time logs_clock.LogsClock) *PerWeekFileLayout {
	return &PerWeekFileLayout{time: time}
}

func (pwf *PerWeekFileLayout) GetLogFileLayoutFormat() string {
	// Right now this format is specifically made for Vector Logs Aggregators format
	// This wil be used my Vector LogsAggregator to determine the path to output to
	// is there a way to get rid of the /var/log/kurtosis?
	return fmt.Sprintf("%s%%Y/%%V/{{ enclave_uuid }}/{{ service_uuid }}.json", volume_consts.LogsStorageDirpath)
}

func (pwf *PerWeekFileLayout) GetLogFilePath(time time.Time, enclaveUuid, serviceUuid string) string {
	year, week := time.ISOWeek()
	return getWeeklyFilePath(year, week, enclaveUuid, serviceUuid)
}

func (pwf *PerWeekFileLayout) GetLogFilePaths(
	filesystem volume_filesystem.VolumeFilesystem,
	retentionPeriod time.Duration,
	retentionPeriodIntervals int,
	enclaveUuid, serviceUuid string) ([]string, error) {
	var paths []string
	retentionPeriodInWeeks := DurationToWeeks(retentionPeriod)

	if retentionPeriodIntervals < 0 {
		return pwf.getLogFilePathsFromNowTillRetentionPeriod(filesystem, retentionPeriodInWeeks, enclaveUuid, serviceUuid)
	} else {
		paths = pwf.getLogFilePathsBeyondRetentionPeriod(filesystem, retentionPeriodInWeeks, retentionPeriodIntervals, enclaveUuid, serviceUuid)
	}

	return paths, nil
}

func (pwf *PerWeekFileLayout) getLogFilePathsFromNowTillRetentionPeriod(fs volume_filesystem.VolumeFilesystem, retentionPeriodInWeeks int, enclaveUuid, serviceUuid string) ([]string, error) {
	var paths []string
	currentTime := pwf.time.Now()

	// scan for first existing log file
	firstWeekWithLogs := 0
	for i := 0; i < retentionPeriodInWeeks; i++ {
		year, week := currentTime.Add(time.Duration(-i) * oneWeekDuration).ISOWeek()
		filePathStr := getWeeklyFilePath(year, week, enclaveUuid, serviceUuid)
		if _, err := fs.Stat(filePathStr); err == nil {
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

	// scan for remaining files as far back as they exist before the retention period
	for i := firstWeekWithLogs + 1; i < retentionPeriodInWeeks; i++ {
		year, week := currentTime.Add(time.Duration(-i) * oneWeekDuration).ISOWeek()
		filePathStr := getWeeklyFilePath(year, week, enclaveUuid, serviceUuid)
		if _, err := fs.Stat(filePathStr); err != nil {
			break
		}
		paths = append(paths, filePathStr)
	}

	// reverse for oldest to most recent
	slices.Reverse(paths)

	return paths, nil
}

func (pwf *PerWeekFileLayout) getLogFilePathsBeyondRetentionPeriod(fs volume_filesystem.VolumeFilesystem, retentionPeriodInWeeks int, retentionPeriodIntervals int, enclaveUuid, serviceUuid string) []string {
	var paths []string
	currentTime := pwf.time.Now()

	// scan for log files just beyond the retention period
	for i := 0; i < retentionPeriodIntervals; i++ {
		numWeeksToGoBack := retentionPeriodInWeeks + i
		year, weekToRemove := currentTime.Add(time.Duration(-numWeeksToGoBack) * oneWeekDuration).ISOWeek()
		filePathStr := getWeeklyFilePath(year, weekToRemove, enclaveUuid, serviceUuid)
		if _, err := fs.Stat(filePathStr); err != nil {
			continue
		}
		paths = append(paths, filePathStr)
	}

	return paths
}

func DurationToWeeks(d time.Duration) int {
	return int(math.Round(d.Hours() / float64(oneWeekInHours)))
}

func getWeeklyFilePath(year, week int, enclaveUuid, serviceUuid string) string {
	formattedWeekNum := fmt.Sprintf("%02d", week)
	return fmt.Sprintf(perWeekFilePathFmtStr, volume_consts.LogsStorageDirpath, strconv.Itoa(year), formattedWeekNum, enclaveUuid, serviceUuid, volume_consts.Filetype)
}
