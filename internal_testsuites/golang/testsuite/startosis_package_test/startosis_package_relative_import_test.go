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
	packageWithRelativeImportTestName = "package-with-relative-import"
	packageWithRelativeImport         = "../../../starlark/valid-package-with-relative-imports"
)

func TestStartosisPackage_RelativeImports(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, packageWithRelativeImportTestName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	currentWorkingDirectory, err := os.Getwd()
	require.Nil(t, err)
	packageDirpath := path.Join(currentWorkingDirectory, packageWithRelativeImport)

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Info("Executing Starlark Package...")

	logrus.Infof("Starlark package path: \n%v", packageDirpath)

	runResult, _ := enclaveCtx.RunStarlarkPackageBlocking(ctx, packageDirpath, useDefaultMainFile, useDefaultFunctionName, emptyRunParams, defaultDryRun, defaultParallelism, noExperimentalFeature)

	require.Nil(t, runResult.InterpretationError)
	require.Empty(t, runResult.ValidationErrors)
	require.Nil(t, runResult.ExecutionError)
	expectedResult := "Files with artifact name 'upload' uploaded with artifact UUID '[a-f0-9]{32}'\nJohn Doe\nOpen Sesame\n"
	require.Regexp(t, expectedResult, string(runResult.RunOutput))
}
