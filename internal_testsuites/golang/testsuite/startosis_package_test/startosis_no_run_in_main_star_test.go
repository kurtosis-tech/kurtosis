package startosis_package_test

import (
	"context"
	"github.com/stretchr/testify/require"
)

const (
	packageWithNoMainInMainStarRelPath = "../../../starlark/no-run-in-main-star"
)

func (suite *StartosisPackageTestSuite) TestStartosisPackage_NoMainInMainStar() {
	ctx := context.Background()
	runResult, _ := suite.RunPackage(ctx, packageWithNoMainInMainStarRelPath)

	t := suite.T()
	expectedInterpretationErr := "No 'run' function found in the main file of package 'github.com/sample/sample-kurtosis-package'; a 'run' entrypoint function with the signature `run(plan, args)` or `run()` is required in the main file of the Kurtosis package"
	require.NotNil(t, runResult.InterpretationError)
	require.Contains(t, runResult.InterpretationError.GetErrorMessage(), expectedInterpretationErr)
	require.Empty(t, runResult.ValidationErrors)
	require.Nil(t, runResult.ExecutionError)
	require.Empty(t, string(runResult.RunOutput))
}
