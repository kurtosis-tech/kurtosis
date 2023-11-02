package test_engine

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/commons/enclave_data_directory"
)

// nolint: gomnd
var (
	testEnclaveUuid = enclave.EnclaveUUID("test-enclave-uuid")

	testServiceName  = service.ServiceName("test-service-name")
	testServiceUuid  = service.ServiceUUID("test-service-uuid")
	testServiceName2 = service.ServiceName("test-service-name-2")
	testServiceUuid2 = service.ServiceUUID("test-service-uuid-2")

	testArtifactName = "artifact-name"
	testArtifactUuid = enclave_data_directory.FilesArtifactUUID("file-artifact-uuid")

	testSrcPath = "/path/to/file.txt"

	testModulePackageId       = "github.com/kurtosistech/test-package"
	testModuleMainFileLocator = "github.com/kurtosistech/test-package/main.star"
	testModuleFileName        = "github.com/kurtosistech/test-package/helpers.star"
	testModuleRelativeLocator = "./helpers.star"

	testContainerImageName = "kurtosistech/example-datastore-server"

	testPrivatePortId              = "grpc"
	testPrivatePortNumber          = uint16(1323)
	testPrivatePortProtocolStr     = "TCP"
	testPrivatePortProtocol        = port_spec.TransportProtocol_TCP
	testPrivateApplicationProtocol = "https"
	testWaitConfiguration          = "2s"
	testWaitDefaultValue           = "2m"

	testPublicPortId              = "endpoints"
	testPublicPortNumber          = uint16(80)
	testPublicPortProtocolStr     = "TCP"
	testPublicPortProtocol        = port_spec.TransportProtocol_TCP
	testPublicApplicationProtocol = "https"

	testFilesArtifactPath1      = "path/to/file/1"
	testFilesArtifactName1      = "file_1"
	testFilesArtifactPath2      = "path/to/file/2"
	testFilesArtifactName2      = "file_2"
	testPersistentDirectoryPath = "path/to/persistent/dir"
	testPersistentDirectoryKey  = "persistent-dir-test"

	testEntryPointSlice = []string{
		"127.0.0.0",
		"1234",
	}

	testCmdSlice = []string{
		"bash",
		"-c",
		"/apps/main.py",
	}

	testEnvVarName1  = "VAR_1"
	testEnvVarValue1 = "VALUE_1"
	testEnvVarName2  = "VAR_2"
	testEnvVarValue2 = "VALUE_2"

	testPrivateIPAddressPlaceholder = "<IP_ADDRESS>"

	testCpuAllocation    = uint64(2000)
	testMemoryAllocation = uint64(1024)

	testMinCpuMilliCores   = uint64(1000)
	testMinMemoryMegabytes = uint64(512)

	testReadyConditionsRecipePortId   = "http"
	testReadyConditionsRecipeEndpoint = "/endpoint?input=data"
	testReadyConditionsRecipeCommand  = []string{"tool", "arg"}
	testReadyConditionsRecipeExtract  = "{}"
	testReadyConditionsField          = "code"
	testReadyConditionsAssertion      = "=="
	testReadyConditionsTarget         = "200"
	testReadyConditionsInterval       = "1s"
	testReadyConditionsTimeout        = "100ms"

	testReadyConditions2RecipePortId   = "https"
	testReadyConditions2RecipeEndpoint = "/user-access"
	testReadyConditions2RecipeExtract  = "{}"
	testReadyConditions2Field          = "code"
	testReadyConditions2Assertion      = "=="
	testReadyConditions2Target         = "201"
	testReadyConditions2Interval       = "500ms"
	testReadyConditions2Timeout        = "2s"

	testGetRequestMethod = "GET"

	testNoPackageReplaceOptions = map[string]string{}

	testServiceConfigLabelsKey1   = "app-version"
	testServiceConfigLabelsValue1 = "2.4"
	testServiceConfigLabelsKey2   = "environment"
	testServiceConfigLabelsValue2 = "production"

	testServiceConfigLabels = map[string]string{
		testServiceConfigLabelsKey1: testServiceConfigLabelsValue1,
		testServiceConfigLabelsKey2: testServiceConfigLabelsValue2,
	}
)
