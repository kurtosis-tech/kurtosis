package kurtosis_instruction

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/recipe"
	"github.com/kurtosis-tech/kurtosis/core/server/commons/enclave_data_directory"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"math"
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
	expectedError := `Error casting value 'key' as element of the struct object 'dict'
	Caused by: 'key' is expected to be a string. Got starlark.Int`
	dict := starlark.StringDict{}
	dict["key"] = starlark.MakeInt(32)
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := extractStringValue(input, "key", "dict")
	require.NotNil(t, err)
	require.Equal(t, expectedError, err.Error())
	require.Equal(t, "", output)
}

func TestSafeCastToInt32_ValidPositiveValue(t *testing.T) {
	input := starlark.MakeInt(32)
	output, err := safeCastToInt32(input, "test")
	require.Nil(t, err)
	require.Equal(t, int32(32), output)
}

func TestSafeCastToInt32_ValidNegativeValue(t *testing.T) {
	input := starlark.MakeInt(-32)
	output, err := safeCastToInt32(input, "test")
	require.Nil(t, err)
	require.Equal(t, int32(-32), output)
}

func TestSafeCastToInt32_ValidZeroValue(t *testing.T) {
	input := starlark.MakeInt(0)
	output, err := safeCastToInt32(input, "test")
	require.Nil(t, err)
	require.Equal(t, int32(0), output)
}

func TestSafeCastToInt32_FailsForValuesGreaterThanMaxInt32(t *testing.T) {
	input := starlark.MakeInt(math.MaxInt32 + 1)
	_, err := safeCastToInt32(input, "test")
	require.NotNil(t, err)
}

func TestSafeCastToInt32_FailsForValuesLowerThanMinInt32(t *testing.T) {
	input := starlark.MakeInt(math.MinInt32 - 1)
	_, err := safeCastToInt32(input, "test")
	require.NotNil(t, err)
}

func TestSafeCastToInt32_FailsForString(t *testing.T) {
	input := starlark.String("hello")
	_, err := safeCastToInt32(input, "test")
	require.NotNil(t, err)
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
	expectedError := `Error casting value 'key' as element of the struct object 'dict'
	Caused by: 'key' argument is expected to be a an integer greater than 0 and lower than 4294967295`
	dict := starlark.StringDict{}
	dict["key"] = starlark.MakeInt64(123456789012345678)
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := extractUint32Value(input, "key", "dict")
	require.NotNil(t, err)
	require.Equal(t, expectedError, err.Error())
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
	output, err := parsePort(input, usedPortsKey)
	require.Nil(t, err)
	require.Equal(t, &kurtosis_core_rpc_api_bindings.Port{Number: 1234, Protocol: kurtosis_core_rpc_api_bindings.Port_TCP}, output)
}

func TestParsePort_FailureWrongProtocol(t *testing.T) {
	dict := starlark.StringDict{}
	dict["number"] = starlark.MakeInt(1234)
	dict["protocol"] = starlark.String("TCPK")
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := parsePort(input, usedPortsKey)
	require.NotNil(t, err)
	require.Equal(t, "Port protocol should be one of TCP, SCTP, UDP", err.Error())
	require.Nil(t, output)
}

func TestParsePort_FailurePortNumberInvalid(t *testing.T) {
	dict := starlark.StringDict{}
	dict["number"] = starlark.MakeInt(123456)
	dict["protocol"] = starlark.String("TCP")
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := parsePort(input, usedPortsKey)
	require.NotNil(t, err)
	require.Equal(t, "Port number should be less than or equal to 65535", err.Error())
	require.Nil(t, output)
}

func TestParseEntryPointArgs_Success(t *testing.T) {
	dict := starlark.StringDict{}
	dict["entrypoint"] = starlark.NewList([]starlark.Value{starlark.String("hello"), starlark.String("world")})
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
	require.Equal(t, []string(nil), output)
}

func TestParseEntryPointArgs_FailureOnListContainingNonStringValues(t *testing.T) {
	expectedError := `Error casting value 'entrypoint' as element of the struct object 'config'
	Caused by: 'entrypoint[0]' is expected to be a string. Got starlark.Int`
	dict := starlark.StringDict{}
	dict["entrypoint"] = starlark.NewList([]starlark.Value{starlark.MakeInt(42)})
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := parseEntryPointArgs(input)
	require.NotNil(t, err)
	require.Equal(t, expectedError, err.Error())
	require.Equal(t, []string(nil), output)
}

func TestParseCommandArgs_Success(t *testing.T) {
	dict := starlark.StringDict{}
	dict["cmd"] = starlark.NewList([]starlark.Value{starlark.String("hello"), starlark.String("world")})
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
	require.Equal(t, []string(nil), output)
}

func TestParseCommandArgs_FailureOnListContainingNonStringValues(t *testing.T) {
	expectedError := `Error casting value 'cmd' as element of the struct object 'config'
	Caused by: 'cmd[0]' is expected to be a string. Got starlark.Int`
	dict := starlark.StringDict{}
	dict["cmd"] = starlark.NewList([]starlark.Value{starlark.MakeInt(42)})
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := parseCmdArgs(input)
	require.NotNil(t, err)
	require.Equal(t, expectedError, err.Error())
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
	expectedError := `Error casting value 'env_vars' as element of the struct object 'config'
	Caused by: 'env_vars.key:42' is expected to be a string. Got starlark.Int`
	subDict := starlark.NewDict(1)
	err := subDict.SetKey(starlark.MakeInt(42), starlark.String("value"))
	require.Nil(t, err)
	dict := starlark.StringDict{}
	dict["env_vars"] = subDict
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := parseEnvVars(input)
	require.NotNil(t, err)
	require.Equal(t, expectedError, err.Error())
	require.Equal(t, map[string]string(nil), output)
}

func TestParseEnvVars_FailureOnNonStringValue(t *testing.T) {
	expectedError := `Error casting value 'env_vars' as element of the struct object 'config'
	Caused by: 'env_vars["key"]' is expected to be a string. Got starlark.Int`
	subDict := starlark.NewDict(1)
	err := subDict.SetKey(starlark.String("key"), starlark.MakeInt(42))
	require.Nil(t, err)
	dict := starlark.StringDict{}
	dict["env_vars"] = subDict
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := parseEnvVars(input)
	require.NotNil(t, err)
	require.Equal(t, expectedError, err.Error())
	require.Equal(t, map[string]string(nil), output)
}

func TestParseExpectedExitCode_ValidValue(t *testing.T) {
	input := starlark.MakeInt(32)
	output, err := ParseExpectedExitCode(input)
	require.Nil(t, err)
	require.Equal(t, int32(32), output)
}

func TestParseExpectedExitCode_OverflowForLargeUnsignedInt64(t *testing.T) {
	input := starlark.MakeUint64(^uint64(0))
	_, err := ParseExpectedExitCode(input)
	require.NotNil(t, err)
}

func TestParseCommand_ValidValue(t *testing.T) {
	input := starlark.NewList([]starlark.Value{starlark.String("foo"), starlark.String("bar")})
	output, err := ParseCommand(input)
	require.Nil(t, err)
	require.Equal(t, []string{"foo", "bar"}, output)
}

func TestParseCommand_InvalidCommandsWithIntegers(t *testing.T) {
	input := starlark.NewList([]starlark.Value{starlark.String("foo"), starlark.MakeInt(42)})
	_, err := ParseCommand(input)
	require.NotNil(t, err)
}

func TestArtifactUuidPathPath_ValidValue(t *testing.T) {
	input := starlark.String("abde-f23dd-1")
	output, err := ParseArtifactId("artifact_id", input)
	require.Nil(t, err)
	require.Equal(t, enclave_data_directory.FilesArtifactID("abde-f23dd-1"), output)
}

func TestArtifactUuidPathPath_EmptyStringFails(t *testing.T) {
	input := starlark.String("")
	_, err := ParseArtifactId("artifact_id", input)
	require.NotNil(t, err)
}

func TestParseFilesArtifactMountDirpaths_Success(t *testing.T) {
	subDict := starlark.NewDict(1)
	err := subDict.SetKey(starlark.String("key"), starlark.String("value"))
	require.Nil(t, err)
	dict := starlark.StringDict{}
	dict["files"] = subDict
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := parseFilesArtifactMountDirpaths(input)
	require.Nil(t, err)
	require.Equal(t, map[string]string{"key": "value"}, output)
}

func TestParseFilesArtifactMountDirpaths_SuccessOnMissingValue(t *testing.T) {
	dict := starlark.StringDict{}
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := parseFilesArtifactMountDirpaths(input)
	require.Nil(t, err)
	require.Equal(t, map[string]string{}, output)
}

func TestParseFilesArtifactMountDirpaths_FailureOnNonStringKey(t *testing.T) {
	expectedError := `Error casting value 'files' as element of the struct object 'config'
	Caused by: 'files.key:42' is expected to be a string. Got starlark.Int`
	subDict := starlark.NewDict(1)
	err := subDict.SetKey(starlark.MakeInt(42), starlark.String("value"))
	require.Nil(t, err)
	dict := starlark.StringDict{}
	dict["files"] = subDict
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := parseFilesArtifactMountDirpaths(input)
	require.NotNil(t, err)
	require.Equal(t, expectedError, err.Error())
	require.Equal(t, map[string]string(nil), output)
}

func TestParseFilesArtifactMountDirpaths_FailureOnNonStringValue(t *testing.T) {
	expectedError := `Error casting value 'files' as element of the struct object 'config'
	Caused by: 'files["key"]' is expected to be a string. Got starlark.Int`
	subDict := starlark.NewDict(1)
	err := subDict.SetKey(starlark.String("key"), starlark.MakeInt(42))
	require.Nil(t, err)
	dict := starlark.StringDict{}
	dict["files"] = subDict
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := parseFilesArtifactMountDirpaths(input)
	require.NotNil(t, err)
	require.Equal(t, expectedError, err.Error())
	require.Equal(t, map[string]string(nil), output)
}

func TestParseTemplatesAndData_SimpleCaseStruct(t *testing.T) {
	dataStringDict := starlark.StringDict{}
	dataStringDict["LargeFloat"] = starlark.Float(1231231243.43)
	dataStringDict["Name"] = starlark.String("John")
	dataStringDict["UnixTs"] = starlark.MakeInt64(1257894000)
	dataStringDict["Boolean"] = starlark.Bool(true)
	data := starlarkstruct.FromStringDict(starlarkstruct.Default, dataStringDict)
	templateDataStrDict := starlark.StringDict{}
	template := "Hello {{.Name}}. {{.LargeFloat}} {{.UnixTs}} {{.Boolean}}"
	templateDataStrDict["template"] = starlark.String(template)
	templateDataStrDict["data"] = data
	input := starlark.NewDict(1)
	err := input.SetKey(starlark.String("/foo/bar"), starlarkstruct.FromStringDict(starlarkstruct.Default, templateDataStrDict))
	require.Nil(t, err)

	expectedTemplateAndData := binding_constructors.NewTemplateAndData(template, `{"Boolean":true,"LargeFloat":1231231243.43,"Name":"John","UnixTs":1257894000}`)
	expectedOutput := map[string]*kurtosis_core_rpc_api_bindings.RenderTemplatesToFilesArtifactArgs_TemplateAndData{
		"/foo/bar": expectedTemplateAndData,
	}

	output, err := ParseTemplatesAndData(input)
	require.Nil(t, err)
	require.Equal(t, expectedOutput, output)
}

func TestParseTemplatesAndData_SimpleCaseDict(t *testing.T) {
	dataDict := starlark.NewDict(4)
	err := dataDict.SetKey(starlark.String("LargeFloat"), starlark.Float(1231231243.43))
	require.Nil(t, err)
	err = dataDict.SetKey(starlark.String("Name"), starlark.String("John"))
	require.Nil(t, err)
	err = dataDict.SetKey(starlark.String("UnixTs"), starlark.MakeInt64(1257894000))
	require.Nil(t, err)
	err = dataDict.SetKey(starlark.String("Boolean"), starlark.Bool(true))
	require.Nil(t, err)
	templateDataStrDict := starlark.StringDict{}
	template := "Hello {{.Name}}. {{.LargeFloat}} {{.UnixTs}} {{.Boolean}}"
	templateDataStrDict["template"] = starlark.String(template)
	templateDataStrDict["data"] = dataDict
	input := starlark.NewDict(1)
	err = input.SetKey(starlark.String("/foo/bar"), starlarkstruct.FromStringDict(starlarkstruct.Default, templateDataStrDict))
	require.Nil(t, err)

	expectedTemplateAndData := binding_constructors.NewTemplateAndData(template, `{"Boolean":true,"LargeFloat":1231231243.43,"Name":"John","UnixTs":1257894000}`)
	expectedOutput := map[string]*kurtosis_core_rpc_api_bindings.RenderTemplatesToFilesArtifactArgs_TemplateAndData{
		"/foo/bar": expectedTemplateAndData,
	}

	output, err := ParseTemplatesAndData(input)
	require.Nil(t, err)
	require.Equal(t, expectedOutput, output)
}

func TestParseTemplatesAndData_FailsForDictWithIntegerKey(t *testing.T) {
	dataDict := starlark.NewDict(1)
	err := dataDict.SetKey(starlark.MakeInt(42), starlark.Float(1231231243.43))
	require.Nil(t, err)
	templateDataStrDict := starlark.StringDict{}
	template := "Hello {{.Name}}"
	templateDataStrDict["template"] = starlark.String(template)
	templateDataStrDict["data"] = dataDict
	input := starlark.NewDict(1)
	err = input.SetKey(starlark.String("/foo/bar"), starlarkstruct.FromStringDict(starlarkstruct.Default, templateDataStrDict))
	require.Nil(t, err)

	_, err = ParseTemplatesAndData(input)
	require.NotNil(t, err)
}

func TestParsePrivateIPAddressPlaceholder_Success(t *testing.T) {
	dict := starlark.StringDict{}
	dict["private_ip_address_placeholder"] = starlark.String("KURTOSIS_IP_ADDRESS")
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	output, err := parsePrivateIPAddressPlaceholder(input)
	require.Nil(t, err)
	require.Equal(t, "KURTOSIS_IP_ADDRESS", output)
}

func TestParsePrivateIPAddressPlaceholder_FailureNonString(t *testing.T) {
	dict := starlark.StringDict{}
	dict["private_ip_address_placeholder"] = starlark.MakeInt(42)
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, dict)
	_, err := parsePrivateIPAddressPlaceholder(input)
	require.NotNil(t, err)
}

func TestParseHttpRequestRecipe_GetRequestWithoutExtractor(t *testing.T) {
	inputDict := starlark.StringDict{}
	expectedRecipe := recipe.NewGetHttpRequestRecipe(
		"service_id",
		"port_id",
		"/",
		map[string]string{})
	inputDict["service_id"] = starlark.String("service_id")
	inputDict["port_id"] = starlark.String("port_id")
	inputDict["method"] = starlark.String("GET")
	inputDict["endpoint"] = starlark.String("/")
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, inputDict)
	actualRecipe, err := ParseHttpRequestRecipe(input)
	require.Nil(t, err)
	require.Equal(t, expectedRecipe, actualRecipe)
}

func TestParseHttpRequestRecipe_GetRequestWithExtractor(t *testing.T) {
	extractorDict := &starlark.Dict{}
	err := extractorDict.SetKey(starlark.String("key"), starlark.String(".value"))
	require.Nil(t, err)
	inputDict := starlark.StringDict{}
	expectedRecipe := recipe.NewGetHttpRequestRecipe(
		"service_id",
		"port_id",
		"/",
		map[string]string{
			"key": ".value",
		})
	inputDict["service_id"] = starlark.String("service_id")
	inputDict["port_id"] = starlark.String("port_id")
	inputDict["method"] = starlark.String("GET")
	inputDict["endpoint"] = starlark.String("/")
	inputDict["extract"] = extractorDict
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, inputDict)
	actualRecipe, err := ParseHttpRequestRecipe(input)
	require.Nil(t, err)
	require.Equal(t, expectedRecipe, actualRecipe)
}

func TestParseHttpRequestRecipe_PostRequestWithoutExtractor(t *testing.T) {
	inputDict := starlark.StringDict{}
	expectedRecipe := recipe.NewPostHttpRequestRecipe(
		"service_id",
		"port_id",
		"content/json",
		"/",
		"body",
		map[string]string{})
	inputDict["service_id"] = starlark.String("service_id")
	inputDict["port_id"] = starlark.String("port_id")
	inputDict["method"] = starlark.String("POST")
	inputDict["endpoint"] = starlark.String("/")
	inputDict["content_type"] = starlark.String("content/json")
	inputDict["body"] = starlark.String("body")
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, inputDict)
	actualRecipe, err := ParseHttpRequestRecipe(input)
	require.Nil(t, err)
	require.Equal(t, expectedRecipe, actualRecipe)
}

func TestParseHttpRequestRecipe_PostRequestWithExtractor(t *testing.T) {
	extractorDict := &starlark.Dict{}
	err := extractorDict.SetKey(starlark.String("key"), starlark.String(".value"))
	require.Nil(t, err)
	inputDict := starlark.StringDict{}
	expectedRecipe := recipe.NewPostHttpRequestRecipe(
		"service_id",
		"port_id",
		"content/json",
		"/",
		"body",
		map[string]string{
			"key": ".value",
		})
	inputDict["service_id"] = starlark.String("service_id")
	inputDict["port_id"] = starlark.String("port_id")
	inputDict["method"] = starlark.String("POST")
	inputDict["endpoint"] = starlark.String("/")
	inputDict["content_type"] = starlark.String("content/json")
	inputDict["body"] = starlark.String("body")
	inputDict["extract"] = extractorDict
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, inputDict)
	actualRecipe, err := ParseHttpRequestRecipe(input)
	require.Nil(t, err)
	require.Equal(t, expectedRecipe, actualRecipe)
}

func TestEncodeStarlarkObjectAsJSON_EncodesStructsCorrectly(t *testing.T) {
	structToJsonifyStrDict := starlark.StringDict{}
	structToJsonifyStrDict["foo"] = starlark.String("bar")
	structToJsonifyStrDict["buzz"] = starlark.MakeInt(42)
	structToJsonifyStrDict["fizz"] = starlark.Bool(false)
	structToJsonify := starlarkstruct.FromStringDict(starlarkstruct.Default, structToJsonifyStrDict)
	require.NotNil(t, structToJsonify)
	structJsonStr, err := encodeStarlarkObjectAsJSON(structToJsonify, "test")
	require.Nil(t, err)
	expectedStr := `{"buzz":42,"fizz":false,"foo":"bar"}`
	require.Equal(t, expectedStr, structJsonStr)
}
