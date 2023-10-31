package test_engine

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_constants"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"github.com/stretchr/testify/suite"
	"go.starlark.net/starlark"
	"testing"
)

const (
	kurtosisHelperThreadName = "kurtosis-helper-test-suite"
)

type KurtosisHelperTestSuite struct {
	suite.Suite

	starlarkThread *starlark.Thread
	starlarkEnv    starlark.StringDict

	packageContentProvider *startosis_packages.MockPackageContentProvider
}

func TestKurtosisHelperSuite(t *testing.T) {
	suite.Run(t, new(KurtosisHelperTestSuite))
}

func (suite *KurtosisHelperTestSuite) SetupTest() {
	suite.starlarkThread = newStarlarkThread(kurtosisHelperThreadName)
	suite.starlarkEnv = getBasePredeclaredDict(suite.T(), suite.starlarkThread)

	suite.packageContentProvider = startosis_packages.NewMockPackageContentProvider(suite.T())
}

func (suite *KurtosisHelperTestSuite) run(builtin KurtosisHelperBaseTest) {
	// Add the KurtosisPlanInstruction that is being tested
	helper := builtin.GetHelper()
	suite.starlarkEnv[helper.GetName()] = starlark.NewBuiltin(helper.GetName(), helper.CreateBuiltin())

	starlarkCode := builtin.GetStarlarkCode()
	globals, err := starlark.ExecFile(suite.starlarkThread, startosis_constants.PackageIdPlaceholderForStandaloneScript, codeToExecute(starlarkCode), suite.starlarkEnv)
	suite.Require().Nil(err, "Error interpreting Starlark code")
	result := extractResultValue(suite.T(), globals)

	builtin.Assert(result)
}

func (suite *KurtosisHelperTestSuite) runShouldFail(moduleName string, builtin KurtosisHelperBaseTest, expectedErrMsg string) {
	// Add the KurtosisPlanInstruction that is being tested
	helper := builtin.GetHelper()
	suite.starlarkEnv[helper.GetName()] = starlark.NewBuiltin(helper.GetName(), helper.CreateBuiltin())

	starlarkCode := builtin.GetStarlarkCode()
	_, err := starlark.ExecFile(suite.starlarkThread, moduleName, codeToExecute(starlarkCode), suite.starlarkEnv)
	suite.Require().Error(err, "Expected to fail running starlark code %s, but it didn't fail", builtin.GetStarlarkCode())
	suite.Require().Equal(expectedErrMsg, err.Error())
}
