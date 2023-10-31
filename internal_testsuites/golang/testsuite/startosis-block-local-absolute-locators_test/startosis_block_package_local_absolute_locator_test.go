package startosis_block_local_absolute_locators

import (
	"context"
	"github.com/stretchr/testify/require"
)

const (
	packageWithLocalAbsoluteLocatorRelPath = "../../../starlark/local-absolute-locators/block-package-local-absolute-locator"
)

func (suite *StartosisBlockLocalAbsoluteLocatorsTestSuite) TestStartosisBlockPackageLocalAbsoluteLocator() {
	ctx := context.Background()
	runResult, err := suite.RunPackage(ctx, packageWithLocalAbsoluteLocatorRelPath)

	t := suite.T()
	require.ErrorContains(t, err, localAbsoluteLocatorNotAllowedMsg)
	require.NotNil(t, runResult)

	require.NotNil(t, runResult.InterpretationError)
	require.Contains(t, runResult.InterpretationError.GetErrorMessage(), localAbsoluteLocatorNotAllowedMsg)
	require.Empty(t, runResult.ValidationErrors)
	require.Nil(t, runResult.ExecutionError)
}
