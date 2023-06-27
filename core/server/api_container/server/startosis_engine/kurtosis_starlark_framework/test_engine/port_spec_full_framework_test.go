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

func newPortSpecFullTestCase(t *testing.T) *portSpecFullTestCase {
	return &portSpecFullTestCase{
		T: t,
	}
}

func (t *portSpecFullTestCase) GetId() string {
	return fmt.Sprintf("%s_%s", port_spec.PortSpecTypeName, "full")
}

func (t *portSpecFullTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%d, %s=%q, %s=%q, %s=%q)",
		port_spec.PortSpecTypeName,
		port_spec.PortNumberAttr,
		TestPrivatePortNumber,
		port_spec.TransportProtocolAttr,
		TestPrivatePortProtocolStr,
		port_spec.PortApplicationProtocolAttr,
		TestPrivateApplicationProtocol,
		port_spec.WaitAttr,
		TestWaitConfiguration,
	)
}

func (t *portSpecFullTestCase) Assert(typeValue builtin_argument.KurtosisValueType) {
	portSpecStarlark, ok := typeValue.(*port_spec.PortSpec)
	require.True(t, ok)
	portSpec, err := portSpecStarlark.ToKurtosisType()
	require.Nil(t, err)

	waitDuration, errParsingDuration := time.ParseDuration(TestWaitConfiguration)
	require.NoError(t, errParsingDuration)
	expectedPortSpec, errPortCreation := port_spec2.NewPortSpec(TestPrivatePortNumber, TestPrivatePortProtocol, TestPrivateApplicationProtocol, port_spec2.NewWait(waitDuration))
	require.NoError(t, errPortCreation)
	require.Equal(t, expectedPortSpec, portSpec)

}
