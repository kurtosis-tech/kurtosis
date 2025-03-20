package kubernetes

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
)

const (
	IngressHttpRuleTypeName = "IngressHttpRule"

	PathAttr     = "path"
	PathTypeAttr = "path_type"
	PortAttr     = "port"
)

func NewIngressHttpRuleType() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return &kurtosis_type_constructor.KurtosisTypeConstructor{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: IngressHttpRuleTypeName,
			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              PathAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						// TODO: Use k8s validator if exists?
						return builtin_argument.NonEmptyString(value, PathAttr)
					},
				},
				{
					Name:              PathTypeAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						str, ok := value.(starlark.String)
						if !ok {
							return startosis_errors.NewInterpretationError("Expected '%s' to be a string", PathTypeAttr)
						}
						pathType := str.GoString()
						if pathType != "Prefix" && pathType != "Exact" && pathType != "ImplementationSpecific" {
							return startosis_errors.NewInterpretationError("PathType must be one of: Prefix, Exact, ImplementationSpecific")
						}
						return nil
					},
				},
			},
			Deprecation: nil,
		},
		Instantiate: instantiateIngressPathConfig,
	}
}

func instantiateIngressPathConfig(arguments *builtin_argument.ArgumentValuesSet) (
	builtin_argument.KurtosisValueType, *startosis_errors.InterpretationError) {
	kurtosisValueType, interpretationErr := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(IngressHttpRuleTypeName, arguments)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return &IngressHttpRule{
		KurtosisValueTypeDefault: kurtosisValueType,
	}, nil
}

type IngressHttpRule struct {
	*kurtosis_type_constructor.KurtosisValueTypeDefault
}

func (config *IngressHttpRule) Copy() (builtin_argument.KurtosisValueType, error) {
	copiedValueType, err := config.KurtosisValueTypeDefault.Copy()
	if err != nil {
		return nil, err
	}
	return &IngressHttpRule{
		KurtosisValueTypeDefault: copiedValueType,
	}, nil
}

func (config *IngressHttpRule) GetPath() (string, *startosis_errors.InterpretationError) {
	path, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](
		config.KurtosisValueTypeDefault, PathAttr)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	if !found {
		return "", startosis_errors.NewInterpretationError("Required attribute '%v' couldn't be found on '%v' type",
			PathAttr, config)
	}
	return path.GoString(), nil
}

func (config *IngressHttpRule) GetPathType() (string, *startosis_errors.InterpretationError) {
	pathType, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](
		config.KurtosisValueTypeDefault, PathTypeAttr)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	if !found {
		return "", startosis_errors.NewInterpretationError("Required attribute '%v' couldn't be found on '%v' type",
			PathTypeAttr, config)
	}
	return pathType.GoString(), nil
}

func (config *IngressHttpRule) GetPortConfig() (*IngressPortConfig, *startosis_errors.InterpretationError) {
	portConfig, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*IngressPortConfig](
		config.KurtosisValueTypeDefault, PortAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if !found {
		return nil, startosis_errors.NewInterpretationError("Required attribute '%v' couldn't be found on '%v' type",
			PortAttr, config)
	}

	return portConfig, nil
}
