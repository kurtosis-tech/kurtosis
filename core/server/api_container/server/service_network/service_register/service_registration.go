package service_register

import (
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"net"
)

const (
	privateIpKey = "private_ip"
	hostnameKey  = "hostname"
	statusKey    = "service_status"
	configKey    = "service_config"
)

type serviceRegistration struct {
	privateIp net.IP
	hostname  string
	status    service.ServiceStatus
	config    service.ServiceConfig
}

func NewServiceRegistration(privateIp net.IP, hostname string, status service.ServiceStatus, config service.ServiceConfig) *serviceRegistration {
	return &serviceRegistration{privateIp: privateIp, hostname: hostname, status: status, config: config}
}

func (serviceRegistration *serviceRegistration) GetPrivateIp() net.IP {
	return serviceRegistration.privateIp
}

func (serviceRegistration *serviceRegistration) GetHostname() string {
	return serviceRegistration.hostname
}

func (serviceRegistration *serviceRegistration) GetStatus() service.ServiceStatus {
	return serviceRegistration.status
}

func (serviceRegistration *serviceRegistration) GetConfig() service.ServiceConfig {
	return serviceRegistration.config
}

func (serviceRegistration *serviceRegistration) MarshalJSON() ([]byte, error) {

	marshalledConfig, err := json.Marshal(serviceRegistration.config)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while marshalling service config '%+v'", serviceRegistration.config)
	}

	data := map[string][]byte{
		privateIpKey: []byte(serviceRegistration.privateIp.String()),
		hostnameKey:  []byte(serviceRegistration.hostname),
		statusKey:    []byte(serviceRegistration.status.String()),
		configKey:    marshalledConfig,
	}

	return json.Marshal(data)
}

func (serviceRegistration *serviceRegistration) UnmarshalJSON(data []byte) error {

	unmarshalledMapPtr := &map[string][]byte{}

	if err := json.Unmarshal(data, unmarshalledMapPtr); err != nil {
		return stacktrace.Propagate(err, "An error occurred unmarshalling map")
	}

	unmarshalledMap := *unmarshalledMapPtr

	privateIpStrBytes, found := unmarshalledMap[privateIpKey]
	if !found {
		return stacktrace.NewError("Expected to find key '%v' on map '%+v' but it was not found, this is a bug in Kurtosis", privateIpKey, unmarshalledMap)
	}
	privateIpStr := string(privateIpStrBytes)
	privateIp := net.ParseIP(privateIpStr)
	if privateIp == nil {
		return stacktrace.NewError("An error occurred while parsing private ip address string '%s', a nil value was returned, this is a bug in Kurtosis", privateIpStr)
	}

	hostnameBytes, found := unmarshalledMap[hostnameKey]
	if !found {
		return stacktrace.NewError("Expected to find key '%v' on map '%+v' but it was not found, this is a bug in Kurtosis", hostnameKey, unmarshalledMap)
	}
	hostname := string(hostnameBytes)

	serviceStatusStrBytes, found := unmarshalledMap[statusKey]
	if !found {
		return stacktrace.NewError("Expected to find key '%v' on map '%+v' but it was not found, this is a bug in Kurtosis", statusKey, unmarshalledMap)
	}
	serviceStatusStr := string(serviceStatusStrBytes)
	serviceStatus, err := service.ServiceStatusString(serviceStatusStr)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while casting string '%s' to service status", serviceStatusStr)
	}

	configBytes, found := unmarshalledMap[configKey]
	if !found {
		return stacktrace.NewError("Expected to find key '%v' on map '%+v' but it was not found, this is a bug in Kurtosis", configKey, unmarshalledMap)
	}
	serviceConfig := &service.ServiceConfig{}
	if err := json.Unmarshal(configBytes, serviceConfig); err != nil {
		return stacktrace.Propagate(err, "An error occurred unmarshalling service config")
	}

	serviceRegistration.privateIp = privateIp
	serviceRegistration.hostname = hostname
	serviceRegistration.status = serviceStatus
	serviceRegistration.config = *serviceConfig

	return nil
}
