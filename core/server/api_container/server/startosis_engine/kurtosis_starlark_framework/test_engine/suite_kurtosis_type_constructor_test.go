package test_engine

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_constants"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"github.com/stretchr/testify/suite"
	"go.starlark.net/starlark"
	"reflect"
	"testing"
)

const (
	kurtosisTypeConstructorThreadName = "kurtosis-type-constructor-test-suite"
)

type KurtosisTypeConstructorTestSuite struct {
	suite.Suite

	starlarkThread *starlark.Thread
	starlarkEnv    starlark.StringDict

	serviceNetwork         *service_network.MockServiceNetwork
	runtimeValueStore      *runtime_value_store.RuntimeValueStore
	packageContentProvider *startosis_packages.MockPackageContentProvider
}

func TestKurtosisTypeConstructorSuite(t *testing.T) {
	suite.Run(t, new(KurtosisTypeConstructorTestSuite))
}

func (suite *KurtosisTypeConstructorTestSuite) SetupTest() {
	suite.starlarkThread = newStarlarkThread(kurtosisTypeConstructorThreadName)
	suite.starlarkEnv = getBasePredeclaredDict(suite.T(), suite.starlarkThread)

	suite.serviceNetwork = service_network.NewMockServiceNetwork(suite.T())

	enclaveDb := getEnclaveDBForTest(suite.T())
	serde := kurtosis_types.NewStarlarkValueSerde(suite.starlarkThread, suite.starlarkEnv)
	runtimeValueStoreForTest, err := runtime_value_store.CreateRuntimeValueStore(serde, enclaveDb)
	suite.Require().NoError(err)
	suite.runtimeValueStore = runtimeValueStoreForTest

	suite.packageContentProvider = startosis_packages.NewMockPackageContentProvider(suite.T())
}

func (suite *KurtosisTypeConstructorTestSuite) run(builtin KurtosisTypeConstructorBaseTest) {
	starlarkCode := builtin.GetStarlarkCode()
	starlarkCodeToExecute := codeToExecute(starlarkCode)
	globals, err := starlark.ExecFile(suite.starlarkThread, startosis_constants.PackageIdPlaceholderForStandaloneScript, starlarkCodeToExecute, suite.starlarkEnv)
	suite.Require().Nil(err, "Error interpreting Starlark code. Code was: \n%v", starlarkCodeToExecute)
	result := extractResultValue(suite.T(), globals)

	kurtosisValue, ok := result.(builtin_argument.KurtosisValueType)
	suite.Require().True(ok, "Error casting the Kurtosis Type to a KurtosisValueType. This is unexpected as all "+
		"typed defined in Kurtosis should implement KurtosisValueType. Its type was: '%s'", reflect.TypeOf(kurtosisValue))

	builtin.Assert(kurtosisValue)

	copiedKurtosisValue, copyErr := kurtosisValue.Copy()
	suite.Require().NoError(copyErr)
	suite.Require().Equal(kurtosisValue, copiedKurtosisValue)
	suite.Require().NotSame(kurtosisValue, copiedKurtosisValue)

	serializedType := result.String()
	suite.Require().Equal(starlarkCode, serializedType)
}
