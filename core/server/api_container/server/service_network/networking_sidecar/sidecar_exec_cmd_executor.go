/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package networking_sidecar

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/networking_sidecar"
	"github.com/kurtosis-tech/stacktrace"
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

	// GUID of the networking sidecar in which exec commands should run
	networkingSidecarGuid networking_sidecar.NetworkingSidecarGUID

	enclaveId enclave.EnclaveID
}

func newStandardSidecarExecCmdExecutor(kurtosisBackend backend_interface.KurtosisBackend, networkingSidecarGuid networking_sidecar.NetworkingSidecarGUID, enclaveId enclave.EnclaveID) *standardSidecarExecCmdExecutor {
	return &standardSidecarExecCmdExecutor{kurtosisBackend: kurtosisBackend, networkingSidecarGuid: networkingSidecarGuid, enclaveId: enclaveId}
}

func (executor standardSidecarExecCmdExecutor) exec(ctx context.Context, cmd []string) error {

	var (
		networkingSidecarCommands = map[networking_sidecar.NetworkingSidecarGUID][]string{
			executor.networkingSidecarGuid: cmd,
		}
	)

	_, erroredNetworkingSidecars, err := executor.kurtosisBackend.RunNetworkingSidecarsExecCommand(
		ctx,
		executor.enclaveId,
		networkingSidecarCommands,
	)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred running exec command in networking sidecar with GUID '%v'", executor.networkingSidecarGuid)
	}
	if len(erroredNetworkingSidecars) > 0 {
		sidecarError := erroredNetworkingSidecars[executor.networkingSidecarGuid]
		return stacktrace.Propagate(sidecarError, "An error occurred running exec command in networking sidecar with GUID '%v'", executor.networkingSidecarGuid)
	}

	return nil
}
