package duplicate_files_artifact_mount_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

const (
	testName              = "duplicate-files-artifact-mount"
	isPartitioningEnabled = false

	image                        = "flashspys/nginx-static"
	serviceId services.ServiceID = "file-server"

	testFilesArtifactUrl = "https://kurtosis-public-access.s3.us-east-1.amazonaws.com/test-artifacts/static-fileserver-files.tgz"

	userServiceMountPointForTestFilesArtifact = "/static"

	duplicateMountpointKubernetesErrMsgSentence   = "Invalid value: \"/static\": must be unique"
	duplicateMountpointDockerDaemonErrMsgSentence = "Duplicate mount point"
)

func TestStoreWebFiles(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, stopEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer stopEnclaveFunc()

	// ------------------------------------- TEST SETUP ----------------------------------------------
	firstFilesArtifactUuid, err := enclaveCtx.StoreWebFiles(context.Background(), testFilesArtifactUrl)
	require.NoError(t, err, "An error occurred storing the first files artifact")
	secondFilesArtifactUuid, err := enclaveCtx.StoreWebFiles(context.Background(), testFilesArtifactUrl)
	require.NoError(t, err, "An error occurred storing the second files artifact")

	filesArtifactMountpoints := map[services.FilesArtifactUUID]string{
		firstFilesArtifactUuid:  userServiceMountPointForTestFilesArtifact,
		secondFilesArtifactUuid: userServiceMountPointForTestFilesArtifact,
	}
	fileServerContainerConfig := getFileServerContainerConfig(filesArtifactMountpoints)

	// ------------------------------------- TEST RUN ----------------------------------------------
	_, err = enclaveCtx.AddService(serviceId, fileServerContainerConfig)
	require.Errorf(
		t,
		err,
		"Adding service '%v' should have failed and did not, because duplicated files artifact mountpoints "+
			"'%v' should throw an error",
		serviceId,
		filesArtifactMountpoints,
	)
	isExpectedErrorMsg := strings.Contains(err.Error(), duplicateMountpointKubernetesErrMsgSentence) || strings.Contains(err.Error(), duplicateMountpointDockerDaemonErrMsgSentence)

	require.True(t, isExpectedErrorMsg, "Adding service '%v' has failed, but the error is not the duplicated-files-artifact-mountpoints error "+
		"that we expected, this is throwing this error instead:\n%v",
		serviceId,
		err.Error(),
	)
}

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================
func getFileServerContainerConfig(filesArtifactMountpoints map[services.FilesArtifactUUID]string) *services.ContainerConfig {
	containerConfig := services.NewContainerConfigBuilder(
		image,
	).WithFiles(
		filesArtifactMountpoints,
	).Build()
	return containerConfig
}
