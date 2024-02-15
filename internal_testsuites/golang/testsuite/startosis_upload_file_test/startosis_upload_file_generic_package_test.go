package startosis_upload_file_test

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"os"
	"path"
)

const (
	packageWithUploadFileGenericRelPath = "../../../starlark/upload-file/upload-file-generic-package"
)

func (suite *StartosisUploadFileTestSuite) TestStartosisPackage_UploadFileGenericPackage() {
	ctx := context.Background()
	t := suite.T()
	// ------------------------------------- ENGINE SETUP ----------------------------------------------

	// ------------------------------------- TEST RUN ----------------------------------------------
	currentWorkingDirectory, err := os.Getwd()
	require.Nil(t, err)
	packageDirpath := path.Join(currentWorkingDirectory, packageWithUploadFileGenericRelPath)

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Info("Executing Startosis Package...")

	// Note: the result extracted from the recipe inside Starlark contains a newline char at the end.
	// We need to add it here manually to have matching hashes
	isRemotePackage := false
	runResult, err := suite.RunPackage(ctx, packageDirpath, nil, isRemotePackage)
	require.NoError(t, err, "Unexpected error executing Starlark package")

	// the package itself runs the assertion here. If the file hash computed withing the enclave with md5sum differs
	// from the one passed as parameter to the package (computed in Go above), an Execution Error will be returned
	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.Nil(t, runResult.ExecutionError, "Unexpected execution error")

	expectedScriptOutputSubstring := `uploaded with artifact UUID`
	require.Contains(t, string(runResult.RunOutput), expectedScriptOutputSubstring)

	logrus.Info("Successfully ran Startosis module")
}
