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
	// basepath year/week/day/hour/
	perHourDirPathFmtStr = "%s%s/%s/%s/%s/"

	// ... enclave-uuid/service-uuid<filetype>
	perHourFilePathFmtSt = perHourDirPathFmtStr + "%s/%s%s"
)

type PerHourFileLayout struct {
	time logs_clock.LogsClock
}

func NewPerHourFileLayout(time logs_clock.LogsClock) *PerHourFileLayout {
	return &PerHourFileLayout{
		time: time,
	}
}

func (phf *PerHourFileLayout) GetLogFileLayoutFormat() string {
	// Right now this format is specifically made for Vector Logs Aggregators format
	// This wil be used my Vector LogsAggregator to determine the path to output to
	return fmt.Sprintf("%s%%Y/%%V/%%u/%%H/{{ enclave_uuid }}/{{ service_uuid }}.json", volume_consts.LogsStorageDirpath)
}

func (phf *PerHourFileLayout) GetLogFilePath(time time.Time, enclaveUuid, serviceUuid string) string {
	year, week := time.ISOWeek()
	day := time.Weekday()
	hour := time.Hour()
	return getHourlyLogFilePath(year, week, int(day), hour, enclaveUuid, serviceUuid)
}

func (phf *PerHourFileLayout) GetLogFilePaths(
	filesystem volume_filesystem.VolumeFilesystem,
	retentionPeriod time.Duration,
	retentionPeriodIntervals int,
	enclaveUuid, serviceUuid string,
) ([]string, error) {
	var paths []string
	retentionPeriodInHours := DurationToHours(retentionPeriod)

	if retentionPeriodIntervals < 0 {
		return phf.getLogFilePathsFromNowTillRetentionPeriod(filesystem, retentionPeriodInHours, enclaveUuid, serviceUuid)
	} else {
		paths = phf.getLogFilePathsBeyondRetentionPeriod(filesystem, retentionPeriodInHours, retentionPeriodIntervals, enclaveUuid, serviceUuid)
	}

	return paths, nil
}

func (phf *PerHourFileLayout) getLogFilePathsFromNowTillRetentionPeriod(fs volume_filesystem.VolumeFilesystem, retentionPeriodInHours int, enclaveUuid, serviceUuid string) ([]string, error) {
	var paths []string
	currentTime := phf.time.Now()

	// scan for first existing log file
	firstHourWithLogs := 0
	for i := 0; i < retentionPeriodInHours; i++ {
		year, week, day, hour := TimeToWeekDayHour(currentTime.Add(time.Duration(-i) * time.Hour))
		filePathStr := getHourlyLogFilePath(year, week, day, hour, enclaveUuid, serviceUuid)
		if _, err := fs.Stat(filePathStr); err == nil {
			paths = append(paths, filePathStr)
			firstHourWithLogs = i
			break
		} else {
			// return if error is not due to nonexistent file path
			if !os.IsNotExist(err) {
				return paths, err
			}
		}
	}

	// scan for remaining files as far back as they exist before the retention period
	for i := firstHourWithLogs + 1; i < retentionPeriodInHours; i++ {
		year, week, day, hour := TimeToWeekDayHour(currentTime.Add(time.Duration(-i) * time.Hour))
		filePathStr := getHourlyLogFilePath(year, week, day, hour, enclaveUuid, serviceUuid)
		if _, err := fs.Stat(filePathStr); err != nil {
			break
		}
		paths = append(paths, filePathStr)
	}

	// reverse for oldest to most recent
	slices.Reverse(paths)

	return paths, nil
}

func (phf *PerHourFileLayout) getLogFilePathsBeyondRetentionPeriod(fs volume_filesystem.VolumeFilesystem, retentionPeriodInHours int, retentionPeriodIntervals int, enclaveUuid, serviceUuid string) []string {
	var paths []string
	currentTime := phf.time.Now()

	// scan for log files just beyond the retention period
	for i := 0; i < retentionPeriodIntervals; i++ {
		numHoursToGoBack := retentionPeriodInHours + i
		year, week, day, hour := TimeToWeekDayHour(currentTime.Add(time.Duration(-numHoursToGoBack) * time.Hour))
		filePathStr := getHourlyLogFilePath(year, week, day, hour, enclaveUuid, serviceUuid)
		if _, err := fs.Stat(filePathStr); err != nil {
			continue
		}
		paths = append(paths, filePathStr)
	}

	return paths
}

func getHourlyLogFilePath(year, week, day, hour int, enclaveUuid, serviceUuid string) string {
	// these match the format in which Vector outputs week, hours, days
	formattedWeekNum := fmt.Sprintf("%02d", week)
	formattedDayNum := fmt.Sprintf("%02d", day)
	formattedHourNum := fmt.Sprintf("%02d", hour)
	return fmt.Sprintf(perHourFilePathFmtSt, volume_consts.LogsStorageDirpath, strconv.Itoa(year), formattedWeekNum, formattedDayNum, formattedHourNum, enclaveUuid, serviceUuid, volume_consts.Filetype)
}

func TimeToWeekDayHour(time time.Time) (int, int, int, int) {
	year, week := time.ISOWeek()
	hour := time.Hour()
	day := int(time.Weekday())
	return year, week, day, hour
}

func DurationToHours(duration time.Duration) int {
	return int(math.Ceil(duration.Hours()))
}