/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package networking_sidecar

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/libs/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/libs/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/libs/container-engine-lib/lib/backend_interface/objects/networking_sidecar"
	"github.com/kurtosis-tech/kurtosis/libs/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
)

// ==========================================================================================
//
//	Interface
//
// ==========================================================================================
type NetworkingSidecarManager interface {
	Add(ctx context.Context, serviceId service.ServiceUUID) (NetworkingSidecarWrapper, error)
	Remove(ctx context.Context, sidecar NetworkingSidecarWrapper) error
}

// ==========================================================================================
//                                      Implementation
// ==========================================================================================

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
// This class's methods are NOT thread-safe - it's up to the caller to ensure that
//
//	only one change at a time is run on a given sidecar container!!!
//
// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
type StandardNetworkingSidecarManager struct {
	kurtosisBackend backend_interface.KurtosisBackend

	enclaveUuid enclave.EnclaveUUID
}

func NewStandardNetworkingSidecarManager(kurtosisBackend backend_interface.KurtosisBackend, enclaveId enclave.EnclaveUUID) *StandardNetworkingSidecarManager {
	return &StandardNetworkingSidecarManager{kurtosisBackend: kurtosisBackend, enclaveUuid: enclaveId}
}

// Adds a sidecar container attached to the given service ID
func (manager *StandardNetworkingSidecarManager) Add(
	ctx context.Context,
	serviceUUID service.ServiceUUID,
) (NetworkingSidecarWrapper, error) {

	networkingSidecar, err := manager.kurtosisBackend.CreateNetworkingSidecar(ctx, manager.enclaveUuid, serviceUUID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating networking sidecar for service with UUID '%v' in enclave with ID '%v'", serviceUUID, manager.enclaveUuid)
	}

	execCmdExecutor := newStandardSidecarExecCmdExecutor(
		manager.kurtosisBackend,
		networkingSidecar.GetServiceUUID(),
		networkingSidecar.GetEnclaveUUID())

	networkingSidecarWrapper, err := NewStandardNetworkingSidecarWrapper(networkingSidecar, execCmdExecutor)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating networking sidecar wrapper for networking sidecar with service UUID '%v'", networkingSidecar.GetServiceUUID())
	}

	return networkingSidecarWrapper, nil
}

func (manager *StandardNetworkingSidecarManager) Remove(
	ctx context.Context,
	networkingSidecarWrapper NetworkingSidecarWrapper) error {
	networkingSidecarServiceUUID := networkingSidecarWrapper.GetServiceUUID()

	filters := &networking_sidecar.NetworkingSidecarFilters{
		EnclaveUUIDs: nil,
		UserServiceUUIDs: map[service.ServiceUUID]bool{
			networkingSidecarServiceUUID: true,
		},
		Statuses: nil,
	}

	_, erroredNetworkingSidecars, err := manager.kurtosisBackend.StopNetworkingSidecars(ctx, filters)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred stopping networking sidecars using filter '%+v'", filters)
	}
	if len(erroredNetworkingSidecars) > 0 {
		sidecarError, sidecarErrorFound := erroredNetworkingSidecars[networkingSidecarServiceUUID]
		if !sidecarErrorFound {
			return stacktrace.NewError("Unable to find error for networking sidecar with GUID '%v'. This is a bug in kurtosis", networkingSidecarServiceUUID)
		}
		return stacktrace.Propagate(sidecarError, "An error occurred stopping networking sidecar with GUID '%v'", networkingSidecarServiceUUID)
	}

	return nil
}
