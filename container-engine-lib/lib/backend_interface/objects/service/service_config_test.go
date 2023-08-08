package service

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestServiceConfigMarshallers(t *testing.T) {

	imageName := "imageNameTest"
	originalServiceConfig := GetServiceConfigForTest(t, imageName)

	marshaledServiceConfig, err := json.Marshal(originalServiceConfig)
	require.NoError(t, err)
	require.NotNil(t, marshaledServiceConfig)

	newServiceConfig := &ServiceConfig{}

	err = json.Unmarshal(marshaledServiceConfig, newServiceConfig)
	require.NoError(t, err)

	require.Equal(t, originalServiceConfig.GetContainerImageName(), newServiceConfig.GetContainerImageName())

	originalServiceConfigPrivatePorts := originalServiceConfig.GetPrivatePorts()
	for privatePortId, privatePortSpec := range newServiceConfig.GetPrivatePorts() {
		originalPrivetPortSpec, found := originalServiceConfigPrivatePorts[privatePortId]
		require.True(t, found)
		require.EqualValues(t, privatePortSpec, originalPrivetPortSpec)
	}

	originalServiceConfigPublicPorts := originalServiceConfig.GetPublicPorts()
	for privatePortId, publicPortSpec := range newServiceConfig.GetPublicPorts() {
		originalPublicPortSpec, found := originalServiceConfigPublicPorts[privatePortId]
		require.True(t, found)
		require.EqualValues(t, publicPortSpec, originalPublicPortSpec)
	}

	require.Equal(t, originalServiceConfig.GetEnvVars(), newServiceConfig.GetEnvVars())
	require.Equal(t, originalServiceConfig.GetCmdArgs(), newServiceConfig.GetCmdArgs())
	require.Equal(t, originalServiceConfig.GetEnvVars(), newServiceConfig.GetEnvVars())
	require.EqualValues(t, originalServiceConfig.GetPersistentDirectories(), newServiceConfig.GetPersistentDirectories())
	require.EqualValues(t, originalServiceConfig.GetPersistentDirectories(), newServiceConfig.GetPersistentDirectories())
	require.Equal(t, originalServiceConfig.GetCPUAllocationMillicpus(), newServiceConfig.GetCPUAllocationMillicpus())
	require.Equal(t, originalServiceConfig.GetMemoryAllocationMegabytes(), newServiceConfig.GetMemoryAllocationMegabytes())
	require.Equal(t, originalServiceConfig.GetPrivateIPAddrPlaceholder(), newServiceConfig.GetPrivateIPAddrPlaceholder())
	require.Equal(t, originalServiceConfig.GetMinCPUAllocationMillicpus(), newServiceConfig.GetMinCPUAllocationMillicpus())
	require.Equal(t, originalServiceConfig.GetMinMemoryAllocationMegabytes(), newServiceConfig.GetMinMemoryAllocationMegabytes())
}
