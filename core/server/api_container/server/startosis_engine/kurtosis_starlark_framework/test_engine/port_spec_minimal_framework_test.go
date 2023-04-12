package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/port_spec"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	portWaitTimeout = ""
)

type portSpecMinimalTestCase struct {
	*testing.T
}

func newPortSpecMinimalTestCase(t *testing.T) *portSpecMinimalTestCase {
	return &portSpecMinimalTestCase{
		T: t,
	}
}

func (t *portSpecMinimalTestCase) GetId() string {
	return fmt.Sprintf("%s_%s", port_spec.PortSpecTypeName, "full")
}

func (t *portSpecMinimalTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%d)", port_spec.PortSpecTypeName, port_spec.PortNumberAttr, TestPrivatePortNumber)
}

func (t *portSpecMinimalTestCase) Assert(typeValue builtin_argument.KurtosisValueType) {
	portSpecStarlark, ok := typeValue.(*port_spec.PortSpec)
	require.True(t, ok)
	portSpec, err := portSpecStarlark.ToKurtosisType()
	require.Nil(t, err)

	expectedPortSpec := binding_constructors.NewPort(TestPrivatePortNumber, kurtosis_core_rpc_api_bindings.Port_TCP, "", portWaitTimeout)
	require.Equal(t, expectedPortSpec, portSpec)

}
