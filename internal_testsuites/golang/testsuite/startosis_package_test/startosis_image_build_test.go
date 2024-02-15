//go:build !kubernetes
// +build !kubernetes

// We don't run this test in Kubernetes because, as of 2023-12-18, image building is not implemented in Kubernetes yet

package startosis_package_test

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

const (
	imageBuildSpecPackageRelPath              = "../../../starlark/image-build-package"
	imageBuildSpecDockerComposePackageRelPath = "../../../starlark/docker-compose-package-img-build"
)

func (suite *StartosisPackageTestSuite) TestStartosisPackage_ImageBuildSpec() {
	ctx := context.Background()
	runResult, err := suite.RunPackage(ctx, imageBuildSpecPackageRelPath)

	t := suite.T()
	require.Nil(t, err, "Unexpected error executing Starlark package")

	require.NotNil(t, runResult)
	require.Nil(t, runResult.InterpretationError)
	require.Empty(t, runResult.ValidationErrors)
	require.Nil(t, runResult.ExecutionError)

	expectedScriptOutputSubstring := `Service 'service' added with service UUID`

	require.Contains(t, string(runResult.RunOutput), expectedScriptOutputSubstring)
	require.Len(t, runResult.Instructions, 1)

	// TODO: Figure out a way to clean image
	logrus.Warnf("THIS TEST GENERATES A SMALL DOCKER IMAGE. IF YOU ARE RUNNING TESTSUITE LOCALLY(NOT IN CI), YOU MUST MANUALLY REMOVE IT!")
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
