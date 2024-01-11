package startosis_directory_test

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

const (
	multipleArtifactsDirectoryPackage = "../../../starlark/service-directory/multiple-artifacts-directory"
)

func (suite *StartosisDirectoryTestSuite) TestAddServiceWithMultipleArtifactsDirectory() {
	ctx := context.Background()
	runResult, err := suite.RunPackage(ctx, multipleArtifactsDirectoryPackage)

	t := suite.T()
	logrus.Infof("Test Output: %v", runResult.RunOutput)
	require.NoError(t, err, "Unexpected error executing starlark package")
	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error.")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.Nil(t, runResult.ExecutionError, "Unexpected execution error")

	// This checks the output of the `cat both files` at the end.
	require.Contains(t, runResult.RunOutput, "hello")
	require.Contains(t, runResult.RunOutput, "world")
}
