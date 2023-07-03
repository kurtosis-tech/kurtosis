package startosis_package_test

import (
	"context"
	"github.com/stretchr/testify/require"
)

const (
	validPackageNoTypeRelPath = "../../../starlark/valid-kurtosis-package-no-input"
)

func (suite *StartosisPackageTestSuite) TestStartosisPackage_ValidPackageNoInput() {
	ctx := context.Background()
	runResult, err := suite.RunPackage(ctx, validPackageNoTypeRelPath)

	t := suite.T()
	require.Nil(t, err, "Unexpected error executing Starlark package")

	require.Nil(t, runResult.InterpretationError)
	require.Empty(t, runResult.ValidationErrors)
	require.Nil(t, runResult.ExecutionError)

	expectedScriptOutput := `package with no input
{
	"message": "package with no input"
}
`
	require.Equal(t, expectedScriptOutput, string(runResult.RunOutput))
	require.Len(t, runResult.Instructions, 1)
}
