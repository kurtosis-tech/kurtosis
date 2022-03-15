package service

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"net"
)

type ServiceID string
type ServiceGUID string

// Object that represents POINT-IN-TIME information about an user service
// Store this object and continue to reference it at your own risk!!!
type Service struct {
	id ServiceID
	guid ServiceGUID
	enclaveId enclave.
	maybePublicIpAddr net.IP // The ip exposed in the host machine. Will be nil if the service doesn't declare any private ports
	publicPorts map[string]*port_spec.PortSpec //Mapping of port-used-by-service -> port-on-the-host-machine where the user can make requests to the port to access the port. If a used port doesn't have a host port bound, then the value will be nil.
}

func NewService(id ServiceID, guid ServiceGUID, maybePublicIpAddr net.IP, publicPorts map[string]*port_spec.PortSpec) *Service {
	return &Service{id: id, guid: guid, maybePublicIpAddr: maybePublicIpAddr, publicPorts: publicPorts}
}

func (service *Service) GetID() ServiceID {
	return service.id
}

func (service *Service) GetGUID() ServiceGUID {
	return service.guid
}

func (service *Service) GetMaybePublicIpAddr() net.IP {
	return service.maybePublicIpAddr
}

func (service *Service) GetPublicPorts() map[string]*port_spec.PortSpec {
	return service.publicPorts
}
