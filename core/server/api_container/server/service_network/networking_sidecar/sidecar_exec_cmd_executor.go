/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package networking_sidecar

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/libs/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/libs/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/libs/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"strings"
)

const (
	successExitCode = 0
)

// ==========================================================================================
//
//	Interface
//
// ==========================================================================================
// Extracted as an interface for testing
type sidecarExecCmdExecutor interface {
	exec(ctx context.Context, cmd []string) error
}

// ==========================================================================================
//
//	Implementation
//
// ==========================================================================================
// The API for the NetworkingSidecar class run exec commands against the Kurtosis Backend
// This is a separate class because NetworkingSidecar we need to create also a mock to test purpose
type standardSidecarExecCmdExecutor struct {
	kurtosisBackend backend_interface.KurtosisBackend

	// Service GUID of the networking sidecar in which exec commands should run
	serviceUUID service.ServiceUUID

	enclaveUuid enclave.EnclaveUUID
}

func newStandardSidecarExecCmdExecutor(kurtosisBackend backend_interface.KurtosisBackend, serviceUUID service.ServiceUUID, enclaveUuid enclave.EnclaveUUID) *standardSidecarExecCmdExecutor {
	return &standardSidecarExecCmdExecutor{kurtosisBackend: kurtosisBackend, serviceUUID: serviceUUID, enclaveUuid: enclaveUuid}
}

func (executor standardSidecarExecCmdExecutor) exec(ctx context.Context, notShWrappedCmd []string) error {

	shWrappedCmd := shWrapCommand(notShWrappedCmd)
	var (
		networkingSidecarCommands = map[service.ServiceUUID][]string{
			executor.serviceUUID: shWrappedCmd,
		}
	)

	successfulNetworkingSidecarExecResults, erroredNetworkingSidecars, err := executor.kurtosisBackend.RunNetworkingSidecarExecCommands(
		ctx,
		executor.enclaveUuid,
		networkingSidecarCommands,
	)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred running exec command in networking sidecar with UUID '%v'", executor.serviceUUID)
	}
	if len(erroredNetworkingSidecars) > 0 {
		sidecarError, sidecarErrorFound := erroredNetworkingSidecars[executor.serviceUUID]
		if !sidecarErrorFound {
			return stacktrace.NewError("Unable to find error for networking sidecar with UUID '%v'. This is a bug in kurtosis", executor.serviceUUID)
		}

		return stacktrace.Propagate(sidecarError, "An error occurred running exec command in networking sidecar with UUID '%v'", executor.serviceUUID)
	}
	execResult, found := successfulNetworkingSidecarExecResults[executor.serviceUUID]
	if !found {
		return stacktrace.NewError("Expected to receive the execution result information after running commands from '%+v' for service with UUID '%v'; but none was found", successfulNetworkingSidecarExecResults, executor.serviceUUID)
	}

	if execResult.GetExitCode() != successExitCode {
		return stacktrace.NewError("Executing commands '%+v' returned an failing exit code with output:\n%v", networkingSidecarCommands, execResult.GetOutput())
	}

	return nil
}

// Embeds the given command in a call to sh shell, so that a command with things
//
//	like '&&' will get executed as expected
func shWrapCommand(unwrappedCmd []string) []string {
	return []string{
		"sh",
		"-c",
		strings.Join(unwrappedCmd, " "),
	}
}
