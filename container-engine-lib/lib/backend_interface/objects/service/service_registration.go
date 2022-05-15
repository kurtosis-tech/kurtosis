package service

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"net"
)

// A ServiceRegistration is a stub for a soon-to-be-started service
// We have this as an independent object so we can return the container's IP
// address to the user before the container exists
type ServiceRegistration struct {
	id				 ServiceID
	guid             ServiceGUID
	enclaveId        enclave.EnclaveID
	privateIp        net.IP
}
