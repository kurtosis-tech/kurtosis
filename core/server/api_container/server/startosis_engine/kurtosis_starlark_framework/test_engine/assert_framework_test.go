package test_engine

import (
	"fmt"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/verify"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

const (
	runtimeValueValue = "Hello World!"
)

type verificationTestcase struct {
	*testing.T

	runtimeValueStore *runtime_value_store.RuntimeValueStore
	runtimeValueUuid  string
}

func (suite *KurtosisPlanInstructionTestSuite) TestVerify() {
	runtimeValueUuid, err := uuid_generator.GenerateUUIDString()
	suite.Require().Nil(err)
	err = suite.runtimeValueStore.SetValue(runtimeValueUuid, map[string]starlark.Comparable{
		"value": starlark.String(runtimeValueValue),
	})
	suite.Require().NoError(err)

	suite.run(&verificationTestcase{
		T:                 suite.T(),
		runtimeValueUuid:  runtimeValueUuid,
		runtimeValueStore: suite.runtimeValueStore,
	})
}

func (t *verificationTestcase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return verify.NewVerify(t.runtimeValueStore)
}

func (t *verificationTestcase) GetStarlarkCode() string {
	runtimeValue := fmt.Sprintf("{{kurtosis:%s:value.runtime_value}}", t.runtimeValueUuid)
	assertion := "=="
	targetValue := runtimeValueValue
	return fmt.Sprintf("%s(%s=%q, %s=%q, %s=%q)", verify.VerifyBuiltinName, verify.RuntimeValueArgName, runtimeValue, verify.AssertionArgName, assertion, verify.TargetArgName, targetValue)
}

func (t *verificationTestcase) GetStarlarkCodeForAssertion() string {
	return ""
}

func (t *verificationTestcase) Assert(interpretationResult starlark.Value, executionResult *string) {
	require.Equal(t, starlark.None, interpretationResult)
	expectedExecutionResult := fmt.Sprintf(`Verification succeeded. Value is '%q'.`, runtimeValueValue)
	require.Equal(t, expectedExecutionResult, *executionResult)
}
