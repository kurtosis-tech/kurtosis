package test_engine

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
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
	TestPrivatePortNumber          = uint16(1323)
	TestPrivatePortProtocolStr     = "TCP"
	TestPrivatePortProtocol        = port_spec.TransportProtocol_TCP
	TestPrivateApplicationProtocol = "https"
	TestWaitConfiguration          = "2s"
	TestWaitDefaultValue           = "2m"
	TestWaitNotValidEmptyString    = ""

	TestPublicPortId              = "endpoints"
	TestPublicPortNumber          = uint16(80)
	TestPublicPortProtocolStr     = "TCP"
	TestPublicPortProtocol        = port_spec.TransportProtocol_TCP
	TestPublicApplicationProtocol = "https"

	TestFilesArtifactPath1      = "path/to/file/1"
	TestFilesArtifactName1      = "file_1"
	TestFilesArtifactPath2      = "path/to/file/2"
	TestFilesArtifactName2      = "file_2"
	TestPersistentDirectoryPath = "path/to/persistent/dir"
	TestPersistentDirectoryKey  = "persistent-dir-test"

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

	TestCpuAllocation    = uint64(2000)
	TestMemoryAllocation = uint64(1024)

	TestMinCpuMilliCores   = uint64(1000)
	TestMinMemoryMegabytes = uint64(512)

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
