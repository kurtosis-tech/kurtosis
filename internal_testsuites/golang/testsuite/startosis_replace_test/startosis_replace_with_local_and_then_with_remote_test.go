package startosis_replace_test

import (
	"context"
	"github.com/stretchr/testify/require"
)

const (
	packageWithoutReplaceRelPath = "../../../starlark/packages-with-replace/without-replace"
	packageWithoutReplaceParams  = `{ "message_origin" : "main" }`
)

func (suite *StartosisReplaceTestSuite) TestStartosisReplaceWithLocalAndThenWithRemote() {
	ctx := context.Background()
	runResult, err := suite.RunPackageWithParams(ctx, packageWithLocalReplaceRelPath, packageWithLocalReplaceParams)

	t := suite.T()
	require.NoError(t, err)
	require.Nil(t, runResult.InterpretationError)
	require.Empty(t, runResult.ValidationErrors)
	require.Nil(t, runResult.ExecutionError)
	expectedResult := "Replace with local package loaded.\nVerification succeeded. Value is '\"msg-loaded-from-local-dependency\"'.\n"
	require.Equal(t, expectedResult, string(runResult.RunOutput))

	runResult2, err := suite.RunPackageWithParams(ctx, packageWithoutReplaceRelPath, packageWithoutReplaceParams)
	require.NoError(t, err)
	require.Nil(t, runResult.InterpretationError)
	require.Empty(t, runResult.ValidationErrors)
	require.Nil(t, runResult.ExecutionError)
	expectedResult2 := "Without replace package loaded.\nVerification succeeded. Value is '\"dependency-loaded-from-main\"'.\n"
	require.Equal(t, expectedResult2, string(runResult2.RunOutput))
}
