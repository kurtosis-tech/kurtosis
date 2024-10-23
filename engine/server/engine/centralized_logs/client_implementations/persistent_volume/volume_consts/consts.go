package volume_consts

const (
	// Location of logs on the filesystem of the engine
	LogsStorageDirpath = "/var/log/kurtosis/"

	Filetype = ".json"

	NewLineRune = '\n'

	LogLabel       = "log"
	TimestampLabel = "timestamp"

	EndOfJsonLine = "}"

	// basepath/enclave uuid/service uuid <filetype>
	PerFileFmtStr = "%s%s/%s%s"
)
