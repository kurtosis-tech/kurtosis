package shared_helpers

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_executor"
	"github.com/stretchr/testify/require"
	"net"
	"regexp"
	"testing"
)

const (
	testServiceId                   = "tesT-SerVice-id"
	testServiceDependence1ServiceId = "test-service-id-1"
	testServiceDependence1IPAddress = "172.17.13.3"
	testServiceDependence2ServiceId = "test-service-id-2"
	testServiceDependence2IPAddress = "172.17.13.45"

	unknownServiceId = "unknown_service"
)

func TestReplaceMagicStringWithValue_SimpleCase(t *testing.T) {
	instruction := kurtosis_instruction.NewInstructionPosition(5, 3, "dummyFile")
	inputStr := instruction.MagicString(ArtifactUUIDSuffix)
	environment := startosis_executor.NewExecutionEnvironment()
	testUuid := "test-uuid"
	environment.SetArtifactUuid(inputStr, testUuid)

	expectedOutput := testUuid
	replacedStr, err := ReplaceArtifactUuidMagicStringWithValue(inputStr, testServiceId, environment)

	require.Nil(t, err)
	require.Equal(t, expectedOutput, replacedStr)
}

func TestReplaceMagicStringWithValue_ValidMultipleReplaces(t *testing.T) {
	instructionA := kurtosis_instruction.NewInstructionPosition(45, 60, "github.com/kurtosis-tech/eth2-module/src/participant_network/prelaunch_data_generator/el_genesis/el_genesis_data_generator.star")
	instructionB := kurtosis_instruction.NewInstructionPosition(56, 33, "github.com/kurtosis-tech/eth2-module/src/participant_network/prelaunch_data_generator/cl_genesis/cl_genesis_data_generator.star")
	magicStringA := instructionA.MagicString(ArtifactUUIDSuffix)
	magicStringB := instructionB.MagicString(ArtifactUUIDSuffix)
	inputStr := fmt.Sprintf("%v %v %v", magicStringB, magicStringA, magicStringB)

	environment := startosis_executor.NewExecutionEnvironment()
	testUuidA := "test-uuid-a"
	testUuidB := "test-uuid-b"
	expectedOutput := fmt.Sprintf("%v %v %v", testUuidB, testUuidA, testUuidB)

	environment.SetArtifactUuid(magicStringA, testUuidA)
	environment.SetArtifactUuid(magicStringB, testUuidB)

	replacedStr, err := ReplaceArtifactUuidMagicStringWithValue(inputStr, testServiceId, environment)
	require.Nil(t, err)
	require.Equal(t, expectedOutput, replacedStr)
}

func TestReplaceMagicStringWithValue_MagicStringNotInEnvironment(t *testing.T) {
	instruction := kurtosis_instruction.NewInstructionPosition(5, 3, "dummyFile")
	magicString := instruction.MagicString(ArtifactUUIDSuffix)
	emptyEnvironment := startosis_executor.NewExecutionEnvironment()
	_, err := ReplaceArtifactUuidMagicStringWithValue(magicString, testServiceId, emptyEnvironment)
	require.NotNil(t, err)
}

func TestReplaceIPAddressInString_MultipleOccurrencesOfSameStringReplaced(t *testing.T) {
	ipAddresses := map[service.ServiceID]net.IP{
		testServiceDependence1ServiceId: net.ParseIP(testServiceDependence1IPAddress),
	}
	serviceNetwork := service_network.NewMockServiceNetwork(ipAddresses)
	originalString := fmt.Sprintf("{{kurtosis:%v.ip_address}} something in the middle {{kurtosis:%v.ip_address}}", testServiceDependence1ServiceId, testServiceDependence1ServiceId)

	expectedString := fmt.Sprintf("%v something in the middle %v", testServiceDependence1IPAddress, testServiceDependence1IPAddress)
	replacedString, err := ReplaceIPAddressInString(originalString, serviceNetwork, testServiceId)
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
	replacedString, err := ReplaceIPAddressInString(originalString, serviceNetwork, testServiceId)
	require.Nil(t, err)
	require.Equal(t, expectedString, replacedString)
}

func TestReplaceIPAddressInString_ReplacementFailsForUnknownServiceId(t *testing.T) {
	ipAddresses := map[service.ServiceID]net.IP{}
	serviceNetwork := service_network.NewMockServiceNetwork(ipAddresses)
	originalString := fmt.Sprintf("{{kurtosis:%v.ip_address}}", unknownServiceId)

	expectedErr := fmt.Sprintf("'%v' depends on the IP address of '%v' but we don't have any registrations for it", testServiceId, unknownServiceId)
	_, err := ReplaceIPAddressInString(originalString, serviceNetwork, testServiceId)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), expectedErr)
}

func TestReplaceIPAddressInString_EnforceRegexAndPlaceholderAlign(t *testing.T) {
	ipAddressPlaceholder := fmt.Sprintf(IpAddressReplacementPlaceholderFormat, testServiceId)
	regex := regexp.MustCompile(ipAddressReplacementRegex)
	hasMatches := regex.MatchString(ipAddressPlaceholder)
	require.True(t, hasMatches)
}
