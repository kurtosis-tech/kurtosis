package kurtosis_instruction

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/recipe"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"testing"
)

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

func TestParseHttpRequestRecipe_GetRequestWithoutExtractor(t *testing.T) {
	inputDict := starlark.StringDict{}
	expectedRecipe := recipe.NewGetHttpRequestRecipe(
		"service_name",
		"port_id",
		"/",
		map[string]string{})
	inputDict["service_name"] = starlark.String("service_name")
	inputDict["port_id"] = starlark.String("port_id")
	inputDict["method"] = starlark.String("GET")
	inputDict["endpoint"] = starlark.String("/")
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, inputDict)
	actualRecipe, err := ParseHttpRequestRecipe(input)
	require.Nil(t, err)
	require.Equal(t, expectedRecipe, actualRecipe)
}

func TestParseExecRecipe(t *testing.T) {
	inputDict := starlark.StringDict{}
	expectedRecipe := recipe.NewExecRecipe(
		"service_name",
		[]string{"cd", ".."})
	inputDict["service_name"] = starlark.String("service_name")
	inputDict["command"] = starlark.NewList([]starlark.Value{starlark.String("cd"), starlark.String("..")})
	input := starlarkstruct.FromStringDict(starlarkstruct.Default, inputDict)
	actualRecipe, err := ParseExecRecipe(input)
	require.Nil(t, err)
	require.Equal(t, expectedRecipe, actualRecipe)
}

func TestParseHttpRequestRecipe_GetRequestWithExtractor(t *testing.T) {
	extractorDict := &starlark.Dict{}
	err := extractorDict.SetKey(starlark.String("key"), starlark.String(".value"))
	require.Nil(t, err)
	inputDict := starlark.StringDict{}
	expectedRecipe := recipe.NewGetHttpRequestRecipe(
		"service_name",
		"port_id",
		"/",
		map[string]string{
			"key": ".value",
		})
	inputDict["service_name"] = starlark.String("service_name")
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
		"service_name",
		"port_id",
		"content/json",
		"/",
		"body",
		map[string]string{})
	inputDict["service_name"] = starlark.String("service_name")
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
		"service_name",
		"port_id",
		"content/json",
		"/",
		"body",
		map[string]string{
			"key": ".value",
		})
	inputDict["service_name"] = starlark.String("service_name")
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

func TestParseSubnetworks_ValidArg(t *testing.T) {
	expectedPartition1 := "subnetwork_1"
	expectedPartition2 := "subnetwork_2"
	subnetworks := starlark.Tuple([]starlark.Value{
		starlark.String(expectedPartition1),
		starlark.String(expectedPartition2),
	})
	partition1, partition2, err := ParseSubnetworks(subnetworks)
	require.Nil(t, err)
	require.Equal(t, service_network_types.PartitionID(expectedPartition1), partition1)
	require.Equal(t, service_network_types.PartitionID(expectedPartition2), partition2)
}

func TestParseSubnetworks_TooManySubnetworks(t *testing.T) {
	expectedPartition1 := "subnetwork_1"
	expectedPartition2 := "subnetwork_2"
	expectedPartition3 := "subnetwork_3"
	subnetworks := starlark.Tuple([]starlark.Value{
		starlark.String(expectedPartition1),
		starlark.String(expectedPartition2),
		starlark.String(expectedPartition3),
	})
	partition1, partition2, err := ParseSubnetworks(subnetworks)
	require.Contains(t, err.Error(), "Subnetworks tuple should contain exactly 2 subnetwork names. 3 was/were provided")
	require.Empty(t, partition1)
	require.Empty(t, partition2)
}

func TestParseSubnetworks_TooFewSubnetworks(t *testing.T) {
	expectedPartition1 := "subnetwork_1"
	subnetworks := starlark.Tuple([]starlark.Value{
		starlark.String(expectedPartition1),
	})
	partition1, partition2, err := ParseSubnetworks(subnetworks)
	require.Contains(t, err.Error(), "Subnetworks tuple should contain exactly 2 subnetwork names. 1 was/were provided")
	require.Empty(t, partition1)
	require.Empty(t, partition2)
}
