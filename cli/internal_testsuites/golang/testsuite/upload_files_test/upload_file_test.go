package upload_files_test

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
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	archiveDirectoryTestPattern    = "upload-test-golang-"
	archiveSubDirectoryTestPattern = "sub-folder-"
	archiveFileTestPattern         = "test-file-"
	archiveTestFileContent         = "This file is for testing purposes."

	numberOfTempTestFilesToCreateInSubDir     = 3
	numberOfTempTestFilesToCreateInArchiveDir = 1

	enclaveTestName       = "upload-files-test"
	isPartitioningEnabled = true

	fileServerServiceImage                      = "flashspys/nginx-static"
	fileServerServiceId      services.ServiceID = "file-server"
	fileServerPortId                            = "http"
	fileServerPrivatePortNum                    = 80

	waitForStartupTimeBetweenPolls = 500
	waitForStartupMaxRetries       = 15
	waitInitialDelayMilliseconds   = 0

	// Filenames & contents for the files stored in the files artifact
	diskDirKeyword                            = "diskDir"
	archiveDirKeyword                         = "archiveDir"
	subDirKeyword                             = "subDir"
	subFilePatternKeyword                     = "subFile"
	archiveRootFilePatternKeyword             = "archiveRootFile"
	userServiceMountPointForTestFilesArtifact = "/static"
)

var fileServerPortSpec = services.NewPortSpec(
	fileServerPrivatePortNum,
	services.PortProtocol_TCP,
)

func TestUploadFiles(t *testing.T) {
	ctx := context.Background()
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, enclaveTestName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	filePathsMap, err := createTestFolderToUpload()
	require.NoError(t, err)
	uuid, err := enclaveCtx.UploadFiles(filePathsMap[diskDirKeyword])
	require.NoError(t, err)

	filesArtifactMountPoints := map[services.FilesArtifactID]string{
		uuid: userServiceMountPointForTestFilesArtifact,
	}
	fileServerContainerConfigSupplier := getFileServerContainerConfigSupplier(filesArtifactMountPoints)

	serviceCtx, err := enclaveCtx.AddService(fileServerServiceId, fileServerContainerConfigSupplier)
	require.NoError(t, err, "An error occurred adding the file server service")
	publicPort, found := serviceCtx.GetPublicPorts()[fileServerPortId]
	require.True(t, found, "Expected to find public port for ID '%v', but none was found", fileServerPortId)
	fileServerPublicIp := serviceCtx.GetMaybePublicIPAddress()
	fileServerPublicPortNum := publicPort.GetNumber()

	println(filePathsMap[archiveRootFilePatternKeyword+"0"])
	require.NoError(t,
		enclaveCtx.WaitForHttpGetEndpointAvailability(
			fileServerServiceId,
			fileServerPrivatePortNum,
			filePathsMap[archiveRootFilePatternKeyword+"0"],
			waitInitialDelayMilliseconds,
			waitForStartupMaxRetries,
			waitForStartupTimeBetweenPolls,
			""),
		"An error occurred waiting for the file server service to become available.",
	)
	logrus.Infof("Added file server service with public IP '%v' and port '%v'", fileServerPublicIp,
		fileServerPublicPortNum)

	err = testContents(filePathsMap, serviceCtx, publicPort)
	require.NoError(t, err)
}

//========================================================================
// Helpers
//========================================================================
func testContents(pathMap map[string]string, serviceCtx *services.ServiceContext, publicPort *services.PortSpec) error {
	ipAddress := serviceCtx.GetMaybePublicIPAddress()
	portNum := publicPort.GetNumber()

	//Test files archive root directory.
	if err := testDirectoryContents(pathMap, numberOfTempTestFilesToCreateInArchiveDir, archiveRootFilePatternKeyword, ipAddress,
		portNum); err != nil {
		return stacktrace.Propagate(err, "File contents and or folder names in '%s' could not be verified.",
			pathMap[archiveDirKeyword])
	}

	//Test subdirectory.
	if err := testDirectoryContents(pathMap, numberOfTempTestFilesToCreateInSubDir, subFilePatternKeyword, ipAddress,
		portNum); err != nil {
		return stacktrace.Propagate(err, "File contents and or folder names in '%s' could not be verified.",
			pathMap[subDirKeyword])
	}
	return nil
}

//Check all a directories mapped files and ensure they contain the same content as archiveTestFileContent
func testDirectoryContents(pathMap map[string]string, fileCount int, fileKeywordPattern string,
	ipAddress string, portNum uint16) error {
	for i := 0; i < fileCount; i++ {
		fileKeyword := fmt.Sprintf("%s%v", fileKeywordPattern, i)
		relativePath := pathMap[fileKeyword]

		if err := testFileContents(ipAddress, portNum, relativePath); err != nil {
			return stacktrace.Propagate(err, "There was an error testing the content of file '%s'.", relativePath)
		}
	}
	return nil
}

//Compare the file contents of the directories against archiveTestFileContent and see if they match.
func testFileContents(serverIP string, port uint16, filename string) error {
	fileContents, err := getFileContents(
		serverIP,
		port,
		filename,
	)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting '%s' contents", filename)
	}
	if archiveTestFileContent != fileContents {
		return stacktrace.NewError("The contents of '%s' do not match the test content '%s'", fileContents,
			archiveTestFileContent)
	}
	return nil
}

func createTestFiles(pathToCreateAt string, fileCount int) ([]string, error) {
	filenames := make([]string, 0)

	for i := 0; i < fileCount; i++ {
		tempFile, err := ioutil.TempFile(pathToCreateAt, archiveFileTestPattern)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to create archive test files in '%s'.", pathToCreateAt)
		}
		defer tempFile.Close()

		err = os.Chmod(tempFile.Name(), 0644) //Change permission for nginx access.
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to set file permissions for '%s'.", tempFile.Name())
		}

		_, err = tempFile.Write([]byte(archiveTestFileContent))
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to archive test file '%s' at '%s'.", tempFile.Name(),
				pathToCreateAt)
		}
		filenames = append(filenames, tempFile.Name())
	}

	return filenames, nil
}

//Creates a temporary folder with x files and 1 sub folder that has y files each.
//Where x is numberOfTempTestFilesToCreateInArchiveDir
//Where y is numberOfTempTestFilesToCreateInSubDir
func createTestFolderToUpload() (map[string]string, error) {
	baseTempDirPath, err := ioutil.TempDir("", archiveDirectoryTestPattern)
	println(baseTempDirPath)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create a temporary directory for testing files.")
	}
	if err = os.Chmod(baseTempDirPath, 0755); err != nil {
		return nil, stacktrace.Propagate(err, "Failed to set file permissions for '%s'.", baseTempDirPath)
	}

	//Create a temporary subdirectory.
	tempSubDirectory, err := ioutil.TempDir(baseTempDirPath, archiveSubDirectoryTestPattern)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create a temporary archive directory within '%s'.",
			baseTempDirPath)
	}

	err = os.Chmod(tempSubDirectory, 0755) ////Change permission for nginx access.
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to set file permissions for '%s'.", tempSubDirectory)
	}

	subFilenames, err := createTestFiles(tempSubDirectory, numberOfTempTestFilesToCreateInSubDir)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create archive test files at '%s'.",
			tempSubDirectory)
	}

	archiveRootFilenames, err := createTestFiles(baseTempDirPath, numberOfTempTestFilesToCreateInArchiveDir)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create archive test files in the root of your temporary "+
			"archive directory at '%s'.", baseTempDirPath)
	}

	allPaths := map[string]string{}
	allPaths[diskDirKeyword] = baseTempDirPath //The full disk path before getting relative endpoints.
	allPaths[archiveDirKeyword] = filepath.Base(baseTempDirPath)
	clientTempDirToStrip := strings.Replace(baseTempDirPath, allPaths[archiveDirKeyword], "", 1)

	tempSubDirectory = strings.Replace(tempSubDirectory, clientTempDirToStrip, "", 1)
	allPaths[subDirKeyword] = tempSubDirectory

	for i := 0; i < len(subFilenames); i++ {
		keyword := fmt.Sprintf("%s%v", subFilePatternKeyword, i)
		relativePath := strings.Replace(subFilenames[i], clientTempDirToStrip, "", 1)
		allPaths[keyword] = relativePath
	}
	for i := 0; i < len(archiveRootFilenames); i++ {
		keyword := fmt.Sprintf("%s%v", archiveRootFilePatternKeyword, i)
		relativePath := strings.Replace(archiveRootFilenames[i], clientTempDirToStrip, "", 1)
		allPaths[keyword] = relativePath
	}
	return allPaths, nil
}

func getFileServerContainerConfigSupplier(filesArtifactMountPoints map[services.FilesArtifactID]string) func(ipAddr string) (*services.ContainerConfig, error) {
	containerConfigSupplier := func(ipAddr string) (*services.ContainerConfig, error) {

		containerConfig := services.NewContainerConfigBuilder(
			fileServerServiceImage,
		).WithUsedPorts(map[string]*services.PortSpec{
			fileServerPortId: fileServerPortSpec,
		}).WithFiles(
			filesArtifactMountPoints,
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
		return "", stacktrace.Propagate(err,
			"An error occurred reading the response body when getting the contents of file '%v'", filename)
	}

	bodyStr := string(bodyBytes)
	return bodyStr, nil
}
