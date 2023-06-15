package startosis_package_test

import (
	"context"
	"github.com/stretchr/testify/require"
)

const (
	packageWithInvalidKurtosisYamlRelPath = "../../../starlark/invalid-yaml-file"
)

func (suite *StartosisPackageTestSuite) TestStartosisPackage_InvalidYamlFile() {
	ctx := context.Background()
	_, err := suite.RunPackage(ctx, packageWithInvalidKurtosisYamlRelPath)

	t := suite.T()
	expectedErrorContents := "Field 'name', which is the Starlark package's name, in kurtosis.yml needs to be set and cannot be empty"
	require.NotNil(t, err, "Unexpected error executing Starlark package")
	require.Contains(t, err.Error(), expectedErrorContents)
}
