package test_engine

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/kurtosis/core/server/commons/enclave_data_directory"
)

var (
	TestEnclaveUuid = enclave.EnclaveUUID("test-enclave-uuid")

	TestServiceName  = service.ServiceName("test-service-name")
	TestServiceUuid  = service.ServiceUUID("test-service-uuid")
	TestServiceName2 = service.ServiceName("test-service-name-2")
	TestServiceUuid2 = service.ServiceUUID("test-service-uuid-2")

	TestArtifactName = "artifact-name"
	TestArtifactUuid = enclave_data_directory.FilesArtifactUUID("file-artifact-uuid")

	TestSrcPath = "/path/to/file.txt"

	TestModuleFileName = "github.com/kurtosistech/test-package/helpers.star"

	TestContainerImageName = "kurtosistech/example-datastore-server"

	TestPrivatePortId              = "grpc"
	TestPrivatePortNumber          = uint32(1323)
	TestPrivatePortProtocolStr     = "TCP"
	TestPrivatePortProtocol        = kurtosis_core_rpc_api_bindings.Port_TCP
	TestPrivateApplicationProtocol = "https"

	TestPublicPortId              = "endpoints"
	TestPublicPortNumber          = uint32(80)
	TestPublicPortProtocolStr     = "TCP"
	TestPublicPortProtocol        = kurtosis_core_rpc_api_bindings.Port_TCP
	TestPublicApplicationProtocol = "https"

	TestFilesArtifactPath1 = "path/to/file/1"
	TestFilesArtifactName1 = "file_1"
	TestFilesArtifactPath2 = "path/to/file/2"
	TestFilesArtifactName2 = "file_2"

	TestEntryPointSlice = []string{
		"127.0.0.0",
		"1234",
	}

	TestCmdSlice = []string{
		"bash",
		"-c",
		"/apps/main.py",
	}

	TestEnvVarName1  = "VAR_1"
	TestEnvVarValue1 = "VALUE_1"
	TestEnvVarName2  = "VAR_2"
	TestEnvVarValue2 = "VALUE_2"

	TestPrivateIPAddressPlaceholder = "<IP_ADDRESS>"

	TestSubnetwork  = service_network_types.PartitionID("test-subnetwork")
	TestSubnetwork2 = service_network_types.PartitionID("test-subnetwork-2")

	TestCpuAllocation = uint64(2000)

	TestMemoryAllocation = uint64(1024)

	TestReadyConditionsRecipePortId   = "http"
	TestReadyConditionsRecipeEndpoint = "/endpoint?input=data"
	TestReadyConditionsRecipeCommand  = []string{"tool", "arg"}
	TestReadyConditionsRecipeExtract  = "{}"
	TestReadyConditionsField          = "code"
	TestReadyConditionsAssertion      = "=="
	TestReadyConditionsTarget         = "200"
	TestReadyConditionsInterval       = "1s"
	TestReadyConditionsTimeout        = "100ms"

	TestReadyConditions2RecipePortId   = "https"
	TestReadyConditions2RecipeEndpoint = "/user-access"
	TestReadyConditions2RecipeExtract  = "{}"
	TestReadyConditions2Field          = "code"
	TestReadyConditions2Assertion      = "=="
	TestReadyConditions2Target         = "201"
	TestReadyConditions2Interval       = "500ms"
	TestReadyConditions2Timeout        = "2s"

	TestGetRequestMethod = "GET"
)
