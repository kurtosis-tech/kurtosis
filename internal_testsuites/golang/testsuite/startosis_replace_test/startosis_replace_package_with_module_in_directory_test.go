package startosis_replace_test

import (
	"context"
	"github.com/stretchr/testify/require"
)

const (
	packageWithReplaceModuleInDirectoryRelPath = "../../../starlark/packages-with-replace/replace-with-module-in-directory"
	packageWithReplaceModuleInDirectoryParams  = `{ "message_origin" : "another-sample" }`
)

func (suite *StartosisReplaceTestSuite) TestStartosisReplaceWithModuleInDirectory() {
	ctx := context.Background()
	runResult, _ := suite.RunPackageWithParams(ctx, packageWithReplaceModuleInDirectoryRelPath, packageWithReplaceModuleInDirectoryParams)

	t := suite.T()
	require.Nil(t, runResult.InterpretationError)
	require.Empty(t, runResult.ValidationErrors)
	require.Nil(t, runResult.ExecutionError)
	expectedResult := "Replace with module in directory sample package loaded.\nVerification succeeded. Value is '\"another-dependency-loaded-from-internal-module-in-main-branch\"'.\n"
	require.Regexp(t, expectedResult, string(runResult.RunOutput))

}
