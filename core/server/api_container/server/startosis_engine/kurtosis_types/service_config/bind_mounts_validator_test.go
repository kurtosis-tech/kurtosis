package service_config

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
)

func TestValidateBindMounts_AllowsDockerSocket(t *testing.T) {
	dict := starlark.NewDict(1)
	require.NoError(t, dict.SetKey(starlark.String("/var/run/docker.sock"), starlark.String("/var/run/docker.sock")))

	err := ValidateBindMounts(dict)
	require.Nil(t, err)
}

func TestValidateBindMounts_AllowsDockerSocketWithCustomContainerPath(t *testing.T) {
	dict := starlark.NewDict(1)
	require.NoError(t, dict.SetKey(starlark.String("/var/run/docker.sock"), starlark.String("/run/docker.sock")))

	err := ValidateBindMounts(dict)
	require.Nil(t, err)
}

func TestValidateBindMounts_RejectsArbitraryHostPath(t *testing.T) {
	dict := starlark.NewDict(1)
	require.NoError(t, dict.SetKey(starlark.String("/etc/passwd"), starlark.String("/x")))

	err := ValidateBindMounts(dict)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "/etc/passwd")
	require.Contains(t, err.Error(), "not permitted")
}

func TestValidateBindMounts_RejectsHostRoot(t *testing.T) {
	dict := starlark.NewDict(1)
	require.NoError(t, dict.SetKey(starlark.String("/"), starlark.String("/host")))

	err := ValidateBindMounts(dict)
	require.NotNil(t, err)
}

func TestValidateBindMounts_RejectsNonDictValue(t *testing.T) {
	err := ValidateBindMounts(starlark.String("not a dict"))
	require.NotNil(t, err)
}

func TestValidateBindMounts_EmptyDictIsValid(t *testing.T) {
	dict := starlark.NewDict(0)

	err := ValidateBindMounts(dict)
	require.Nil(t, err)
}
