package kurtosis_types

import (
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

const (
	testProtocol = starlark.String("TCP")
)

var testPortNumber = starlark.MakeUint(123)

func TestPortSpec_StringRepresentation(t *testing.T) {
	portSpec := NewPortSpec(testPortNumber, testProtocol)
	expectedStr := "port_spec(number=123, protocol=\"TCP\")"
	require.Equal(t, expectedStr, portSpec.String())
}

func TestPortSpec_PortType(t *testing.T) {
	portSpec := NewPortSpec(testPortNumber, testProtocol)
	require.Equal(t, portSpecTypeName, portSpec.Type())
}

func TestPort_Freeze(t *testing.T) {
	portSpec := NewPortSpec(testPortNumber, testProtocol)
	// just checking it doesn't panic as it's a no-op
	require.NotPanics(t, portSpec.Freeze)
}

func TestPort_TruthValidService(t *testing.T) {
	portSpec := NewPortSpec(testPortNumber, testProtocol)
	// just checking it doesn't panic as it's a no-op
	require.Equal(t, starlark.Bool(true), portSpec.Truth())
}

func TestPort_TruthFalsyService(t *testing.T) {
	portSpec := PortSpec{}
	require.Equal(t, starlark.Bool(false), portSpec.Truth())
}

func TestPort_HashThrowsError(t *testing.T) {
	portSpec := NewPortSpec(testPortNumber, testProtocol)
	hash, err := portSpec.Hash()
	require.NotNil(t, err)
	require.Equal(t, uint32(0), hash)
}

func TestPort_TestValidAttr(t *testing.T) {
	portSpec := NewPortSpec(testPortNumber, testProtocol)
	attrValue, err := portSpec.Attr(portNumberAttr)
	require.Nil(t, err)
	require.Equal(t, testPortNumber, attrValue)
}

func TestPort_TestInvalidAttr(t *testing.T) {
	portSpec := NewPortSpec(testPortNumber, testProtocol)
	attrValue, err := portSpec.Attr(testInvalidAttr)
	require.NotNil(t, err)
	require.Nil(t, attrValue)
}

func TestPort_TestAttrNames(t *testing.T) {
	portSpec := NewPortSpec(testPortNumber, testProtocol)
	attrNames := portSpec.AttrNames()
	require.Equal(t, []string{portNumberAttr, portProtocolAttr}, attrNames)
}
