package service

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"net"
)

type ServiceID string
type ServiceGUID string

// Object that represents a ServiceRegistration that has had a container bonded
// to it (in essence, Service is a "full" service where ServiceRegistration is a service stub)
type Service struct {
	registration *ServiceRegistration

	status           container_status.ContainerStatus

	// Keyed by user-provided port ID
	privatePorts map[string]*port_spec.PortSpec

	// These will only be non-nil if the service's status is running
	maybePublicIp    net.IP                         // The ip exposed in the host machine. Will be nil if the service doesn't declare any private ports
	maybePublicPorts map[string]*port_spec.PortSpec //Mapping of port-used-by-service -> port-on-the-host-machine where the user can make requests to the port to access the port. If a used port doesn't have a host port bound, then the value will be nil.
}

func NewService(registration *ServiceRegistration, status container_status.ContainerStatus, privatePorts map[string]*port_spec.PortSpec, maybePublicIp net.IP, maybePublicPorts map[string]*port_spec.PortSpec) *Service {
	return &Service{registration: registration, status: status, privatePorts: privatePorts, maybePublicIp: maybePublicIp, maybePublicPorts: maybePublicPorts}
}

func (service *Service) GetRegistration() *ServiceRegistration {
	return service.registration
}

func (service *Service) GetStatus() container_status.ContainerStatus {
	return service.status
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