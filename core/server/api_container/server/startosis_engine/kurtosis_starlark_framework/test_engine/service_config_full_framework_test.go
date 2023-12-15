package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service_directory"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/directory"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_constants"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages/mock_package_content_provider"
	"github.com/stretchr/testify/require"
	"net"
	"testing"
	"time"
)

type serviceConfigFullTestCase struct {
	*testing.T
	serviceNetwork         *service_network.MockServiceNetwork
	packageContentProvider *mock_package_content_provider.MockPackageContentProvider
}

func (suite *KurtosisTypeConstructorTestSuite) TestServiceConfigFull() {
	suite.serviceNetwork.EXPECT().GetApiContainerInfo().Times(1).Return(
		service_network.NewApiContainerInfo(net.IPv4(0, 0, 0, 0), 0, "0.0.0"),
	)

	suite.run(&serviceConfigFullTestCase{
		T:                      suite.T(),
		serviceNetwork:         suite.serviceNetwork,
		packageContentProvider: suite.packageContentProvider,
	})
}

func (t *serviceConfigFullTestCase) GetStarlarkCode() string {
	fileArtifact1 := fmt.Sprintf("%s(%s=%q)", directory.DirectoryTypeName, directory.ArtifactNameAttr, testFilesArtifactName1)
	fileArtifact2 := fmt.Sprintf("%s(%s=%q)", directory.DirectoryTypeName, directory.ArtifactNameAttr, testFilesArtifactName2)
	persistentDirectory := fmt.Sprintf("%s(%s=%q)", directory.DirectoryTypeName, directory.PersistentKeyAttr, testPersistentDirectoryKey)
	starlarkCode := fmt.Sprintf("%s(%s=%q, %s=%s, %s=%s, %s=%s, %s=%s, %s=%s, %s=%s, %s=%q, %s=%d, %s=%d, %s=%d, %s=%d, %s=%s, %s=%v)",
		service_config.ServiceConfigTypeName,
		service_config.ImageAttr, testContainerImageName,
		service_config.PortsAttr, fmt.Sprintf("{%q: PortSpec(number=%d, transport_protocol=%q, application_protocol=%q, wait=%q)}", testPrivatePortId, testPrivatePortNumber, testPrivatePortProtocolStr, testPrivateApplicationProtocol, testWaitConfiguration),
		service_config.PublicPortsAttr, fmt.Sprintf("{%q: PortSpec(number=%d, transport_protocol=%q, application_protocol=%q, wait=%q)}", testPublicPortId, testPublicPortNumber, testPublicPortProtocolStr, testPublicApplicationProtocol, testWaitConfiguration),
		service_config.FilesAttr, fmt.Sprintf("{%q: %s, %q: %s, %q: %s}", testFilesArtifactPath1, fileArtifact1, testFilesArtifactPath2, fileArtifact2, testPersistentDirectoryPath, persistentDirectory),
		service_config.EntrypointAttr, fmt.Sprintf("[%q, %q]", testEntryPointSlice[0], testEntryPointSlice[1]),
		service_config.CmdAttr, fmt.Sprintf("[%q, %q, %q]", testCmdSlice[0], testCmdSlice[1], testCmdSlice[2]),
		service_config.EnvVarsAttr, fmt.Sprintf("{%q: %q, %q: %q}", testEnvVarName1, testEnvVarValue1, testEnvVarName2, testEnvVarValue2),
		service_config.PrivateIpAddressPlaceholderAttr, testPrivateIPAddressPlaceholder,
		service_config.MaxCpuMilliCoresAttr, testCpuAllocation,
		service_config.MinCpuMilliCoresAttr, testMinCpuMilliCores,
		service_config.MaxMemoryMegaBytesAttr, testMemoryAllocation,
		service_config.MinMemoryMegaBytesAttr, testMinMemoryMegabytes,
		service_config.ReadyConditionsAttr,
		getDefaultReadyConditionsScriptPart(),
		service_config.LabelsAttr, fmt.Sprintf("{%q: %q, %q: %q}", testServiceConfigLabelsKey1, testServiceConfigLabelsValue1, testServiceConfigLabelsKey2, testServiceConfigLabelsValue2))
	return starlarkCode
}

func (t *serviceConfigFullTestCase) Assert(typeValue builtin_argument.KurtosisValueType) {
	serviceConfigStarlark, ok := typeValue.(*service_config.ServiceConfig)
	require.True(t, ok)

	serviceConfig, err := serviceConfigStarlark.ToKurtosisType(
		t.serviceNetwork,
		"",
		"",
		t.packageContentProvider,
		map[string]string{})
	require.Nil(t, err)

	require.Equal(t, testContainerImageName, serviceConfig.GetContainerImageName())

	waitDuration, errParseDuration := time.ParseDuration(testWaitConfiguration)
	require.NoError(t, errParseDuration)

	privatePortSpec, errPrivatePortSpec := port_spec.NewPortSpec(testPrivatePortNumber, testPrivatePortProtocol, testPrivateApplicationProtocol, port_spec.NewWait(waitDuration))
	require.NoError(t, errPrivatePortSpec)
	expectedPrivatePorts := map[string]*port_spec.PortSpec{
		testPrivatePortId: privatePortSpec,
	}
	require.Equal(t, expectedPrivatePorts, serviceConfig.GetPrivatePorts())

	portSpec, errPublicPortSpec := port_spec.NewPortSpec(testPublicPortNumber, testPublicPortProtocol, testPrivateApplicationProtocol, port_spec.NewWait(waitDuration))
	require.NoError(t, errPublicPortSpec)
	expectedPublicPorts := map[string]*port_spec.PortSpec{
		testPublicPortId: portSpec,
	}
	require.Equal(t, expectedPublicPorts, serviceConfig.GetPublicPorts())

	expectedFilesArtifactMap := map[string]string{
		testFilesArtifactPath1: testFilesArtifactName1,
		testFilesArtifactPath2: testFilesArtifactName2,
	}
	require.NotNil(t, serviceConfig.GetFilesArtifactsExpansion())
	require.Equal(t, expectedFilesArtifactMap, serviceConfig.GetFilesArtifactsExpansion().ServiceDirpathsToArtifactIdentifiers)

	expectedPersistentDirectoryMap := map[string]service_directory.PersistentDirectory{
		testPersistentDirectoryPath: {
			PersistentKey: service_directory.DirectoryPersistentKey(testPersistentDirectoryKey),
			Size:          service_directory.DirectoryPersistentSize(startosis_constants.DefaultPersistentDirectorySize),
		},
	}
	require.NotNil(t, serviceConfig.GetPersistentDirectories())
	require.Equal(t, expectedPersistentDirectoryMap, serviceConfig.GetPersistentDirectories().ServiceDirpathToPersistentDirectory)

	require.Equal(t, testEntryPointSlice, serviceConfig.GetEntrypointArgs())
	require.Equal(t, testCmdSlice, serviceConfig.GetCmdArgs())

	expectedEnvVars := map[string]string{
		testEnvVarName1: testEnvVarValue1,
		testEnvVarName2: testEnvVarValue2,
	}
	require.Equal(t, expectedEnvVars, serviceConfig.GetEnvVars())

	require.Equal(t, testPrivateIPAddressPlaceholder, serviceConfig.GetPrivateIPAddrPlaceholder())
	require.Equal(t, testMemoryAllocation, serviceConfig.GetMemoryAllocationMegabytes())
	require.Equal(t, testCpuAllocation, serviceConfig.GetCPUAllocationMillicpus())
	require.Equal(t, testMinMemoryMegabytes, serviceConfig.GetMinMemoryAllocationMegabytes())
	require.Equal(t, testMinCpuMilliCores, serviceConfig.GetMinCPUAllocationMillicpus())

	require.Equal(t, testServiceConfigLabels, serviceConfig.GetLabels())
}
