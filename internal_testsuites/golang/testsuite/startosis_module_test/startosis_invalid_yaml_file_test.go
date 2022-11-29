package startosis_module_test

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
	invalidCaseYamlFileTestName          = "invalid-module-invalid-yaml-file"
	moduleWithInvalidKurtosisYamlRelPath = "../../../startosis/invalid-yaml-file"
)

func TestStartosisModule_InvalidYamlFile(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, invalidCaseYamlFileTestName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	currentWorkingDirectory, err := os.Getwd()
	require.Nil(t, err)
	moduleDirpath := path.Join(currentWorkingDirectory, moduleWithInvalidKurtosisYamlRelPath)

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Info("Executing Startosis Module...")

	logrus.Infof("Startosis module path: \n%v", moduleDirpath)

	expectedErrorContents := "Field 'name', which is the Starlark package's name, in kurtosis.yml needs to be set and cannot be empty"
	_, _, err = enclaveCtx.RunStarlarkPackage(ctx, moduleDirpath, emptyExecuteParams, defaultDryRun)
	require.NotNil(t, err, "Unexpected error executing startosis module")
	require.Contains(t, err.Error(), expectedErrorContents)
}
