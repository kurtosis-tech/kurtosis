package startosis_subpackage_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/dzobbe/PoTE-kurtosis/api/golang/core/lib/enclaves"
	"github.com/dzobbe/PoTE-kurtosis/api/golang/core/lib/starlark_run_config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

const (
	name = "startosis-ports-wait"
)

type StartosisSubpackageTestSuite struct {
	suite.Suite
	enclaveCtx         *enclaves.EnclaveContext
	destroyEnclaveFunc func() error
}

func TestStartosisSubpackageTestSuite(t *testing.T) {
	suite.Run(t, new(StartosisSubpackageTestSuite))
}

func (suite *StartosisSubpackageTestSuite) SetupTest() {
	ctx := context.Background()
	t := suite.T()
	enclaveCtx, _, destroyEnclaveFunc, err := test_helpers.CreateEnclave(t, ctx, name)
	require.NoError(t, err, "An error occurred creating an enclave")
	suite.enclaveCtx = enclaveCtx
	suite.destroyEnclaveFunc = destroyEnclaveFunc
}

func (suite *StartosisSubpackageTestSuite) TearDownTest() {
	err := suite.destroyEnclaveFunc()
	require.NoError(suite.T(), err, "Destroying the test suite's enclave process has failed, you will have to remove it manually")
}

func (suite *StartosisSubpackageTestSuite) RunPackage(ctx context.Context, packageLocation string, remote bool) (*enclaves.StarlarkRunResult, error) {
	logrus.Infof("Executing Startosis package...")

	logrus.Debugf("Startosis package location: %v", packageLocation)

	if remote {
		return suite.enclaveCtx.RunStarlarkRemotePackageBlocking(
			ctx,
			packageLocation,
			starlark_run_config.NewRunStarlarkConfig(),
		)
	}

	return suite.enclaveCtx.RunStarlarkPackageBlocking(
		ctx,
		packageLocation,
		starlark_run_config.NewRunStarlarkConfig(),
	)
}
