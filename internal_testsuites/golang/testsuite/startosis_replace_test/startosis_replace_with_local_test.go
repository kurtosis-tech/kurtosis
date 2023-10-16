package startosis_replace_test

import (
	"context"
	"github.com/stretchr/testify/require"
)

const (
	packageWithLocalReplaceRelPath = "../../../starlark/packages-with-replace/replace-with-local"
	packageWithLocalReplaceParams  = `{ "message_origin" : "local" }`
)

func (suite *StartosisReplaceTestSuite) TestStartosisReplaceWithLocal() {
	ctx := context.Background()
	runResult, err := suite.RunPackageWithParams(ctx, packageWithLocalReplaceRelPath, packageWithLocalReplaceParams)

	t := suite.T()
	require.NoError(t, err)
	require.Nil(t, runResult.InterpretationError)
	require.Empty(t, runResult.ValidationErrors)
	require.Nil(t, runResult.ExecutionError)
	expectedResult := "Replace with local package loaded.\nVerification succeeded. Value is '\"msg-loaded-from-local-dependency\"'.\n"
	require.Equal(t, expectedResult, string(runResult.RunOutput))
}
