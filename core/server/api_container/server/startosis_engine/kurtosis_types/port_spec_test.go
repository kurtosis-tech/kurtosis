package kurtosis_types

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

const (
	testTransportProtocol    = kurtosis_core_rpc_api_bindings.Port_TCP
	emptyApplicationProtocol = ""
)

var testPortNumber = uint32(123)

func TestPortSpec_StringRepresentation(t *testing.T) {
	portSpec := NewPortSpec(testPortNumber, testTransportProtocol, emptyApplicationProtocol)
	expectedStr := `PortSpec(number=123, transport_protocol="TCP", application_protocol="")`
	require.Equal(t, expectedStr, portSpec.String())
}

func TestPortSpec_StringRepresentationWithAppProtocol(t *testing.T) {
	portSpec := NewPortSpec(testPortNumber, testTransportProtocol, "http")
	expectedStr := `PortSpec(number=123, transport_protocol="TCP", application_protocol="http")`
	require.Equal(t, expectedStr, portSpec.String())
}

func TestPortSpec_PortType(t *testing.T) {
	portSpec := NewPortSpec(testPortNumber, testTransportProtocol, emptyApplicationProtocol)
	require.Equal(t, PortSpecTypeName, portSpec.Type())
}

func TestPort_Freeze(t *testing.T) {
	portSpec := NewPortSpec(testPortNumber, testTransportProtocol, emptyApplicationProtocol)
	// just checking it doesn't panic as it's a no-op
	require.NotPanics(t, portSpec.Freeze)
}

func TestPort_TruthValidService(t *testing.T) {
	portSpec := NewPortSpec(testPortNumber, testTransportProtocol, emptyApplicationProtocol)
	// just checking it doesn't panic as it's a no-op
	require.Equal(t, starlark.Bool(true), portSpec.Truth())
}

func TestPort_TruthFalsyService(t *testing.T) {
	portSpec := PortSpec{
		number:                   0,
		transportProtocol:        0,
		maybeApplicationProtocol: "",
	}
	require.Equal(t, starlark.Bool(false), portSpec.Truth())
}

func TestPort_HashThrowsError(t *testing.T) {
	portSpec := NewPortSpec(testPortNumber, testTransportProtocol, emptyApplicationProtocol)
	hash, err := portSpec.Hash()
	require.NotNil(t, err)
	require.Equal(t, uint32(0), hash)
}

func TestPort_TestValidAttr(t *testing.T) {
	portSpec := NewPortSpec(testPortNumber, testTransportProtocol, emptyApplicationProtocol)
	attrValue, err := portSpec.Attr(portNumberAttr)
	require.Nil(t, err)
	require.Equal(t, starlark.MakeInt(int(testPortNumber)), attrValue)
}

func TestPort_TestInvalidAttr(t *testing.T) {
	portSpec := NewPortSpec(testPortNumber, testTransportProtocol, emptyApplicationProtocol)
	attrValue, err := portSpec.Attr(testInvalidAttr)
	require.NotNil(t, err)
	require.Nil(t, attrValue)
}

func TestPort_TestAttrNames(t *testing.T) {
	portSpec := NewPortSpec(testPortNumber, testTransportProtocol, emptyApplicationProtocol)
	attrNames := portSpec.AttrNames()
	require.Equal(t, []string{portNumberAttr, transportProtocolAttr, portApplicationProtocolAttr}, attrNames)
}

func TestParseTransportProtocol_TCP(t *testing.T) {
	input := "TCP"
	output, err := parseTransportProtocol(input)
	require.Nil(t, err)
	require.Equal(t, kurtosis_core_rpc_api_bindings.Port_TCP, output)
}

func TestParseTransportProtocol_UDP(t *testing.T) {
	input := "UDP"
	output, err := parseTransportProtocol(input)
	require.Nil(t, err)
	require.Equal(t, kurtosis_core_rpc_api_bindings.Port_UDP, output)
}

func TestParseTransportProtocol_SCTP(t *testing.T) {
	input := "SCTP"
	output, err := parseTransportProtocol(input)
	require.Nil(t, err)
	require.Equal(t, kurtosis_core_rpc_api_bindings.Port_SCTP, output)
}

func TestParseTransportProtocol_DefaultValue(t *testing.T) {
	input := ""
	output, err := parseTransportProtocol(input)
	require.Nil(t, err)
	require.Equal(t, kurtosis_core_rpc_api_bindings.Port_TCP, output)
}

func TestParseTransportProtocol_Unknown(t *testing.T) {
	input := "BLAH"
	output, err := parseTransportProtocol(input)
	require.NotNil(t, err)
	require.Equal(t, "Port protocol should be one of TCP, SCTP, UDP", err.Error())
	require.Equal(t, kurtosis_core_rpc_api_bindings.Port_TransportProtocol(-1), output)
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

func TestValidateApplicationProtocol_FailureWithInvalidChars(t *testing.T) {
	maybeApplicationProtocol := "H-ttp.2 "
	err := validateApplicationProtocol(maybeApplicationProtocol)
	require.NotNil(t, err)
	require.ErrorContains(t, err, "application protocol 'H-ttp.2 ' contains invalid character(s). It must only contain [a-zA-Z0-9+.-]")
}

func TestValidateApplicationProtocol_Success(t *testing.T) {
	maybeApplicationProtocol := "H-ttp.2"
	err := validateApplicationProtocol(maybeApplicationProtocol)
	require.Nil(t, err)
}
