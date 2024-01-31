package service_config

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
	v1 "k8s.io/api/core/v1"
)

const (
	TolerationTypeName = "Toleration"

	KeyAttr               = "key"
	OperatorAttr          = "operator"
	ValueAttr             = "value"
	EffectAttr            = "effect"
	TolerationSecondsAttr = "toleration_seconds"
)

var allowedOperatorValues = []string{string(v1.TolerationOpExists), string(v1.TolerationOpEqual)}
var allowedEffectValues = []string{string(v1.TaintEffectNoSchedule), string(v1.TaintEffectNoExecute), string(v1.TaintEffectPreferNoSchedule)}

func NewTolerationType() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return &kurtosis_type_constructor.KurtosisTypeConstructor{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: TolerationTypeName,
			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              KeyAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, KeyAttr)
					},
				},
				{
					Name:              OperatorAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.StringValues(value, OperatorAttr, allowedOperatorValues)
					},
				},
				{
					Name:              ValueAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
				},
				{
					Name:              EffectAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.StringValues(value, EffectAttr, allowedEffectValues)
					},
				},
				{
					Name:              TolerationSecondsAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Int],
					Validator:         nil,
				},
			},
			Deprecation: nil,
		},
		Instantiate: instantiateToleration,
	}
}

func instantiateToleration(arguments *builtin_argument.ArgumentValuesSet) (builtin_argument.KurtosisValueType, *startosis_errors.InterpretationError) {
	kurtosisValueType, interpretationErr := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(TolerationTypeName, arguments)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return &Toleration{
		kurtosisValueType,
	}, nil
}

type Toleration struct {
	*kurtosis_type_constructor.KurtosisValueTypeDefault
}

func (toleration *Toleration) Copy() (builtin_argument.KurtosisValueType, error) {
	copiedValueType, err := toleration.KurtosisValueTypeDefault.Copy()
	if err != nil {
		return nil, err
	}
	return &Toleration{
		KurtosisValueTypeDefault: copiedValueType,
	}, nil
}

func (toleration *Toleration) GetKeyIfSet() (string, bool, *startosis_errors.InterpretationError) {
	keyValue, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](
		toleration.KurtosisValueTypeDefault, KeyAttr)
	if interpretationErr != nil {
		return "", false, interpretationErr
	}
	if !found {
		return "", false, nil
	}
	return keyValue.GoString(), true, nil
}

func (toleration *Toleration) GetOperatorIfSet() (string, bool, *startosis_errors.InterpretationError) {
	attrValue, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](
		toleration.KurtosisValueTypeDefault, OperatorAttr)
	if interpretationErr != nil {
		return "", false, interpretationErr
	}
	if !found {
		return "", false, nil
	}
	return attrValue.GoString(), true, nil
}

func (toleration *Toleration) GetValueIfExists() (string, bool, *startosis_errors.InterpretationError) {
	attrValue, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](
		toleration.KurtosisValueTypeDefault, ValueAttr)
	if interpretationErr != nil {
		return "", false, interpretationErr
	}
	if !found {
		return "", false, nil
	}
	return attrValue.GoString(), true, nil
}

func (toleration *Toleration) GetEffectIfExist() (string, bool, *startosis_errors.InterpretationError) {
	attrValue, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](
		toleration.KurtosisValueTypeDefault, EffectAttr)
	if interpretationErr != nil {
		return "", false, interpretationErr
	}
	if !found {
		return "", false, nil
	}
	return attrValue.GoString(), true, nil
}

func (toleration *Toleration) GetTolerationSecondsIfExists() (int64, bool, *startosis_errors.InterpretationError) {
	attrValue, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.Int](
		toleration.KurtosisValueTypeDefault, TolerationSecondsAttr)
	if interpretationErr != nil {
		return 0, false, interpretationErr
	}
	if !found {
		return 0, false, nil
	}
	attr, ok := attrValue.Int64()
	if !ok {
		return 0, false, startosis_errors.NewInterpretationError("Couldn't convert '%v' '%v' to int64", TolerationSecondsAttr, attr)
	}
	return attr, true, nil
}

func (toleration *Toleration) ToKubeType() (*v1.Toleration, *startosis_errors.InterpretationError) {
	//nolint :exhaustruct
	returnValue := &v1.Toleration{}

	key, keyFound, err := toleration.GetKeyIfSet()
	if err != nil {
		return nil, err
	}

	if keyFound {
		returnValue.Key = key
	}

	operator, operatorFound, err := toleration.GetOperatorIfSet()
	if err != nil {
		return nil, err
	}

	if !keyFound && (!operatorFound || operator != string(v1.TolerationOpExists)) {
		return nil, startosis_errors.NewInterpretationError("'%v' expects either '%v' to be set or for '%v' to be '%v'", TolerationTypeName, KeyAttr, OperatorAttr, v1.TolerationOpExists)
	}

	if operatorFound {
		returnValue.Operator = v1.TolerationOperator(operator)
	}

	value, valueFound, err := toleration.GetValueIfExists()
	if err != nil {
		return nil, err
	}

	if operatorFound && operator == string(v1.TolerationOpExists) && valueFound && value != "" {
		return nil, startosis_errors.NewInterpretationError("'%v' cannot have non empty value '%v' for '%v' if '%v' is '%v'", TolerationTypeName, value, ValueAttr, OperatorAttr, operator)
	}

	if valueFound {
		returnValue.Value = value
	}

	effect, effectFound, err := toleration.GetEffectIfExist()
	if err != nil {
		return nil, err
	}

	if effectFound {
		returnValue.Effect = v1.TaintEffect(effect)
	}

	tolerationSeconds, tolerationSecondsFound, err := toleration.GetTolerationSecondsIfExists()
	if err != nil {
		return nil, err
	}

	if tolerationSecondsFound {
		returnValue.TolerationSeconds = &tolerationSeconds
	}

	return returnValue, nil
}
