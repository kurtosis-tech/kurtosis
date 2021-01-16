/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package sidecar_container_manager

import (
	"bytes"
	"context"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
)

// The API for the SidecarContainer class run exec commands against the actual Docker container that backs it
// This is a separate class because SidecarContainer shouldn't know about the underlying DockerManager used to run
//  the exec commands; it should be transparent to SidecarContainer
type SidecarExecCmdExecutor struct {
	dockerManager *docker_manager.DockerManager

	// Container ID of the sidecar container in which exec commands should run
	containerId string

	shWrappingCmd func([]string) []string
}

func (executor SidecarExecCmdExecutor) exec(ctx context.Context, unwrappedCmd []string) error {
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
