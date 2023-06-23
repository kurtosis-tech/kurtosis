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
	runResult, _ := suite.RunPackage(ctx, packageWithNoMainStarRelPath)

	t := suite.T()
	expectedErrorContents := `An error occurred while verifying that 'main.star' exists in the package 'github.com/sample/sample-kurtosis-package' at '/kurtosis-data/startosis-packages/sample/sample-kurtosis-package/main.star'
	Caused by: stat /kurtosis-data/startosis-packages/sample/sample-kurtosis-package/main.star: no such file or directory`
	require.NotNil(t, runResult.InterpretationError)
	require.Equal(t, runResult.InterpretationError.GetErrorMessage(), expectedErrorContents)
	require.Empty(t, runResult.ValidationErrors)
	require.Nil(t, runResult.ExecutionError)
	require.Empty(t, string(runResult.RunOutput))
}
