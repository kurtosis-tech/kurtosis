package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	port_spec_starlark "github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/port_spec"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type portSpecMinimalTestCase struct {
	*testing.T
}

func (suite *KurtosisTypeConstructorTestSuite) TestPortSpecMinimal() {
	suite.run(&portSpecMinimalTestCase{
		T: suite.T(),
	})
}

func (t *portSpecMinimalTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%d)", port_spec_starlark.PortSpecTypeName, port_spec_starlark.PortNumberAttr, testPrivatePortNumber)
}

func (t *portSpecMinimalTestCase) Assert(typeValue builtin_argument.KurtosisValueType) {
	portSpecStarlark, ok := typeValue.(*port_spec_starlark.PortSpec)
	require.True(t, ok)
	portSpec, err := portSpecStarlark.ToKurtosisType()
	require.Nil(t, err)

	waitDuration, errParsingDuration := time.ParseDuration(testWaitDefaultValue)
	require.NoError(t, errParsingDuration)
	expectedPortSpec, errPortCreation := port_spec.NewPortSpec(testPrivatePortNumber, port_spec.TransportProtocol_TCP, "", port_spec.NewWait(waitDuration), "")
	require.NoError(t, errPortCreation)
	require.Equal(t, expectedPortSpec, portSpec)

}
