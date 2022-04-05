package service

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"net"
)

type ServiceID string
type ServiceGUID string

// Object that represents POINT-IN-TIME information about an user service
// Store this object and continue to reference it at your own risk!!!
type Service struct {
	id            ServiceID
	guid          ServiceGUID
	status        container_status.ContainerStatus
	enclaveId     enclave.EnclaveID
	maybePublicIp net.IP                         // The ip exposed in the host machine. Will be nil if the service doesn't declare any private ports
	publicPorts   map[string]*port_spec.PortSpec //Mapping of port-used-by-service -> port-on-the-host-machine where the user can make requests to the port to access the port. If a used port doesn't have a host port bound, then the value will be nil.
	privateIp     net.IP
}

func NewService(id ServiceID, guid ServiceGUID, status container_status.ContainerStatus, enclaveId enclave.EnclaveID, maybePublicIp net.IP, publicPorts map[string]*port_spec.PortSpec, privateIp net.IP) *Service {
	return &Service{id: id, guid: guid, status: status, enclaveId: enclaveId, maybePublicIp: maybePublicIp, publicPorts: publicPorts, privateIp: privateIp}
}

func (service *Service) GetID() ServiceID {
	return service.id
}

func (service *Service) GetGUID() ServiceGUID {
	return service.guid
}

func (service *Service) GetEnclaveID() enclave.EnclaveID {
	return service.enclaveId
}

func (service *Service) GetMaybePublicIP() net.IP {
	return service.maybePublicIp
}

func (service *Service) GetPublicPorts() map[string]*port_spec.PortSpec {
	return service.publicPorts
}

func (service *Service) GetStatus() container_status.ContainerStatus {
	return service.status
}

func (service *Service) GetPrivateIP() net.IP {
	return service.privateIp
}
