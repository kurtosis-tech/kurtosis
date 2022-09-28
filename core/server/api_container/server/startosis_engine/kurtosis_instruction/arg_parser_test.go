package kurtosis_instruction

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/stretchr/testify/assert"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"testing"
)

func TestSafeCastToUint32_Success(t *testing.T) {
	input := starlark.MakeInt(32)
	output, err := safeCastToUint32(input, "test")
	assert.Nil(t, err)
	assert.Equal(t, uint32(32), output)
}

func TestSafeCastToUint32_FailureWrongType(t *testing.T) {
	input := starlark.String("blah")
	output, err := safeCastToUint32(input, "test")
	assert.NotNil(t, err)
	assert.Equal(t, uint32(0), output)
}

func TestSafeCastToUint32_FailureNotUint32(t *testing.T) {
	input := starlark.MakeInt64(1234567890123456789)
	output, err := safeCastToUint32(input, "test")
	assert.NotNil(t, err)
	assert.Equal(t, uint32(0), output)
}

func TestSafeCastToString_Success(t *testing.T) {
	input := starlark.String("blah")
	output, err := safeCastToString(input, "test")
	assert.Nil(t, err)
	assert.Equal(t, "blah", output)
}

func TestSafeCastToString_Failure(t *testing.T) {
	input := starlark.MakeInt(32)
	output, err := safeCastToString(input, "test")
	assert.NotNil(t, err)
	assert.Equal(t, "", output)
}

func TestExtractStringValueFromStruct_Success(t *testing.T) {
	dict := starlark.StringDict{}
	dict["key"] = starlark.String("value")
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := extractStringValue(input, "key")
	assert.Nil(t, err)
	assert.Equal(t, "value", output)
}

func TestExtractStringValueFromStruct_FailureUnknownKey(t *testing.T) {
	dict := starlark.StringDict{}
	dict["key"] = starlark.String("value")
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := extractStringValue(input, "keyWITHATYPO")
	assert.NotNil(t, err)
	assert.Equal(t, "", output)
}

func TestExtractStringValueFromStruct_FailureWrongValue(t *testing.T) {
	dict := starlark.StringDict{}
	dict["key"] = starlark.MakeInt(32)
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := extractStringValue(input, "key")
	assert.NotNil(t, err)
	assert.Equal(t, "", output)
}

func TestExtractUint32ValueFromStruct_Success(t *testing.T) {
	dict := starlark.StringDict{}
	dict["key"] = starlark.MakeInt(32)
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := extractUint32Value(input, "key")
	assert.Nil(t, err)
	assert.Equal(t, uint32(32), output)
}

func TestExtractUint32ValueFromStruct_FailureUnknownKey(t *testing.T) {
	dict := starlark.StringDict{}
	dict["key"] = starlark.MakeInt(32)
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := extractUint32Value(input, "keyWITHATYPO")
	assert.NotNil(t, err)
	assert.Equal(t, uint32(0), output)
}

func TestExtractUint32ValueFromStruct_FailureWrongType(t *testing.T) {
	dict := starlark.StringDict{}
	dict["key"] = starlark.MakeInt64(123456789012345678)
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := extractUint32Value(input, "key")
	assert.NotNil(t, err)
	assert.Equal(t, uint32(0), output)
}

func TestParsePortProtocol_TCP(t *testing.T) {
	input := "TCP"
	output, err := parsePortProtocol(input)
	assert.Nil(t, err)
	assert.Equal(t, kurtosis_core_rpc_api_bindings.Port_TCP, output)
}

func TestParsePortProtocol_UDP(t *testing.T) {
	input := "UDP"
	output, err := parsePortProtocol(input)
	assert.Nil(t, err)
	assert.Equal(t, kurtosis_core_rpc_api_bindings.Port_UDP, output)
}

func TestParsePortProtocol_SCTP(t *testing.T) {
	input := "SCTP"
	output, err := parsePortProtocol(input)
	assert.Nil(t, err)
	assert.Equal(t, kurtosis_core_rpc_api_bindings.Port_SCTP, output)
}

func TestParsePortProtocol_Unknown(t *testing.T) {
	input := "BLAH"
	output, err := parsePortProtocol(input)
	assert.NotNil(t, err)
	assert.Equal(t, kurtosis_core_rpc_api_bindings.Port_Protocol(-1), output)
}

func TestParsePort_Success(t *testing.T) {
	dict := starlark.StringDict{}
	dict["number"] = starlark.MakeInt(1234)
	dict["protocol"] = starlark.String("TCP")
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := parsePort(input)
	assert.Nil(t, err)
	assert.Equal(t, kurtosis_core_rpc_api_bindings.Port{Number: 1234, Protocol: kurtosis_core_rpc_api_bindings.Port_TCP}, *output)
}

func TestParsePort_FailureWrongProtocol(t *testing.T) {
	dict := starlark.StringDict{}
	dict["number"] = starlark.MakeInt(1234)
	dict["protocol"] = starlark.String("TCPK")
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := parsePort(input)
	assert.NotNil(t, err)
	assert.Nil(t, output)
}

func TestParsePort_FailurePortNumberInvalid(t *testing.T) {
	dict := starlark.StringDict{}
	dict["number"] = starlark.MakeInt(123456)
	dict["protocol"] = starlark.String("TCP")
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := parsePort(input)
	assert.NotNil(t, err)
	assert.Nil(t, output)
}
