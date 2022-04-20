/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package networking_sidecar

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"strings"
)

// ==========================================================================================
//                                  Interface
// ==========================================================================================
// Extracted as an interface for testing
type sidecarExecCmdExecutor interface {
	exec(ctx context.Context, cmd []string) error
}

// ==========================================================================================
//                                  Implementation
// ==========================================================================================
// The API for the NetworkingSidecar class run exec commands against the Kurtosis Backend
// This is a separate class because NetworkingSidecar we need to create also a mock to test purpose
type standardSidecarExecCmdExecutor struct {
	kurtosisBackend backend_interface.KurtosisBackend

	// Service GUID of the networking sidecar in which exec commands should run
	serviceGUID service.ServiceGUID

	enclaveId enclave.EnclaveID
}

func newStandardSidecarExecCmdExecutor(kurtosisBackend backend_interface.KurtosisBackend, serviceGUID service.ServiceGUID, enclaveId enclave.EnclaveID) *standardSidecarExecCmdExecutor {
	return &standardSidecarExecCmdExecutor{kurtosisBackend: kurtosisBackend, serviceGUID: serviceGUID, enclaveId: enclaveId}
}

func (executor standardSidecarExecCmdExecutor) exec(ctx context.Context, notShWrappedCmd []string) error {

	shWrappedCmd := shWrapCommand(notShWrappedCmd)
	var (
		networkingSidecarCommands = map[service.ServiceGUID][]string{
			executor.serviceGUID: shWrappedCmd,
		}
	)

	_, erroredNetworkingSidecars, err := executor.kurtosisBackend.RunNetworkingSidecarExecCommands(
		ctx,
		executor.enclaveId,
		networkingSidecarCommands,
	)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred running exec command in networking sidecar with GUID '%v'", executor.serviceGUID)
	}
	if len(erroredNetworkingSidecars) > 0 {
		sidecarError, sidecarErrorFound := erroredNetworkingSidecars[executor.serviceGUID]
		if !sidecarErrorFound {
			return stacktrace.NewError("Unable to find error for networking sidecar with GUID '%v'. This is a bug in kurtosis", executor.serviceGUID)
		}

		return stacktrace.Propagate(sidecarError, "An error occurred running exec command in networking sidecar with GUID '%v'", executor.serviceGUID)
	}

	return nil
}

// Embeds the given command in a call to sh shell, so that a command with things
//  like '&&' will get executed as expected
func shWrapCommand(unwrappedCmd []string) []string {
	return []string{
		"sh",
		"-c",
		strings.Join(unwrappedCmd, " "),
	}
}