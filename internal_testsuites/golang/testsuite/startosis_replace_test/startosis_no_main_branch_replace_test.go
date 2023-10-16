package startosis_replace_test

import (
	"context"
	"github.com/stretchr/testify/require"
)

const (
	packageWithNoMainBranchReplaceRelPath = "../../../starlark/packages-with-replace/replace-with-no-main-branch"
	packageWithNoMainBranchReplaceParams  = `{ "message_origin" : "branch" }`
)

func (suite *StartosisReplaceTestSuite) TestStartosisNoMainBranchReplace() {
	ctx := context.Background()
	runResult, err := suite.RunPackageWithParams(ctx, packageWithNoMainBranchReplaceRelPath, packageWithNoMainBranchReplaceParams)

	t := suite.T()
	require.NoError(t, err)
	require.Nil(t, runResult.InterpretationError)
	require.Empty(t, runResult.ValidationErrors)
	require.Nil(t, runResult.ExecutionError)
	expectedResult := "Replace with no main branch package loaded.\nVerification succeeded. Value is '\"dependency-loaded-from-test-branch\"'.\n"
	require.Equal(t, expectedResult, string(runResult.RunOutput))

}
