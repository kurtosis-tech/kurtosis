package kubernetes

import (
	"math"

	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
)

const (
	IngressPortConfigTypeName = "IngressPortConfig"
	PortNameAttr              = "name"
	PortNumberAttr            = "number"
)

func NewIngressPortConfigType() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return &kurtosis_type_constructor.KurtosisTypeConstructor{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: IngressPortConfigTypeName,
			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              PortNameAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
				},
				{
					Name:              PortNumberAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Int],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.Int64InRange(value, PortNameAttr, 0, int64(math.Pow(2, 16)-1))
					},
				},
			},
			Deprecation: nil,
		},
		Instantiate: instantiateIngressPortConfig,
	}
}

func instantiateIngressPortConfig(arguments *builtin_argument.ArgumentValuesSet) (builtin_argument.KurtosisValueType, *startosis_errors.InterpretationError) {
	kurtosisValueType, interpretationErr := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(IngressPortConfigTypeName, arguments)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return &IngressPortConfig{
		KurtosisValueTypeDefault: kurtosisValueType,
	}, nil
}

type IngressPortConfig struct {
	*kurtosis_type_constructor.KurtosisValueTypeDefault
}

func (config *IngressPortConfig) Copy() (builtin_argument.KurtosisValueType, error) {
	copiedValueType, err := config.KurtosisValueTypeDefault.Copy()
	if err != nil {
		return nil, err
	}
	return &IngressPortConfig{
		KurtosisValueTypeDefault: copiedValueType,
	}, nil
}

func (config *IngressPortConfig) GetPortName() (*string, *startosis_errors.InterpretationError) {
	portName, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*starlark.String](
		config.KurtosisValueTypeDefault, PortNameAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if !found || portName == nil {
		return nil, nil
	}
	s := portName.GoString()
	return &s, nil
}

func (config *IngressPortConfig) GetPortNumber() (*uint16, *startosis_errors.InterpretationError) {
	portNumber, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*starlark.Int](
		config.KurtosisValueTypeDefault, PortNumberAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if !found || portNumber == nil {
		return nil, nil
	}
	var uint16Port uint16
	err := starlark.AsInt(portNumber, &uint16Port)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(
			err,
			"Error interpreting port number %s as uint16",
			portNumber,
		)
	}
	return &uint16Port, nil
}
