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

// ==========================================================================================
//                                  Interface
// ==========================================================================================
// Extracted as an interface for testing
type sidecarExecCmdExecutor interface {
	exec(ctx context.Context, unwrappedCmd []string) error
}

// ==========================================================================================
//                                  Implementation
// ==========================================================================================
// The API for the NetworkingSidecar class run exec commands against the actual Docker container that backs it
// This is a separate class because NetworkingSidecar shouldn't know about the underlying DockerManager used to run
//  the exec commands; it should be transparent to NetworkingSidecar
type standardSidecarExecCmdExecutor struct {
	dockerManager *docker_manager.DockerManager

	// Container ID of the sidecar container in which exec commands should run
	sidecarContainerId string

	shWrappingCmd func([]string) []string
}

func newStandardSidecarExecCmdExecutor(dockerManager *docker_manager.DockerManager, sidecarContainerId string, shWrappingCmd func([]string) []string) *standardSidecarExecCmdExecutor {
	return &standardSidecarExecCmdExecutor{dockerManager: dockerManager, sidecarContainerId: sidecarContainerId, shWrappingCmd: shWrappingCmd}
}



func (executor standardSidecarExecCmdExecutor) exec(ctx context.Context, unwrappedCmd []string) error {
	shWrappedCmd := executor.shWrappingCmd(unwrappedCmd)

	execOutputBuf := &bytes.Buffer{}
	err := executor.dockerManager.RunExecCommand(
			ctx,
			executor.sidecarContainerId,
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
