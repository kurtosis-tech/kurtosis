package service_config

import (
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

func TestFilesInputSimpleString(t *testing.T) {
	dict := newStarlarkDictContainingString("/mount", "solo-file.txt")
	filesArtifacts, _, _ := convertFilesArguments("files", dict)

	expected := map[string]string{
		"/mount": "solo-file.txt",
	}

	require.Equal(t, expected, filesArtifacts)
}

func TestFilesInputWithArray(t *testing.T) {
	dict := newStarlarkDictContainingArray("/mount", []string{"file-one.txt", "file-two.txt"})
	filesArtifacts, _, _ := convertFilesArguments("files", dict)

	expected := map[string]interface{}{
		"/mount": []interface{}{"file-one.txt", "file-two.txt"},
	}

	require.Equal(t, expected, filesArtifacts)
}

func newStarlarkDictContainingString(key string, value string) *starlark.Dict {
	dict := starlark.NewDict(1)
	starlarkKey := starlark.String(key)
	starlarkValue := starlark.String(value)
	dict.SetKey(starlarkKey, starlarkValue)
	return dict
}

func newStarlarkDictContainingArray(key string, values []string) *starlark.Dict {
	dict := starlark.NewDict(1)
	starlarkKey := starlark.String(key)
	starlarkValues := toStarlarkStringArray(values)
	dict.SetKey(starlarkKey, starlark.NewList(starlarkValues))
	return dict
}

func toStarlarkStringArray(strings []string) []starlark.Value {
	starlarkValues := make([]starlark.Value, len(strings))
	for i, str := range strings {
		starlarkValues[i] = starlark.String(str)
	}
	return starlarkValues
}
