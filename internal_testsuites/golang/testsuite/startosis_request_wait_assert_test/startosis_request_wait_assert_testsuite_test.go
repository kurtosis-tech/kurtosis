package startosis_request_wait_assert_test

// Basic imports
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
	name                 = "example-test-suite"
	partitioningDisabled = false
)

type StartosisRequestWaitAssertTestSuite struct {
	suite.Suite
	enclaveCtx         *enclaves.EnclaveContext
	destroyEnclaveFunc func() error
}

func TestStartosisRequestWaitAssertTestSuite(t *testing.T) {
	suite.Run(t, new(StartosisRequestWaitAssertTestSuite))
}

func (suite *StartosisRequestWaitAssertTestSuite) SetupSuite() {
	ctx := context.Background()
	t := suite.T()
	enclaveCtx, _, destroyEnclaveFunc, err := test_helpers.CreateEnclave(t, ctx, name, partitioningDisabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	suite.enclaveCtx = enclaveCtx
	suite.destroyEnclaveFunc = destroyEnclaveFunc
}

func (suite *StartosisRequestWaitAssertTestSuite) TearDownSuite() {
	err := suite.destroyEnclaveFunc()
	require.NoError(suite.T(), err, "Destroying the test suite's enclave process has failed, you will have to remove it manually")
}

func (suite *StartosisRequestWaitAssertTestSuite) RunScript(ctx context.Context, script string) (*enclaves.StarlarkRunResult, error) {
	logrus.Infof("Executing Startosis script...")
	logrus.Debugf("Startosis script content: \n%v", script)

	return test_helpers.RunScriptWithDefaultConfig(ctx, suite.enclaveCtx, script)
}
