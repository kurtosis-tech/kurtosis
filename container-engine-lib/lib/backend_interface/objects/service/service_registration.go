package service

import (
	"net"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
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

	// The hostname differs whether Kurtosis is using a Kubernetes backend of a Docker backend.
	// In Docker, we set the serviceName as the hostname, whereas in Kubernetes the hostname is automatically assigned
	// to the "Kubernetes Service Name", which is something like user-services-<SERVICE_UUID>
	// TODO: for consistency we should probably set the hostname to user-services-<SERVICE_UUID> in docker as well
	hostname string

	// Service state: registered, started or stopped
	status ServiceStatus

	// Service config used during service creation
	// Used when the service is restarted
	config *ServiceConfig
}

func NewServiceRegistration(name ServiceName, guid ServiceUUID, enclaveId enclave.EnclaveUUID, privateIp net.IP, hostname string) *ServiceRegistration {
	return &ServiceRegistration{
		name:      name,
		uuid:      guid,
		enclaveId: enclaveId,
		privateIp: privateIp,
		hostname:  hostname,
		status:    ServiceStatus_Registered,
		config:    nil,
	}
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

func (registration *ServiceRegistration) GetHostname() string {
	return registration.hostname
}

func (registration *ServiceRegistration) GetStatus() ServiceStatus {
	return registration.status
}

func (registration *ServiceRegistration) SetStatus(status ServiceStatus) {
	registration.status = status
}

func (registration *ServiceRegistration) GetConfig() *ServiceConfig {
	return registration.config
}

func (registration *ServiceRegistration) SetConfig(config *ServiceConfig) {
	registration.config = config
}
