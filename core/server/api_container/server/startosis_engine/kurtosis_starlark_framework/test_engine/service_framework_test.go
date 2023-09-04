package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testServiceIpAddress = "192.168.0.43"
	testServiceHostname  = "test-service-hostname"
	testServicePorts     = "{}"
)

type serviceTestCase struct {
	*testing.T
}

func newServiceTestCase(t *testing.T) *serviceTestCase {
	return &serviceTestCase{
		T: t,
	}
}

func (t serviceTestCase) GetId() string {
	return kurtosis_types.ServiceTypeName
}

func (t serviceTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%q, %s=%q, %s=%q, %s=%s)",
		kurtosis_types.ServiceTypeName,
		kurtosis_types.ServiceNameAttr, TestServiceName,
		kurtosis_types.HostnameAttr, testServiceHostname,
		kurtosis_types.IpAddressAttr, testServiceIpAddress,
		kurtosis_types.PortsAttr, testServicePorts)
}

func (t serviceTestCase) Assert(typeValue builtin_argument.KurtosisValueType) {
	serviceStarlark, ok := typeValue.(*kurtosis_types.Service)
	require.True(t, ok)

	resultServiceName, interpretationErr := serviceStarlark.GetName()
	require.Nil(t, interpretationErr)
	require.Equal(t, TestServiceName, resultServiceName)

	resultServiceHostname, interpretationErr := serviceStarlark.GetHostname()
	require.Nil(t, interpretationErr)
	require.Equal(t, testServiceHostname, resultServiceHostname)

	resultServiceIpAddress, interpretationErr := serviceStarlark.GetIpAddress()
	require.Nil(t, interpretationErr)
	require.Equal(t, testServiceIpAddress, resultServiceIpAddress)

	resultPorts, interpretationErr := serviceStarlark.GetPorts()
	require.Nil(t, interpretationErr)
	require.Equal(t, 0, resultPorts.Len())
}
