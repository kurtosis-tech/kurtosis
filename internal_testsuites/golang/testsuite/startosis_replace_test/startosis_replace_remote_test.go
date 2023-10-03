package startosis_replace_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testName = "replace-remote-package-test"

	remotePackage = "github.com/kurtosis-tech/sample-startosis-load/sample-package"

	executeParams          = `{ "message_origin" : "another-main" }`
	defaultParallelism     = 4
	useDefaultMainFile     = ""
	useDefaultFunctionName = ""
	defaultDryRun          = false
)

var (
	noExperimentalFeature = []kurtosis_core_rpc_api_bindings.KurtosisFeatureFlag{}
)

func TestStartosisReplaceRemotePackage(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, _, _, err := test_helpers.CreateEnclave(t, ctx, testName)
	require.NoError(t, err, "An error occurred creating an enclave")
	/*defer func() {
		err = destroyEnclaveFunc()
		require.NoError(t, err, "An error occurred destroying the enclave after the test finished")
	}()*/

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Debugf("Executing Starlark Package: '%v'", remotePackage)

	runResult, err := enclaveCtx.RunStarlarkRemotePackageBlocking(ctx, remotePackage, useDefaultMainFile, useDefaultFunctionName, executeParams, defaultDryRun, defaultParallelism, noExperimentalFeature)
	require.NoError(t, err, "Unexpected error executing starlark package")

	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error. This test requires you to be online for the read_file command to run")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.Empty(t, runResult.ExecutionError, "Unexpected execution error")
	logrus.Infof("Successfully ran Starlark Package")

}
