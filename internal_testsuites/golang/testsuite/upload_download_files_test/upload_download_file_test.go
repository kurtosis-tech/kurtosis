package upload_download_files_test

import (
	"context"
	"crypto/md5"
	"fmt"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis/api/golang/v2/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/v2/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/path-compression"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	archiveDirectoryTestPattern    = "upload-test-golang-"
	archiveSubDirectoryTestPattern = "sub-folder-"
	archiveFileTestPattern         = "test-file-"
	archiveTestFileContent         = "This file is for testing purposes.\n"

	numberOfTempTestFilesToCreateInSubDir     = 3
	numberOfTempTestFilesToCreateInArchiveDir = 1

	enclaveTestName = "upload-files-test"

	// Filenames & contents for the files stored in the files artifact
	diskDirKeyword                = "diskDir"
	archiveDirKeyword             = "archiveDir"
	subDirKeyword                 = "subDir"
	subFileKeywordPattern         = "subFile"
	archiveRootFileKeywordPattern = "archiveRootFile"

	folderPermission = 0755
	filePermission   = 0644

	fileServerServiceName services.ServiceName = "file-server"

	testArtifactName = "test-artifact"

	largeFileSize            = 25 * 1024 * 1024
	emptyFilePathForTempFile = ""
	enforceFileSizeLimit     = true
)

func TestUploadAndDownloadFiles(t *testing.T) {
	filePathsMap, err := createTestFolderToUpload()
	require.NoError(t, err)

	ctx := context.Background()
	enclaveCtx, _, destroyEnclaveFunc, err := test_helpers.CreateEnclave(t, ctx, enclaveTestName)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer func() {
		err = destroyEnclaveFunc()
		require.NoError(t, err, "An error occurred destroying the enclave after the test finished")
	}()

	pathToUpload := filePathsMap[diskDirKeyword]
	require.NotEmptyf(t, pathToUpload, "Failed to store uploadable path in path map.")
	artifactUuid, artifactName, err := enclaveCtx.UploadFiles(filePathsMap[diskDirKeyword], testArtifactName)
	require.NoError(t, err)
	require.Equal(t, string(artifactName), testArtifactName)

	firstArchiveRootKeyword := fmt.Sprintf("%s%v", archiveRootFileKeywordPattern, 0)
	firstArchiveRootFilename := filePathsMap[firstArchiveRootKeyword]

	fileServerPublicIp, fileServerPublicPortNum, err := test_helpers.StartFileServer(ctx, fileServerServiceName, testArtifactName, firstArchiveRootFilename, enclaveCtx)
	require.NoError(t, err)

	err = testAllContents(filePathsMap, fileServerPublicIp, fileServerPublicPortNum)
	require.NoError(t, err)

	archiveBytesViaUuid, err := enclaveCtx.DownloadFilesArtifact(ctx, string(artifactUuid))
	require.Nil(t, err)
	archiveBytesViaShortenedUuid, err := enclaveCtx.DownloadFilesArtifact(ctx, string(artifactUuid)[:12])
	require.Nil(t, err)
	archiveBytesViaName, err := enclaveCtx.DownloadFilesArtifact(ctx, testArtifactName)
	require.Nil(t, err)

	require.Equal(t, archiveBytesViaShortenedUuid, archiveBytesViaUuid)
	require.Equal(t, archiveBytesViaName, archiveBytesViaUuid)

	inspectedFiles, err := enclaveCtx.InspectFilesArtifact(ctx, testArtifactName)
	require.Nil(t, err)
	containsText := slices.ContainsFunc(inspectedFiles.GetFileDescriptions(), func(description *kurtosis_core_rpc_api_bindings.FileArtifactContentsFileDescription) bool {
		return description.GetTextPreview() == archiveTestFileContent &&
			strings.Contains(description.GetPath(), archiveFileTestPattern) &&
			strings.Contains(description.GetPath(), archiveFileTestPattern)
	})
	require.True(t, containsText)
}

func TestUploadAndDownloadLargeFilesCheckingConsistency(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, _, destroyEnclaveFunc, err := test_helpers.CreateEnclave(t, ctx, enclaveTestName)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer func() { _ = destroyEnclaveFunc() }()

	// generate a 10MB random file
	randomFilePath, deleteFile, err := test_helpers.GenerateRandomTempFile(largeFileSize, emptyFilePathForTempFile)
	require.NoError(t, err)
	defer deleteFile()

	// Upload the file to an artifact
	fileArtifactUuid, fileArtifactName, err := enclaveCtx.UploadFiles(randomFilePath, testArtifactName)
	require.NoError(t, err)
	require.Equal(t, testArtifactName, string(fileArtifactName))

	// Download the artifact - Note that the artifact comes as a .tgz here
	downloadedFileContent, err := enclaveCtx.DownloadFilesArtifact(context.Background(), string(fileArtifactUuid))
	require.NoError(t, err)

	// Compute the hash of the initial file, compressed in the same way artifacts are compressed
	initialFileCompressed, _, _, err := path_compression.CompressPath(randomFilePath, enforceFileSizeLimit)
	require.NoError(t, err)
	md5Hash := md5.New()
	_, err = io.Copy(md5Hash, initialFileCompressed)
	require.NoError(t, err)
	initialFileHash := md5Hash.Sum(nil)

	// Compute the hash of the downloaded file
	md5Hash = md5.New()
	_, err = md5Hash.Write(downloadedFileContent)
	require.NoError(t, err)
	downloadedFileMd5 := md5Hash.Sum(nil)

	// Compare the two hashes = they must match
	require.Equal(t, initialFileHash, downloadedFileMd5)
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
			return stacktrace.NewError("The file for keyword '%s' was not mapped in the paths map.", fileKeyword)
		}
		if err := test_helpers.CheckFileContents(ipAddress, portNum, relativePath, archiveTestFileContent); err != nil {
			return stacktrace.Propagate(err, "There was an error testing the content of file '%s'.", relativePath)
		}
	}
	return nil
}

func createTestFiles(pathToCreateAt string, fileCount int) ([]string, error) {
	filenames := []string{}
	tempFiles := []*os.File{}
	defer func() {
		for _, tempFile := range tempFiles {
			if err := tempFile.Close(); err != nil {
				logrus.Errorf("An error occurred closing the temporary file '%v'", tempFile.Name())
			}
		}
	}()
	for i := 0; i < fileCount; i++ {
		tempFile, err := os.CreateTemp(pathToCreateAt, archiveFileTestPattern)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to create archive test files in '%s'.", pathToCreateAt)
		}
		tempFiles = append(tempFiles, tempFile)

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
	baseTempDirPath, err := os.MkdirTemp("", archiveDirectoryTestPattern)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create a temporary directory for testing files.")
	}

	//Create a single subdirectory.
	tempSubDirectory, err := os.MkdirTemp(baseTempDirPath, archiveSubDirectoryTestPattern)
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
