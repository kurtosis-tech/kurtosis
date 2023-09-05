package volume_consts

const (
	// Location of logs on the filesystem of the engine
	LogsStorageDirpath = "/var/log/kurtosis/"

	Filetype = ".json"

	NewLineRune = '\n'

	LogLabel = "log"

	EndOfJsonLine = "}\n"

	LogRetentionPeriodInWeeks = 4

	// basepath/enclave uuid/service uuid <filetype>
	PerFileFmtStr = "%s%s/%s%s"

	// basepath/year/week/enclave uuid/service uuid <filetype>
	PerWeekFmtStr = "%s%s/%s/%s/%s%s"
)
