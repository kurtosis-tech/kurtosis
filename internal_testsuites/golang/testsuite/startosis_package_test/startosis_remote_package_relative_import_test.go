package startosis_package_test

import (
	"context"
	"github.com/stretchr/testify/require"
)

const (
	noMainBranchWithRelativeImportRemotePackage = "github.com/kurtosis-tech/sample-startosis-load@test-branch"
)

func (suite *StartosisPackageTestSuite) TestStartosisNoMainBranchRemotePackage_RelativeImports() {
	ctx := context.Background()
	runResult, _ := suite.RunRemotePackage(ctx, noMainBranchWithRelativeImportRemotePackage)

	t := suite.T()
	require.NotNil(t, runResult)
	require.Nil(t, runResult.InterpretationError)
	require.Empty(t, runResult.ValidationErrors)
	require.Nil(t, runResult.ExecutionError)

	expectedResult := "Package loaded.\n"
	require.Regexp(t, expectedResult, string(runResult.RunOutput))
}
