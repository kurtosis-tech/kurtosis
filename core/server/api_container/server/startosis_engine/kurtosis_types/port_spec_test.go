package kurtosis_types

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

const (
	testProtocol = kurtosis_core_rpc_api_bindings.Port_TCP
)

var testPortNumber = uint32(123)

func TestPortSpec_StringRepresentation(t *testing.T) {
	portSpec := NewPortSpec(testPortNumber, testProtocol)
	expectedStr := `PortSpec(number=123, protocol="TCP")`
	require.Equal(t, expectedStr, portSpec.String())
}

func TestPortSpec_PortType(t *testing.T) {
	portSpec := NewPortSpec(testPortNumber, testProtocol)
	require.Equal(t, PortSpecTypeName, portSpec.Type())
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
	require.Equal(t, starlark.MakeInt(int(testPortNumber)), attrValue)
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

func TestParsePortProtocol_TCP(t *testing.T) {
	input := "TCP"
	output, err := parsePortProtocol(input)
	require.Nil(t, err)
	require.Equal(t, kurtosis_core_rpc_api_bindings.Port_TCP, output)
}

func TestParsePortProtocol_UDP(t *testing.T) {
	input := "UDP"
	output, err := parsePortProtocol(input)
	require.Nil(t, err)
	require.Equal(t, kurtosis_core_rpc_api_bindings.Port_UDP, output)
}

func TestParsePortProtocol_SCTP(t *testing.T) {
	input := "SCTP"
	output, err := parsePortProtocol(input)
	require.Nil(t, err)
	require.Equal(t, kurtosis_core_rpc_api_bindings.Port_SCTP, output)
}

func TestParsePortProtocol_DefaultValue(t *testing.T) {
	input := ""
	output, err := parsePortProtocol(input)
	require.Nil(t, err)
	require.Equal(t, kurtosis_core_rpc_api_bindings.Port_TCP, output)
}

func TestParsePortProtocol_Unknown(t *testing.T) {
	input := "BLAH"
	output, err := parsePortProtocol(input)
	require.NotNil(t, err)
	require.Equal(t, "Port protocol should be one of TCP, SCTP, UDP", err.Error())
	require.Equal(t, kurtosis_core_rpc_api_bindings.Port_Protocol(-1), output)
}

func TestValidatePort_Success(t *testing.T) {
	number := 1234
	_, err := parsePortNumber(number)
	require.Nil(t, err)
}

func TestParsePort_FailurePortNumberInvalid(t *testing.T) {
	number := 123456
	_, err := parsePortNumber(number)
	require.NotNil(t, err)
	require.Equal(t, fmt.Sprintf("Port number should be in range [%d - %d]", 1, 65535), err.Error())
}

func TestParsePort_FailurePortNumberInvalidNegativeNumber(t *testing.T) {
	number := -1
	_, err := parsePortNumber(number)
	require.NotNil(t, err)
	require.Equal(t, fmt.Sprintf("Port number should be in range [%d - %d]", 1, 65535), err.Error())
}
