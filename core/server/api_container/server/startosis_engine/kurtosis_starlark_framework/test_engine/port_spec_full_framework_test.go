package test_engine

import (
	"fmt"
	port_spec2 "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/port_spec"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type portSpecFullTestCase struct {
	*testing.T
}

func (suite *KurtosisTypeConstructorTestSuite) TestPortSpecFull() {
	suite.run(&portSpecFullTestCase{
		T: suite.T(),
	})
}

func (t *portSpecFullTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%d, %s=%q, %s=%q, %s=%q)",
		port_spec.PortSpecTypeName,
		port_spec.PortNumberAttr,
		testPrivatePortNumber,
		port_spec.TransportProtocolAttr,
		testPrivatePortProtocolStr,
		port_spec.PortApplicationProtocolAttr,
		testPrivateApplicationProtocol,
		port_spec.WaitAttr,
		testWaitConfiguration,
	)
}

func (t *portSpecFullTestCase) Assert(typeValue builtin_argument.KurtosisValueType) {
	portSpecStarlark, ok := typeValue.(*port_spec.PortSpec)
	require.True(t, ok)
	portSpec, err := portSpecStarlark.ToKurtosisType()
	require.Nil(t, err)

	waitDuration, errParsingDuration := time.ParseDuration(testWaitConfiguration)
	require.NoError(t, errParsingDuration)
	expectedPortSpec, errPortCreation := port_spec2.NewPortSpec(testPrivatePortNumber, testPrivatePortProtocol, testPrivateApplicationProtocol, port_spec2.NewWait(waitDuration), "")
	require.NoError(t, errPortCreation)
	require.Equal(t, expectedPortSpec, portSpec)

}
