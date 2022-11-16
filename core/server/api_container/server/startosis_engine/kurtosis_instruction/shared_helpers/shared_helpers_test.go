package shared_helpers

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
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
