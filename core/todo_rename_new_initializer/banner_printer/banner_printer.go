package banner_printer

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"strings"
)

/*
Little helper function to print a container logfile with with banners indicating the start and end of the logs

Args:
	containerDescription: Short, human-readable description of the container whose logs are being printed
	logFilepath: Filepath of the file containing the container's logs
*/
func PrintContainerLogsWithBanners(containerDescription string, logFilepath string) {
	containerDescUppercase := strings.ToUpper(containerDescription)
	logrus.Info("- - - - - - - - - - - - - " + containerDescUppercase + " LOGS - - - - - - - - - - - - -")
	fp, err := os.Open(logFilepath)
	if err != nil {
		logrus.Errorf("Could not print the test suite container's logs due to the following error when opening the file:")
		fmt.Fprintln(logrus.StandardLogger().Out, err)
	}
	defer fp.Close()
	if _, err := io.Copy(logrus.StandardLogger().Out, fp); err != nil {
		logrus.Errorf("Could not print the test suite container's logs due to the following error when copying logfile contents:")
		fmt.Fprintln(logrus.StandardLogger().Out, err)
	}
	logrus.Info("- - - - - - - - - - - - " + containerDescUppercase + " LOGS - - - - - - - - - - - - -")
}

