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
	isPartitioningEnabled = false

	fileServerServiceImage                      = "flashspys/nginx-static"
	fileServerServiceId      services.ServiceID = "file-server"
	fileServerPortId                            = "http"
	fileServerPrivatePortNum                    = 80

	waitForStartupTimeBetweenPolls = 500
	waitForStartupMaxRetries       = 15
	waitInitialDelayMilliseconds   = 0
	waitForAvailabilityBodyText    = ""

	// Filenames & contents for the files stored in the files artifact
	diskDirKeyword                            = "diskDir"
	archiveDirKeyword                         = "archiveDir"
	subDirKeyword                             = "subDir"
	subFileKeywordPattern                     = "subFile"
	archiveRootFileKeywordPattern             = "archiveRootFile"
	userServiceMountPointForTestFilesArtifact = "/static"

	folderPermission = 0755
	filePermission   = 0644
)

var fileServerPortSpec = services.NewPortSpec(
	fileServerPrivatePortNum,
	services.PortProtocol_TCP,
)

func TestUploadFiles(t *testing.T) {
	filePathsMap, err := createTestFolderToUpload()
	require.NoError(t, err)

	ctx := context.Background()
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, enclaveTestName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	pathToUpload := filePathsMap[diskDirKeyword]
	if pathToUpload == "" {
		stacktrace.NewError("Failed to store uploadable path in path map.")
	}
	filesArtifactUUID, err := enclaveCtx.UploadFiles(filePathsMap[diskDirKeyword])
	require.NoError(t, err)

	filesArtifactMountPoints := map[services.FilesArtifactUUID]string{
		filesArtifactUUID: userServiceMountPointForTestFilesArtifact,
	}
	fileServerContainerConfigSupplier := getFileServerContainerConfigSupplier(filesArtifactMountPoints)
	serviceCtx, err := enclaveCtx.AddService(fileServerServiceId, fileServerContainerConfigSupplier)
	require.NoError(t, err, "An error occurred adding the file server service")

	publicPort, found := serviceCtx.GetPublicPorts()[fileServerPortId]
	require.True(t, found, "Expected to find public port for ID '%v', but none was found", fileServerPortId)

	fileServerPublicIp := serviceCtx.GetMaybePublicIPAddress()
	fileServerPublicPortNum := publicPort.GetNumber()
	firstArchiveRootKeyword := fmt.Sprintf("%s%v", archiveRootFileKeywordPattern, 0)
	firstArchiveRootFilename := filePathsMap[firstArchiveRootKeyword]

	require.NoError(t,
		enclaveCtx.WaitForHttpGetEndpointAvailability(
			fileServerServiceId,
			fileServerPrivatePortNum,
			firstArchiveRootFilename,
			waitInitialDelayMilliseconds,
			waitForStartupMaxRetries,
			waitForStartupTimeBetweenPolls,
			waitForAvailabilityBodyText,
		),
		"An error occurred waiting for the file server service to become available.",
	)
	logrus.Infof("Added file server service with public IP '%v' and port '%v'", fileServerPublicIp,
		fileServerPublicPortNum)

	err = testAllContents(filePathsMap, fileServerPublicIp, fileServerPublicPortNum)
	require.NoError(t, err)
}

// ========================================================================
// Helpers
// ========================================================================
func testAllContents(pathMap map[string]string, ipAddress string, portNum uint16) error {
	//Test files in archive root directory.
	if err := testDirectoryContents(
		pathMap,
		numberOfTempTestFilesToCreateInArchiveDir,
		archiveRootFileKeywordPattern,
		ipAddress,
		portNum,
	); err != nil {
		return stacktrace.Propagate(err, "File contents and or folder names in '%s' could not be verified.",
			pathMap[archiveDirKeyword])
	}

	//Test files in subdirectory.
	if err := testDirectoryContents(
		pathMap,
		numberOfTempTestFilesToCreateInSubDir,
		subFileKeywordPattern,
		ipAddress,
		portNum,
	); err != nil {
		return stacktrace.Propagate(err, "File contents and or folder names in '%s' could not be verified.",
			pathMap[subDirKeyword])
	}
	return nil
}

// Check all a directories mapped files and ensure they contain the same content as archiveTestFileContent
func testDirectoryContents(
	pathsMap map[string]string,
	fileCount int,
	fileKeywordPattern string,
	ipAddress string,
	portNum uint16,
) error {

	for i := 0; i < fileCount; i++ {
		fileKeyword := fmt.Sprintf("%s%v", fileKeywordPattern, i)
		relativePath := pathsMap[fileKeyword]
		if relativePath == "" {
			stacktrace.NewError("The file for keyword '%s' was not mapped in the paths map.", fileKeyword)
		}
		if err := testFileContents(ipAddress, portNum, relativePath); err != nil {
			return stacktrace.Propagate(err, "There was an error testing the content of file '%s'.", relativePath)
		}
	}
	return nil
}

// Compare the file contents of the directories against archiveTestFileContent and see if they match.
func testFileContents(serverIP string, port uint16, relativeFilepath string) error {
	fileContents, err := getFileContents(serverIP, port, relativeFilepath)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting '%s' contents", relativeFilepath)
	}
	if archiveTestFileContent != fileContents {
		return stacktrace.NewError(
			"The contents of '%s' do not match the test content '%s'",
			fileContents,
			archiveTestFileContent,
		)
	}
	return nil
}

func createTestFiles(pathToCreateAt string, fileCount int) ([]string, error) {
	filenames := []string{}

	for i := 0; i < fileCount; i++ {
		tempFile, err := ioutil.TempFile(pathToCreateAt, archiveFileTestPattern)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to create archive test files in '%s'.", pathToCreateAt)
		}
		defer tempFile.Close()

		err = os.Chmod(tempFile.Name(), filePermission) //Change permission for nginx access.
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to set file permissions for '%s'.", tempFile.Name())
		}

		_, err = tempFile.Write([]byte(archiveTestFileContent))
		if err != nil {
			return nil, stacktrace.Propagate(
				err,
				"Failed to write test file '%s' at '%s'.",
				tempFile.Name(),
				pathToCreateAt,
			)
		}
		filenames = append(filenames, tempFile.Name())
	}
	return filenames, nil
}

// Creates a temporary folder with x files and 1 sub folder that has y files each.
// Where x is numberOfTempTestFilesToCreateInArchiveDir
// Where y is numberOfTempTestFilesToCreateInSubDir
func createTestFolderToUpload() (map[string]string, error) {
	//Create base directory.
	baseTempDirPath, err := ioutil.TempDir("", archiveDirectoryTestPattern)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create a temporary directory for testing files.")
	}

	//Create a single subdirectory.
	tempSubDirectory, err := ioutil.TempDir(baseTempDirPath, archiveSubDirectoryTestPattern)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create a temporary archive directory within '%s'.",
			baseTempDirPath)
	}

	//Create NUMBER_OF_TEMP_FILES_IN_SUBDIRECTORY
	subDirFilenames, err := createTestFiles(tempSubDirectory, numberOfTempTestFilesToCreateInSubDir)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create archive test files at '%s'.",
			tempSubDirectory)
	}

	//Create NUMBER_OF_TEMP_FILES_IN_ROOT_DIRECTORY
	archiveRootFilenames, err := createTestFiles(baseTempDirPath, numberOfTempTestFilesToCreateInArchiveDir)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create archive test files in the root of your temporary "+
			"archive directory at '%s'.", baseTempDirPath)
	}

	//Set folder permissions.
	if err = os.Chmod(baseTempDirPath, folderPermission); err != nil {
		return nil, stacktrace.Propagate(err, "Failed to set file permissions for '%s'.", baseTempDirPath)
	}
	err = os.Chmod(tempSubDirectory, folderPermission) //Change permission for nginx access.
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to set file permissions for '%s'.", tempSubDirectory)
	}

	archiveRootDir := filepath.Base(baseTempDirPath)
	subDir := filepath.Base(tempSubDirectory)

	relativeDiskPaths := map[string]string{}
	relativeDiskPaths[diskDirKeyword] = baseTempDirPath //The full disk path before getting relative endpoints.
	relativeDiskPaths[archiveDirKeyword] = archiveRootDir
	testDirectory := strings.Replace(baseTempDirPath, relativeDiskPaths[archiveDirKeyword], "", 1)

	tempSubDirectory = strings.Replace(tempSubDirectory, testDirectory, "", 1)
	relativeDiskPaths[subDirKeyword] = tempSubDirectory

	for i := 0; i < len(subDirFilenames); i++ {
		keyword := fmt.Sprintf("%s%v", subFileKeywordPattern, i)
		basename := filepath.Base(subDirFilenames[i])
		relativePath := filepath.Join(subDir, basename)
		relativeDiskPaths[keyword] = relativePath
	}
	for i := 0; i < len(archiveRootFilenames); i++ {
		keyword := fmt.Sprintf("%s%v", archiveRootFileKeywordPattern, i)
		basename := filepath.Base(archiveRootFilenames[i])
		relativeDiskPaths[keyword] = basename
	}
	return relativeDiskPaths, nil
}

func getFileServerContainerConfigSupplier(filesArtifactMountPoints map[services.FilesArtifactUUID]string) func(ipAddr string) (*services.ContainerConfig, error) {
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

func getFileContents(ipAddress string, portNum uint16, realtiveFilepath string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("http://%v:%v/%v", ipAddress, portNum, realtiveFilepath))
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting the contents of file '%v'", realtiveFilepath)
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
			"An error occurred reading the response body when getting the contents of file '%v'", realtiveFilepath)
	}

	bodyStr := string(bodyBytes)
	return bodyStr, nil
}
