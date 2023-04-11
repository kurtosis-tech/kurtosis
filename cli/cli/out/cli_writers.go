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
	fileLogger := GetFileLogger()
	fileLogger.Println(msg)
}
