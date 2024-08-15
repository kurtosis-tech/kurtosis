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
	oneWeek = 7 * 24 * time.Hour

	// basepath /year/week
	PerWeekDirPathStr = "%s%s/%s/"

	// ... enclave uuid/service uuid <filetype>
	PerWeekFilePathFmtStr = PerWeekDirPathStr + "%s/%s%s"
)

type PerWeekFileLayout struct {
	time logs_clock.LogsClock
}

func NewPerWeekFileLayout(time logs_clock.LogsClock) *PerWeekFileLayout {
	return &PerWeekFileLayout{time: time}
}

func (pwf *PerWeekFileLayout) GetLogFileLayoutFormat() string {
	return "/var/log/kurtosis/%%Y/%%V/{{ enclave_uuid }}/{{ service_uuid }}.json"
}

func (pwf *PerWeekFileLayout) GetLogFilePath(time time.Time, enclaveUuid, serviceUuid string) string {
	year, week := time.ISOWeek()
	return getLogFilePath(year, week, enclaveUuid, serviceUuid)
}

// TODO: adjust to support getting log file paths beyond retention period
func (pwf *PerWeekFileLayout) GetLogFilePaths(
	filesystem volume_filesystem.VolumeFilesystem,
	retentionPeriod time.Duration,
	retentionPeriodIntervals int,
	enclaveUuid, serviceUuid string) ([]string, error) {
	var paths []string
	currentTime := pwf.time.Now()

	retentionPeriodInWeeks := DurationToWeeks(retentionPeriod)

	// scan for first existing log file
	firstWeekWithLogs := 0
	for i := 0; i < retentionPeriodInWeeks; i++ {
		year, week := currentTime.Add(time.Duration(-i) * oneWeek).ISOWeek()
		filePathStr := getLogFilePath(year, week, enclaveUuid, serviceUuid)
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
		filePathStr := getLogFilePath(year, week, enclaveUuid, serviceUuid)
		if _, err := filesystem.Stat(filePathStr); err != nil {
			break
		}
		paths = append(paths, filePathStr)
	}

	// reverse for oldest to most recent
	slices.Reverse(paths)

	return paths, nil
}

func DurationToWeeks(d time.Duration) int {
	return int(math.Round(d.Hours() / float64(7*24)))
}

func getLogFilePath(year, week int, enclaveUuid, serviceUuid string) string {
	formattedWeekNum := fmt.Sprintf("%02d", week)
	return fmt.Sprintf(PerWeekFilePathFmtStr, volume_consts.LogsStorageDirpath, strconv.Itoa(year), formattedWeekNum, enclaveUuid, serviceUuid, volume_consts.Filetype)
}
