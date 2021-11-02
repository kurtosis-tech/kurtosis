/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package networking_sidecar

import (
	"bytes"
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
)

const (
	succesfulExecCmdExitCode = 0
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
	exitCode, err := executor.dockerManager.RunExecCommand(
			ctx,
			executor.sidecarContainerId,
			shWrappedCmd,
			execOutputBuf)
	if err != nil || exitCode != succesfulExecCmdExitCode {
		logrus.Errorf("------------------ Exec command output for command '%v' --------------------", shWrappedCmd)
		if _, err := io.Copy(logrus.StandardLogger().Out, execOutputBuf); err != nil {
			logrus.Errorf("An error occurred printing the exec logs: %v", err)
		}
		logrus.Errorf("------------------ END Exec command output for command '%v' --------------------", shWrappedCmd)
		var resultErr error
		if err != nil {
			resultErr = stacktrace.Propagate(err, "An error occurred running exec command '%v'", shWrappedCmd)
		}
		if exitCode != succesfulExecCmdExitCode {
			resultErr = stacktrace.NewError(
				"Expected exit code '%v' when running exec command '%v', but got exit code '%v' instead",
				succesfulExecCmdExitCode,
				shWrappedCmd,
				exitCode)
		}
		return resultErr
	}
	return nil
}
