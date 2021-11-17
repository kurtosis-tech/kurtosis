package files_artifact_mounting_test

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/services"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"testing"
)

const (
	testName = "files-artifact-mounting-test"
	isPartitioningEnabled = false

	fileServerServiceImage                          = "flashspys/nginx-static"
	fileServerServiceId          services.ServiceID = "file-server"
	fileServerListenPortNum                         = 80
	fileServerListenPortProtocol                    = "tcp"

	waitForStartupTimeBetweenPolls = 500
	waitForStartupMaxRetries       = 15
	waitInitialDelayMilliseconds   = 0

	testFilesArtifactId  services.FilesArtifactID = "test-files-artifact"
	testFilesArtifactUrl                          = "https://kurtosis-public-access.s3.us-east-1.amazonaws.com/test-artifacts/static-fileserver-files.tgz"

	// Filenames & contents for the files stored in the files artifact
	file1Filename = "file1.txt"
	file2Filename = "file2.txt"

	expectedFile1Contents = "file1\n"
	expectedFile2Contents = "file2\n"
)
var fileServerListenPortStr = fmt.Sprintf("%v/%v", fileServerListenPortNum, fileServerListenPortProtocol)

func TestFilesArtifactMounting(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, stopEnclaveFunc, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer stopEnclaveFunc()

	// ------------------------------------- TEST SETUP ----------------------------------------------
	filesArtifacts := map[services.FilesArtifactID]string{
		testFilesArtifactId: testFilesArtifactUrl,
	}
	require.NoError(t, enclaveCtx.RegisterFilesArtifacts(filesArtifacts), "An error occurred registering the files artifacts")

	fileServerContainerConfigSupplier := getFileServerContainerConfigSupplier()
	fileServerServiceContext, hostPortBindings, err := enclaveCtx.AddService(fileServerServiceId, fileServerContainerConfigSupplier)
	require.NoError(t, err, "An error occurred adding the file server service")

	require.NoError(t,
		enclaveCtx.WaitForHttpGetEndpointAvailability(fileServerServiceId, fileServerListenPortNum, file1Filename, waitInitialDelayMilliseconds, waitForStartupMaxRetries, waitForStartupTimeBetweenPolls, ""),
		"An error occurred waiting for the file server service to become available",
	)
	logrus.Infof("Added file server service with host port bindings: %+v", hostPortBindings)

	// ------------------------------------- TEST RUN ----------------------------------------------
	file1Contents, err := getFileContents(fileServerServiceContext.GetIPAddress(), fileServerListenPortNum, file1Filename)
	require.NoError(t, err, "An error occurred getting file 1's contents")
	require.Equal(
		t,
		expectedFile1Contents,
		file1Contents,
		"Actual file 1 contents '%v' != expected file 1 contents '%v'",
		file1Contents,
		expectedFile1Contents,
	)

	file2Contents, err := getFileContents(fileServerServiceContext.GetIPAddress(), fileServerListenPortNum, file2Filename)
	require.NoError(t, err, "An error occurred getting file 2's contents")
	require.Equal(
		t,
		expectedFile2Contents,
		file2Contents,
		"Actual file 2 contents '%v' != expected file 2 contents '%v'",
		file2Contents,
		expectedFile2Contents,
	)
}

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================

func getFileServerContainerConfigSupplier() func(ipAddr string, sharedDirectory *services.SharedPath) (*services.ContainerConfig, error) {
	containerConfigSupplier  := func(ipAddr string, sharedDirectory *services.SharedPath) (*services.ContainerConfig, error) {

		containerConfig := services.NewContainerConfigBuilder(
			fileServerServiceImage,
		).WithUsedPorts(map[string]bool{
			fileServerListenPortStr: true,
		}).WithFilesArtifacts(map[services.FilesArtifactID]string{
			testFilesArtifactId: "/static",
		}).Build()
		return containerConfig, nil
	}
	return containerConfigSupplier
}

func getFileContents(ipAddress string, port uint32, filename string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("http://%v:%v/%v", ipAddress, port, filename))
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting the contents of file '%v'", filename)
	}
	body := resp.Body
	defer func() {
		if err := body.Close(); err != nil {
			logrus.Warnf("We tried to close the response body, but doing so threw an error:\n%v", err)
		}
	}()

	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred reading the response body when getting the contents of file '%v'", filename)
	}

	bodyStr := string(bodyBytes)
	return bodyStr, nil
}

