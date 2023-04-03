package out

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os"
)

var (
	std = New()
)

type CliWriter struct {
	out io.Writer
	err io.Writer
}

func New() *CliWriter {
	return &CliWriter{
		out: os.Stdout,
		err: os.Stderr,
	}
}

func SetOut(writer io.Writer) {
	std.out = writer
}

func GetOut() io.Writer {
	return std.out
}

func SetErr(writer io.Writer) {
	std.err = writer
}

func GetErr() io.Writer {
	return std.err
}

func PrintOutLn(msg string) {
	if _, printErr := fmt.Fprintln(std.out, msg); printErr != nil {
		logrus.Errorf("Error printing message to StdOut. Message was:\n%s\nError was:\n%v", msg, printErr.Error())
	}

	printLogsToFile(msg)

}

func PrintErrLn(msg string) {
	if _, printErr := fmt.Fprintln(std.err, msg); printErr != nil {
		logrus.Errorf("Error printing message to StdErr. Message was:\n%s\nError was:\n%v", msg, printErr.Error())
	}

	printLogsToFile(msg)
}

func printLogsToFile(msg string) {
	fileLogger, err := GetFileLogger()
	if err != nil {
		// errors with file-logger needs to be logged carefully
		// file logger is used to log onto kurtosis-cli file which gets called whenever logrus default
		// logger is used due to the hook. It can be found in this method: `setupFileLogger`
		logrus.StandardLogger().Warnf("Error using the file logger as it failed with following error: %v. "+
			"This is a bug where logs are not being collected properly, please file an issue using `kurtosis feedback`", err)
	} else {
		fileLogger.Println(msg)
	}
}
