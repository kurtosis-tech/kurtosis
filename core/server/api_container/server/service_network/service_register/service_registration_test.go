package service_register

import (
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/stretchr/testify/require"
	"net"
	"testing"
)

func TestServiceRegistrationMarshallers(t *testing.T) {
	serviceConfig := service.GetServiceConfigForTest(t, "image-name:tag-version")
	privateIp := net.ParseIP("198.51.100.121")
	hostname := "hostname-test"

	startedServiceStatus := service.ServiceStatus_Started

	originalServiceRegistration := NewServiceRegistration(privateIp, hostname, startedServiceStatus, *serviceConfig)

	marshaledServiceRegistration, err := json.Marshal(originalServiceRegistration)
	require.NoError(t, err)
	require.NotNil(t, marshaledServiceRegistration)

	newServiceRegistration := &serviceRegistration{}

	err = json.Unmarshal(marshaledServiceRegistration, newServiceRegistration)
	require.NoError(t, err)

	require.EqualValues(t, originalServiceRegistration, newServiceRegistration)
}
