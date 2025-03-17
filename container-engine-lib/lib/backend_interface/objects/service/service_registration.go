package service

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/kurtosis-tech/stacktrace"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
)

// A ServiceRegistration is a stub for a started service
// We had created this to return the user's IP before the container exists
// The user doesn't need the private IP address anymore.
// Now this is used for caches in APIC & DockerKurtosisBackend
type ServiceRegistration struct {
	// we do this way in order to have exported fields which can be marshalled
	// and an unexported type for encapsulation
	privateServiceRegistration *privateServiceRegistration
}

type privateServiceRegistration struct {
	Name        ServiceName
	Uuid        ServiceUUID
	EnclaveUuid enclave.EnclaveUUID

	// The private IP is the IP of the service within the enclave, meaning other services can use this IP to communicate
	// with the service
	PrivateIp net.IP

	// The hostname differs whether Kurtosis is using a Kubernetes backend of a Docker backend.
	// In Docker, we set the serviceName as the hostname, whereas in Kubernetes the hostname is automatically assigned
	// to the "Kubernetes Service Name", which is something like user-services-<SERVICE_UUID>
	// TODO: for consistency we should probably set the hostname to user-services-<SERVICE_UUID> in docker as well
	Hostname string

	// Service state: registered, started or stopped
	Status ServiceStatus

	// Service config used during service creation
	// Used when the service is restarted
	Config *ServiceConfig
}

func NewServiceRegistration(name ServiceName, uuid ServiceUUID, enclaveUuid enclave.EnclaveUUID, privateIp net.IP, hostname string) *ServiceRegistration {
	privateServiceRegistrationObj := &privateServiceRegistration{
		Name:        name,
		Uuid:        uuid,
		EnclaveUuid: enclaveUuid,
		PrivateIp:   privateIp,
		Hostname:    hostname,
		Status:      ServiceStatus_Registered,
		Config:      nil,
	}
	return &ServiceRegistration{
		privateServiceRegistration: privateServiceRegistrationObj,
	}
}

func (registration *ServiceRegistration) GetName() ServiceName {
	return registration.privateServiceRegistration.Name
}

func (registration *ServiceRegistration) GetUUID() ServiceUUID {
	return registration.privateServiceRegistration.Uuid
}

func (registration *ServiceRegistration) GetEnclaveID() enclave.EnclaveUUID {
	return registration.privateServiceRegistration.EnclaveUuid
}

func (registration *ServiceRegistration) GetPrivateIP() net.IP {
	return registration.privateServiceRegistration.PrivateIp
}

// GetPrivateIPAddress is an alias for GetPrivateIP for backward compatibility
func (registration *ServiceRegistration) GetPrivateIPAddress() net.IP {
	return registration.GetPrivateIP()
}

func (registration *ServiceRegistration) GetHostname() string {
	return registration.privateServiceRegistration.Hostname
}

func (registration *ServiceRegistration) GetStatus() ServiceStatus {
	return registration.privateServiceRegistration.Status
}

func (registration *ServiceRegistration) SetStatus(status ServiceStatus) {
	registration.privateServiceRegistration.Status = status
}

func (registration *ServiceRegistration) GetConfig() *ServiceConfig {
	fmt.Println("GetConfig() call")
	fmt.Println("reg: ", registration)
	fmt.Println("svcr: ", registration.privateServiceRegistration)
	fmt.Println("Config: ", registration.privateServiceRegistration.Config)
	return registration.privateServiceRegistration.Config
}

func (registration *ServiceRegistration) SetConfig(config *ServiceConfig) {
	registration.privateServiceRegistration.Config = config
}

func (registration *ServiceRegistration) MarshalJSON() ([]byte, error) {
	return json.Marshal(registration.privateServiceRegistration)
}

func (registration *ServiceRegistration) UnmarshalJSON(data []byte) error {

	// Suppressing exhaustruct requirement because we want an object with zero values
	// nolint: exhaustruct
	unmarshalledPrivateStructPtr := &privateServiceRegistration{}

	if err := json.Unmarshal(data, unmarshalledPrivateStructPtr); err != nil {
		return stacktrace.Propagate(err, "An error occurred unmarshalling the private struct")
	}

	registration.privateServiceRegistration = unmarshalledPrivateStructPtr
	return nil
}
