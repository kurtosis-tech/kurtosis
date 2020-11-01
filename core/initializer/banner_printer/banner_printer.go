/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package banner_printer

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/sirupsen/logrus"
	"io"
	"strings"
)

/*
Little helper function to print a container's logs with with banners indicating the start and end of the logs

Args:
	dockerManager: Docker manager to use when retrieving container logs
	context: The context in which to run the log retrieval Docker function
	containerId: ID of the Docker container from which to retrieve logs
	containerDescription: Short, human-readable description of the container whose logs are being printed
	logFilepath: Filepath of the file containing the container's logs
*/
func PrintContainerLogsWithBanners(
		dockerManager commons.DockerManager,
		context context.Context,
		containerId string,
		log *logrus.Logger,
		containerDescription string) {
	var logReader io.Reader
	logReadCloser, err := dockerManager.GetContainerLogs(context, containerId)
	if err != nil {
		errStr := fmt.Sprintf("Could not print container's logs due to the following error: %v", err)
		logReader = strings.NewReader(errStr)
	} else {
		defer logReadCloser.Close()
		logReader = logReadCloser
	}

	containerDescUppercase := strings.ToUpper(containerDescription)
	log.Info("- - - - - - - - - - - - - " + containerDescUppercase + " LOGS - - - - - - - - - - - - -")
	if _, err := io.Copy(log.Out, logReader); err != nil {
		log.Errorf("Could not print the test suite container's logs due to the following error when copying log contents:")
		fmt.Fprintln(log.Out, err)
	}
	log.Info("- - - - - - - - - - - - " + containerDescUppercase + " LOGS - - - - - - - - - - - - -")
}

