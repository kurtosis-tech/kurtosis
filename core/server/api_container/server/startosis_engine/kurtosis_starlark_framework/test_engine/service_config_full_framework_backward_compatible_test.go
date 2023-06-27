package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	"github.com/stretchr/testify/require"
	"net"
	"testing"
	"time"
)

type serviceConfigFullTestCaseBackwardCompatible struct {
	*testing.T
	serviceNetwork *service_network.MockServiceNetwork
}

func newServiceConfigFullTestCaseBackwardCompatible(t *testing.T) *serviceConfigFullTestCaseBackwardCompatible {
	serviceNetwork := service_network.NewMockServiceNetwork(t)
	serviceNetwork.EXPECT().GetApiContainerInfo().Times(1).Return(
		service_network.NewApiContainerInfo(net.IPv4(0, 0, 0, 0), 0, "0.0.0"),
	)

	return &serviceConfigFullTestCaseBackwardCompatible{
		T:              t,
		serviceNetwork: serviceNetwork,
	}
}

func (t *serviceConfigFullTestCaseBackwardCompatible) GetId() string {
	return service_config.ServiceConfigTypeName
}

func (t *serviceConfigFullTestCaseBackwardCompatible) GetStarlarkCode() string {
	starlarkCode := fmt.Sprintf("%s(%s=%q, %s=%s, %s=%s, %s=%s, %s=%s, %s=%s, %s=%s, %s=%q, %s=%q, %s=%d, %s=%d, %s=%s)",
		service_config.ServiceConfigTypeName,
		service_config.ImageAttr, TestContainerImageName,
		service_config.PortsAttr, fmt.Sprintf("{%q: PortSpec(number=%d, transport_protocol=%q, application_protocol=%q, wait=%q)}", TestPrivatePortId, TestPrivatePortNumber, TestPrivatePortProtocolStr, TestPrivateApplicationProtocol, TestWaitConfiguration),
		service_config.PublicPortsAttr, fmt.Sprintf("{%q: PortSpec(number=%d, transport_protocol=%q, application_protocol=%q, wait=%q)}", TestPublicPortId, TestPublicPortNumber, TestPublicPortProtocolStr, TestPublicApplicationProtocol, TestWaitConfiguration),
		service_config.FilesAttr, fmt.Sprintf("{%q: %q, %q: %q}", TestFilesArtifactPath1, TestFilesArtifactName1, TestFilesArtifactPath2, TestFilesArtifactName2),
		service_config.EntrypointAttr, fmt.Sprintf("[%q, %q]", TestEntryPointSlice[0], TestEntryPointSlice[1]),
		service_config.CmdAttr, fmt.Sprintf("[%q, %q, %q]", TestCmdSlice[0], TestCmdSlice[1], TestCmdSlice[2]),
		service_config.EnvVarsAttr, fmt.Sprintf("{%q: %q, %q: %q}", TestEnvVarName1, TestEnvVarValue1, TestEnvVarName2, TestEnvVarValue2),
		service_config.PrivateIpAddressPlaceholderAttr, TestPrivateIPAddressPlaceholder,
		service_config.SubnetworkAttr, TestSubnetwork,
		service_config.CpuAllocationAttr, TestCpuAllocation,
		service_config.MemoryAllocationAttr, TestMemoryAllocation,
		service_config.ReadyConditionsAttr,
		getDefaultReadyConditionsScriptPart(),
	)
	return starlarkCode
}

func (t *serviceConfigFullTestCaseBackwardCompatible) Assert(typeValue builtin_argument.KurtosisValueType) {
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

	require.Equal(t, TestEntryPointSlice, serviceConfig.GetEntrypointArgs())
	require.Equal(t, TestCmdSlice, serviceConfig.GetCmdArgs())

	expectedEnvVars := map[string]string{
		TestEnvVarName1: TestEnvVarValue1,
		TestEnvVarName2: TestEnvVarValue2,
	}
	require.Equal(t, expectedEnvVars, serviceConfig.GetEnvVars())

	require.Equal(t, TestPrivateIPAddressPlaceholder, serviceConfig.GetPrivateIPAddrPlaceholder())
	require.Equal(t, string(TestSubnetwork), serviceConfig.GetSubnetwork())
	require.Equal(t, TestMemoryAllocation, serviceConfig.GetMemoryAllocationMegabytes())
	require.Equal(t, TestCpuAllocation, serviceConfig.GetCPUAllocationMillicpus())
	require.Equal(t, uint64(0), serviceConfig.GetMinMemoryAllocationMegabytes())
	require.Equal(t, uint64(0), serviceConfig.GetMinCPUAllocationMillicpus())
}
