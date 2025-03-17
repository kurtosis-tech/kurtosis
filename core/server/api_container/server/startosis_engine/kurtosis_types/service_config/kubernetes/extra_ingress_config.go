package kubernetes

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/kubernetes"
	//"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/kubernetes"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
)

const (
	ExtraIngressConfigTypeName = "ExtraIngressConfig"
	ExtraIngressConfigAttr     = "extraIngressConfig"
	IngressesAttr              = "ingresses"
)

type ExtraIngressConfig struct {
	*kurtosis_type_constructor.KurtosisValueTypeDefault
}

func NewExtraIngressConfigType() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return &kurtosis_type_constructor.KurtosisTypeConstructor{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: ExtraIngressConfigTypeName,
			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              IngressesAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.List],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						if _, ok := value.(*starlark.List); !ok {
							return startosis_errors.NewInterpretationError("Expected '%s' to be a list, found %s", IngressesAttr, value.String())
						}
						return nil
					},
				},
			},
			Deprecation: nil,
		},
		Instantiate: instantiateExtraIngressConfig,
	}
}

func instantiateExtraIngressConfig(arguments *builtin_argument.ArgumentValuesSet) (builtin_argument.KurtosisValueType, *startosis_errors.InterpretationError) {
	kurtosisValueType, interpretationErr := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(ExtraIngressConfigTypeName, arguments)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return &ExtraIngressConfig{
		KurtosisValueTypeDefault: kurtosisValueType,
	}, nil
}

func (extraIngressConfig *ExtraIngressConfig) Copy() (builtin_argument.KurtosisValueType, error) {
	copiedDefault, err := extraIngressConfig.KurtosisValueTypeDefault.Copy()
	if err != nil {
		return nil, err
	}
	return &ExtraIngressConfig{
		KurtosisValueTypeDefault: copiedDefault,
	}, nil
}

func (extraIngressConfig *ExtraIngressConfig) Validate() error { return nil }

func (extraIngressConfig *ExtraIngressConfig) GetIngresses() ([]*kubernetes.IngressSpec, error) {
	ingressTargetsList, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*starlark.List](
		extraIngressConfig.KurtosisValueTypeDefault, IngressesAttr,
	)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if !found {
		return nil, nil
	}

	var ingressTargets []*kubernetes.IngressSpec
	for idx := 0; idx < ingressTargetsList.Len(); idx++ {
		item := ingressTargetsList.Index(idx)
		ingressTarget, ok := item.(*IngressSpec)
		if !ok {
			return nil, startosis_errors.NewInterpretationError(
				"Item number %d in '%s' list was not a string. Expecting '%s' to be a %s",
				idx, IngressesAttr, ingressTarget.Type(),
			)
		}
		kit, err := ingressTarget.ToKurtosisType()
		if err != nil {
			return nil, err
		}
		ingressTargets = append(ingressTargets, kit)
	}

	return ingressTargets, nil
}
