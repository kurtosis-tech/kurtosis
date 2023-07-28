package out

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/host_machine_directories"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io/fs"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	once                      sync.Once
	kurtosisCliLogsRootFolder = ""
	cliLogNotSet              = ""
)

const (
	fileNamePrefix              = "kurtosis-cli"
	numberOfLogFilesForCommands = 7
	readWriteEveryonePermission = 0666
)

// TODO: In commands like inspect we use out.PrintOutLn - will need to add this fileLogger to print commands' output as well.
// fileLogger this logger will only log to a file
var fileLogger *logrus.Logger
var permission = fs.FileMode(readWriteEveryonePermission)

func GetFileLogger() *logrus.Logger {
	if fileLogger == nil {
		once.Do(func() {
			setupFileLogger()
		})
	}
	return fileLogger
}

func RemoveLogFiles() {
	if kurtosisCliLogsRootFolder == cliLogNotSet {
		return
	}

	var files []os.DirEntry
	var err error

	files, err = os.ReadDir(kurtosisCliLogsRootFolder)
	if err != nil {
		logrus.Warnf("Error occurred while listing the files from the directory at - '%v' with error %+v", kurtosisCliLogsRootFolder, err)
	}

	numberOfFilesToBeDeleted := len(files) - numberOfLogFilesForCommands
	logrus.Debugf("Removing old %v log file(s) from path %v", numberOfFilesToBeDeleted, kurtosisCliLogsRootFolder)

	// filter out log files that starts with fileNamePrefix image in the kurtosis/cli folder
	var filesWithKurtosisCliPrefix []os.DirEntry
	if len(files) > numberOfLogFilesForCommands {
		for i := 0; i < numberOfFilesToBeDeleted; i++ {
			fileName := files[i].Name()
			if strings.HasPrefix(fileName, fileNamePrefix) {
				filesWithKurtosisCliPrefix = append(filesWithKurtosisCliPrefix, files[i])
			}
		}
	}

	// remove all the extra files from the kurtosis/cli folder that starts with fileNamePrefix
	if len(filesWithKurtosisCliPrefix) > 0 {
		for _, file := range filesWithKurtosisCliPrefix {
			fileName := file.Name()
			logFilePath := path.Join(kurtosisCliLogsRootFolder, fileName)

			logrus.Debugf("Removing log file at %v", logFilePath)
			if err := os.Remove(logFilePath); err != nil {
				logrus.Warnf("Error occurred while removing the log file - '%v' with error %+v", logFilePath, err)
				break
			}
		}
	}

	// should have had this file in a cli folder but did not do it
	// added this logic to remove kurtosis-cli.log in kurtosis folder
	// TODO: remove this next week
	kurtosisRootFolder := path.Dir(kurtosisCliLogsRootFolder)
	kurtosisCliLogPath := path.Join(kurtosisRootFolder, fmt.Sprintf("%v.log", fileNamePrefix))
	if _, err := os.Stat(kurtosisCliLogPath); err == nil {
		if err := os.Remove(kurtosisCliLogPath); err != nil {
			logrus.Warnf("Error occurred while removing the log file - '%v' with error %+v", kurtosisCliLogPath, err)
		}
	}
}

// helper method that creates log file for a command invocation
func createLogFile() (*os.File, error) {
	timestamp := strconv.FormatInt(time.Now().UnixMicro(), 10)
	fileName := fmt.Sprintf("%v-%v.log", fileNamePrefix, timestamp)

	logFilePath, err := host_machine_directories.GetKurtosisCliLogsFileDirPath(fileName)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error occurred while getting the path of the cli log dir - '%v'", logFilePath)
	}

	logFile, err := os.OpenFile(
		logFilePath,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		permission)

	if err != nil {
		return nil, stacktrace.Propagate(err, "Error occurred while opening log file found here: '%v", logFilePath)
	}

	return logFile, nil
}

func setupFileLogger() {
	logFile, err := createLogFile()
	// silently log the error as warn if there are issues creating log file
	if err != nil {
		logrus.Warnf("Error occurred while getting the file logger %+v", err)
	}

	// this is the formatter using for fileLogger
	textFormatter := &logrus.TextFormatter{
		ForceColors:               false,
		DisableColors:             true,
		ForceQuote:                false,
		DisableQuote:              true,
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

	// creates a files logger which logs data onto the log file
	fileLogger = logrus.New()
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

	// store the path of the dir where the file was created, this will be used by remove file method to clean
	kurtosisCliLogsRootFolder = path.Dir(logFile.Name())
}
