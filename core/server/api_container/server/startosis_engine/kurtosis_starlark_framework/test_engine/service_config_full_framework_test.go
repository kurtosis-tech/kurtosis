package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service_directory"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/directory"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	"github.com/stretchr/testify/require"
	"net"
	"testing"
	"time"
)

type serviceConfigFullTestCase struct {
	*testing.T
	serviceNetwork *service_network.MockServiceNetwork
}

func newServiceConfigFullTestCase(t *testing.T) *serviceConfigFullTestCase {
	serviceNetwork := service_network.NewMockServiceNetwork(t)
	serviceNetwork.EXPECT().GetApiContainerInfo().Times(1).Return(
		service_network.NewApiContainerInfo(net.IPv4(0, 0, 0, 0), 0, "0.0.0"),
	)

	return &serviceConfigFullTestCase{
		T:              t,
		serviceNetwork: serviceNetwork,
	}
}

func (t *serviceConfigFullTestCase) GetId() string {
	return service_config.ServiceConfigTypeName
}

func (t *serviceConfigFullTestCase) GetStarlarkCode() string {
	fileArtifact1 := fmt.Sprintf("%s(%s=%q)", directory.DirectoryTypeName, directory.ArtifactNameAttr, TestFilesArtifactName1)
	fileArtifact2 := fmt.Sprintf("%s(%s=%q)", directory.DirectoryTypeName, directory.ArtifactNameAttr, TestFilesArtifactName2)
	persistentDirectory := fmt.Sprintf("%s(%s=%q)", directory.DirectoryTypeName, directory.PersistentKeyAttr, TestPersistentDirectoryKey)
	starlarkCode := fmt.Sprintf("%s(%s=%q, %s=%s, %s=%s, %s=%s, %s=%s, %s=%s, %s=%s, %s=%q, %s=%d, %s=%d, %s=%d, %s=%d, %s=%s)",
		service_config.ServiceConfigTypeName,
		service_config.ImageAttr, TestContainerImageName,
		service_config.PortsAttr, fmt.Sprintf("{%q: PortSpec(number=%d, transport_protocol=%q, application_protocol=%q, wait=%q)}", TestPrivatePortId, TestPrivatePortNumber, TestPrivatePortProtocolStr, TestPrivateApplicationProtocol, TestWaitConfiguration),
		service_config.PublicPortsAttr, fmt.Sprintf("{%q: PortSpec(number=%d, transport_protocol=%q, application_protocol=%q, wait=%q)}", TestPublicPortId, TestPublicPortNumber, TestPublicPortProtocolStr, TestPublicApplicationProtocol, TestWaitConfiguration),
		service_config.FilesAttr, fmt.Sprintf("{%q: %s, %q: %s, %q: %s}", TestFilesArtifactPath1, fileArtifact1, TestFilesArtifactPath2, fileArtifact2, TestPersistentDirectoryPath, persistentDirectory),
		service_config.EntrypointAttr, fmt.Sprintf("[%q, %q]", TestEntryPointSlice[0], TestEntryPointSlice[1]),
		service_config.CmdAttr, fmt.Sprintf("[%q, %q, %q]", TestCmdSlice[0], TestCmdSlice[1], TestCmdSlice[2]),
		service_config.EnvVarsAttr, fmt.Sprintf("{%q: %q, %q: %q}", TestEnvVarName1, TestEnvVarValue1, TestEnvVarName2, TestEnvVarValue2),
		service_config.PrivateIpAddressPlaceholderAttr, TestPrivateIPAddressPlaceholder,
		service_config.MaxCpuMilliCoresAttr, TestCpuAllocation,
		service_config.MinCpuMilliCoresAttr, TestMinCpuMilliCores,
		service_config.MaxMemoryMegaBytesAttr, TestMemoryAllocation,
		service_config.MinMemoryMegaBytesAttr, TestMinMemoryMegabytes,
		service_config.ReadyConditionsAttr,
		getDefaultReadyConditionsScriptPart(),
	)
	return starlarkCode
}

func (t *serviceConfigFullTestCase) Assert(typeValue builtin_argument.KurtosisValueType) {
	serviceConfigStarlark, ok := typeValue.(*service_config.ServiceConfig)
	require.True(t, ok)

	serviceConfig, err := serviceConfigStarlark.ToKurtosisType(t.serviceNetwork)
	require.Nil(t, err)

	require.Equal(t, TestContainerImageName, serviceConfig.GetContainerImageName())

	waitDuration, errParseDuration := time.ParseDuration(TestWaitConfiguration)
	require.NoError(t, errParseDuration)

	privatePortSpec, errPrivatePortSpec := port_spec.NewPortSpec(TestPrivatePortNumber, TestPrivatePortProtocol, TestPrivateApplicationProtocol, port_spec.NewWait(waitDuration))
	require.NoError(t, errPrivatePortSpec)
	expectedPrivatePorts := map[string]*port_spec.PortSpec{
		TestPrivatePortId: privatePortSpec,
	}
	require.Equal(t, expectedPrivatePorts, serviceConfig.GetPrivatePorts())

	portSpec, errPublicPortSpec := port_spec.NewPortSpec(TestPublicPortNumber, TestPublicPortProtocol, TestPrivateApplicationProtocol, port_spec.NewWait(waitDuration))
	require.NoError(t, errPublicPortSpec)
	expectedPublicPorts := map[string]*port_spec.PortSpec{
		TestPublicPortId: portSpec,
	}
	require.Equal(t, expectedPublicPorts, serviceConfig.GetPublicPorts())

	expectedFilesArtifactMap := map[string]string{
		TestFilesArtifactPath1: TestFilesArtifactName1,
		TestFilesArtifactPath2: TestFilesArtifactName2,
	}
	require.NotNil(t, serviceConfig.GetFilesArtifactsExpansion())
	require.Equal(t, expectedFilesArtifactMap, serviceConfig.GetFilesArtifactsExpansion().ServiceDirpathsToArtifactIdentifiers)

	expectedPersistentDirectoryMap := map[string]service_directory.DirectoryPersistentKey{
		TestPersistentDirectoryPath: service_directory.DirectoryPersistentKey(TestPersistentDirectoryKey),
	}
	require.NotNil(t, serviceConfig.GetPersistentDirectories())
	require.Equal(t, expectedPersistentDirectoryMap, serviceConfig.GetPersistentDirectories().ServiceDirpathToDirectoryPersistentKey)

	require.Equal(t, TestEntryPointSlice, serviceConfig.GetEntrypointArgs())
	require.Equal(t, TestCmdSlice, serviceConfig.GetCmdArgs())

	expectedEnvVars := map[string]string{
		TestEnvVarName1: TestEnvVarValue1,
		TestEnvVarName2: TestEnvVarValue2,
	}
	require.Equal(t, expectedEnvVars, serviceConfig.GetEnvVars())

	require.Equal(t, TestPrivateIPAddressPlaceholder, serviceConfig.GetPrivateIPAddrPlaceholder())
	require.Equal(t, TestMemoryAllocation, serviceConfig.GetMemoryAllocationMegabytes())
	require.Equal(t, TestCpuAllocation, serviceConfig.GetCPUAllocationMillicpus())
	require.Equal(t, TestMinMemoryMegabytes, serviceConfig.GetMinMemoryAllocationMegabytes())
	require.Equal(t, TestMinCpuMilliCores, serviceConfig.GetMinCPUAllocationMillicpus())
}
