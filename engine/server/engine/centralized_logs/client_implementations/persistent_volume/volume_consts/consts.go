package volume_consts

import "strings"

const (
	// Location of logs on the filesystem of the engine
	LogsStorageDirpath = "/var/log/kurtosis/"

	Filetype = ".json"

	NewLineRune = '\n'

	LogLabel       = "log"
	TimestampLabel = "timestamp"

	EndOfJsonLine = "}"

	LogRetentionPeriodInWeeks = 4

	// basepath/enclave uuid/service uuid <filetype>
	PerFileFmtStr = "%s%s/%s%s"

	// basepath /year/week
	PerWeekDirPathStr = "%s%s/%s/"

	// ... enclave uuid/service uuid <filetype>
	PerWeekFilePathFmtStr = PerWeekDirPathStr + "%s/%s%s"
)

var (
	// We use this storage dir path for tests because fstest.MapFS doesn't like absolute paths
	// when using fstest.MapFS, all paths need to be stored and retrieved as a relative path
	// so we trim the leading forward slash
	LogsStorageDirpathForTests = strings.TrimLeft(LogsStorageDirpath, "/")
)
