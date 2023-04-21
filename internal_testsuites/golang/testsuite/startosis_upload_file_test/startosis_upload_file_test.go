package startosis_upload_files_test

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"testing"
)

const (
	defaultDryRun         = false
	isPartitioningEnabled = false
	defaultParallelism    = 1

	validPackageWithInputTestName = "upload-file-package"
	validPackageWithInputRelPath  = "../../../starlark/upload-file-package"

	fileName = "large-file.bin"
	fileSize = 25 * 1024 * 1024 // 25MB
)

func TestStartosisPackage_ValidPackageWithInput(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, _, destroyEnclaveFunc, err := test_helpers.CreateEnclave(t, ctx, validPackageWithInputTestName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer func() { _ = destroyEnclaveFunc() }()

	currentWorkingDirectory, err := os.Getwd()
	require.Nil(t, err)
	packageDirpath := path.Join(currentWorkingDirectory, validPackageWithInputRelPath)

	// ------------------------------------- TEST RUN ----------------------------------------------
	// we generate a file and place it into the package - as it is a big file, we don't want to check it in GitHub
	fileRelPath := path.Join(validPackageWithInputRelPath, fileName)
	randomFilePath, deleteFile, err := test_helpers.GenerateRandomTempFile(fileSize, fileRelPath)
	require.NoError(t, err)
	defer deleteFile()

	// we also compute the hexadecimal file hash to be able to compare it to the one that will
	// be computed inside the container with a call to `md5sum`
	content, err := os.ReadFile(randomFilePath)
	require.NoError(t, err)
	hash := md5.New()
	_, err = hash.Write(content)
	require.NoError(t, err)
	randomFileHexHash := hex.EncodeToString(hash.Sum(nil))

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Info("Executing Startosis Package...")

	// Note: the result extracted from the recipe inside Starlark contains a newline char at the end.
	// We need to add it here manually to have matching hashes
	params := fmt.Sprintf(`{"file_hash": "%s\n"}`, randomFileHexHash)
	runResult, err := enclaveCtx.RunStarlarkPackageBlocking(ctx, packageDirpath, params, defaultDryRun, defaultParallelism)
	require.NoError(t, err, "Unexpected error executing Starlark package")

	// the package itself runs the assertion here. If the file hash computed withing the enclave with md5sum differs
	// from the one passed as parameter to the package (computed in Go above), an Execution Error will be returned
	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.Nil(t, runResult.ExecutionError, "Unexpected execution error")

	logrus.Info("Successfully ran Startosis module")
}
