package services

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	portWaitTimeout = "2s"
)

func TestServiceConfigBuilderFrom_Invariant(t *testing.T) {
	initialServiceConfig := NewServiceConfigBuilder(
		"test-image",
	).WithPrivatePorts(
		map[string]*kurtosis_core_rpc_api_bindings.Port{
			"grpc": binding_constructors.NewPort(1323, kurtosis_core_rpc_api_bindings.Port_TCP, "https", portWaitTimeout),
		},
	).WithPublicPorts(
		map[string]*kurtosis_core_rpc_api_bindings.Port{
			"grpc": binding_constructors.NewPort(1323, kurtosis_core_rpc_api_bindings.Port_TCP, "https", portWaitTimeout),
		},
	).WithEntryPointArgs(
		[]string{"echo", "'Hello World!'"},
	).WithCmdArgs(
		[]string{"sleep", "999999"},
	).WithEnvVars(
		map[string]string{
			"VAR_1": "VALUE",
		},
	).WithFilesArtifactMountDirpaths(
		map[string]string{
			"/path/to/file": "artifact",
		},
	).WithCpuAllocationMillicpus(
		1000,
	).WithMemoryAllocationMegabytes(
		1024,
	).WithPrivateIPAddressPlaceholder(
		"<IP_ADDRESS>",
	).WithSubnetwork(
		"subnetwork_1",
	).Build()

	newServiceConfigBuilder := NewServiceConfigBuilderFromServiceConfig(initialServiceConfig)
	newServiceConfig := newServiceConfigBuilder.Build()

	require.Equal(t, initialServiceConfig, newServiceConfig)

	// modify random values
	newServiceConfig.ContainerImageName = "new-test-image"
	newServiceConfig.PrivatePorts["grpc"].Number = 9876
	newServiceConfig.PublicPorts["grpc"].MaybeApplicationProtocol = "ftp"
	newServiceConfig.EntrypointArgs[0] = "new-echo"
	newServiceConfig.CmdArgs[1] = "1234"
	newServiceConfig.EnvVars["VAR_1"] = "NEW_VALUE"
	newServiceConfig.FilesArtifactMountpoints["new/path/to/file"] = "new-artifact"
	newServiceConfig.CpuAllocationMillicpus = 500
	newServiceConfig.MemoryAllocationMegabytes = 512
	newSubnetwork := "new-subnetwork"
	newServiceConfig.Subnetwork = &newSubnetwork

	// test that initial value has not changed
	require.Equal(t, "test-image", initialServiceConfig.ContainerImageName)
	require.Equal(t, binding_constructors.NewPort(1323, kurtosis_core_rpc_api_bindings.Port_TCP, "https", portWaitTimeout), initialServiceConfig.PrivatePorts["grpc"])
	require.Equal(t, binding_constructors.NewPort(1323, kurtosis_core_rpc_api_bindings.Port_TCP, "https", portWaitTimeout), initialServiceConfig.PublicPorts["grpc"])
	require.Equal(t, []string{"echo", "'Hello World!'"}, initialServiceConfig.EntrypointArgs)
	require.Equal(t, []string{"sleep", "999999"}, initialServiceConfig.CmdArgs)
	require.Equal(t, map[string]string{"VAR_1": "VALUE"}, initialServiceConfig.EnvVars)
	require.Equal(t, map[string]string{"/path/to/file": "artifact"}, initialServiceConfig.FilesArtifactMountpoints)
	require.Equal(t, uint64(1000), initialServiceConfig.CpuAllocationMillicpus)
	require.Equal(t, uint64(1024), initialServiceConfig.MemoryAllocationMegabytes)
	require.Equal(t, "subnetwork_1", *initialServiceConfig.Subnetwork)

	// test that new value are as expected
	require.Equal(t, "new-test-image", newServiceConfig.ContainerImageName)
	require.Equal(t, binding_constructors.NewPort(9876, kurtosis_core_rpc_api_bindings.Port_TCP, "https", portWaitTimeout), newServiceConfig.PrivatePorts["grpc"])
	require.Equal(t, binding_constructors.NewPort(1323, kurtosis_core_rpc_api_bindings.Port_TCP, "ftp", portWaitTimeout), newServiceConfig.PublicPorts["grpc"])
	require.Equal(t, []string{"new-echo", "'Hello World!'"}, newServiceConfig.EntrypointArgs)
	require.Equal(t, []string{"sleep", "1234"}, newServiceConfig.CmdArgs)
	require.Equal(t, map[string]string{"VAR_1": "NEW_VALUE"}, newServiceConfig.EnvVars)
	require.Equal(t, map[string]string{"/path/to/file": "artifact", "new/path/to/file": "new-artifact"}, newServiceConfig.FilesArtifactMountpoints)
	require.Equal(t, uint64(500), newServiceConfig.CpuAllocationMillicpus)
	require.Equal(t, uint64(512), newServiceConfig.MemoryAllocationMegabytes)
	require.Equal(t, "new-subnetwork", *newServiceConfig.Subnetwork)
}
