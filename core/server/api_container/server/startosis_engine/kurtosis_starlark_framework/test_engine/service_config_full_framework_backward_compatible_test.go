package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_download_mode"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"github.com/stretchr/testify/require"
	"net"
	"testing"
	"time"
)

type serviceConfigFullTestCaseBackwardCompatible struct {
	*testing.T
	serviceNetwork         *service_network.MockServiceNetwork
	packageContentProvider *startosis_packages.MockPackageContentProvider
}

func (suite *KurtosisTypeConstructorTestSuite) TestServiceConfigFullBackwardCompatible() {
	suite.serviceNetwork.EXPECT().GetApiContainerInfo().Times(1).Return(
		service_network.NewApiContainerInfo(net.IPv4(0, 0, 0, 0), 0, "0.0.0"),
	)

	suite.run(&serviceConfigFullTestCaseBackwardCompatible{
		T:                      suite.T(),
		serviceNetwork:         suite.serviceNetwork,
		packageContentProvider: suite.packageContentProvider,
	})
}

func (t *serviceConfigFullTestCaseBackwardCompatible) GetStarlarkCode() string {
	starlarkCode := fmt.Sprintf("%s(%s=%q, %s=%s, %s=%s, %s=%s, %s=%s, %s=%s, %s=%s, %s=%q, %s=%d, %s=%d, %s=%s)",
		service_config.ServiceConfigTypeName,
		service_config.ImageAttr, testContainerImageName,
		service_config.PortsAttr, fmt.Sprintf("{%q: PortSpec(number=%d, transport_protocol=%q, application_protocol=%q, wait=%q)}", testPrivatePortId, testPrivatePortNumber, testPrivatePortProtocolStr, testPrivateApplicationProtocol, testWaitConfiguration),
		service_config.PublicPortsAttr, fmt.Sprintf("{%q: PortSpec(number=%d, transport_protocol=%q, application_protocol=%q, wait=%q)}", testPublicPortId, testPublicPortNumber, testPublicPortProtocolStr, testPublicApplicationProtocol, testWaitConfiguration),
		service_config.FilesAttr, fmt.Sprintf("{%q: %q, %q: %q}", testFilesArtifactPath1, testFilesArtifactName1, testFilesArtifactPath2, testFilesArtifactName2),
		service_config.EntrypointAttr, fmt.Sprintf("[%q, %q]", testEntryPointSlice[0], testEntryPointSlice[1]),
		service_config.CmdAttr, fmt.Sprintf("[%q, %q, %q]", testCmdSlice[0], testCmdSlice[1], testCmdSlice[2]),
		service_config.EnvVarsAttr, fmt.Sprintf("{%q: %q, %q: %q}", testEnvVarName1, testEnvVarValue1, testEnvVarName2, testEnvVarValue2),
		service_config.PrivateIpAddressPlaceholderAttr, testPrivateIPAddressPlaceholder,
		service_config.CpuAllocationAttr, testCpuAllocation,
		service_config.MemoryAllocationAttr, testMemoryAllocation,
		service_config.ReadyConditionsAttr,
		getDefaultReadyConditionsScriptPart(),
	)
	return starlarkCode
}

func (t *serviceConfigFullTestCaseBackwardCompatible) Assert(typeValue builtin_argument.KurtosisValueType) {
	serviceConfigStarlark, ok := typeValue.(*service_config.ServiceConfig)
	require.True(t, ok)

	serviceConfig, err := serviceConfigStarlark.ToKurtosisType(
		t.serviceNetwork,
		testModulePackageId,
		testModuleMainFileLocator,
		t.packageContentProvider,
		map[string]string{},
		image_download_mode.ImageDownloadMode_Missing)
	require.Nil(t, err)

	require.Equal(t, testContainerImageName, serviceConfig.GetContainerImageName())
	require.Nil(t, serviceConfig.GetImageBuildSpec())

	waitDuration, errParseDuration := time.ParseDuration(testWaitConfiguration)
	require.NoError(t, errParseDuration)

	privatePortSpec, errPrivatePortSpec := port_spec.NewPortSpec(testPrivatePortNumber, testPrivatePortProtocol, testPrivateApplicationProtocol, port_spec.NewWait(waitDuration), "")
	require.NoError(t, errPrivatePortSpec)
	expectedPrivatePorts := map[string]*port_spec.PortSpec{
		testPrivatePortId: privatePortSpec,
	}
	require.Equal(t, expectedPrivatePorts, serviceConfig.GetPrivatePorts())

	portSpec, errPublicPortSpec := port_spec.NewPortSpec(testPublicPortNumber, testPublicPortProtocol, testPrivateApplicationProtocol, port_spec.NewWait(waitDuration), "")
	require.NoError(t, errPublicPortSpec)
	expectedPublicPorts := map[string]*port_spec.PortSpec{
		testPublicPortId: portSpec,
	}
	require.Equal(t, expectedPublicPorts, serviceConfig.GetPublicPorts())

	expectedFilesArtifactMap := map[string][]string{
		testFilesArtifactPath1: {testFilesArtifactName1},
		testFilesArtifactPath2: {testFilesArtifactName2},
	}
	require.NotNil(t, serviceConfig.GetFilesArtifactsExpansion())
	require.Equal(t, expectedFilesArtifactMap, serviceConfig.GetFilesArtifactsExpansion().ServiceDirpathsToArtifactIdentifiers)

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
	require.Equal(t, uint64(0), serviceConfig.GetMinMemoryAllocationMegabytes())
	require.Equal(t, uint64(0), serviceConfig.GetMinCPUAllocationMillicpus())
}
