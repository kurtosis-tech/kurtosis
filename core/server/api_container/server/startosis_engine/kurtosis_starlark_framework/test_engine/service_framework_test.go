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

type serviceObjectTestCase struct {
	*testing.T
}

func (suite *KurtosisTypeConstructorTestSuite) TestServiceObject() {
	suite.run(&serviceObjectTestCase{
		T: suite.T(),
	})
}

func (t serviceObjectTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%q, %s=%q, %s=%q, %s=%s)",
		kurtosis_types.ServiceTypeName,
		kurtosis_types.ServiceNameAttr, testServiceName,
		kurtosis_types.HostnameAttr, testServiceHostname,
		kurtosis_types.IpAddressAttr, testServiceIpAddress,
		kurtosis_types.PortsAttr, testServicePorts)
}

func (t serviceObjectTestCase) Assert(typeValue builtin_argument.KurtosisValueType) {
	serviceStarlark, ok := typeValue.(*kurtosis_types.Service)
	require.True(t, ok)

	resultServiceName, interpretationErr := serviceStarlark.GetName()
	require.Nil(t, interpretationErr)
	require.Equal(t, testServiceName, resultServiceName)

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
