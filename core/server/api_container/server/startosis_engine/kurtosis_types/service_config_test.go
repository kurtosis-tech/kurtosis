package kurtosis_types

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

var (
	image       = "nginx"
	fakeBuiltin = &starlark.Builtin{}
)

func TestServiceConfig_StringRepresentation(t *testing.T) {
	serviceConfig := newServiceConfig(image)
	expectedRepresentation := fmt.Sprintf(`%s(%s="%s")`, ServiceConfigTypeName, serviceConfigImageAttr, image)
	require.Equal(t, expectedRepresentation, serviceConfig.String())
}

func TestServiceConfig_Type(t *testing.T) {
	serviceConfig := newServiceConfig(image)
	require.Equal(t, ServiceConfigTypeName, serviceConfig.Type())
}

func TestServiceConfig_Truth_True(t *testing.T) {
	serviceConfig := newServiceConfig(image)
	require.Equal(t, starlark.Bool(true), serviceConfig.Truth())
}

func TestServiceConfig_Truth_False(t *testing.T) {
	serviceConfig := newServiceConfig("")
	require.Equal(t, starlark.Bool(false), serviceConfig.Truth())
}

func TestServiceConfig_Attr_Exists(t *testing.T) {
	serviceConfig := newServiceConfig(image)
	attr, err := serviceConfig.Attr(serviceConfigImageAttr)
	require.Nil(t, err)
	require.Equal(t, starlark.String(image), attr)

	attr, err = serviceConfig.Attr(serviceConfigPortsAttr)
	require.Nil(t, err)
	require.Nil(t, attr)

	attr, err = serviceConfig.Attr(serviceConfigPublicPortsAttr)
	require.Nil(t, err)
	require.Nil(t, attr)

	attr, err = serviceConfig.Attr(serviceConfigFilesAttr)
	require.Nil(t, err)
	require.Nil(t, attr)

	attr, err = serviceConfig.Attr(serviceConfigEntrypointAttr)
	require.Nil(t, err)
	require.Nil(t, attr)

	attr, err = serviceConfig.Attr(serviceConfigCmdAttr)
	require.Nil(t, err)
	require.Nil(t, attr)

	attr, err = serviceConfig.Attr(serviceConfigEnvVarsAttr)
	require.Nil(t, err)
	require.Nil(t, attr)

	attr, err = serviceConfig.Attr(serviceConfigPrivateIpAddressPlaceholderAttr)
	require.Nil(t, err)
	require.Nil(t, attr)

	attr, err = serviceConfig.Attr(serviceConfigSubnetworkAttr)
	require.Nil(t, err)
	require.Nil(t, attr)
}

func TestServiceConfig_Attr_DoesNotExist(t *testing.T) {
	serviceConfig := newServiceConfig(image)
	attr, err := serviceConfig.Attr("do-not-exist")
	expectedError := fmt.Sprintf("'%s' has no attribute 'do-not-exist'", ServiceConfigTypeName)
	require.Equal(t, expectedError, err.Error())
	require.Nil(t, attr)
}

func TestServiceConfig_AttrNames(t *testing.T) {
	serviceConfig := newServiceConfig(image)
	attrs := serviceConfig.AttrNames()
	expectedAttrs := []string{
		serviceConfigImageAttr,
		serviceConfigPortsAttr,
		serviceConfigPublicPortsAttr,
		serviceConfigFilesAttr,
		serviceConfigEntrypointAttr,
		serviceConfigCmdAttr,
		serviceConfigEnvVarsAttr,
		serviceConfigPrivateIpAddressPlaceholderAttr,
		serviceConfigSubnetworkAttr,
	}
	require.Equal(t, expectedAttrs, attrs)
}

func TestServiceConfig_MakeWithArgs_Minimal(t *testing.T) {
	args := starlark.Tuple([]starlark.Value{
		starlark.String(image),
	})
	serviceConfig, err := MakeServiceConfig(nil, fakeBuiltin, args, noKwargs)
	require.Nil(t, err)
	expectedConnectionResult := newServiceConfig(image)
	require.Equal(t, expectedConnectionResult, serviceConfig)
}

func TestServiceConfig_MakeWithArgs_Full(t *testing.T) {
	privatePorts := newPortsMap(t, 1323)
	publicPorts := newPortsMap(t, 80)
	files := newStarlarkDict(t, "/path/to/file", "file1")
	entrypoint := newStarlarkList("bash")
	cmd := newStarlarkList("-c sleep 99")
	envVars := newStarlarkDict(t, "VAR", "VALUE")
	privateIpAddressPlaceholder := starlark.String("<IP_ADDRESS>")
	subnetwork := starlark.String("subnetwork_1")

	args := starlark.Tuple([]starlark.Value{
		starlark.String(image),
		privatePorts,
		publicPorts,
		files,
		entrypoint,
		cmd,
		envVars,
		privateIpAddressPlaceholder,
		subnetwork,
	})
	serviceConfig, err := MakeServiceConfig(nil, fakeBuiltin, args, noKwargs)
	require.Nil(t, err)
	expectedConnectionResult := NewServiceConfig(starlark.String(image), privatePorts, publicPorts, files, entrypoint, cmd, envVars, &privateIpAddressPlaceholder, &subnetwork)
	require.Equal(t, expectedConnectionResult, serviceConfig)
}

func TestServiceConfig_MakeWithKwargs_Minimal(t *testing.T) {
	kwargs := []starlark.Tuple{
		starlark.Tuple([]starlark.Value{
			starlark.String(serviceConfigImageAttr),
			starlark.String(image),
		}),
	}
	serviceConfig, err := MakeServiceConfig(nil, fakeBuiltin, noArgs, kwargs)
	require.Nil(t, err)
	expectedConnectionResult := newServiceConfig(image)
	require.Equal(t, expectedConnectionResult, serviceConfig)
}

func TestServiceConfig_MakeWithKwargs_Full(t *testing.T) {
	privatePorts := newPortsMap(t, 1323)
	publicPorts := newPortsMap(t, 80)
	files := newStarlarkDict(t, "/path/to/file", "file1")
	entrypoint := newStarlarkList("bash")
	cmd := newStarlarkList("-c sleep 99")
	envVars := newStarlarkDict(t, "VAR", "VALUE")
	privateIpAddressPlaceholder := starlark.String("<IP_ADDRESS>")
	subnetwork := starlark.String("subnetwork_1")

	kwargs := []starlark.Tuple{
		starlark.Tuple([]starlark.Value{
			starlark.String(serviceConfigImageAttr),
			starlark.String(image),
		}),
		starlark.Tuple([]starlark.Value{
			starlark.String(serviceConfigPortsAttr),
			privatePorts,
		}),
		starlark.Tuple([]starlark.Value{
			starlark.String(serviceConfigPublicPortsAttr),
			publicPorts,
		}),
		starlark.Tuple([]starlark.Value{
			starlark.String(serviceConfigFilesAttr),
			files,
		}),
		starlark.Tuple([]starlark.Value{
			starlark.String(serviceConfigEntrypointAttr),
			entrypoint,
		}),
		starlark.Tuple([]starlark.Value{
			starlark.String(serviceConfigCmdAttr),
			cmd,
		}),
		starlark.Tuple([]starlark.Value{
			starlark.String(serviceConfigEnvVarsAttr),
			envVars,
		}),
		starlark.Tuple([]starlark.Value{
			starlark.String(serviceConfigPrivateIpAddressPlaceholderAttr),
			privateIpAddressPlaceholder,
		}),
		starlark.Tuple([]starlark.Value{
			starlark.String(serviceConfigSubnetworkAttr),
			subnetwork,
		}),
	}

	serviceConfig, err := MakeServiceConfig(nil, fakeBuiltin, noArgs, kwargs)
	require.Nil(t, err)
	expectedConnectionResult := NewServiceConfig(starlark.String(image), privatePorts, publicPorts, files, entrypoint, cmd, envVars, &privateIpAddressPlaceholder, &subnetwork)
	require.Equal(t, expectedConnectionResult, serviceConfig)
}

func TestServiceConfig_ToKurtosisType(t *testing.T) {
	serviceConfig := newServiceConfig(image)
	expectedKurtosisType := services.NewServiceConfigBuilder(image).Build()
	convertedServiceConfig, err := serviceConfig.ToKurtosisType()
	require.Nil(t, err)
	require.Equal(t, expectedKurtosisType, convertedServiceConfig)
}

func newServiceConfig(image string) *ServiceConfig {
	return NewServiceConfig(
		starlark.String(image),
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
}

func newPortsMap(t *testing.T, portNum uint32) *starlark.Dict {
	privatePorts := starlark.NewDict(1)
	require.Nil(t, privatePorts.SetKey(starlark.String("grpc"), NewPortSpec(portNum, kurtosis_core_rpc_api_bindings.Port_TCP, "http")))
	return privatePorts
}

func newStarlarkDict(t *testing.T, key string, value string) *starlark.Dict {
	dict := starlark.NewDict(1)
	require.Nil(t, dict.SetKey(starlark.String(key), starlark.String(value)))
	return dict
}

func newStarlarkList(element string) *starlark.List {
	list := starlark.NewList([]starlark.Value{
		starlark.String(element),
	})
	return list
}
