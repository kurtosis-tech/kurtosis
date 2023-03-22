package out

import (
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/host_machine_directories"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io/fs"
	"os"
	"sync"
)

var once sync.Once

// FileLogger this logger will only log to a file
// TODO: In commands like inspect we use out.PrintOutLn - will need to add this fileLogger
//  to print commands' output as well.
var fileLogger = logrus.New()
var permission = fs.FileMode(0666)

func SetupFileLogger() error {
	var err error
	once.Do(func() {
		err = setupFileLogger()
	})
	return err
}

func GetFileLogger() *logrus.Logger {
	if fileLogger == nil {
		SetupFileLogger()
	}
	return fileLogger
}

func setupFileLogger() error {
	// kurtosis-cli.log can be found in the same directory as kurtosis-config.yml
	filePath, err := host_machine_directories.GetKurtosisCliLogsFilePath()
	if err != nil {
		return stacktrace.Propagate(err, "Error occurred while getting the path of the log file - '%v'", filePath)
	}

	//TODO: store at least x number of commands in the file
	//TODO: print the command, and then the output of the command
	logFile, err := os.OpenFile(
		filePath,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		permission)

	if err != nil {
		return stacktrace.Propagate(err, "Error occurred while opening log file found here: '%v", filePath)
	}

	// this is the formatter using for fileLogger
	// TODO: Seeing weird characters in the log file - will look into fixing it in next PR
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

	fileLogger.SetOutput(logFile)
	fileLogger.SetLevel(logrus.InfoLevel)
	fileLogger.SetFormatter(textFormatter)

	logLevels := []logrus.Level{
		logrus.InfoLevel,
		logrus.WarnLevel,
		logrus.DebugLevel,
		logrus.ErrorLevel,
	}
	// added logrus hook which automatically make sure that all the logging is being done in the file as well
	// currently only info, warn, debug and error level logs are added to the file but we can add more later.
	logsHook := NewHook(logFile, logLevels, textFormatter)
	logrus.AddHook(&logsHook)

	return nil
}
