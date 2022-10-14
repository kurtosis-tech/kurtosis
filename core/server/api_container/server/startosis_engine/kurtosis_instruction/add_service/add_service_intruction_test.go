package add_service

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/stretchr/testify/require"
	"net"
	"regexp"
	"testing"
)

const (
	testContainerImageName = "kurtosistech/example-datastore-server"

	testServiceId                   = "tesT-SerVice-id"
	testServiceDependence1ServiceId = "test-service-id-1"
	testServiceDependence1IPAddress = "172.17.13.3"
	testServiceDependence2ServiceId = "test-service-id-2"
	testServiceDependence2IPAddress = "172.17.13.45"

	unknownServiceId = "unknown_service"
)

func TestAddServiceInstruction_GetCanonicalizedInstruction(t *testing.T) {
	addServiceInstruction := NewAddServiceInstruction(
		nil,
		*kurtosis_instruction.NewInstructionPosition(22, 26),
		"example-datastore-server-2",
		services.NewServiceConfigBuilder(
			testContainerImageName,
		).WithPrivatePorts(
			map[string]*kurtosis_core_rpc_api_bindings.Port{
				"grpc": {
					Number:   1323,
					Protocol: kurtosis_core_rpc_api_bindings.Port_TCP,
				},
			},
		).Build(),
	)

	// TODO: Update when we implement this
	expectedOutput := "add_service(...)"
	require.Equal(t, expectedOutput, addServiceInstruction.GetCanonicalInstruction())
}

func TestAddServiceInstruction_EntryPointArgsAreReplaced(t *testing.T) {
	ipAddresses := map[service.ServiceID]net.IP{
		"foo_service": net.ParseIP("172.17.3.13"),
	}
	serviceNetwork := service_network.NewMockServiceNetwork(ipAddresses)
	addServiceInstruction := NewAddServiceInstruction(
		serviceNetwork,
		*kurtosis_instruction.NewInstructionPosition(22, 26),
		"example-datastore-server-2",
		services.NewServiceConfigBuilder(
			testContainerImageName,
		).WithPrivatePorts(
			map[string]*kurtosis_core_rpc_api_bindings.Port{
				"grpc": {
					Number:   1323,
					Protocol: kurtosis_core_rpc_api_bindings.Port_TCP,
				},
			},
		).WithEntryPointArgs(
			[]string{"-- {{kurtosis:foo_service.ip_address}}"},
		).Build(),
	)

	err := addServiceInstruction.replaceIPAddress()
	require.Nil(t, err)
	require.Equal(t, "-- 172.17.3.13", addServiceInstruction.serviceConfig.EntrypointArgs[0])
}

func TestReplaceIPAddressInString_MultipleOccurrencesOfSameStringReplaced(t *testing.T) {
	ipAddresses := map[service.ServiceID]net.IP{
		testServiceDependence1ServiceId: net.ParseIP(testServiceDependence1IPAddress),
	}
	serviceNetwork := service_network.NewMockServiceNetwork(ipAddresses)
	originalString := fmt.Sprintf("{{kurtosis:%v.ip_address}} something in the middle {{kurtosis:%v.ip_address}}", testServiceDependence1ServiceId, testServiceDependence1ServiceId)

	expectedString := fmt.Sprintf("%v something in the middle %v", testServiceDependence1IPAddress, testServiceDependence1IPAddress)
	replacedString, err := replaceIPAddressInString(originalString, serviceNetwork, testServiceId)
	require.Nil(t, err)
	require.Equal(t, expectedString, replacedString)
}

func TestReplaceIPAddressInString_MultipleReplacesOfDifferentStrings(t *testing.T) {
	ipAddresses := map[service.ServiceID]net.IP{
		testServiceDependence1ServiceId: net.ParseIP(testServiceDependence1IPAddress),
		testServiceDependence2ServiceId: net.ParseIP(testServiceDependence2IPAddress),
	}
	serviceNetwork := service_network.NewMockServiceNetwork(ipAddresses)
	originalString := fmt.Sprintf("{{kurtosis:%v.ip_address}} {{kurtosis:%v.ip_address}} {{kurtosis:%v.ip_address}}", testServiceDependence1ServiceId, testServiceDependence2ServiceId, testServiceDependence1ServiceId)

	expectedString := fmt.Sprintf("%v %v %v", testServiceDependence1IPAddress, testServiceDependence2IPAddress, testServiceDependence1IPAddress)
	replacedString, err := replaceIPAddressInString(originalString, serviceNetwork, testServiceId)
	require.Nil(t, err)
	require.Equal(t, expectedString, replacedString)
}

func TestReplaceIPAddressInString_ReplacementFailsForUnknownServiceId(t *testing.T) {
	ipAddresses := map[service.ServiceID]net.IP{}
	serviceNetwork := service_network.NewMockServiceNetwork(ipAddresses)
	originalString := fmt.Sprintf("{{kurtosis:%v.ip_address}}", unknownServiceId)

	expectedErr := fmt.Sprintf("'%v' depends on the IP address of '%v' but we don't have any registrations for it", testServiceId, unknownServiceId)
	_, err := replaceIPAddressInString(originalString, serviceNetwork, testServiceId)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), expectedErr)
}

func TestReplaceIPAddressInString_EnforceRegexAndPlaceholderAlign(t *testing.T) {
	ipAddressPlaceholder := fmt.Sprintf(ipAddressReplacementPlaceholderFormat, testServiceId)
	regex := regexp.MustCompile(ipAddressReplacementRegex)
	hasMatches := regex.MatchString(ipAddressPlaceholder)
	require.True(t, hasMatches)
}
