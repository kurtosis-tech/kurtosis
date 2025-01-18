package startosis_upload_file_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis/api/golang/v2/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/v2/core/lib/starlark_run_config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

const (
	name = "startosis-upload-file"
)

type StartosisUploadFileTestSuite struct {
	suite.Suite
	enclaveCtx         *enclaves.EnclaveContext
	destroyEnclaveFunc func() error
}

func TestStartosisUploadFileTestSuite(t *testing.T) {
	suite.Run(t, new(StartosisUploadFileTestSuite))
}

func (suite *StartosisUploadFileTestSuite) SetupTest() {
	ctx := context.Background()
	t := suite.T()
	enclaveCtx, _, destroyEnclaveFunc, err := test_helpers.CreateEnclave(t, ctx, name)
	require.NoError(t, err, "An error occurred creating an enclave")
	suite.enclaveCtx = enclaveCtx
	suite.destroyEnclaveFunc = destroyEnclaveFunc
}

func (suite *StartosisUploadFileTestSuite) TearDownTest() {
	err := suite.destroyEnclaveFunc()
	require.NoError(suite.T(), err, "Destroying the test suite's enclave process has failed, you will have to remove it manually")
}

func (suite *StartosisUploadFileTestSuite) RunPackage(
	ctx context.Context,
	packageLocation string,
	runConfig *starlark_run_config.StarlarkRunConfig,
	remote bool,
) (*enclaves.StarlarkRunResult, error) {
	logrus.Infof("Executing Startosis package...")

	logrus.Debugf("Startosis package location: %v", packageLocation)

	if runConfig == nil {
		runConfig = starlark_run_config.NewRunStarlarkConfig()
	}

	if remote {
		return suite.enclaveCtx.RunStarlarkRemotePackageBlocking(
			ctx,
			packageLocation,
			runConfig,
		)
	}

	return suite.enclaveCtx.RunStarlarkPackageBlocking(
		ctx,
		packageLocation,
		runConfig,
	)
}
