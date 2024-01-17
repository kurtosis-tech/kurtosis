package service

import (
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_build_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service_directory"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestServiceConfigMarshallers(t *testing.T) {

	imageName := "imageNameTest"
	originalServiceConfig := getServiceConfigForTest(t, imageName)

	marshaledServiceConfig, err := json.Marshal(originalServiceConfig)
	require.NoError(t, err)
	require.NotNil(t, marshaledServiceConfig)

	// Suppressing exhaustruct requirement because we want an object with zero values
	// nolint: exhaustruct
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
	require.Equal(t, originalServiceConfig.GetLabels(), newServiceConfig.GetLabels())
	require.Equal(t, originalServiceConfig.GetImageBuildSpec(), newServiceConfig.GetImageBuildSpec())
}

func getServiceConfigForTest(t *testing.T, imageName string) *ServiceConfig {
	serviceConfig, err := CreateServiceConfig(
		imageName,
		testImageBuildSpec(),
		nil,
		testPrivatePorts(t),
		testPublicPorts(t),
		[]string{"bin", "bash", "ls"},
		[]string{"-l", "-a"},
		testEnvVars(),
		testFilesArtifactExpansion(),
		testPersistentDirectory(),
		500,
		1024,
		"IP-ADDRESS",
		100,
		512,
		map[string]string{
			"test-label-key":        "test-label-value",
			"test-second-label-key": "test-second-label-value",
		},
		nil,
	)
	require.NoError(t, err)
	return serviceConfig
}

func testPersistentDirectory() *service_directory.PersistentDirectories {
	persistentDirectoriesMap := map[string]service_directory.PersistentDirectory{
		"dirpath1": {PersistentKey: service_directory.DirectoryPersistentKey("dirpath1_persistent_directory_key"), Size: service_directory.DirectoryPersistentSize(int64(0))},
		"dirpath2": {PersistentKey: service_directory.DirectoryPersistentKey("dirpath2_persistent_directory_key"), Size: service_directory.DirectoryPersistentSize(int64(0))},
	}

	return service_directory.NewPersistentDirectories(persistentDirectoriesMap)
}

func testFilesArtifactExpansion() *service_directory.FilesArtifactsExpansion {
	return &service_directory.FilesArtifactsExpansion{
		ExpanderImage: "expander-image:tag-version",
		ExpanderEnvVars: map[string]string{
			"ENV_VAR1": "env_var1_value",
			"ENV_VAR2": "env_var2_value",
		},
		ServiceDirpathsToArtifactIdentifiers: map[string][]string{
			"/path/number1": {"first_identifier"},
			"/path/number2": {"second_identifier"},
		},
		ExpanderDirpathsToServiceDirpaths: map[string]string{
			"/expander/dir1": "/service/dir1",
			"/expander/dir2": "/service/dir2",
		},
	}
}

func testPrivatePorts(t *testing.T) map[string]*port_spec.PortSpec {

	port1Id := "port1"
	port1Num := uint16(23)
	port1Protocol := port_spec.TransportProtocol_TCP
	appProtocol1 := "app-protocol1"
	wait1 := port_spec.NewWait(5 * time.Minute)
	port1Spec, err := port_spec.NewPortSpec(port1Num, port1Protocol, appProtocol1, wait1)
	require.NoError(t, err, "An unexpected error occurred creating port 1 spec")

	port2Id := "port2"
	port2Num := uint16(45)
	port2Protocol := port_spec.TransportProtocol_TCP
	appProtocol2 := "app-protocol2"
	wait2 := port_spec.NewWait(24 * time.Second)
	port2Spec, err := port_spec.NewPortSpec(port2Num, port2Protocol, appProtocol2, wait2)
	require.NoError(t, err, "An unexpected error occurred creating port 2 spec")

	input := map[string]*port_spec.PortSpec{
		port1Id: port1Spec,
		port2Id: port2Spec,
	}

	return input
}

func testPublicPorts(t *testing.T) map[string]*port_spec.PortSpec {

	port1Id := "port1"
	port1Num := uint16(23)
	port1Protocol := port_spec.TransportProtocol_TCP
	appProtocol1 := "app-protocol1-public"
	wait1 := port_spec.NewWait(5 * time.Minute)
	port1Spec, err := port_spec.NewPortSpec(port1Num, port1Protocol, appProtocol1, wait1)
	require.NoError(t, err, "An unexpected error occurred creating port 1 spec")

	port2Id := "port2"
	port2Num := uint16(45)
	port2Protocol := port_spec.TransportProtocol_TCP
	appProtocol2 := "app-protocol2-public"
	wait2 := port_spec.NewWait(24 * time.Second)
	port2Spec, err := port_spec.NewPortSpec(port2Num, port2Protocol, appProtocol2, wait2)
	require.NoError(t, err, "An unexpected error occurred creating port 2 spec")

	input := map[string]*port_spec.PortSpec{
		port1Id: port1Spec,
		port2Id: port2Spec,
	}

	return input
}

func testEnvVars() map[string]string {
	return map[string]string{
		"HTTP_PORT":  "80",
		"HTTPS_PORT": "443",
	}
}

func testImageBuildSpec() *image_build_spec.ImageBuildSpec {
	return image_build_spec.NewImageBuildSpec(
		"test-image",
		"path",
		"")
}
