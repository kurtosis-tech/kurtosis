package log_remover

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/consts"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
	"time"
)

const (
	maxWeekNum = 54
)

// LogRemover removes logs one week older than the log retention period.
type LogRemover struct {
}

// Run implements the Job cron interface. It removes logs a week older than the log retention period.
func (remover LogRemover) Run() {
	// get the current week
	_, week := time.Now().UTC().ISOWeek()

	// compute the next oldest week
	diff := week - (consts.LogRetentionPeriodInWeeks + 1) - 1
	var weekPastRetentionPeriod int
	if diff >= 0 {
		weekPastRetentionPeriod = diff
	} else {
		weekPastRetentionPeriod = maxWeekNum + diff
	}

	// create directory path for that week
	oldLogsDirPath := fmt.Sprintf("%s%s/", consts.LogsStorageDirpath, strconv.Itoa(weekPastRetentionPeriod))
	if _, err := os.Stat(oldLogsDirPath); !os.IsNotExist(err) {
		err = os.RemoveAll(oldLogsDirPath)
		if err != nil {
			logrus.Warnf("An error occurred removing old logs at the following path: %v\n", oldLogsDirPath)
		}
	}
}
