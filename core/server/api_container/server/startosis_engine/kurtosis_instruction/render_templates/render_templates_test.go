package render_templates

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/render_templates"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"testing"
)

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

	expectedTemplateAndData, err := render_templates.CreateTemplateData(template, `{"Boolean":true,"LargeFloat":1231231243.43,"Name":"John","UnixTs":1257894000}`)
	require.Nil(t, err)
	expectedOutput := map[string]*render_templates.TemplateData{
		"/foo/bar": expectedTemplateAndData,
	}

	output, err := parseTemplatesAndData(input)
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

	expectedTemplateAndData, err := render_templates.CreateTemplateData(template, `{"Boolean":true,"LargeFloat":1231231243.43,"Name":"John","UnixTs":1257894000}`)
	require.Nil(t, err)
	expectedOutput := map[string]*render_templates.TemplateData{
		"/foo/bar": expectedTemplateAndData,
	}

	output, err := parseTemplatesAndData(input)
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

	_, err = parseTemplatesAndData(input)
	require.NotNil(t, err)
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
