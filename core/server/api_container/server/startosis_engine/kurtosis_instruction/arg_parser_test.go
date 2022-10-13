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
	require.Equal(t, "'test' is expected to be a string. Got starlark.Int", err.Error())
	require.Equal(t, "", output)
}

func TestSafeCastToStringSlice_Success(t *testing.T) {
	input := starlark.NewList([]starlark.Value{starlark.String("string_1"), starlark.String("string_2")})
	output, err := safeCastToStringSlice(input, "test")
	require.Nil(t, err)
	require.Equal(t, []string{"string_1", "string_2"}, output)
}

func TestSafeCastToStringSlice_FailureWrongTypeInsideList(t *testing.T) {
	input := starlark.NewList([]starlark.Value{starlark.String("string_1"), starlark.MakeInt(42)})
	output, err := safeCastToStringSlice(input, "test")
	require.NotNil(t, err)
	require.Equal(t, "'test[1]' is expected to be a string. Got starlark.Int", err.Error())
	require.Equal(t, []string(nil), output)
}

func TestSafeCastToStringSlice_FailureNotList(t *testing.T) {
	input := starlark.MakeInt(42)
	output, err := safeCastToStringSlice(input, "test")
	require.NotNil(t, err)
	require.Equal(t, "'test' argument is expected to be a list. Got starlark.Int", err.Error())
	require.Equal(t, []string(nil), output)
}

func TestSafeCastToMapStringString_Success(t *testing.T) {
	input := starlark.NewDict(1)
	err := input.SetKey(starlark.String("key"), starlark.String("value"))
	require.Nil(t, err)
	output, err := safeCastToMapStringString(input, "test")
	require.Nil(t, err)
	require.Equal(t, map[string]string{"key": "value"}, output)
}

func TestSafeCastToMapStringString_Failure(t *testing.T) {
	input := starlark.MakeInt(32)
	output, err := safeCastToMapStringString(input, "test")
	require.NotNil(t, err)
	require.Equal(t, "'test' argument is expected to be a dict. Got starlark.Int", err.Error())
	require.Equal(t, map[string]string(nil), output)
}

func TestSafeCastToMapStringString_FailureValueIsNotString(t *testing.T) {
	input := starlark.NewDict(1)
	err := input.SetKey(starlark.String("key"), starlark.MakeInt(42))
	require.Nil(t, err)
	output, err := safeCastToMapStringString(input, "test")
	require.NotNil(t, err)
	require.Equal(t, "'test[\"key\"]' is expected to be a string. Got starlark.Int", err.Error())
	require.Equal(t, map[string]string(nil), output)
}

func TestSafeCastToMapStringString_FailureKeyIsNotString(t *testing.T) {
	input := starlark.NewDict(1)
	err := input.SetKey(starlark.MakeInt(42), starlark.String("value"))
	require.Nil(t, err)
	output, err := safeCastToMapStringString(input, "test")
	require.NotNil(t, err)
	require.Equal(t, "'test.key:42' is expected to be a string. Got starlark.Int", err.Error())
	require.Equal(t, map[string]string(nil), output)
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
	require.Equal(t, "'key' is expected to be a string. Got starlark.Int", err.Error())
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

func TestExtractSliceValue_Success(t *testing.T) {
	dict := starlark.StringDict{}
	dict["key"] = starlark.NewList([]starlark.Value{starlark.String("test")})
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := extractStringSliceValue(input, "key", "dict")
	require.Nil(t, err)
	require.Equal(t, []string{"test"}, output)
}

func TestExtractSliceValue_FailureMissing(t *testing.T) {
	dict := starlark.StringDict{}
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := extractStringSliceValue(input, "missingKey", "dict")
	require.NotNil(t, err)
	require.Equal(t, "Missing value 'missingKey' as element of the struct object 'dict'", err.Error())
	require.Equal(t, []string(nil), output)
}

func TestExtractMapStringString_Success(t *testing.T) {
	subDict := starlark.NewDict(1)
	err := subDict.SetKey(starlark.String("key"), starlark.String("value"))
	require.Nil(t, err)
	dict := starlark.StringDict{}
	dict["key"] = subDict
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := extractMapStringStringValue(input, "key", "dict")
	require.Nil(t, err)
	require.Equal(t, map[string]string{"key": "value"}, output)
}

func TestExtractMapStringString_FailureMissing(t *testing.T) {
	dict := starlark.StringDict{}
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := extractMapStringStringValue(input, "missingKey", "dict")
	require.NotNil(t, err)
	require.Equal(t, "Missing value 'missingKey' as element of the struct object 'dict'", err.Error())
	require.Equal(t, map[string]string(nil), output)
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
	require.Equal(t, "Port protocol should be one of TCP, SCTP, UDP", err.Error())
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
	require.Equal(t, "Port protocol should be one of TCP, SCTP, UDP", err.Error())
	require.Nil(t, output)
}

func TestParsePort_FailurePortNumberInvalid(t *testing.T) {
	dict := starlark.StringDict{}
	dict["number"] = starlark.MakeInt(123456)
	dict["protocol"] = starlark.String("TCP")
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := parsePort(input)
	require.NotNil(t, err)
	require.Equal(t, "Port number should be less than or equal to 65535", err.Error())
	require.Nil(t, output)
}

func TestParseEntryPointArgs_Success(t *testing.T) {
	dict := starlark.StringDict{}
	dict["entry_point_args"] = starlark.NewList([]starlark.Value{starlark.String("hello"), starlark.String("world")})
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := parseEntryPointArgs(input)
	require.Nil(t, err)
	require.Equal(t, []string{"hello", "world"}, output)
}

func TestParseEntryPointArgs_SuccessOnMissingValue(t *testing.T) {
	dict := starlark.StringDict{}
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := parseEntryPointArgs(input)
	require.Nil(t, err)
	require.Equal(t, []string{}, output)
}

func TestParseEntryPointArgs_FailureOnListContainingNonStringValues(t *testing.T) {
	dict := starlark.StringDict{}
	dict["entry_point_args"] = starlark.NewList([]starlark.Value{starlark.MakeInt(42)})
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := parseEntryPointArgs(input)
	require.NotNil(t, err)
	require.Equal(t, "'entry_point_args[0]' is expected to be a string. Got starlark.Int", err.Error())
	require.Equal(t, []string(nil), output)
}

func TestParseCommandArgs_Success(t *testing.T) {
	dict := starlark.StringDict{}
	dict["cmd_args"] = starlark.NewList([]starlark.Value{starlark.String("hello"), starlark.String("world")})
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := parseCmdArgs(input)
	require.Nil(t, err)
	require.Equal(t, []string{"hello", "world"}, output)
}

func TestParseCommandArgs_SuccessOnMissingValue(t *testing.T) {
	dict := starlark.StringDict{}
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := parseCmdArgs(input)
	require.Nil(t, err)
	require.Equal(t, []string{}, output)
}

func TestParseCommandArgs_FailureOnListContainingNonStringValues(t *testing.T) {
	dict := starlark.StringDict{}
	dict["cmd_args"] = starlark.NewList([]starlark.Value{starlark.MakeInt(42)})
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := parseCmdArgs(input)
	require.NotNil(t, err)
	require.Equal(t, "'cmd_args[0]' is expected to be a string. Got starlark.Int", err.Error())
	require.Equal(t, []string(nil), output)
}

func TestParseEnvVars_Success(t *testing.T) {
	subDict := starlark.NewDict(1)
	err := subDict.SetKey(starlark.String("key"), starlark.String("value"))
	require.Nil(t, err)
	dict := starlark.StringDict{}
	dict["env_vars"] = subDict
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := parseEnvVars(input)
	require.Nil(t, err)
	require.Equal(t, map[string]string{"key": "value"}, output)
}

func TestParseEnvVars_SuccessOnMissingValue(t *testing.T) {
	dict := starlark.StringDict{}
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := parseEnvVars(input)
	require.Nil(t, err)
	require.Equal(t, map[string]string{}, output)
}

func TestParseEnvVars_FailureOnNonStringKey(t *testing.T) {
	subDict := starlark.NewDict(1)
	err := subDict.SetKey(starlark.MakeInt(42), starlark.String("value"))
	require.Nil(t, err)
	dict := starlark.StringDict{}
	dict["env_vars"] = subDict
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := parseEnvVars(input)
	require.NotNil(t, err)
	require.Equal(t, "'env_vars.key:42' is expected to be a string. Got starlark.Int", err.Error())
	require.Equal(t, map[string]string(nil), output)
}

func TestParseEnvVars_FailureOnNonStringValue(t *testing.T) {
	subDict := starlark.NewDict(1)
	err := subDict.SetKey(starlark.String("key"), starlark.MakeInt(42))
	require.Nil(t, err)
	dict := starlark.StringDict{}
	dict["env_vars"] = subDict
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := parseEnvVars(input)
	require.NotNil(t, err)
	require.Equal(t, "'env_vars[\"key\"]' is expected to be a string. Got starlark.Int", err.Error())
	require.Equal(t, map[string]string(nil), output)
}
