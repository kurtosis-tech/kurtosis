package startosis_package_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"testing"
)

const (
	invalidCaseYamlFileTestName           = "invalid-package-invalid-yaml-file"
	packageWithInvalidKurtosisYamlRelPath = "../../../starlark/invalid-yaml-file"
)

func TestStartosisPackage_InvalidYamlFile(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, invalidCaseYamlFileTestName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	currentWorkingDirectory, err := os.Getwd()
	require.Nil(t, err)
	packageDirpath := path.Join(currentWorkingDirectory, packageWithInvalidKurtosisYamlRelPath)

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Info("Executing Starlark Package...")

	logrus.Infof("Starlark package path: \n%v", packageDirpath)

	expectedErrorContents := "Field 'name', which is the Starlark package's name, in kurtosis.yml needs to be set and cannot be empty"
	_, err = enclaveCtx.RunStarlarkPackageBlocking(ctx, packageDirpath, emptyRunParams, defaultDryRun, defaultParallelism)
	require.NotNil(t, err, "Unexpected error executing Starlark package")
	require.Contains(t, err.Error(), expectedErrorContents)
}
