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
	runResult, err := suite.RunPackage(ctx, packageWithRelativeImport)

	t := suite.T()
	require.NoError(t, err)
	require.Nil(t, runResult.InterpretationError)
	require.Empty(t, runResult.ValidationErrors)
	require.Nil(t, runResult.ExecutionError)
	expectedResult := "Files with artifact name 'upload' uploaded with artifact UUID '[a-f0-9]{32}'\nJohn Doe\nOpen Sesame\n"
	require.Regexp(t, expectedResult, string(runResult.RunOutput))
}
