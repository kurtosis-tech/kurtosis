package add_service

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_constants"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"net"
	"testing"
)

const (
	testServiceName          = "example-datastore-server-2"
	testContainerImageName   = "kurtosistech/example-datastore-server"
	testSubnetwork           = "subnetwork_1"
	testIpAddressPlaceholder = "<IP_ADDRESS>"
	testMemoryAllocation     = 1024
	testCpuAllocation        = 2000
)

var (
	thread = shared_helpers.NewStarlarkThread("test-add-service")
)

func TestAddServiceInstruction_GetCanonicalizedInstruction(t *testing.T) {
	addServiceInstruction := newEmptyAddServiceInstruction(
		nil,
		kurtosis_instruction.NewInstructionPosition(22, 26, "dummyFile"),
		nil,
	)
	addServiceInstruction.starlarkKwargs[serviceNameArgName] = starlark.String(testServiceName)
	addServiceInstruction.starlarkKwargs[serviceConfigArgName] = newTestStarlarkServiceConfig(t)

	expectedOutput := `add_service(config=ServiceConfig(image="kurtosistech/example-datastore-server", ports={"grpc": PortSpec(number=1234, transport_protocol="TCP", application_protocol="http")}, public_ports={"grpc": PortSpec(number=80, transport_protocol="TCP", application_protocol="http")}, files={"path/to/file/1": "file_1", "path/to/file/2": "file_2"}, entrypoint=["127.0.0.0", "1234"], cmd=["bash", "-c", "/apps/main.py", "1234"], env_vars={"VAR_1": "VALUE_1", "VAR_2": "VALUE_2"}, private_ip_address_placeholder="<IP_ADDRESS>", subnetwork="subnetwork_1", cpu_allocation=2000, memory_allocation=1024), service_name="example-datastore-server-2")`
	require.Equal(t, expectedOutput, addServiceInstruction.String())
}

func TestAddServiceInstruction_EntryPointArgsAreReplaced(t *testing.T) {
	ipAddresses := map[service.ServiceName]net.IP{
		"foo_service": net.ParseIP("172.17.3.13"),
	}
	serviceNetwork := service_network.NewMockServiceNetwork(ipAddresses)
	addServiceInstruction := NewAddServiceInstruction(
		serviceNetwork,
		kurtosis_instruction.NewInstructionPosition(22, 26, "dummyFile"),
		"example-datastore-server-2",
		services.NewServiceConfigBuilder(
			testContainerImageName,
		).WithPrivatePorts(
			map[string]*kurtosis_core_rpc_api_bindings.Port{
				"grpc": {
					Number:            1323,
					TransportProtocol: kurtosis_core_rpc_api_bindings.Port_TCP,
				},
			},
		).WithEntryPointArgs(
			[]string{"-- {{kurtosis:foo_service.ip_address}}"},
		).Build(),
		starlark.StringDict{}, // Unused
		nil,
	)

	err := addServiceInstruction.replaceMagicStrings()
	require.Nil(t, err)
	require.Equal(t, "-- 172.17.3.13", addServiceInstruction.serviceConfig.EntrypointArgs[0])
}

func TestAddServiceInstruction_SerializeAndParseAgain(t *testing.T) {
	serviceConfigBuilder := services.NewServiceConfigBuilder(
		testContainerImageName,
	).WithPrivatePorts(map[string]*kurtosis_core_rpc_api_bindings.Port{
		"grpc": binding_constructors.NewPort(1234, kurtosis_core_rpc_api_bindings.Port_TCP, "http"),
	}).WithPublicPorts(map[string]*kurtosis_core_rpc_api_bindings.Port{
		"grpc": binding_constructors.NewPort(80, kurtosis_core_rpc_api_bindings.Port_TCP, "http"),
	}).WithEntryPointArgs([]string{
		"127.0.0.0",
		"1234",
	}).WithCmdArgs([]string{
		"bash",
		"-c",
		"/apps/main.py",
		"1234",
	}).WithEnvVars(map[string]string{
		"VAR_1": "VALUE_1",
		"VAR_2": "VALUE_2",
	}).WithFilesArtifactMountDirpaths(map[string]string{
		"path/to/file/1": "file_1",
		"path/to/file/2": "file_2",
	}).WithPrivateIPAddressPlaceholder(
		testIpAddressPlaceholder,
	).WithSubnetwork(
		"subnetwork_1",
	).WithMemoryAllocationMegabytes(
		testMemoryAllocation,
	).WithCpuAllocationMillicpus(
		testCpuAllocation,
	)
	starlarkArgs := starlark.StringDict{
		serviceNameArgName:   starlark.String(testServiceName),
		serviceConfigArgName: newTestStarlarkServiceConfig(t),
	}
	initialInstruction := NewAddServiceInstruction(
		nil,
		kurtosis_instruction.NewInstructionPosition(1, 12, startosis_constants.PackageIdPlaceholderForStandaloneScript),
		testServiceName,
		serviceConfigBuilder.Build(),
		starlarkArgs,
		nil)

	canonicalizedInstruction := initialInstruction.String()

	var instructions []kurtosis_instruction.KurtosisInstruction
	_, err := starlark.ExecFile(thread, startosis_constants.PackageIdPlaceholderForStandaloneScript, canonicalizedInstruction, starlark.StringDict{
		kurtosis_types.ServiceConfigTypeName: starlark.NewBuiltin(kurtosis_types.ServiceConfigTypeName, kurtosis_types.MakeServiceConfig),
		kurtosis_types.PortSpecTypeName:      starlark.NewBuiltin(kurtosis_types.PortSpecTypeName, kurtosis_types.MakePortSpec),
		AddServiceBuiltinName:                starlark.NewBuiltin(AddServiceBuiltinName, GenerateAddServiceBuiltin(&instructions, nil, nil)),
	})
	require.Nil(t, err)

	require.Len(t, instructions, 1)
	require.Equal(t, initialInstruction, instructions[0])
}

func newTestStarlarkServiceConfig(t *testing.T) *kurtosis_types.ServiceConfig {
	privatePortsDict := starlark.NewDict(1)
	port1 := kurtosis_types.NewPortSpec(1234, kurtosis_core_rpc_api_bindings.Port_TCP, "http")
	require.Nil(t, privatePortsDict.SetKey(starlark.String("grpc"), port1))

	publicPortsDict := starlark.NewDict(1)
	publicPort1 := kurtosis_types.NewPortSpec(80, kurtosis_core_rpc_api_bindings.Port_TCP, "http")
	require.Nil(t, publicPortsDict.SetKey(starlark.String("grpc"), publicPort1))

	fileArtifactMountDirPath := starlark.NewDict(2)
	require.Nil(t, fileArtifactMountDirPath.SetKey(starlark.String("path/to/file/1"), starlark.String("file_1")))
	require.Nil(t, fileArtifactMountDirPath.SetKey(starlark.String("path/to/file/2"), starlark.String("file_2")))

	entrypoint := starlark.NewList([]starlark.Value{starlark.String("127.0.0.0"), starlark.String("1234")})

	cmd := starlark.NewList([]starlark.Value{starlark.String("bash"), starlark.String("-c"), starlark.String("/apps/main.py"), starlark.String("1234")})

	envVar := starlark.NewDict(2)
	require.Nil(t, envVar.SetKey(starlark.String("VAR_1"), starlark.String("VALUE_1")))
	require.Nil(t, envVar.SetKey(starlark.String("VAR_2"), starlark.String("VALUE_2")))

	ipAddressPlaceholder := starlark.String(testIpAddressPlaceholder)
	subnetwork := starlark.String(testSubnetwork)

	cpuAllocation := starlark.MakeUint64(testCpuAllocation)
	memoryAllocation := starlark.MakeUint64(testMemoryAllocation)

	return kurtosis_types.NewServiceConfig(
		testContainerImageName,
		privatePortsDict,
		publicPortsDict,
		fileArtifactMountDirPath,
		entrypoint,
		cmd,
		envVar,
		&ipAddressPlaceholder,
		&subnetwork,
		&cpuAllocation,
		&memoryAllocation,
	)
}
