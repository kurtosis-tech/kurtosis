package service

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"net"
	"regexp"
)

const (
	// ServiceNameRegex implements RFC-1035 for naming services, namely:
	// * contain at most 63 characters
	// * contain only lowercase alphanumeric characters or '-'
	// * start with an alphabetic character
	// * end with an alphanumeric character
	// The adoption of RFC-1035 is to maintain compatability with current Kubernetes service and pod naming standards:
	// We use this over RFC-1035 as Service Names require 1035 to be followed
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-label-names
	// https://kubernetes.io/docs/concepts/services-networking/service/
	ServiceNameRegex            = "[a-z]([-a-z0-9]{0,61}[a-z0-9])?"
	WordWrappedServiceNameRegex = "^" + ServiceNameRegex + "$"
	serviceNameMaxLength        = 63
)

var (
	compiledWordWrappedServiceNameRegex = regexp.MustCompile(WordWrappedServiceNameRegex)
)

type ServiceName string
type ServiceUUID string

// Service represents a ServiceRegistration that has had a container bonded
// to it (in essence, Service is a "full" service where ServiceRegistration is a service stub)
type Service struct {
	registration *ServiceRegistration

	// Keyed by user-provided port ID
	privatePorts map[string]*port_spec.PortSpec

	// When running in Docker, the IP on the user's machine (outside the Docker VM) where this service can be reached
	// This will only be non-nil if both:
	// - The service's status is running
	// - The backend type is Docker
	maybePublicIp net.IP

	// When running in Docker, a mapping of service_port_id -> port_on_host_machine where the user can make requests to
	//  access the service (where host machine == outside the Docker VM)
	// This will only be non-nil if both:
	// - The service's status is running
	// - The backend type is Docker
	maybePublicPorts map[string]*port_spec.PortSpec

	// Docker container or Kubernetes pod container
	container *container.Container
}

func NewService(registration *ServiceRegistration, privatePorts map[string]*port_spec.PortSpec, maybePublicIp net.IP, maybePublicPorts map[string]*port_spec.PortSpec, container *container.Container) *Service {
	return &Service{registration: registration, privatePorts: privatePorts, maybePublicIp: maybePublicIp, maybePublicPorts: maybePublicPorts, container: container}
}

func (service *Service) GetRegistration() *ServiceRegistration {
	return service.registration
}

func (service *Service) GetContainer() *container.Container {
	return service.container
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

func IsServiceNameValid(serviceName ServiceName) bool {
	return compiledWordWrappedServiceNameRegex.MatchString(string(serviceName))
}
