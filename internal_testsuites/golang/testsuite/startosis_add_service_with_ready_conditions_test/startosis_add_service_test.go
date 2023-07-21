package startosis_add_service_with_ready_conditions_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

const (
	name                  = "startosis-add-service"
	isPartitioningEnabled = false
)

type StartosisAddServiceReadyTestSuite struct {
	suite.Suite
	enclaveCtx         *enclaves.EnclaveContext
	destroyEnclaveFunc func() error
}

func TestStartosisAddServiceTestSuite(t *testing.T) {
	suite.Run(t, new(StartosisAddServiceReadyTestSuite))
}

func (suite *StartosisAddServiceReadyTestSuite) SetupSuite() {
	ctx := context.Background()
	t := suite.T()
	enclaveCtx, _, destroyEnclaveFunc, err := test_helpers.CreateEnclave(t, ctx, name, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	suite.enclaveCtx = enclaveCtx
	suite.destroyEnclaveFunc = destroyEnclaveFunc
}

func (suite *StartosisAddServiceReadyTestSuite) TearDownSuite() {
	err := suite.destroyEnclaveFunc()
	require.NoError(suite.T(), err, "Destroying the test suite's enclave process has failed, you will have to remove it manually")
}

func (suite *StartosisAddServiceReadyTestSuite) RunScript(ctx context.Context, script string) (*enclaves.StarlarkRunResult, error) {
	logrus.Infof("Executing Startosis script...")
	logrus.Debugf("Startosis script content: \n%v", script)

	return test_helpers.RunScriptWithDefaultConfig(ctx, suite.enclaveCtx, script)
}
