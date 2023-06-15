package startosis_subpackage_test

import (
	"context"
	"github.com/stretchr/testify/require"
)

const (
	relPathToKurtosisSubpackage = "../../../starlark/kurtosis-sub-package/subpackage-module"
)

func (suite *StartosisSubpackageTestSuite) TestStarlarkValidLocalSubPackage() {
	ctx := context.Background()
	isRemotePackage := false
	runResult, err := suite.RunPackage(ctx, relPathToKurtosisSubpackage, isRemotePackage)

	t := suite.T()
	require.NoError(t, err, "Unexpected error executing starlark package")

	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.Empty(t, runResult.ExecutionError, "Unexpected execution error")
	require.Equal(t, "Successfully ran kurtosis package from a subfolder\n", string(runResult.RunOutput))
}
