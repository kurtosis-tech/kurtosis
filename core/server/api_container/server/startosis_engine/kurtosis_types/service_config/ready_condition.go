package service_config

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/verify"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/recipe"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
	"reflect"
	"time"
)

const (
	ReadyConditionTypeName = "ReadyCondition"

	RecipeAttr    = "recipe"
	FieldAttr     = "field"
	AssertionAttr = "assertion"
	TargetAttr    = "target_value"
	IntervalAttr  = "interval"
	TimeoutAttr   = "timeout"

	defaultInterval = 1 * time.Second
	defaultTimeout  = 15 * time.Minute //TODO we could move these two to the service helpers method
)

func NewReadyConditionType() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return &kurtosis_type_constructor.KurtosisTypeConstructor{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: ReadyConditionTypeName,
			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              RecipeAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Value],
					Validator:         validateRecipe,
				},
				{
					Name:              FieldAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, FieldAttr)
					},
				},
				{
					Name:              AssertionAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator:         verify.ValidateVerificationToken,
				},
				{
					Name:              TargetAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Comparable],
					Validator:         nil,
				},
				{
					Name:              IntervalAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.Duration(value, IntervalAttr)
					},
				},
				{
					Name:              TimeoutAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.Duration(value, TimeoutAttr)
					},
				},
			},
		},
		Instantiate: instantiateReadyCondition,
	}
}

func instantiateReadyCondition(arguments *builtin_argument.ArgumentValuesSet) (builtin_argument.KurtosisValueType, *startosis_errors.InterpretationError) {
	kurtosisValueType, err := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(ReadyConditionTypeName, arguments)
	if err != nil {
		return nil, err
	}
	return &ReadyCondition{
		KurtosisValueTypeDefault: kurtosisValueType,
	}, nil
}

// ReadyCondition is a starlark.Value that holds all the information needed for ensuring service readiness
type ReadyCondition struct {
	*kurtosis_type_constructor.KurtosisValueTypeDefault
}

func (readyCondition *ReadyCondition) Copy() (builtin_argument.KurtosisValueType, error) {
	copiedValueType, err := readyCondition.KurtosisValueTypeDefault.Copy()
	if err != nil {
		return nil, err
	}
	return &ReadyCondition{
		KurtosisValueTypeDefault: copiedValueType,
	}, nil
}

func (readyCondition *ReadyCondition) GetRecipe() (recipe.Recipe, *startosis_errors.InterpretationError) {
	//TODO we should rework the recipe types to inherit a single common type, this will avoid the double parsing here.
	httpRecipe, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[recipe.HttpRequestRecipe](readyCondition.KurtosisValueTypeDefault, RecipeAttr)
	if !found {
		return nil, startosis_errors.NewInterpretationError("Required attribute '%s' could not be found on type '%s'",
			RecipeAttr, ReadyConditionTypeName)
	}
	if interpretationErr == nil {
		return httpRecipe, nil
	}
	execRecipe, _, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*recipe.ExecRecipe](readyCondition.KurtosisValueTypeDefault, RecipeAttr)
	if interpretationErr == nil {
		return execRecipe, nil
	}
	return nil, interpretationErr
}

func (readyCondition *ReadyCondition) GetField() (string, *startosis_errors.InterpretationError) {
	field, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](readyCondition.KurtosisValueTypeDefault, FieldAttr)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	if !found {
		return "", startosis_errors.NewInterpretationError("Required attribute '%s' could not be found on type '%s'",
			FieldAttr, ReadyConditionTypeName)
	}
	fieldStr := field.GoString()

	return fieldStr, nil
}

func (readyCondition *ReadyCondition) GetAssertion() (string, *startosis_errors.InterpretationError) {
	assertion, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](readyCondition.KurtosisValueTypeDefault, AssertionAttr)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	if !found {
		return "", startosis_errors.NewInterpretationError("Required attribute '%s' could not be found on type '%s'",
			AssertionAttr, ReadyConditionTypeName)
	}
	assertionStr := assertion.GoString()

	return assertionStr, nil
}

func (readyCondition *ReadyCondition) GetTarget() (starlark.Comparable, *startosis_errors.InterpretationError) {
	target, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.Comparable](readyCondition.KurtosisValueTypeDefault, TargetAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if !found {
		return nil, startosis_errors.NewInterpretationError("Required attribute '%s' could not be found on type '%s'",
			TargetAttr, ReadyConditionTypeName)
	}

	return target, nil
}

func (readyCondition *ReadyCondition) GetInterval() (time.Duration, *startosis_errors.InterpretationError) {
	interval := defaultInterval

	intervalStr, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](readyCondition.KurtosisValueTypeDefault, IntervalAttr)
	if interpretationErr != nil {
		return interval, interpretationErr
	}
	if found {
		parsedInterval, parseErr := time.ParseDuration(intervalStr.GoString())
		if parseErr != nil {
			return interval, startosis_errors.WrapWithInterpretationError(parseErr, "An error occurred when parsing interval '%v'", intervalStr.GoString())
		}
		interval = parsedInterval
	}

	return interval, nil
}

func (readyCondition *ReadyCondition) GetTimeout() (time.Duration, *startosis_errors.InterpretationError) {
	timeout := defaultTimeout

	timeoutStr, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](readyCondition.KurtosisValueTypeDefault, TimeoutAttr)
	if interpretationErr != nil {
		return timeout, interpretationErr
	}
	if found {
		parsedTimeout, parseErr := time.ParseDuration(timeoutStr.GoString())
		if parseErr != nil {
			return timeout, startosis_errors.WrapWithInterpretationError(parseErr, "An error occurred when parsing timeout '%v'", timeoutStr.GoString())
		}
		timeout = parsedTimeout
	}

	return timeout, nil
}

func validateRecipe(value starlark.Value) *startosis_errors.InterpretationError {
	_, ok := value.(recipe.HttpRequestRecipe)
	if !ok {
		//TODO we should rework the recipe types to inherit a single common type, this will avoid the double parsing here.
		_, ok := value.(*recipe.ExecRecipe)
		if !ok {
			return startosis_errors.NewInterpretationError("The '%s' attribute is not a Recipe (was '%s').", RecipeAttr, reflect.TypeOf(value))
		}
	}
	return nil
}
