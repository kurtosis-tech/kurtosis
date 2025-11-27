package startosis_directory_test

import (
	"context"
	"github.com/dzobbe/PoTE-kurtosis/api/golang/core/lib/starlark_run_config"
	"os"
	"path"
	"testing"

	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/dzobbe/PoTE-kurtosis/api/golang/core/lib/enclaves"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	name           = "startosis-directory"
	emptyRunParams = "{}"
)

type StartosisDirectoryTestSuite struct {
	suite.Suite
	enclaveCtx         *enclaves.EnclaveContext
	destroyEnclaveFunc func() error
}

func TestStartosisDirectoryTestSuite(t *testing.T) {
	suite.Run(t, new(StartosisDirectoryTestSuite))
}

func (suite *StartosisDirectoryTestSuite) SetupTest() {
	ctx := context.Background()
	t := suite.T()
	enclaveCtx, _, destroyEnclaveFunc, err := test_helpers.CreateEnclave(t, ctx, name)
	require.NoError(t, err, "An error occurred creating an enclave")
	suite.enclaveCtx = enclaveCtx
	suite.destroyEnclaveFunc = destroyEnclaveFunc
}

func (suite *StartosisDirectoryTestSuite) TearDownTest() {
	err := suite.destroyEnclaveFunc()
	require.NoError(suite.T(), err, "Destroying the test suite's enclave process has failed, you will have to remove it manually")
}

func (suite *StartosisDirectoryTestSuite) RunScript(ctx context.Context, script string) (*enclaves.StarlarkRunResult, error) {
	logrus.Infof("Executing Startosis script...")
	logrus.Debugf("Startosis script content: \n%v", script)

	return test_helpers.RunScriptWithDefaultConfig(ctx, suite.enclaveCtx, script)
}

func (suite *StartosisDirectoryTestSuite) RunPackage(ctx context.Context, packageRelativeDirpath string) (*enclaves.StarlarkRunResult, error) {
	return suite.RunPackageWithParams(ctx, packageRelativeDirpath, emptyRunParams)
}

func (suite *StartosisDirectoryTestSuite) RunPackageWithParams(ctx context.Context, packageRelativeDirpath string, params string) (*enclaves.StarlarkRunResult, error) {
	logrus.Infof("Executing Startosis package...")

	currentWorkingDirectory, err := os.Getwd()
	require.Nil(suite.T(), err)
	packageDirpath := path.Join(currentWorkingDirectory, packageRelativeDirpath)

	logrus.Debugf("Startosis package dirpath: %v", packageDirpath)

	return suite.enclaveCtx.RunStarlarkPackageBlocking(
		ctx,
		packageDirpath,
		starlark_run_config.NewRunStarlarkConfig(starlark_run_config.WithSerializedParams(params)),
	)
}
