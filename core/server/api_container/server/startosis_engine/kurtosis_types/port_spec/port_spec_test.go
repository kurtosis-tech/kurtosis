package port_spec

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

func TestParseTransportProtocol_TCP(t *testing.T) {
	input := starlark.String("TCP")
	output, err := parseTransportProtocol(true, input)
	require.Nil(t, err)
	require.Equal(t, port_spec.TransportProtocol_TCP, output)
}

func TestParseTransportProtocol_UDP(t *testing.T) {
	input := starlark.String("UDP")
	output, err := parseTransportProtocol(true, input)
	require.Nil(t, err)
	require.Equal(t, port_spec.TransportProtocol_UDP, output)
}

func TestParseTransportProtocol_SCTP(t *testing.T) {
	input := starlark.String("SCTP")
	output, err := parseTransportProtocol(true, input)
	require.Nil(t, err)
	require.Equal(t, port_spec.TransportProtocol_SCTP, output)
}

func TestParseTransportProtocol_DefaultValue(t *testing.T) {
	input := starlark.String("")
	output, err := parseTransportProtocol(true, input)
	require.Nil(t, err)
	require.Equal(t, port_spec.TransportProtocol_TCP, output)
}

func TestParseTransportProtocol_DefaultValueUnset(t *testing.T) {
	input := starlark.String("")
	output, err := parseTransportProtocol(false, input)
	require.Nil(t, err)
	require.Equal(t, port_spec.TransportProtocol_TCP, output)
}

func TestParseTransportProtocol_Unknown(t *testing.T) {
	input := starlark.String("BLAH")
	output, err := parseTransportProtocol(true, input)
	require.NotNil(t, err)
	require.Equal(t, "Port protocol should be one of TCP, SCTP, UDP", err.Error())
	require.Equal(t, port_spec.TransportProtocol(-1), output)
}

func TestValidatePort_Success(t *testing.T) {
	number := starlark.MakeInt(1234)
	_, err := parsePortNumber(true, number)
	require.Nil(t, err)
}

func TestParsePort_FailurePortNumberInvalid(t *testing.T) {
	number := starlark.MakeInt(123456)
	_, err := parsePortNumber(true, number)
	require.NotNil(t, err)
	require.Equal(t, fmt.Sprintf("Port number should be in range [%d - %d]", 1, 65535), err.Error())
}

func TestParsePort_FailurePortNumberInvalidNegativeNumber(t *testing.T) {
	number := starlark.MakeInt(-1)
	_, err := parsePortNumber(true, number)
	require.NotNil(t, err)
	require.Equal(t, fmt.Sprintf("Port number should be in range [%d - %d]", 1, 65535), err.Error())
}

func TestValidateApplicationProtocol_Success(t *testing.T) {
	maybeApplicationProtocol := starlark.String("H-ttp.2")
	applicationProtocol, err := parsePortApplicationProtocol(true, maybeApplicationProtocol)
	require.Nil(t, err)
	require.Equal(t, maybeApplicationProtocol.GoString(), applicationProtocol)
}
