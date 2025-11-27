package files_artifact_mounting_test

import (
	"context"
	"fmt"
	"github.com/dzobbe/PoTE-kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/dzobbe/PoTE-kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"testing"
)

const (
	testName = "files-artifact-mounting"

	fileServerServiceName services.ServiceName = "file-server"

	testFilesArtifactUrl = "https://kurtosis-public-access.s3.us-east-1.amazonaws.com/test-artifacts/static-fileserver-files.tgz"
	testArtifactName     = "test-artifact-name"

	// Filenames & contents for the files stored in the files artifact
	file1Filename = "file1.txt"
	file2Filename = "file2.txt"

	expectedFile1Contents = "file1\n"
	expectedFile2Contents = "file2\n"
)

// TODO: remove or re-enable this test in future. This functionality seems to be deprecating soon.
func TestStoreWebFiles(t *testing.T) {
	t.Skip()
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, _, destroyEnclaveFunc, err := test_helpers.CreateEnclave(t, ctx, testName)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer func() {
		err = destroyEnclaveFunc()
		require.NoError(t, err, "An error occurred destroying the enclave after the test finished")
	}()
	// ------------------------------------- TEST SETUP ----------------------------------------------
	filesArtifactUuid, err := enclaveCtx.StoreWebFiles(context.Background(), testFilesArtifactUrl, testArtifactName)
	require.NoError(t, err, "An error occurred storing the files artifact")
	fileServerPublicIp, fileServerPublicPortNum, err := test_helpers.StartFileServer(ctx, fileServerServiceName, filesArtifactUuid, file1Filename, enclaveCtx)
	require.NoError(t, err, "An error occurred waiting for the file server service to become available")
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
//
//	Private helper functions
//
// ====================================================================================================
// nolint
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

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred reading the response body when getting the contents of file '%v'", filename)
	}

	bodyStr := string(bodyBytes)
	return bodyStr, nil
}
