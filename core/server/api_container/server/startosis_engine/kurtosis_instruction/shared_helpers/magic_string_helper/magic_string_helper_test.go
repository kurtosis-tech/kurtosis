package magic_string_helper

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
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

	testFactName = "test-fact-name"

	unknownServiceId = "unknown_service"

	testStringRuntimeValue         = starlark.String("test_string")
	testRuntimeValueField          = "field"
	testExpectedInterpolatedString = starlark.String("test_string is not 0")
)

var testIntRuntimeValue = starlark.MakeInt(0)

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

func TestReplaceFactInString(t *testing.T) {
	returnValue := MakeWaitInterpretationReturnValue(testServiceId, testFactName)
	require.Equal(t, returnValue.String(), "\"{{kurtosis:tesT-SerVice-id:test-fact-name.fact}}\"")
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

func TestGetRuntimeValueFromString_BasicFetch(t *testing.T) {
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	stringValueUuid := runtimeValueStore.CreateValue()
	runtimeValueStore.SetValue(stringValueUuid, map[string]starlark.Comparable{testRuntimeValueField: testStringRuntimeValue})
	intValueUuid := runtimeValueStore.CreateValue()
	runtimeValueStore.SetValue(intValueUuid, map[string]starlark.Comparable{testRuntimeValueField: testIntRuntimeValue})
	fetchedStringValue, err := GetRuntimeValueFromString(fmt.Sprintf(RuntimeValueReplacementPlaceholderFormat, stringValueUuid, testRuntimeValueField), runtimeValueStore)
	require.Nil(t, err)
	require.Equal(t, fetchedStringValue, testStringRuntimeValue)
	fetchedIntValue, err := GetRuntimeValueFromString(fmt.Sprintf(RuntimeValueReplacementPlaceholderFormat, intValueUuid, testRuntimeValueField), runtimeValueStore)
	require.Nil(t, err)
	require.Equal(t, fetchedIntValue, testIntRuntimeValue)
}

func TestGetRuntimeValueFromString_Interpolated(t *testing.T) {
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	stringValueUuid := runtimeValueStore.CreateValue()
	runtimeValueStore.SetValue(stringValueUuid, map[string]starlark.Comparable{testRuntimeValueField: testStringRuntimeValue})
	intValueUuid := runtimeValueStore.CreateValue()
	runtimeValueStore.SetValue(intValueUuid, map[string]starlark.Comparable{testRuntimeValueField: testIntRuntimeValue})
	stringRuntimeValue := fmt.Sprintf(RuntimeValueReplacementPlaceholderFormat, stringValueUuid, testRuntimeValueField)
	intRuntimeValue := fmt.Sprintf(RuntimeValueReplacementPlaceholderFormat, intValueUuid, testRuntimeValueField)
	interpolatedString := fmt.Sprintf("%v is not %v", stringRuntimeValue, intRuntimeValue)
	resolvedInterpolatedString, err := GetRuntimeValueFromString(interpolatedString, runtimeValueStore)
	require.Nil(t, err)
	require.Equal(t, resolvedInterpolatedString, testExpectedInterpolatedString)
}
