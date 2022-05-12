package user_service_registration

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"net"
)

type UserServiceRegistration struct {
	enclaveId enclave.EnclaveID
	id ServiceID
	ipAddress net.IP
}

func NewUserServiceRegistration(enclaveId enclave.EnclaveID, id ServiceID, ipAddress net.IP) *UserServiceRegistration {
	return &UserServiceRegistration{enclaveId: enclaveId, id: id, ipAddress: ipAddress}
}

func (registration *UserServiceRegistration) GetEnclaveID() enclave.EnclaveID {
	return registration.enclaveId
}
func (registration *UserServiceRegistration) GetID() ServiceID {
	return registration.id
}
func (registration *UserServiceRegistration) GetIPAddress() net.IP {
	return registration.ipAddress
}