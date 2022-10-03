package kurtosis_instruction

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"testing"
)

func TestSafeCastToUint32_Success(t *testing.T) {
	input := starlark.MakeInt(32)
	output, err := safeCastToUint32(input, "test")
	require.Nil(t, err)
	require.Equal(t, uint32(32), output)
}

func TestSafeCastToUint32_FailureWrongType(t *testing.T) {
	input := starlark.String("blah")
	output, err := safeCastToUint32(input, "test")
	require.NotNil(t, err)
	require.Equal(t, "Argument 'test' is expected to be an integer. Got starlark.String", err.Error())
	require.Equal(t, uint32(0), output)
}

func TestSafeCastToUint32_FailureNotUint32(t *testing.T) {
	input := starlark.MakeInt64(1234567890123456789)
	output, err := safeCastToUint32(input, "test")
	require.NotNil(t, err)
	require.Equal(t, "'test' argument is expected to be a an integer greater than 0 and lower than 4294967295", err.Error())
	require.Equal(t, uint32(0), output)
}

func TestSafeCastToString_Success(t *testing.T) {
	input := starlark.String("blah")
	output, err := safeCastToString(input, "test")
	require.Nil(t, err)
	require.Equal(t, "blah", output)
}

func TestSafeCastToString_Failure(t *testing.T) {
	input := starlark.MakeInt(32)
	output, err := safeCastToString(input, "test")
	require.NotNil(t, err)
	require.Equal(t, "'test' argument is expected to be a string. Got starlark.Int", err.Error())
	require.Equal(t, "", output)
}

func TestExtractStringValueFromStruct_Success(t *testing.T) {
	dict := starlark.StringDict{}
	dict["key"] = starlark.String("value")
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := extractStringValue(input, "key", "dict")
	require.Nil(t, err)
	require.Equal(t, "value", output)
}

func TestExtractStringValueFromStruct_FailureUnknownKey(t *testing.T) {
	dict := starlark.StringDict{}
	dict["key"] = starlark.String("value")
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := extractStringValue(input, "keyWITHATYPO", "dict")
	require.NotNil(t, err)
	require.Equal(t, "Missing value 'keyWITHATYPO' as element of the struct object 'dict'", err.Error())
	require.Equal(t, "", output)
}

func TestExtractStringValueFromStruct_FailureWrongValue(t *testing.T) {
	dict := starlark.StringDict{}
	dict["key"] = starlark.MakeInt(32)
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := extractStringValue(input, "key", "dict")
	require.NotNil(t, err)
	require.Equal(t, "'key' argument is expected to be a string. Got starlark.Int", err.Error())
	require.Equal(t, "", output)
}

func TestExtractUint32ValueFromStruct_Success(t *testing.T) {
	dict := starlark.StringDict{}
	dict["key"] = starlark.MakeInt(32)
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := extractUint32Value(input, "key", "dict")
	require.Nil(t, err)
	require.Equal(t, uint32(32), output)
}

func TestExtractUint32ValueFromStruct_FailureUnknownKey(t *testing.T) {
	dict := starlark.StringDict{}
	dict["key"] = starlark.MakeInt(32)
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := extractUint32Value(input, "keyWITHATYPO", "dict")
	require.NotNil(t, err)
	require.Equal(t, "Missing value 'keyWITHATYPO' as element of the struct object 'dict'", err.Error())
	require.Equal(t, uint32(0), output)
}

func TestExtractUint32ValueFromStruct_FailureWrongType(t *testing.T) {
	dict := starlark.StringDict{}
	dict["key"] = starlark.MakeInt64(123456789012345678)
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := extractUint32Value(input, "key", "dict")
	require.NotNil(t, err)
	require.Equal(t, "'key' argument is expected to be a an integer greater than 0 and lower than 4294967295", err.Error())
	require.Equal(t, uint32(0), output)
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

func TestParsePortProtocol_Unknown(t *testing.T) {
	input := "BLAH"
	output, err := parsePortProtocol(input)
	require.NotNil(t, err)
	require.Equal(t, "Port protocol should be either TCP, SCTP, UDP", err.Error())
	require.Equal(t, kurtosis_core_rpc_api_bindings.Port_Protocol(-1), output)
}

func TestParsePort_Success(t *testing.T) {
	dict := starlark.StringDict{}
	dict["number"] = starlark.MakeInt(1234)
	dict["protocol"] = starlark.String("TCP")
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := parsePort(input)
	require.Nil(t, err)
	require.Equal(t, &kurtosis_core_rpc_api_bindings.Port{Number: 1234, Protocol: kurtosis_core_rpc_api_bindings.Port_TCP}, output)
}

func TestParsePort_FailureWrongProtocol(t *testing.T) {
	dict := starlark.StringDict{}
	dict["number"] = starlark.MakeInt(1234)
	dict["protocol"] = starlark.String("TCPK")
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := parsePort(input)
	require.NotNil(t, err)
	require.Equal(t, "Port protocol should be either TCP, SCTP, UDP", err.Error())
	require.Nil(t, output)
}

func TestParsePort_FailurePortNumberInvalid(t *testing.T) {
	dict := starlark.StringDict{}
	dict["number"] = starlark.MakeInt(123456)
	dict["protocol"] = starlark.String("TCP")
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := parsePort(input)
	require.NotNil(t, err)
	require.Equal(t, "Port number should be strictly lower than 65535", err.Error())
	require.Nil(t, output)
}
