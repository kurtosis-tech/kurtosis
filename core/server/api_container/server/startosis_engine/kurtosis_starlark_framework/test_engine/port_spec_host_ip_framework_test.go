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

const testPortSpecHostIp = "0.0.0.0"

type portSpecHostIpTestCase struct {
	*testing.T
}

func (suite *KurtosisTypeConstructorTestSuite) TestPortSpecHostIp() {
	suite.run(&portSpecHostIpTestCase{
		T: suite.T(),
	})
}

func (t *portSpecHostIpTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%d, %s=%q)",
		port_spec.PortSpecTypeName,
		port_spec.PortNumberAttr,
		testPrivatePortNumber,
		port_spec.HostIpAttr,
		testPortSpecHostIp,
	)
}

func (t *portSpecHostIpTestCase) Assert(typeValue builtin_argument.KurtosisValueType) {
	portSpecStarlark, ok := typeValue.(*port_spec.PortSpec)
	require.True(t, ok)
	portSpec, err := portSpecStarlark.ToKurtosisType()
	require.Nil(t, err)

	waitDuration, errParsingDuration := time.ParseDuration(testWaitDefaultValue)
	require.NoError(t, errParsingDuration)
	expectedPortSpec, errPortCreation := port_spec2.NewPortSpec(testPrivatePortNumber, port_spec2.TransportProtocol_TCP, "", port_spec2.NewWait(waitDuration), "")
	require.NoError(t, errPortCreation)
	require.Equal(t, expectedPortSpec.WithHostIp(testPortSpecHostIp), portSpec)
}
