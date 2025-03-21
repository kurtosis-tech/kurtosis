package kubernetes

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
)

const (
	IngressTLSConfigTypeName = "IngressTLSConfig"
	SecretNameAttr           = "secret_name"
)

func NewIngressTLSConfigType() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return &kurtosis_type_constructor.KurtosisTypeConstructor{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: IngressTLSConfigTypeName,
			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              SecretNameAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, SecretNameAttr)
					},
				},
			},
			Deprecation: nil,
		},
		Instantiate: instantiateIngressTLSConfig,
	}
}

func instantiateIngressTLSConfig(arguments *builtin_argument.ArgumentValuesSet) (builtin_argument.KurtosisValueType, *startosis_errors.InterpretationError) {
	kurtosisValueType, interpretationErr := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(IngressTLSConfigTypeName, arguments)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return &IngressTLSConfig{
		KurtosisValueTypeDefault: kurtosisValueType,
	}, nil
}

type IngressTLSConfig struct {
	*kurtosis_type_constructor.KurtosisValueTypeDefault
}

func (config *IngressTLSConfig) Copy() (builtin_argument.KurtosisValueType, error) {
	copiedValueType, err := config.KurtosisValueTypeDefault.Copy()
	if err != nil {
		return nil, err
	}
	return &IngressTLSConfig{
		KurtosisValueTypeDefault: copiedValueType,
	}, nil
}

func (config *IngressTLSConfig) GetSecretName() (string, *startosis_errors.InterpretationError) {
	secretName, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](
		config.KurtosisValueTypeDefault, SecretNameAttr)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	if !found {
		return "", startosis_errors.NewInterpretationError("Required attribute '%v' couldn't be found on '%v' type",
			SecretNameAttr, IngressTLSConfigTypeName)
	}
	return secretName.GoString(), nil
}
