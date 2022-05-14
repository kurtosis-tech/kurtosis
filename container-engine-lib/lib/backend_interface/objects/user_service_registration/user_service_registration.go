package user_service_registration

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"net"
)

// Unique across all enclaves
type UserServiceRegistrationGUID string

// A user service represents a fixed IP and a domain name for a given user service
// This can be thought of as a Kubernetes Service, though in Docker it's implemented a bit differently
type UserServiceRegistration struct {
	// Unique ID representing this registration
	// It's important that each registration has its own GUID so that we can tag containers/pods with the registration
	// GUID which will:
	//  1) allow us to delete any containers/pods consuming the registration if we delete the registration
	//  2) give a unique ID to return registrations by
	guid UserServiceRegistrationGUID

	// The ID to which the registration is associated
	enclaveId enclave.EnclaveID

	// The ID of the service within the enclave that the registration represents
	id ServiceID

	ipAddress net.IP
}

func NewUserServiceRegistration(guid UserServiceRegistrationGUID, enclaveId enclave.EnclaveID, id ServiceID, ipAddress net.IP) *UserServiceRegistration {
	return &UserServiceRegistration{guid: guid, enclaveId: enclaveId, id: id, ipAddress: ipAddress}
}

func (registration *UserServiceRegistration) GetGUID() UserServiceRegistrationGUID {
	return registration.guid
}
func (registration *UserServiceRegistration) GetEnclaveID() enclave.EnclaveID {
	return registration.enclaveId
}
func (registration *UserServiceRegistration) GetServiceID() ServiceID {
	return registration.id
}
func (registration *UserServiceRegistration) GetIPAddress() net.IP {
	return registration.ipAddress
}