package service

import (
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/stretchr/testify/require"
	"net"
	"testing"
)

func TestServiceRegistrationMarshallers(t *testing.T) {
	serviceName := ServiceName("service-name-test")
	uuid := ServiceUUID("cddc2ea3948149d9afa2ef93abb4ec54")
	enclaveUuid := enclave.EnclaveUUID("9e5c8bf2fbeb4de68f647280b1c79cbb")
	privateIp := net.ParseIP("198.51.100.121")
	hostname := "hostname-test"

	originalServiceRegistration := NewServiceRegistration(serviceName, uuid, enclaveUuid, privateIp, hostname)

	startedServiceStatus := ServiceStatus_Started
	serviceConfig := getServiceConfigForTest(t, "image-name:tag-version")

	originalServiceRegistration.SetStatus(startedServiceStatus)
	originalServiceRegistration.SetConfig(serviceConfig)

	marshaledServiceRegistration, err := json.Marshal(originalServiceRegistration)
	require.NoError(t, err)
	require.NotNil(t, marshaledServiceRegistration)

	// Suppressing exhaustruct requirement because we want an object with zero values
	// nolint: exhaustruct
	newServiceRegistration := &ServiceRegistration{}

	err = json.Unmarshal(marshaledServiceRegistration, newServiceRegistration)
	require.NoError(t, err)

	// Compare fields individually instead of entire object
	require.Equal(t, originalServiceRegistration.GetName(), newServiceRegistration.GetName())
	require.Equal(t, originalServiceRegistration.GetUUID(), newServiceRegistration.GetUUID())
	require.Equal(t, originalServiceRegistration.GetEnclaveID(), newServiceRegistration.GetEnclaveID())
	require.Equal(t, originalServiceRegistration.GetPrivateIP(), newServiceRegistration.GetPrivateIP())
	require.Equal(t, originalServiceRegistration.GetHostname(), newServiceRegistration.GetHostname())
	require.Equal(t, originalServiceRegistration.GetStatus(), newServiceRegistration.GetStatus())
	
	// Check that config exists
	require.NotNil(t, newServiceRegistration.GetConfig())
	
	// Compare important fields of the config
	require.Equal(t, originalServiceRegistration.GetConfig().GetContainerImageName(), 
		newServiceRegistration.GetConfig().GetContainerImageName())
}
