package log_remover

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/consts"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/logs_clock"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
	"time"
)

const (
	oneWeekDuration = 7 * 24 * time.Hour
)

// LogRemover removes logs one week older than the log retention period.
type LogRemover struct {
	time logs_clock.LogsClock
}

func NewLogRemover(time logs_clock.LogsClock) *LogRemover {
	return &LogRemover{
		time: time,
	}
}

// Run implements the Job cron interface. It removes logs a week older than the log retention period.
func (remover LogRemover) Run() {
	// [LogRetentionPeriodInWeeks] weeks plus an extra week of logs are retained so remove logs a week past that, hence +2
	numWeeksBack := consts.LogRetentionPeriodInWeeks + 2

	// compute the next oldest week
	_, weekToRemove := remover.time.Now().Add(time.Duration(-numWeeksBack) * oneWeekDuration).ISOWeek()

	// remove directory for that week
	oldLogsDirPath := fmt.Sprintf("%s%s/", consts.LogsStorageDirpath, strconv.Itoa(weekToRemove))
	if err := os.RemoveAll(oldLogsDirPath); err != nil {
		logrus.Warnf("An error occurred removing old logs at the following path '%v': %v\n", oldLogsDirPath, err)
	}
}
