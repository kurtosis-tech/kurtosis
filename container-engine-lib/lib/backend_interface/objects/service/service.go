package service

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/user_service_registration"
	"net"
)

type ServiceGUID string

// Object that represents POINT-IN-TIME information about an user service
// Store this object and continue to reference it at your own risk!!!
type Service struct {
	// The GUID of the registration that the service consumed to start
	registrationGuid user_service_registration.UserServiceRegistrationGUID
	guid             ServiceGUID
	status           container_status.ContainerStatus
	enclaveId        enclave.EnclaveID
	privateIp        net.IP
	privatePorts     map[string]*port_spec.PortSpec // Keyed by user-provided port ID
	maybePublicIp    net.IP                         // The ip exposed in the host machine. Will be nil if the service doesn't declare any private ports
	maybePublicPorts map[string]*port_spec.PortSpec //Mapping of port-used-by-service -> port-on-the-host-machine where the user can make requests to the port to access the port. If a used port doesn't have a host port bound, then the value will be nil.
}

func NewService(registrationGuid user_service_registration.UserServiceRegistrationGUID, guid ServiceGUID, status container_status.ContainerStatus, enclaveId enclave.EnclaveID, privateIp net.IP, privatePorts map[string]*port_spec.PortSpec, maybePublicIp net.IP, maybePublicPorts map[string]*port_spec.PortSpec) *Service {
	return &Service{registrationGuid: registrationGuid, guid: guid, status: status, enclaveId: enclaveId, privateIp: privateIp, privatePorts: privatePorts, maybePublicIp: maybePublicIp, maybePublicPorts: maybePublicPorts}
}

func (service *Service) GetRegistrationGUID() user_service_registration.UserServiceRegistrationGUID {
	return service.registrationGuid
}

func (service *Service) GetGUID() ServiceGUID {
	return service.guid
}

func (service *Service) GetStatus() container_status.ContainerStatus {
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

