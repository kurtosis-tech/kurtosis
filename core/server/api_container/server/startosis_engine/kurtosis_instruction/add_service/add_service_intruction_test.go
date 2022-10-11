package add_service

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/stretchr/testify/require"
	"net"
	"testing"
)

const (
	testServiceID                   = "test-service-id"
	testServiceDependence1ServiceID = "test-service-id-1"
	testServiceDependence1IPAddress = "172.17.13.3"
	testServiceDependence2ServiceID = "test-service-id-2"
	testServiceDependence2IPAddress = "172.17.13.45"

	unknownServiceID = "unknown_service"
)

func TestAddServiceInstruction_GetCanonicalizedInstruction(t *testing.T) {
	addServiceInstruction := NewAddServiceInstruction(
		nil,
		*kurtosis_instruction.NewInstructionPosition(22, 26),
		service.ServiceID("example-datastore-server-2"),
		&kurtosis_core_rpc_api_bindings.ServiceConfig{
			ContainerImageName: "kurtosistech/example-datastore-server",
			PrivatePorts: map[string]*kurtosis_core_rpc_api_bindings.Port{
				"grpc": {
					Number:   1325,
					Protocol: kurtosis_core_rpc_api_bindings.Port_TCP,
				},
			},
		})

	// TODO: Update when we implement this
	expectedOutput := "add_service(...)"
	require.Equal(t, expectedOutput, addServiceInstruction.GetCanonicalInstruction())
}

func TestAddServiceInstruction_EntryPointArgsAreReplaced(t *testing.T) {
	ipAddresses := map[service.ServiceID]net.IP{}
	ipAddresses["foo_service"] = net.ParseIP("172.17.3.13")
	serviceNetwork := service_network.NewMockServiceNetwork(ipAddresses)
	addServiceInstruction := NewAddServiceInstruction(
		serviceNetwork,
		*kurtosis_instruction.NewInstructionPosition(22, 26),
		service.ServiceID("example-datastore-server-2"),
		&kurtosis_core_rpc_api_bindings.ServiceConfig{
			ContainerImageName: "kurtosistech/example-datastore-server",
			PrivatePorts: map[string]*kurtosis_core_rpc_api_bindings.Port{
				"grpc": {
					Number:   1325,
					Protocol: kurtosis_core_rpc_api_bindings.Port_TCP,
				},
			},
			EntrypointArgs: []string{"-- {{foo_service.ip_address}}"},
		})

	err := addServiceInstruction.replaceIPAddress()
	require.Nil(t, err)
	require.Equal(t, addServiceInstruction.serviceConfig.EntrypointArgs[0], "-- 172.17.3.13")
}

func TestAddServiceInstruction_MultipleOccurrencesOfSameStringReplaced(t *testing.T) {
	ipAddresses := map[service.ServiceID]net.IP{
		testServiceDependence1ServiceID: net.ParseIP(testServiceDependence1IPAddress),
	}
	serviceNetwork := service_network.NewMockServiceNetwork(ipAddresses)
	originalString := fmt.Sprintf("{{%v.ip_address}} something in the middle {{%v.ip_address}}", testServiceDependence1ServiceID, testServiceDependence1ServiceID)

	expectedString := fmt.Sprintf("%v something in the middle %v", testServiceDependence1IPAddress, testServiceDependence1IPAddress)
	replacedString, err := replaceIPAddressInString(originalString, serviceNetwork, testServiceID)
	require.Nil(t, err)
	require.Equal(t, expectedString, replacedString)
}

func TestReplaceIPAddressInString_MultipleReplacesOfDifferentStrings(t *testing.T) {
	ipAddresses := map[service.ServiceID]net.IP{
		testServiceDependence1ServiceID: net.ParseIP(testServiceDependence1IPAddress),
		testServiceDependence2ServiceID: net.ParseIP(testServiceDependence2IPAddress),
	}
	serviceNetwork := service_network.NewMockServiceNetwork(ipAddresses)
	originalString := fmt.Sprintf("{{%v.ip_address}} {{%v.ip_address}} {{%v.ip_address}}", testServiceDependence1ServiceID, testServiceDependence2ServiceID, testServiceDependence1ServiceID)

	expectedString := fmt.Sprintf("%v %v %v", testServiceDependence1IPAddress, testServiceDependence2IPAddress, testServiceDependence1IPAddress)
	replacedString, err := replaceIPAddressInString(originalString, serviceNetwork, testServiceID)
	require.Nil(t, err)
	require.Equal(t, expectedString, replacedString)
}

func TestReplaceIPAddressInString_ReplacementFailsForUnknownServiceID(t *testing.T) {
	ipAddresses := map[service.ServiceID]net.IP{}
	serviceNetwork := service_network.NewMockServiceNetwork(ipAddresses)
	originalString := fmt.Sprintf("{{%v.ip_address}}", unknownServiceID)

	expectedErr := fmt.Sprintf("'%v' depends on the IP address of '%v' but we don't have any registrations for it", testServiceID, unknownServiceID)
	_, err := replaceIPAddressInString(originalString, serviceNetwork, testServiceID)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), expectedErr)
}
