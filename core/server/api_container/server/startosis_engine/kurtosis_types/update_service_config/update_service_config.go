package update_service_config

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
)

const (
	UpdateServiceConfigTypeName = "UpdateServiceConfig"
	SubnetworkAttr              = "subnetwork"
)

func NewUpdateServiceConfigType() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return &kurtosis_type_constructor.KurtosisTypeConstructor{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: UpdateServiceConfigTypeName,

			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              SubnetworkAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, SubnetworkAttr)
					},
				},
			},
		},

		Instantiate: instantiate,
	}
}

func instantiate(arguments *builtin_argument.ArgumentValuesSet) (kurtosis_type_constructor.KurtosisValueType, *startosis_errors.InterpretationError) {
	kurtosisValueType, err := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(UpdateServiceConfigTypeName, arguments)
	if err != nil {
		return nil, err
	}
	return &UpdateServiceConfig{
		KurtosisValueTypeDefault: kurtosisValueType,
	}, nil
}

type UpdateServiceConfig struct {
	*kurtosis_type_constructor.KurtosisValueTypeDefault
}

func (config *UpdateServiceConfig) ToKurtosisType() (*kurtosis_core_rpc_api_bindings.UpdateServiceConfig, *startosis_errors.InterpretationError) {
	subnetworkStarlarkValue, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](
		config.KurtosisValueTypeDefault, SubnetworkAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if !found {
		return nil, startosis_errors.NewInterpretationError("Required attribute '%s' could not be found on type '%s'",
			SubnetworkAttr, UpdateServiceConfigTypeName)
	}
	return binding_constructors.NewUpdateServiceConfig(subnetworkStarlarkValue.GoString()), nil
}
