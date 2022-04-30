package duplicate_files_artifact_mount_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/services"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testName = "files-artifact-mounting"
	isPartitioningEnabled = false

	image                        = "flashspys/nginx-static"
	serviceId services.ServiceID = "file-server"

	testFilesArtifactUrl                          = "https://kurtosis-public-access.s3.us-east-1.amazonaws.com/test-artifacts/static-fileserver-files.tgz"

	userServiceMountPointForTestFilesArtifact = "/static"

	duplicateMountpointDockerDaemonErrMsgSentence = "Duplicate mount point"
)

func TestStoreWebFiles(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, stopEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer stopEnclaveFunc()

	// ------------------------------------- TEST SETUP ----------------------------------------------
	firstFilesArtifactId, err := enclaveCtx.StoreWebFiles(context.Background(), testFilesArtifactUrl)
	require.NoError(t, err, "An error occurred storing the first files artifact")
	secondFilesArtifactId, err := enclaveCtx.StoreWebFiles(context.Background(), testFilesArtifactUrl)
	require.NoError(t, err, "An error occurred storing the second files artifact")

	// ------------------------------------- TEST RUN ----------------------------------------------
	filesArtifactMountpoints := map[services.FilesArtifactID]string{
		firstFilesArtifactId: userServiceMountPointForTestFilesArtifact,
		secondFilesArtifactId: userServiceMountPointForTestFilesArtifact,
	}
	fileServerContainerConfigSupplier := getFileServerContainerConfigSupplier(filesArtifactMountpoints)

	_, err = enclaveCtx.AddService(serviceId, fileServerContainerConfigSupplier)
	require.Errorf(
		t,
		err,
		"Adding service '%v' should have failed and did not, because duplicated files artifact mountpoints " +
			"'%v' should throw an error",
		serviceId,
		filesArtifactMountpoints,
	)
	require.Contains(
		t,
		err.Error(),
		duplicateMountpointDockerDaemonErrMsgSentence,
		"Adding service '%v' has failed, but the error is not the duplicated-files-artifact-mountpoints error " +
			"that we expected, this is throwing this error instead:\n%v",
		serviceId,
		err.Error(),
	)
}

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================
func getFileServerContainerConfigSupplier(filesArtifactMountpoints map[services.FilesArtifactID]string) func(ipAddr string) (*services.ContainerConfig, error) {
	containerConfigSupplier  := func(ipAddr string) (*services.ContainerConfig, error) {

		containerConfig := services.NewContainerConfigBuilder(
			image,
		).WithFiles(
			filesArtifactMountpoints,
		).Build()
		return containerConfig, nil
	}
	return containerConfigSupplier
}