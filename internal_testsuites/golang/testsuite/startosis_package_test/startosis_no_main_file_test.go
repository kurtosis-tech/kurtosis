package startosis_package_test

import (
	"context"
	"github.com/stretchr/testify/require"
)

const (
	packageWithNoMainStarRelPath = "../../../starlark/no-main-star"
)

func (suite *StartosisPackageTestSuite) TestStartosisPackage_NoMainFile() {
	ctx := context.Background()
	runResult, err := suite.RunPackage(ctx, packageWithNoMainStarRelPath)

	t := suite.T()

	require.NoError(t, err)

	expectedErrorContents := `An error occurred while verifying that 'main.star' exists in the package 'github.com/kurtosis-tech/kurtosis/internal_testsuites/starlark/no-main-star' at '/kurtosis-data/startosis-packages/kurtosis-tech/kurtosis/internal_testsuites/starlark/no-main-star/main.star'
	Caused by: stat /kurtosis-data/startosis-packages/kurtosis-tech/kurtosis/internal_testsuites/starlark/no-main-star/main.star: no such file or directory`
	require.NotNil(t, runResult.InterpretationError)
	require.Equal(t, expectedErrorContents, runResult.InterpretationError.GetErrorMessage())
	require.Empty(t, runResult.ValidationErrors)
	require.Nil(t, runResult.ExecutionError)
	require.Empty(t, string(runResult.RunOutput))
}
