package startosis_package_test

import (
	"context"
	"github.com/stretchr/testify/require"
)

const (
	packageWithRelativeImport = "../../../starlark/valid-package-with-relative-imports"
)

func (suite *StartosisPackageTestSuite) TestStartosisPackage_RelativeImports() {
	ctx := context.Background()
	runResult, _ := suite.RunPackage(ctx, packageWithRelativeImport)

	t := suite.T()
	require.Nil(t, runResult.InterpretationError)
	require.Empty(t, runResult.ValidationErrors)
	require.Nil(t, runResult.ExecutionError)
	expectedResult := "John Doe\nOpen Sesame\n"
	require.Equal(t, expectedResult, string(runResult.RunOutput))
}
