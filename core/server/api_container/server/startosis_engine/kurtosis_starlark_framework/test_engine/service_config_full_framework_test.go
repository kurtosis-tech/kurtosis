package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	"github.com/stretchr/testify/require"
	"testing"
)

type serviceConfigFullTestCase struct {
	*testing.T
}

func newServiceConfigFullTestCase(t *testing.T) *serviceConfigFullTestCase {
	return &serviceConfigFullTestCase{
		T: t,
	}
}

func (t *serviceConfigFullTestCase) GetId() string {
	return service_config.ServiceConfigTypeName
}

func (t *serviceConfigFullTestCase) GetStarlarkCode() string {
	starlarkCode := fmt.Sprintf("%s(%s=%q, %s=%s, %s=%s, %s=%s, %s=%s, %s=%s, %s=%s, %s=%q, %s=%q, %s=%d, %s=%d, %s=%s)",
		service_config.ServiceConfigTypeName,
		service_config.ImageAttr, TestContainerImageName,
		service_config.PortsAttr, fmt.Sprintf("{%q: PortSpec(number=%d, transport_protocol=%q, application_protocol=%q)}", TestPrivatePortId, TestPrivatePortNumber, TestPrivatePortProtocolStr, TestPrivateApplicationProtocol),
		service_config.PublicPortsAttr, fmt.Sprintf("{%q: PortSpec(number=%d, transport_protocol=%q, application_protocol=%q)}", TestPublicPortId, TestPublicPortNumber, TestPublicPortProtocolStr, TestPublicApplicationProtocol),
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

func (t *serviceConfigFullTestCase) Assert(typeValue builtin_argument.KurtosisValueType) {
	serviceConfigStarlark, ok := typeValue.(*service_config.ServiceConfig)
	require.True(t, ok)

	serviceConfig, err := serviceConfigStarlark.ToKurtosisType()
	require.Nil(t, err)

	expectedServiceConfig := services.NewServiceConfigBuilder(
		TestContainerImageName,
	).WithPrivatePorts(map[string]*kurtosis_core_rpc_api_bindings.Port{
		TestPrivatePortId: binding_constructors.NewPort(TestPrivatePortNumber, TestPrivatePortProtocol, TestPrivateApplicationProtocol, TestWaitConfiguration),
	}).WithPublicPorts(map[string]*kurtosis_core_rpc_api_bindings.Port{
		TestPublicPortId: binding_constructors.NewPort(TestPublicPortNumber, TestPublicPortProtocol, TestPrivateApplicationProtocol, TestWaitConfiguration),
	}).WithFilesArtifactMountDirpaths(map[string]string{
		TestFilesArtifactPath1: TestFilesArtifactName1,
		TestFilesArtifactPath2: TestFilesArtifactName2,
	}).WithEntryPointArgs(
		TestEntryPointSlice,
	).WithCmdArgs(
		TestCmdSlice,
	).WithEnvVars(map[string]string{
		TestEnvVarName1: TestEnvVarValue1,
		TestEnvVarName2: TestEnvVarValue2,
	}).WithPrivateIPAddressPlaceholder(
		TestPrivateIPAddressPlaceholder,
	).WithSubnetwork(
		string(TestSubnetwork),
	).WithCpuAllocationMillicpus(
		TestCpuAllocation,
	).WithMemoryAllocationMegabytes(
		TestMemoryAllocation,
	)
	require.Equal(t, expectedServiceConfig.Build(), serviceConfig)
}
