package service

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"net"
)

// A ServiceRegistration is a stub for a soon-to-be-started service
// We had created this to return the user's IP before the container exists
// Now this is used for caches in APIC & DockerKurtosisBackend
// ToDo visit removing this  after partitioning is moved to container-engine-lib and data is stored in a database
type ServiceRegistration struct {
	id				 ServiceID
	guid             ServiceGUID
	enclaveId        enclave.EnclaveID

	// The private IP is the IP of the service within the enclave, meaning other services can use this IP to communicate
	// with the service
	privateIp        net.IP
}

func NewServiceRegistration(id ServiceID, guid ServiceGUID, enclaveId enclave.EnclaveID, privateIp net.IP) *ServiceRegistration {
	return &ServiceRegistration{id: id, guid: guid, enclaveId: enclaveId, privateIp: privateIp}
}

func (registration *ServiceRegistration) GetID() ServiceID {
	return registration.id
}

func (registration *ServiceRegistration) GetGUID() ServiceGUID {
	return registration.guid
}

func (registration *ServiceRegistration) GetEnclaveID() enclave.EnclaveID {
	return registration.enclaveId
}

func (registration *ServiceRegistration) GetPrivateIP() net.IP {
	return registration.privateIp
}
