package test_engine

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_structure"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan/resolver"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/interpretation_time_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_constants"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages/mock_package_content_provider"
	"github.com/stretchr/testify/suite"
	"go.starlark.net/starlark"
	"testing"
)

const (
	kurtosisPlanInstructionThreadName = "kurtosis-plan-instruction-test-suite"
)

type KurtosisPlanInstructionTestSuite struct {
	suite.Suite

	starlarkThread *starlark.Thread
	starlarkEnv    starlark.StringDict

	serviceNetwork               *service_network.MockServiceNetwork
	runtimeValueStore            *runtime_value_store.RuntimeValueStore
	packageContentProvider       *mock_package_content_provider.MockPackageContentProvider
	interpretationTimeValueStore *interpretation_time_value_store.InterpretationTimeValueStore
}

func TestKurtosisPlanInstructionSuite(t *testing.T) {
	suite.Run(t, new(KurtosisPlanInstructionTestSuite))
}

func (suite *KurtosisPlanInstructionTestSuite) SetupTest() {
	suite.starlarkThread = newStarlarkThread(kurtosisPlanInstructionThreadName)
	suite.starlarkEnv = getBasePredeclaredDict(suite.T(), suite.starlarkThread)

	suite.serviceNetwork = service_network.NewMockServiceNetwork(suite.T())

	enclaveDb := getEnclaveDBForTest(suite.T())
	serde := kurtosis_types.NewStarlarkValueSerde(suite.starlarkThread, suite.starlarkEnv)
	runtimeValueStoreForTest, err := runtime_value_store.CreateRuntimeValueStore(serde, enclaveDb)
	suite.Require().NoError(err)
	suite.runtimeValueStore = runtimeValueStoreForTest

	interpretationTimeValueStore, err := interpretation_time_value_store.CreateInterpretationTimeValueStore(enclaveDb, serde)
	suite.Require().NoError(err)
	suite.interpretationTimeValueStore = interpretationTimeValueStore

	suite.packageContentProvider = mock_package_content_provider.NewMockPackageContentProvider()
}

func (suite *KurtosisPlanInstructionTestSuite) run(builtin KurtosisPlanInstructionBaseTest) {
	instructionsPlan := instructions_plan.NewInstructionsPlan()

	// Add the KurtosisPlanInstruction that is being tested
	instructionFromBuiltin := builtin.GetInstruction()
	emptyEnclaveComponents := enclave_structure.NewEnclaveComponents()
	emptyInstructionsPlanMask := resolver.NewInstructionsPlanMask(0)
	instructionWrapper := kurtosis_plan_instruction.NewKurtosisPlanInstructionWrapper(instructionFromBuiltin, emptyEnclaveComponents, nil, emptyInstructionsPlanMask, instructionsPlan)
	suite.starlarkEnv[instructionWrapper.GetName()] = starlark.NewBuiltin(instructionWrapper.GetName(), instructionWrapper.CreateBuiltin())

	starlarkCode := builtin.GetStarlarkCode()
	globals, err := starlark.ExecFile(suite.starlarkThread, startosis_constants.PackageIdPlaceholderForStandaloneScript, codeToExecute(starlarkCode), suite.starlarkEnv)
	suite.Require().Nil(err, "Error interpreting Starlark code")
	interpretationResult := extractResultValue(suite.T(), globals)

	suite.Require().Equal(1, instructionsPlan.Size())
	instructionsSequence, err := instructionsPlan.GeneratePlan()
	suite.Require().Nil(err)
	instructionToExecute := instructionsSequence[0].GetInstruction()

	// execute the instruction and run custom builtin assertions
	executionResult, err := instructionToExecute.Execute(context.WithValue(context.Background(), startosis_constants.ParallelismParam, 1))
	suite.Require().Nil(err)
	builtin.Assert(interpretationResult, executionResult)

	// check serializing the obtained instruction falls back to the initial one
	serializedInstruction := instructionToExecute.String()

	starlarkCodeForAssertion := builtin.GetStarlarkCodeForAssertion()
	if starlarkCodeForAssertion == "" {
		starlarkCodeForAssertion = starlarkCode
	}

	suite.Require().Equal(starlarkCodeForAssertion, serializedInstruction)
}

func (suite *KurtosisPlanInstructionTestSuite) runShouldFail(packageId string, builtin KurtosisPlanInstructionBaseTest, expectedErrMsg string) {
	instructionsPlan := instructions_plan.NewInstructionsPlan()

	// Add the KurtosisPlanInstruction that is being tested
	instructionFromBuiltin := builtin.GetInstruction()
	emptyEnclaveComponents := enclave_structure.NewEnclaveComponents()
	emptyInstructionsPlanMask := resolver.NewInstructionsPlanMask(0)
	instructionWrapper := kurtosis_plan_instruction.NewKurtosisPlanInstructionWrapper(instructionFromBuiltin, emptyEnclaveComponents, nil, emptyInstructionsPlanMask, instructionsPlan)
	suite.starlarkEnv[instructionWrapper.GetName()] = starlark.NewBuiltin(instructionWrapper.GetName(), instructionWrapper.CreateBuiltin())

	starlarkCode := builtin.GetStarlarkCode()
	_, err := starlark.ExecFile(suite.starlarkThread, packageId, codeToExecute(starlarkCode), suite.starlarkEnv)
	suite.Require().Error(err, "Expected to fail running starlark code %s, but it didn't fail", builtin.GetStarlarkCode())
	suite.Require().Equal(expectedErrMsg, err.Error())
}
