/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package networking_sidecar

import (
	"bytes"
	"context"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
)

// The API for the NetworkingSidecar class run exec commands against the actual Docker container that backs it
// This is a separate class because NetworkingSidecar shouldn't know about the underlying DockerManager used to run
//  the exec commands; it should be transparent to NetworkingSidecar
type sidecarExecCmdExecutor struct {
	dockerManager *docker_manager.DockerManager

	// Container ID of the sidecar container in which exec commands should run
	containerId string

	shWrappingCmd func([]string) []string
}

func newSidecarExecCmdExecutor(dockerManager *docker_manager.DockerManager, containerId string, shWrappingCmd func([]string) []string) *sidecarExecCmdExecutor {
	return &sidecarExecCmdExecutor{dockerManager: dockerManager, containerId: containerId, shWrappingCmd: shWrappingCmd}
}



func (executor sidecarExecCmdExecutor) exec(ctx context.Context, unwrappedCmd []string) error {
	shWrappedCmd := executor.shWrappingCmd(unwrappedCmd)

	execOutputBuf := &bytes.Buffer{}
	err := executor.dockerManager.RunExecCommand(
			ctx,
			executor.containerId,
			shWrappedCmd,
			execOutputBuf)
	if err !=  nil {
		logrus.Errorf("------------------ Exec command output for command '%v' --------------------", shWrappedCmd)
		if _, err := io.Copy(logrus.StandardLogger().Out, execOutputBuf); err != nil {
			logrus.Errorf("An error occurred printing the exec logs: %v", err)
		}
		logrus.Errorf("------------------ END Exec command output for command '%v' --------------------", shWrappedCmd)
		return stacktrace.Propagate(err, "An error occurred running exec command '%v'", shWrappedCmd)
	}
	return nil
}
