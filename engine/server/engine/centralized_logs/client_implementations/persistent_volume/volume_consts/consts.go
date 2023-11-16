package volume_consts

import (
	"time"
)

const (
	// Location of logs on the filesystem of the engine
	LogsStorageDirpath = "/var/log/kurtosis/"

	Filetype = ".json"

	NewLineRune = '\n'

	LogLabel       = "log"
	TimestampLabel = "timestamp"

	EndOfJsonLine = "}"

	// promise to keep 4 weeks of logs for users, but store an additional week for safety
	LogRetentionPeriodInWeeks = 5

	RemoveLogsWaitHours = 6 * time.Hour

	CreateLogsWaitMinutes = 1 * time.Minute

	// basepath/enclave uuid/service uuid <filetype>
	PerFileFmtStr = "%s%s/%s%s"

	// basepath /year/week
	PerWeekDirPathStr = "%s%s/%s/"

	// ... enclave uuid/service uuid <filetype>
	PerWeekFilePathFmtStr = PerWeekDirPathStr + "%s/%s%s"
)
