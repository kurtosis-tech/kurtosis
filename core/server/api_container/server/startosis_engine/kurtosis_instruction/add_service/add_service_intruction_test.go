package add_service

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/stretchr/testify/require"
	"net"
	"testing"
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

	err := addServiceInstruction.ReplaceIPAddress()
	require.Equal(t, addServiceInstruction.serviceConfig.EntrypointArgs[0], "-- 172.17.3.13")
	require.Nil(t, err)

}

func TestAddServiceInstruction_MultipleOccurrencesOfSameStringReplaced(t *testing.T) {
	ipAddresses := map[service.ServiceID]net.IP{}
	ipAddresses["example_service"] = net.ParseIP("172.17.3.13")
	serviceNetwork := service_network.NewMockServiceNetwork(ipAddresses)
	serviceID := "test service id"
	originalString := "{{example_service.ip_address}} something in the middle {{example_service.ip_address}}"
	replacedString, err := replaceIPAddressInString(originalString, serviceNetwork, serviceID)
	require.Nil(t, err)
	require.Equal(t, "172.17.3.13 something in the middle 172.17.3.13", replacedString)
}

func TestAddServiceInstruction_MultipleReplacesOfStrings(t *testing.T) {
	ipAddresses := map[service.ServiceID]net.IP{}
	ipAddresses["example_service"] = net.ParseIP("172.17.3.13")
	ipAddresses["different_service"] = net.ParseIP("172.17.3.169")
	serviceNetwork := service_network.NewMockServiceNetwork(ipAddresses)
	placeHolderServiceID := "test service id"
	originalString := "{{example_service.ip_address}} {{different_service.ip_address}} {{example_service.ip_address}}"
	replacedString, err := replaceIPAddressInString(originalString, serviceNetwork, placeHolderServiceID)
	require.Nil(t, err)
	require.Equal(t, "172.17.3.13 172.17.3.169 172.17.3.13", replacedString)
}
