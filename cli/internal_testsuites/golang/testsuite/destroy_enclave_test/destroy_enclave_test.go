package destroy_enclave_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/services"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testName = "destroy-enclave"
	isPartitioningEnabled = false

	fileServerServiceImage                          = "flashspys/nginx-static"
	fileServerServiceId          services.ServiceID = "file-server"
	fileServerPortId = "http"
	fileServerPrivatePortNum = 80


	testFilesArtifactId  services.FilesArtifactID = "test-files-artifact"
	testFilesArtifactUrl                          = "https://kurtosis-public-access.s3.us-east-1.amazonaws.com/test-artifacts/static-fileserver-files.tgz"

)
var fileServerPortSpec = services.NewPortSpec(
	fileServerPrivatePortNum,
	services.PortProtocol_TCP,
)

func TestDestroyEnclave(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, stopEnclaveFunc, destroyEnclaveFunc, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	shouldStopEnclaveAtTheEnd := true
	defer func() {
		if shouldStopEnclaveAtTheEnd {
			stopEnclaveFunc()
		}
	}()

	// ------------------------------------- TEST SETUP ----------------------------------------------
	filesArtifacts := map[services.FilesArtifactID]string{
		testFilesArtifactId: testFilesArtifactUrl,
	}
	require.NoError(t, enclaveCtx.RegisterFilesArtifacts(filesArtifacts), "An error occurred registering the files artifacts")


	fileServerContainerConfigSupplier := getFileServerContainerConfigSupplier()
	_, err = enclaveCtx.AddService(fileServerServiceId, fileServerContainerConfigSupplier)
	require.NoError(t, err, "An error occurred adding the file server service")

	err = destroyEnclaveFunc()
	require.NoErrorf(t, err, "An error occurred destroying enclave with ID '%v'", enclaveCtx.GetEnclaveID())
	shouldStopEnclaveAtTheEnd = false
}

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================

func getFileServerContainerConfigSupplier() func(ipAddr string, sharedDirectory *services.SharedPath) (*services.ContainerConfig, error) {
	containerConfigSupplier  := func(ipAddr string, sharedDirectory *services.SharedPath) (*services.ContainerConfig, error) {

		containerConfig := services.NewContainerConfigBuilder(
			fileServerServiceImage,
		).WithUsedPorts(map[string]*services.PortSpec{
			fileServerPortId: fileServerPortSpec,
		}).WithFilesArtifacts(map[services.FilesArtifactID]string{
			testFilesArtifactId: "/static",
		}).Build()
		return containerConfig, nil
	}
	return containerConfigSupplier
}
