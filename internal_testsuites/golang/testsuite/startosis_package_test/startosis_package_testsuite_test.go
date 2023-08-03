package startosis_package_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"os"
	"path"
	"testing"
)

const (
	name                   = "startosis-package"
	emptyRunParams         = "{}"
	defaultDryRun          = false
	defaultParallelism     = 4
	useDefaultMainFile     = ""
	useDefaultFunctionName = ""
)

var (
	noExperimentalFeature = []kurtosis_core_rpc_api_bindings.KurtosisFeatureFlag{}
)

type StartosisPackageTestSuite struct {
	suite.Suite
	enclaveCtx         *enclaves.EnclaveContext
	destroyEnclaveFunc func() error
}

func TestStartosisPackageTestSuite(t *testing.T) {
	suite.Run(t, new(StartosisPackageTestSuite))
}

func (suite *StartosisPackageTestSuite) SetupTest() {
	ctx := context.Background()
	t := suite.T()
	enclaveCtx, _, destroyEnclaveFunc, err := test_helpers.CreateEnclave(t, ctx, name)
	require.NoError(t, err, "An error occurred creating an enclave")
	suite.enclaveCtx = enclaveCtx
	suite.destroyEnclaveFunc = destroyEnclaveFunc
}

func (suite *StartosisPackageTestSuite) TearDownTest() {
	err := suite.destroyEnclaveFunc()
	require.NoError(suite.T(), err, "Destroying the test suite's enclave process has failed, you will have to remove it manually")
}

func (suite *StartosisPackageTestSuite) RunPackage(ctx context.Context, packageRelativeDirpath string) (*enclaves.StarlarkRunResult, error) {
	return suite.RunPackageWithParams(ctx, packageRelativeDirpath, emptyRunParams)
}

func (suite *StartosisPackageTestSuite) RunPackageWithParams(ctx context.Context, packageRelativeDirpath string, params string) (*enclaves.StarlarkRunResult, error) {
	logrus.Infof("Executing Startosis package...")

	currentWorkingDirectory, err := os.Getwd()
	require.Nil(suite.T(), err)
	packageDirpath := path.Join(currentWorkingDirectory, packageRelativeDirpath)

	logrus.Debugf("Startosis package dirpath: %v", packageDirpath)

	return suite.enclaveCtx.RunStarlarkPackageBlocking(
		ctx,
		packageDirpath,
		useDefaultMainFile,
		useDefaultFunctionName,
		params,
		defaultDryRun,
		defaultParallelism,
		noExperimentalFeature,
	)
}
