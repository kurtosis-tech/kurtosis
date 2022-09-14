package files_artifact_mounting_test

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis-core/api/golang/lib/services"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"testing"
)

const (
	testName = "files-artifact-mounting"
	isPartitioningEnabled = false

	fileServerServiceImage                          = "flashspys/nginx-static"
	fileServerServiceId          services.ServiceID = "file-server"
	fileServerPortId = "http"
	fileServerPrivatePortNum = 80

	waitForStartupTimeBetweenPolls = 500
	waitForStartupMaxRetries       = 15
	waitInitialDelayMilliseconds   = 0

	testFilesArtifactUrl                          = "https://kurtosis-public-access.s3.us-east-1.amazonaws.com/test-artifacts/static-fileserver-files.tgz"

	// Filenames & contents for the files stored in the files artifact
	file1Filename = "file1.txt"
	file2Filename = "file2.txt"

	expectedFile1Contents = "file1\n"
	expectedFile2Contents = "file2\n"

	userServiceMountPointForTestFilesArtifact = "/static"
)
var fileServerPortSpec = services.NewPortSpec(
	fileServerPrivatePortNum,
	services.PortProtocol_TCP,
)

func TestStoreWebFiles(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, stopEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer stopEnclaveFunc()

	// ------------------------------------- TEST SETUP ----------------------------------------------
	filesArtifactUuid, err := enclaveCtx.StoreWebFiles(context.Background(), testFilesArtifactUrl)
	require.NoError(t, err, "An error occurred storing the files artifact")

	filesArtifactMountpoints := map[services.FilesArtifactUUID]string{
		filesArtifactUuid: userServiceMountPointForTestFilesArtifact,
	}
	fileServerContainerConfigSupplier := getFileServerContainerConfigSupplier(filesArtifactMountpoints)

	serviceCtx, err := enclaveCtx.AddService(fileServerServiceId, fileServerContainerConfigSupplier)
	require.NoError(t, err, "An error occurred adding the file server service")
	publicPort, found := serviceCtx.GetPublicPorts()[fileServerPortId]
	require.True(t, found, "Expected to find public port for ID '%v', but none was found", fileServerPortId)
	fileServerPublicIp := serviceCtx.GetMaybePublicIPAddress()
	fileServerPublicPortNum := publicPort.GetNumber()

	require.NoError(t,
		// TODO It's suuuuuuuuuuper confusing that we have to pass the private port in here!!!! We should just require the user
		//  to pass in the port ID and the API container will translate that to the private port automatically!!!
		enclaveCtx.WaitForHttpGetEndpointAvailability(fileServerServiceId, fileServerPrivatePortNum, file1Filename, waitInitialDelayMilliseconds, waitForStartupMaxRetries, waitForStartupTimeBetweenPolls, ""),
		"An error occurred waiting for the file server service to become available",
	)
	logrus.Infof("Added file server service with public IP '%v' and port '%v'", fileServerPublicIp, fileServerPublicPortNum)

	// ------------------------------------- TEST RUN ----------------------------------------------
	file1Contents, err := getFileContents(
		fileServerPublicIp,
		fileServerPublicPortNum,
		file1Filename,
	)
	require.NoError(t, err, "An error occurred getting file 1's contents")
	require.Equal(
		t,
		expectedFile1Contents,
		file1Contents,
		"Actual file 1 contents '%v' != expected file 1 contents '%v'",
		file1Contents,
		expectedFile1Contents,
	)

	file2Contents, err := getFileContents(
		fileServerPublicIp,
		fileServerPublicPortNum,
		file2Filename,
	)
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
func getFileServerContainerConfigSupplier(filesArtifactMountpoints map[services.FilesArtifactUUID]string) func(ipAddr string) (*services.ContainerConfig, error) {
	containerConfigSupplier  := func(ipAddr string) (*services.ContainerConfig, error) {

		containerConfig := services.NewContainerConfigBuilder(
			fileServerServiceImage,
		).WithUsedPorts(map[string]*services.PortSpec{
			fileServerPortId: fileServerPortSpec,
		}).WithFiles(
			filesArtifactMountpoints,
		).Build()
		return containerConfig, nil
	}
	return containerConfigSupplier
}

func getFileContents(ipAddress string, portNum uint16, filename string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("http://%v:%v/%v", ipAddress, portNum, filename))
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
