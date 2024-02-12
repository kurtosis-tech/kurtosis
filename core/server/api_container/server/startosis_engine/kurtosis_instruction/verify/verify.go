package verify

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_plan_persistence"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_structure"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers/magic_string_helper"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
	"reflect"
	"strings"
)

const (
	VerifyBuiltinName = "verify"

	RuntimeValueArgName = "value"
	AssertionArgName    = "assertion"
	TargetArgName       = "target_value"

	InCollectionAssertionToken    = "IN"
	NotInCollectionAssertionToken = "NOT_IN"

	expectedValuesSeparator = ", "
)

var StringTokenToComparisonStarlarkToken = map[string]syntax.Token{
	"==": syntax.EQL,
	"!=": syntax.NEQ,
	">=": syntax.GE,
	">":  syntax.GT,
	"<=": syntax.LE,
	"<":  syntax.LT,
}

func NewVerify(runtimeValueStore *runtime_value_store.RuntimeValueStore) *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return &kurtosis_plan_instruction.KurtosisPlanInstruction{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: VerifyBuiltinName,

			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              RuntimeValueArgName,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator:         nil,
				},
				{
					Name:              AssertionArgName,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator:         ValidateVerificationToken,
				},
				{
					Name:              TargetArgName,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Comparable],
					Validator:         nil,
				},
			},
		},

		Capabilities: func() kurtosis_plan_instruction.KurtosisPlanInstructionCapabilities {
			return &VerifyCapabilities{
				runtimeValueStore: runtimeValueStore,

				runtimeValue: "",  // populated at interpretation time
				assertion:    "",  // populated at interpretation time
				target:       nil, // populated at interpretation time
			}
		},

		DefaultDisplayArguments: map[string]bool{
			RuntimeValueArgName: true,
			AssertionArgName:    true,
			TargetArgName:       true,
		},
	}
}

type VerifyCapabilities struct {
	runtimeValueStore *runtime_value_store.RuntimeValueStore

	runtimeValue string
	assertion    string
	target       starlark.Comparable
}

func (builtin *VerifyCapabilities) Interpret(_ string, arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError) {
	runtimeValue, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, RuntimeValueArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", RuntimeValueArgName)
	}
	verification, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, AssertionArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", AssertionArgName)
	}
	target, err := builtin_argument.ExtractArgumentValue[starlark.Comparable](arguments, TargetArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", TargetArgName)
	}

	builtin.assertion = verification.GoString()
	builtin.runtimeValue = runtimeValue.GoString()
	builtin.target = target

	if _, ok := builtin.target.(starlark.Iterable); (builtin.assertion == InCollectionAssertionToken || builtin.assertion == NotInCollectionAssertionToken) && !ok {
		return nil, startosis_errors.NewInterpretationError("'%v' assertion requires an iterable for target values, got '%v'", builtin.assertion, builtin.target.Type())
	}
	return starlark.None, nil
}

func (builtin *VerifyCapabilities) Validate(_ *builtin_argument.ArgumentValuesSet, _ *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
	return nil
}

func (builtin *VerifyCapabilities) Execute(_ context.Context, _ *builtin_argument.ArgumentValuesSet) (string, error) {
	currentValue, err := magic_string_helper.GetOrReplaceRuntimeValueFromString(builtin.runtimeValue, builtin.runtimeValueStore)
	if err != nil {
		return "", err
	}
	targetWithReplacedRuntimeValuesMaybe := builtin.target
	targetStr, ok := builtin.target.(starlark.String)
	if ok {
		// target ws a string. Apply runtime value replacement in case it contains one
		targetWithReplacedRuntimeValuesMaybe, err = magic_string_helper.GetOrReplaceRuntimeValueFromString(targetStr.GoString(), builtin.runtimeValueStore)
		if err != nil {
			return "", err
		}
	}
	err = Verify(currentValue, builtin.assertion, targetWithReplacedRuntimeValuesMaybe)
	if err != nil {
		return "", err
	}
	instructionResult := fmt.Sprintf("Verification succeeded. Value is '%s'.", currentValue.String())
	return instructionResult, nil
}

func (builtin *VerifyCapabilities) TryResolveWith(instructionsAreEqual bool, _ *enclave_plan_persistence.EnclavePlanInstruction, _ *enclave_structure.EnclaveComponents) enclave_structure.InstructionResolutionStatus {
	if instructionsAreEqual {
		return enclave_structure.InstructionIsEqual
	}
	return enclave_structure.InstructionIsUnknown
}

func (builtin *VerifyCapabilities) FillPersistableAttributes(builder *enclave_plan_persistence.EnclavePlanInstructionBuilder) {
	builder.SetType(VerifyBuiltinName)
}

func (builtin *VerifyCapabilities) Description() string {
	return fmt.Sprintf("Verifying whether '%v' '%v' '%v'", builtin.runtimeValue, builtin.assertion, builtin.target)
}

// Verify verifies whether the currentValue matches the targetValue w.r.t. the assertion operator
// TODO: This and ValidateVerificationToken below are used by both verify and wait. Refactor it to a better place
func Verify(currentValue starlark.Comparable, assertion string, targetValue starlark.Comparable) error {
	if comparisonToken, found := StringTokenToComparisonStarlarkToken[assertion]; found {
		if currentValue.Type() != targetValue.Type() {
			return stacktrace.NewError("Verify failed because '%v' is type '%v' and '%v' is type '%v'", currentValue, currentValue.Type(), targetValue, targetValue.Type())
		}
		result, err := currentValue.CompareSameType(comparisonToken, targetValue, 1)
		if err != nil {
			return stacktrace.Propagate(err, "Verify comparison failed '%v' '%v' '%v'", currentValue, assertion, targetValue)
		}
		if !result {
			return stacktrace.NewError("Verification failed '%v' '%v' '%v'", currentValue, assertion, targetValue)
		}
		return nil
	} else if assertion == InCollectionAssertionToken || assertion == NotInCollectionAssertionToken {
		iterableTarget, ok := targetValue.(starlark.Iterable)
		if !ok {
			return stacktrace.NewError("Verification failed, expected an iterable object but got '%v'", targetValue.Type())
		}

		iterator := iterableTarget.Iterate()
		defer iterator.Done()
		var item starlark.Value
		currentValuePresentInIterable := false
		for idx := 0; iterator.Next(&item); idx++ {
			if item == currentValue {
				if assertion == InCollectionAssertionToken {
					return nil
				}
				currentValuePresentInIterable = true
				break
			}
		}
		if assertion == NotInCollectionAssertionToken && !currentValuePresentInIterable {
			return nil
		}
		return stacktrace.NewError("Verification failed '%v' '%v' '%v'", currentValue, assertion, targetValue)
	}
	return stacktrace.NewError("The '%s' token '%s' seems invalid. This is a Kurtosis bug as it should have been validated earlier", AssertionArgName, assertion)
}

func ValidateVerificationToken(value starlark.Value) *startosis_errors.InterpretationError {
	strValue, ok := value.(starlark.String)
	if !ok {
		return startosis_errors.NewInterpretationError("'%s' argument should be a 'starlark.String', got '%s'", AssertionArgName, reflect.TypeOf(value))
	}
	var validTokens []string
	for stringComparisonToken := range StringTokenToComparisonStarlarkToken {
		validTokens = append(validTokens, stringComparisonToken)
	}
	validTokens = append(validTokens, InCollectionAssertionToken)
	validTokens = append(validTokens, NotInCollectionAssertionToken)
	found := false
	for _, validToken := range validTokens {
		if validToken == strValue.GoString() {
			found = true
			break
		}
	}
	if !found {
		return startosis_errors.NewInterpretationError("'%s' argument is invalid, valid values are: '%s'", AssertionArgName, strings.Join(validTokens, expectedValuesSeparator))
	}
	return nil
}
