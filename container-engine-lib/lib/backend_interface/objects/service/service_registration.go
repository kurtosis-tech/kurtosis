package service

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"net"
)

// A ServiceRegistration is a stub for a started service
// We had created this to return the user's IP before the container exists
// The user doesn't need the private IP address anymore.
// Now this is used for caches in APIC & DockerKurtosisBackend
// Also partitioning in APIC is based on the private IP address this returns
// TODO visit removing this  after partitioning is moved to container-engine-lib and data is stored in a database
type ServiceRegistration struct {
	name      ServiceName
	uuid      ServiceUUID
	enclaveId enclave.EnclaveUUID

	// The private IP is the IP of the service within the enclave, meaning other services can use this IP to communicate
	// with the service
	privateIp net.IP
}

func NewServiceRegistration(name ServiceName, guid ServiceUUID, enclaveId enclave.EnclaveUUID, privateIp net.IP) *ServiceRegistration {
	return &ServiceRegistration{name: name, uuid: guid, enclaveId: enclaveId, privateIp: privateIp}
}

func (registration *ServiceRegistration) GetName() ServiceName {
	return registration.name
}

func (registration *ServiceRegistration) GetUUID() ServiceUUID {
	return registration.uuid
}

func (registration *ServiceRegistration) GetEnclaveID() enclave.EnclaveUUID {
	return registration.enclaveId
}

func (registration *ServiceRegistration) GetPrivateIP() net.IP {
	return registration.privateIp
}
