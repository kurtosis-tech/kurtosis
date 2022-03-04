package api_container

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"net"
)

// Represents point-in-time information about an API container
// WARNING: Store this at your own risk!
type APIContainer struct {
	// The ID of the enclave the API container manages
	enclaveId string

	status container_status.ContainerStatus

	// Public (i.e. external to Kurtosis) information about the API container
	// This information will be nil if the API container isn't running
	publicIpAddr net.IP
	publicGrpcPort *port_spec.PortSpec
	publicGrpcProxyPort *port_spec.PortSpec
}
