package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/assert"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

const (
	runtimeValueValue = "Hello World!"
)

type assertTestCase struct {
	*testing.T

	runtimeValueStore *runtime_value_store.RuntimeValueStore
	runtimeValueUuid  string
}

func (suite *KurtosisPlanInstructionTestSuite) TestAssert() {
	runtimeValueUuid, err := uuid_generator.GenerateUUIDString()
	suite.Require().Nil(err)
	err = suite.runtimeValueStore.SetValue(runtimeValueUuid, map[string]starlark.Comparable{
		"value": starlark.String(runtimeValueValue),
	})
	suite.Require().NoError(err)

	suite.run(&assertTestCase{
		T:                 suite.T(),
		runtimeValueUuid:  runtimeValueUuid,
		runtimeValueStore: suite.runtimeValueStore,
	})
}

func (t *assertTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return assert.NewAssert(t.runtimeValueStore)
}

func (t *assertTestCase) GetStarlarkCode() string {
	runtimeValue := fmt.Sprintf("{{kurtosis:%s:value.runtime_value}}", t.runtimeValueUuid)
	assertion := "=="
	targetValue := runtimeValueValue
	return fmt.Sprintf("%s(%s=%q, %s=%q, %s=%q)", assert.AssertBuiltinName, assert.RuntimeValueArgName, runtimeValue, assert.AssertionArgName, assertion, assert.TargetArgName, targetValue)
}

func (t *assertTestCase) GetStarlarkCodeForAssertion() string {
	return ""
}

func (t *assertTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	require.Equal(t, starlark.None, interpretationResult)
	expectedExecutionResult := fmt.Sprintf(`Assertion succeeded. Value is '%q'.`, runtimeValueValue)
	require.Equal(t, expectedExecutionResult, *executionResult)
}
