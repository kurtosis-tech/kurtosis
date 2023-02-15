package kurtosis_types

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

func TestUpdateServiceConfig_StringRepresentation(t *testing.T) {
	updateServiceConfig := NewUpdateServiceConfig("subnetwork_1")
	expectedRepresentation := `UpdateServiceConfig(subnetwork="subnetwork_1")`
	require.Equal(t, expectedRepresentation, updateServiceConfig.String())
}

func TestUpdateServiceConfig_Type(t *testing.T) {
	updateServiceConfig := NewUpdateServiceConfig("subnetwork_1")
	require.Equal(t, UpdateServiceConfigTypeName, updateServiceConfig.Type())
}

func TestUpdateServiceConfig_Truth(t *testing.T) {
	updateServiceConfig := NewUpdateServiceConfig("subnetwork_1")
	require.Equal(t, starlark.Bool(true), updateServiceConfig.Truth())
}

func TestUpdateServiceConfig_Truth_EmptyString(t *testing.T) {
	updateServiceConfig := NewUpdateServiceConfig("")
	require.Equal(t, starlark.Bool(true), updateServiceConfig.Truth())
}

func TestUpdateServiceConfig_Attr_Exists(t *testing.T) {
	updateServiceConfig := NewUpdateServiceConfig("subnetwork_1")
	attr, err := updateServiceConfig.Attr(subnetworkAttr)
	require.Nil(t, err)
	require.Equal(t, starlark.String("subnetwork_1"), attr)
}

func TestUpdateServiceConfig_Attr_DoesNotExist(t *testing.T) {
	updateServiceConfig := NewUpdateServiceConfig("subnetwork_1")
	attr, err := updateServiceConfig.Attr("do-not-exist")
	expectedError := fmt.Sprintf("'%s' has no attribute 'do-not-exist'", UpdateServiceConfigTypeName)
	require.Equal(t, expectedError, err.Error())
	require.Nil(t, attr)
}

func TestUpdateServiceConfig_AttrNames(t *testing.T) {
	updateServiceConfig := NewUpdateServiceConfig("subnetwork_1")
	attrs := updateServiceConfig.AttrNames()
	expectedAttrs := []string{
		subnetworkAttr,
	}
	require.Equal(t, expectedAttrs, attrs)
}

func TestUpdateServiceConfig_MakeWithArgs(t *testing.T) {
	builtin := &starlark.Builtin{}
	args := starlark.Tuple([]starlark.Value{
		starlark.String("subnetwork_1"),
	})
	updateServiceConfig, err := MakeUpdateServiceConfig(nil, builtin, args, noKwargs)
	require.Nil(t, err)
	expectedUpdateServiceConfig := NewUpdateServiceConfig("subnetwork_1")
	require.Equal(t, expectedUpdateServiceConfig, updateServiceConfig)
}

func TestUpdateServiceConfig_MakeWithArgs_FailureBadArgument(t *testing.T) {
	builtin := &starlark.Builtin{}
	args := starlark.Tuple([]starlark.Value{
		starlark.Float(0),
	})
	updateServiceConfig, err := MakeUpdateServiceConfig(nil, builtin, args, noKwargs)
	expectedError := fmt.Sprintf(`Cannot construct '%s' from the provided arguments. Expecting a single argument '%s'
	Caused by: : for parameter %s: got float, want string`, UpdateServiceConfigTypeName, subnetworkAttr, subnetworkAttr)
	require.Equal(t, expectedError, err.Error())
	require.Nil(t, updateServiceConfig)
}

func TestUpdateServiceConfig_MakeWithKwargs(t *testing.T) {
	builtin := &starlark.Builtin{}
	kwargs := []starlark.Tuple{
		starlark.Tuple([]starlark.Value{
			starlark.String(subnetworkAttr),
			starlark.String("subnetwork_1"),
		}),
	}
	updateServiceConfig, err := MakeUpdateServiceConfig(nil, builtin, noArgs, kwargs)
	require.Nil(t, err)
	expectedUpdateServiceConfig := NewUpdateServiceConfig("subnetwork_1")
	require.Equal(t, expectedUpdateServiceConfig, updateServiceConfig)
}

func TestUpdateServiceConfig_MakeWithKwargs_FailureWrongArg(t *testing.T) {
	builtin := &starlark.Builtin{}
	kwargs := []starlark.Tuple{
		starlark.Tuple([]starlark.Value{
			starlark.String(subnetworkAttr),
			starlark.String("subnetwork_1"),
		}),
		starlark.Tuple([]starlark.Value{
			starlark.String("unknown-kwarg"),
			starlark.Float(50),
		}),
	}
	updateServiceConfig, err := MakeUpdateServiceConfig(nil, builtin, noArgs, kwargs)
	expectedError := fmt.Sprintf(`Cannot construct '%s' from the provided arguments. Expecting a single argument '%s'
	Caused by: : unexpected keyword argument "unknown-kwarg"`, UpdateServiceConfigTypeName, subnetworkAttr)
	require.Equal(t, expectedError, err.Error())
	require.Nil(t, updateServiceConfig)
}

func TestUpdateServiceConfig_ToKurtosisType(t *testing.T) {
	updateServiceConfig := NewUpdateServiceConfig("subnetwork_1")
	expectedKurtosisType := binding_constructors.NewUpdateServiceConfig("subnetwork_1")
	require.Equal(t, expectedKurtosisType, updateServiceConfig.ToKurtosisType())
}
