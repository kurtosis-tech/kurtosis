package upload_files_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
)

const archiveRootDirectoryTestPattern = "upload-test-"
const archiveSubDirectoryTestPattern = "sub-folder-"
const archiveFileTestPattern = "test-file-"
const archiveTestFileContent = "This file is for testing purposes."

const numberOfTempTestFilesToCreateInSubDir = 3
const numberOfTempTestFilesToCreateInRootDir = 1

const enclaveTestName = "upload-files-test"
const isPartitioningEnabled = true

func TestUploadFiles(t *testing.T) {
	ctx := context.Background()
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, enclaveTestName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	pathToUpload, err := createTestFolderToUpload()
	require.NoError(t, err)
	uuid, err := enclaveCtx.UploadFiles(pathToUpload)
	require.NoError(t, err)
	println(uuid)
}

//========================================================================
// Helpers
//========================================================================
func createTestFiles(pathToCreateAt string, fileCount int) error {
	for i := 0; i < fileCount; i++ {
		tempFile, err := ioutil.TempFile(pathToCreateAt, archiveFileTestPattern)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to create archive test files in '%s'.", pathToCreateAt)
		}
		defer tempFile.Close()

		_, err = tempFile.Write([]byte(archiveTestFileContent))
		if err != nil {
			return stacktrace.Propagate(err, "Failed to archive test file '%s' at '%s'.", tempFile.Name(), pathToCreateAt)
		}
	}
	return nil
}

//Creates a temporary folder with x files and 1 sub folder that has y files each.
//Where x is numberOfTempTestFilesToCreateInRootDir
//Where y is numberOfTempTestFilesToCreateInSubDir
func createTestFolderToUpload() (string, error) {
	baseTempDirPath, err := ioutil.TempDir("", archiveRootDirectoryTestPattern)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to create a temporary root directory for testing.")
	}

	//Create a temporary subdirectory.
	tempSubDirectory, err := ioutil.TempDir(baseTempDirPath, archiveSubDirectoryTestPattern)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to create a temporary archive directory within '%s'.",
			baseTempDirPath)
	}

	if err = createTestFiles(tempSubDirectory, numberOfTempTestFilesToCreateInSubDir); err != nil {
		return "", stacktrace.Propagate(err, "Failed to create archive test files at '%s'.",
			tempSubDirectory)
	}

	if err := createTestFiles(baseTempDirPath, numberOfTempTestFilesToCreateInRootDir); err != nil {
		return "", stacktrace.Propagate(err, "Failed to create archive test files in your root directory at '%s'.",
			baseTempDirPath)
	}
	return baseTempDirPath, nil
}
