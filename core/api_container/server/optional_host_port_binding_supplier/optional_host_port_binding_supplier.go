/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package optional_host_port_binding_supplier

import (
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/commons/free_host_port_binding_supplier"
	"github.com/palantir/stacktrace"
)

// This struct wraps a potentially-nil FreeHostPortBindingSupplier, and will translate a container's used ports and
//  provide host port bindings if necessary
type OptionalHostPortBindingSupplier struct {
	// If nil, no free host port bindings will be supplied
	hostPortBindingSupplier *free_host_port_binding_supplier.FreeHostPortBindingSupplier
}

func NewOptionalHostPortBindingSupplier(hostPortBindingSupplier *free_host_port_binding_supplier.FreeHostPortBindingSupplier) *OptionalHostPortBindingSupplier {
	return &OptionalHostPortBindingSupplier{hostPortBindingSupplier: hostPortBindingSupplier}
}

// Translates a container's used port set into the necessary map for starting a container, filling in host port bindings
//  if appropriate
func (supplier OptionalHostPortBindingSupplier) BindPortsToHostIfNeeded(
		usedPorts map[nat.Port]bool) (map[nat.Port]*nat.PortBinding, error) {
	// The value to this map can be nil to indicate no host port binding
	hostPortBindingsForDocker := map[nat.Port]*nat.PortBinding{} // Docker requires a present key to declare a used port, and a possibly-optional nil value
	for port, _ := range usedPorts {
		var dockerBindingToUse *nat.PortBinding = nil
		if supplier.hostPortBindingSupplier != nil {
			freeBinding, err := supplier.hostPortBindingSupplier.GetFreePortBinding()
			if err != nil {
				return nil, stacktrace.Propagate(
					err,
					"Host port binding was requested for port '%v', but an error occurred getting a free host port to bind to it",
					port.Port(),
				)
			}
			dockerBindingToUse = freeBinding
		}
		hostPortBindingsForDocker[port] = dockerBindingToUse
	}
	return hostPortBindingsForDocker, nil
}


