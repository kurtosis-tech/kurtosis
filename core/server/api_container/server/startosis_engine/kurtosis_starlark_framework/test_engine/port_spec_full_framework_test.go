package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/port_spec"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
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

func (t *portSpecFullTestCase) GetTypeConstructor() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return port_spec.NewPortSpecType()
}

func (t *portSpecFullTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%d, %s=%q, %s=%q)", port_spec.PortSpecTypeName, port_spec.PortNumberAttr, TestPrivatePortNumber, port_spec.TransportProtocolAttr, TestPrivatePortProtocolStr, port_spec.PortApplicationProtocolAttr, TestPrivateApplicationProtocol)
}

func (t *portSpecFullTestCase) Assert(typeValue starlark.Value) {
	portSpecStarlark, ok := typeValue.(*port_spec.PortSpec)
	require.True(t, ok)
	portSpec, err := portSpecStarlark.ToKurtosisType()
	require.Nil(t, err)

	expectedPortSpec := binding_constructors.NewPort(TestPrivatePortNumber, TestPrivatePortProtocol, TestPrivateApplicationProtocol)
	require.Equal(t, expectedPortSpec, portSpec)

}
