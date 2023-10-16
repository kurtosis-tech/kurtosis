package startosis_replace_test

import (
	"context"
	"github.com/stretchr/testify/require"
)

const (
	packageWithRegularReplaceRelPath = "../../../starlark/packages-with-replace/regular-replace"
	packageWithRegularReplaceParams  = `{ "message_origin" : "another-main" }`
)

func (suite *StartosisReplaceTestSuite) TestStartosisRegularReplace() {
	ctx := context.Background()
	runResult, err := suite.RunPackageWithParams(ctx, packageWithRegularReplaceRelPath, packageWithRegularReplaceParams)

	t := suite.T()
	require.NoError(t, err)
	require.NotNil(t, runResult)

	require.Nil(t, runResult.InterpretationError)
	require.Empty(t, runResult.ValidationErrors)
	require.Nil(t, runResult.ExecutionError)
	expectedResult := "Regular replace package loaded.\nVerification succeeded. Value is '\"another-dependency-loaded-from-main\"'.\n"
	require.Equal(t, expectedResult, string(runResult.RunOutput))

}
