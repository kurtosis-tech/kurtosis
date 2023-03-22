package out

import (
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/host_machine_directories"
	"github.com/sirupsen/logrus"
	"os"
)

var FileLogger = logrus.New()

func InitFileLogger() {
	filePath, err := host_machine_directories.GetKurtosisCliLogsFilePath()
	if err != nil {
		panic(err)
	}

	//TODO: store at least x number of commands in the file
	//TODO: print the command, and then the output of the command
	logFile, err := os.OpenFile(
		filePath,
		os.O_WRONLY|os.O_CREATE,
		0666)

	if err != nil {
		panic(err)
	}

	textFormatter := &logrus.TextFormatter{
		ForceColors:               true,
		DisableColors:             false,
		ForceQuote:                false,
		DisableQuote:              false,
		EnvironmentOverrideColors: false,
		DisableTimestamp:          false,
		FullTimestamp:             false,
		TimestampFormat:           "",
		DisableSorting:            true,
		SortingFunc:               nil,
		DisableLevelTruncation:    false,
		PadLevelText:              false,
		QuoteEmptyFields:          false,
		FieldMap:                  nil,
		CallerPrettyfier:          nil,
	}

	FileLogger.SetOutput(logFile)
	FileLogger.SetLevel(logrus.InfoLevel)
	FileLogger.SetFormatter(textFormatter)

	logrus.AddHook(&Hook{
		Writer: logFile,
		LogLevels: []logrus.Level{
			logrus.InfoLevel,
			logrus.WarnLevel,
			logrus.DebugLevel,
			logrus.ErrorLevel,
		},
		Formatter: textFormatter,
	})
}
