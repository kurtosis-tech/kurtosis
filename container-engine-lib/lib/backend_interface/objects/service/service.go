package service

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"net"
)

type ServiceID string
type ServiceGUID string

// Object that represents POINT-IN-TIME information about an user service
// Store this object and continue to reference it at your own risk!!!
type Service struct {
	id				 ServiceID
	guid             ServiceGUID
	status           UserServiceStatus
	enclaveId        enclave.EnclaveID

	privateIp        net.IP

	// Will be nil if the service is in UserServiceStatus_Registered state
	privatePorts     map[string]*port_spec.PortSpec // Keyed by user-provided port ID

	// These will only be non-nil if all of the following are true:
	//  - The backend is Docker
	//  - The service is in UserServiceStatus_Running state
	maybePublicIp    net.IP                         // The ip exposed in the host machine. Will be nil if the service doesn't declare any private ports
	maybePublicPorts map[string]*port_spec.PortSpec //Mapping of port-used-by-service -> port-on-the-host-machine where the user can make requests to the port to access the port. If a used port doesn't have a host port bound, then the value will be nil.
}

func (service *Service) GetID() ServiceID {
	return service.id
}

func (service *Service) GetGUID() ServiceGUID {
	return service.guid
}

func (service *Service) GetStatus() UserServiceStatus {
	return service.status
}

func (service *Service) GetEnclaveID() enclave.EnclaveID {
	return service.enclaveId
}

func (service *Service) GetPrivateIP() net.IP {
	return service.privateIp
}

func (service *Service) GetPrivatePorts() map[string]*port_spec.PortSpec {
	return service.privatePorts
}

func (service *Service) GetMaybePublicIP() net.IP {
	return service.maybePublicIp
}

func (service *Service) GetMaybePublicPorts() map[string]*port_spec.PortSpec {
	return service.maybePublicPorts
}