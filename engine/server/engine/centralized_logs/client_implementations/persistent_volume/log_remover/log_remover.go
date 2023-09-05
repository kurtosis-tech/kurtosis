package log_remover

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/logs_clock"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_consts"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_filesystem"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
)

const (
	oneWeekDuration = 7 * 24 * time.Hour
)

// LogRemover removes logs one week older than the log retention period.
type LogRemover struct {
	filesystem volume_filesystem.VolumeFilesystem

	time logs_clock.LogsClock
}

func NewLogRemover(filesystem volume_filesystem.VolumeFilesystem, time logs_clock.LogsClock) *LogRemover {
	return &LogRemover{
		filesystem: filesystem,
		time:       time,
	}
}

// Run implements the Job cron interface. It removes logs a week older than the log retention period.
func (remover LogRemover) Run() {
	// [LogRetentionPeriodInWeeks] weeks plus an extra week of logs are retained so remove logs a week past that, hence +1
	numWeeksBack := volume_consts.LogRetentionPeriodInWeeks + 1

	// compute the next oldest week
	_, weekToRemove := remover.time.Now().Add(time.Duration(-numWeeksBack) * oneWeekDuration).ISOWeek()

	// remove directory for that week
	oldLogsDirPath := fmt.Sprintf("%s%s/", volume_consts.LogsStorageDirpath, strconv.Itoa(weekToRemove))
	if err := remover.filesystem.RemoveAll(oldLogsDirPath); err != nil {
		logrus.Warnf("An error occurred removing old logs at the following path '%v': %v\n", oldLogsDirPath, err)
	}
}
