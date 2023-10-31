package startosis_block_local_absolute_locators

import (
	"context"
	"github.com/stretchr/testify/require"
)

const (
	packageImportingFileFromAnotherPackageInSameRepositoryRelPath = "../../../starlark/local-absolute-locators/subpackages-in-same-repository/package1"
)

func (suite *StartosisBlockLocalAbsoluteLocatorsTestSuite) TestStartosisDontBlockPackageInSameRepository() {
	ctx := context.Background()
	runResult, err := suite.RunPackage(ctx, packageImportingFileFromAnotherPackageInSameRepositoryRelPath)

	t := suite.T()
	require.NoError(t, err)
	require.NotNil(t, runResult)

	require.Nil(t, runResult.InterpretationError)
	require.Empty(t, runResult.ValidationErrors)
	require.Nil(t, runResult.ExecutionError)
	require.Equal(t, expectedRunOutput, string(runResult.RunOutput))
}
