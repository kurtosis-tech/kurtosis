package startosis_package_test

import (
	"context"
	"github.com/stretchr/testify/require"
)

const (
	simpleDockerComposePackageRelPath         = "../../../starlark/docker-compose-package"
	imageBuildSpecDockerComposePackageRelPath = "../../../starlark/docker-compose-package-img-build"
)

func (suite *StartosisPackageTestSuite) TestStartosisPackage_DockerComposePackage() {
	ctx := context.Background()
	runResult, err := suite.RunPackage(ctx, simpleDockerComposePackageRelPath)

	t := suite.T()
	require.Nil(t, err, "Unexpected error executing Starlark package")

	require.NotNil(t, runResult)
	require.Nil(t, runResult.InterpretationError)
	require.Empty(t, runResult.ValidationErrors)
	require.Nil(t, runResult.ExecutionError)

	expectedScriptOutputSubstring := `Service 'simple' added with service UUID `

	require.Contains(t, string(runResult.RunOutput), expectedScriptOutputSubstring)
	require.Len(t, runResult.Instructions, 4)
}

func (suite *StartosisPackageTestSuite) TestStartosisPackage_DockerComposePackageWithImageBuildSpec() {
	ctx := context.Background()
	runResult, err := suite.RunPackage(ctx, imageBuildSpecDockerComposePackageRelPath)

	t := suite.T()
	require.Nil(t, err, "Unexpected error executing Starlark package")

	require.NotNil(t, runResult)
	require.Nil(t, runResult.InterpretationError)
	require.Empty(t, runResult.ValidationErrors)
	require.Nil(t, runResult.ExecutionError)

	expectedScriptOutputSubstring := `Service 'web' added with service UUID `

	require.Contains(t, string(runResult.RunOutput), expectedScriptOutputSubstring)
	require.Len(t, runResult.Instructions, 2)
}

func (suite *StartosisPackageTestSuite) TestStartosisPackage_DockerComposePackageWithInvalidPathInVolumeErrors() {
	ctx := context.Background()
	runResult, err := suite.RunPackage(ctx, imageBuildSpecDockerComposePackageRelPath)

	t := suite.T()
	require.Nil(t, err, "Unexpected error executing Starlark package")

	require.NotNil(t, runResult)
	require.Nil(t, runResult.InterpretationError)
	require.Empty(t, runResult.ValidationErrors)
	require.Nil(t, runResult.ExecutionError)

	expectedScriptOutputSubstring := `Service 'web' added with service UUID `

	require.Contains(t, string(runResult.RunOutput), expectedScriptOutputSubstring)
	require.Len(t, runResult.Instructions, 2)
}
