package service

import (
	"encoding/json"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
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

	require.EqualValues(t, originalServiceRegistration, newServiceRegistration)
}
