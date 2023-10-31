package startosis_block_local_absolute_locators

import (
	"context"
	"github.com/stretchr/testify/require"
)

const (
	packageWithChildPackageImportingParentPackageFileRelPath = "../../../starlark/local-absolute-locators/locator-in-child-package-points-to-parent-package/parent-package"
)

func (suite *StartosisBlockLocalAbsoluteLocatorsTestSuite) TestStartosisDontBlockChildImportingParentFile() {
	ctx := context.Background()
	runResult, err := suite.RunPackage(ctx, packageWithChildPackageImportingParentPackageFileRelPath)

	t := suite.T()
	require.NoError(t, err)
	require.NotNil(t, runResult)

	require.Nil(t, runResult.InterpretationError)
	require.Empty(t, runResult.ValidationErrors)
	require.Nil(t, runResult.ExecutionError)
	require.Equal(t, expectedRunOutput, string(runResult.RunOutput))
}
