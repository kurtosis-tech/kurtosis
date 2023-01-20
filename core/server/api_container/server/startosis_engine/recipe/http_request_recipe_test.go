package recipe

import (
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

var (
	noArgs = starlark.Tuple{}
)

func TestGetHttpRequestRecipe_String(t *testing.T) {
	builtin := &starlark.Builtin{}
	builtin.Name()
	kwargs := []starlark.Tuple{
		starlark.Tuple([]starlark.Value{
			starlark.String(serviceNameKey),
			starlark.String("web-server"),
		}),
		starlark.Tuple([]starlark.Value{
			starlark.String(endpointAttr),
			starlark.String("?input=output"),
		}),
		starlark.Tuple([]starlark.Value{
			starlark.String(portIdAttr),
			starlark.String("portId"),
		}),
	}
	getHttpRequestRecipe, err := MakeGetHttpRequestRecipe(nil, builtin, noArgs, kwargs)
	require.Nil(t, err, "Unexpected error occurred")

	getHttpRequestRecipeString := getHttpRequestRecipe.String()
	expectedStringOutput := `GetHttpRequestRecipe(port_id="portId", service_name="web-server", endpoint="?input=output", extract="")`
	require.NotNil(t, expectedStringOutput, getHttpRequestRecipeString)

	extractors := starlark.NewDict(1)
	err = extractors.SetKey(starlark.String("field"), starlark.String(".input.*"))
	require.Nil(t, err)
	kwargsWithExtractors := []starlark.Tuple{
		starlark.Tuple([]starlark.Value{
			starlark.String(serviceNameKey),
			starlark.String("web-server"),
		}),
		starlark.Tuple([]starlark.Value{
			starlark.String(endpointAttr),
			starlark.String("?input=output"),
		}),
		starlark.Tuple([]starlark.Value{
			starlark.String(portIdAttr),
			starlark.String("portId"),
		}),
		starlark.Tuple([]starlark.Value{
			starlark.String(extractKeyPrefix),
			extractors,
		}),
	}

	getHttpRequestRecipeWithExtractors, err := MakeGetHttpRequestRecipe(nil, builtin, noArgs, kwargsWithExtractors)
	require.Nil(t, err, "Unexpected error occurred")

	getHttpRequestRecipeWithExtractorsString := getHttpRequestRecipeWithExtractors.String()
	expectedStringOutputWithExtractors := `GetHttpRequestRecipe(port_id="portId", service_name="web-server", endpoint="?input=output", extract="{\"field\": \".input.*\"}")`
	require.NotNil(t, expectedStringOutputWithExtractors, getHttpRequestRecipeWithExtractorsString)
}

func TestPostHttpRequestRecipe_String(t *testing.T) {
	builtin := &starlark.Builtin{}
	kwargs := []starlark.Tuple{
		starlark.Tuple([]starlark.Value{
			starlark.String(serviceNameKey),
			starlark.String("web-server"),
		}),
		starlark.Tuple([]starlark.Value{
			starlark.String(endpointAttr),
			starlark.String("?input=output"),
		}),
		starlark.Tuple([]starlark.Value{
			starlark.String(portIdAttr),
			starlark.String("portId"),
		}),
		starlark.Tuple([]starlark.Value{
			starlark.String(bodyKey),
			starlark.String("body"),
		}),
		starlark.Tuple([]starlark.Value{
			starlark.String(contentTypeAttr),
			starlark.String("content-type"),
		}),
	}
	postHttpRequestRecipe, err := MakePostHttpRequestRecipe(nil, builtin, noArgs, kwargs)
	require.Nil(t, err, "Unexpected error occurred")

	postHttpRequestRecipeString := postHttpRequestRecipe.String()
	expectedStringOutput := `PostHttpRequestRecipe(port_id="portId", service_name="web-server", endpoint="?input=output", body="body", content_type="content-type", extract="")`
	require.NotNil(t, expectedStringOutput, postHttpRequestRecipeString)

	extractors := starlark.NewDict(1)
	err = extractors.SetKey(starlark.String("field"), starlark.String(".input.*"))
	require.Nil(t, err)
	kwargsWithExtractors := []starlark.Tuple{
		starlark.Tuple([]starlark.Value{
			starlark.String(serviceNameKey),
			starlark.String("web-server"),
		}),
		starlark.Tuple([]starlark.Value{
			starlark.String(endpointAttr),
			starlark.String("?input=output"),
		}),
		starlark.Tuple([]starlark.Value{
			starlark.String(portIdAttr),
			starlark.String("portId"),
		}),
		starlark.Tuple([]starlark.Value{
			starlark.String(extractKeyPrefix),
			extractors,
		}),
		starlark.Tuple([]starlark.Value{
			starlark.String(bodyKey),
			starlark.String("body"),
		}),
		starlark.Tuple([]starlark.Value{
			starlark.String(contentTypeAttr),
			starlark.String("content-type"),
		}),
	}

	postHttpRequestRecipeWithExtractors, err := MakePostHttpRequestRecipe(nil, builtin, noArgs, kwargsWithExtractors)
	require.Nil(t, err, "Unexpected error occurred")

	postHttpRequestRecipeWithExtractorsString := postHttpRequestRecipeWithExtractors.String()
	expectedStringOutputWithExtractors := `PostHttpRequestRecipe(port_id="portId", service_name="web-server", endpoint="?input=output", body="body", content_type="content-type", extract="{\"field\": \".input.*\"}")`
	require.NotNil(t, expectedStringOutputWithExtractors, postHttpRequestRecipeWithExtractorsString)
}

func TestStartosisInterpreter_HttpRequestMissingRequiredFields(t *testing.T) {
	builtin := &starlark.Builtin{}
	kwargs := []starlark.Tuple{
		starlark.Tuple([]starlark.Value{
			starlark.String(serviceNameKey),
			starlark.String("web-server"),
		}),
		starlark.Tuple([]starlark.Value{
			starlark.String(endpointAttr),
			starlark.String("?input=output"),
		}),
	}
	getHttpRequestRecipe, err := MakeGetHttpRequestRecipe(nil, builtin, noArgs, kwargs)
	expectedError := "missing argument for port_id"
	require.Contains(t, err.Error(), expectedError)
	require.Nil(t, getHttpRequestRecipe)
}

func TestStartosisInterpreter_MissingRequiredFieldForHttpRecipeWithPostMethod(t *testing.T) {
	builtin := &starlark.Builtin{}
	extractors := starlark.NewDict(1)
	err := extractors.SetKey(starlark.String("field"), starlark.String(".input.*"))
	require.Nil(t, err)
	kwargsWithoutBody := []starlark.Tuple{
		starlark.Tuple([]starlark.Value{
			starlark.String(serviceNameKey),
			starlark.String("web-server"),
		}),
		starlark.Tuple([]starlark.Value{
			starlark.String(endpointAttr),
			starlark.String("?input=output"),
		}),
		starlark.Tuple([]starlark.Value{
			starlark.String(portIdAttr),
			starlark.String("portId"),
		}),
		starlark.Tuple([]starlark.Value{
			starlark.String(extractKeyPrefix),
			extractors,
		}),
		starlark.Tuple([]starlark.Value{
			starlark.String(contentTypeAttr),
			starlark.String("content-type"),
		}),
	}

	postHttpRequestRecipe, err := MakePostHttpRequestRecipe(nil, builtin, noArgs, kwargsWithoutBody)
	expectedError := "missing argument for body"
	require.Contains(t, err.Error(), expectedError)
	require.Nil(t, postHttpRequestRecipe)
}
