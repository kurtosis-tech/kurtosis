package kubernetes

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_value"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
)

const (
	IngressSpecTypeName  = "IngressSpec"
	AnnotationsAttr      = "annotations"
	IngressClassNameAttr = "ingress_class_name"
	HostAttr             = "host"
	IngressTlsAttr       = "tls"
	IngressHttpRuleAttr  = "http_rules"
)

type IngressSpec struct {
	*kurtosis_type_constructor.KurtosisValueTypeDefault
}

func NewIngressSpecType() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return &kurtosis_type_constructor.KurtosisTypeConstructor{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: IngressSpecTypeName,
			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              HostAttr,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, HostAttr)
					},
				},
				{
					Name:              IngressClassNameAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, IngressClassNameAttr)
					},
				},
				{
					Name:              AnnotationsAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.Dict],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						if _, ok := value.(*starlark.Dict); !ok {
							return startosis_errors.NewInterpretationError("Expected '%s' to be a dict of string annotations", AnnotationsAttr)
						}
						for index, t := range value.(*starlark.Dict).Items() {
							key, ok := t[0].(starlark.String)
							if !ok {
								return startosis_errors.NewInterpretationError(
									"Expected key at index %d of '%s' to be a string",
									index,
									AnnotationsAttr,
								)
							}
							_, err := kubernetes_annotation_key.CreateNewKubernetesAnnotationKey(key.GoString())
							if err != nil {
								return startosis_errors.WrapWithInterpretationError(
									err,
									"Error occurred while creating annotation key '%v'",
									key,
								)
							}

							value, ok := t[1].(starlark.String)
							if !ok {
								return startosis_errors.NewInterpretationError(
									"Expected value at index %d of '%s' to be a string",
									index,
									AnnotationsAttr,
								)
							}
							_, err = kubernetes_annotation_value.CreateNewKubernetesAnnotationValue(value.GoString())
							if err != nil {
								return startosis_errors.WrapWithInterpretationError(
									err,
									"Error occurred while creating annotation value '%v'",
									value,
								)
							}
						}
						return nil
					},
				},
				{
					Name:              IngressTlsAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*IngressTLSConfig],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						if _, ok := value.(*IngressTLSConfig); !ok {
							return startosis_errors.NewInterpretationError("Expected %s to be of type IngressTlsConfig", IngressTlsAttr)
						}
						return nil
					},
				},
				{
					Name:              IngressHttpRuleAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.List],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						if _, ok := value.(*starlark.List); !ok {
							return startosis_errors.NewInterpretationError("Expected '%s' to be a list, found %s", IngressHttpRuleAttr, value.String())
						}
						return nil
					},
				},
			},
			Deprecation: nil,
		},
		Instantiate: instantiateIngressSpec,
	}
}

func instantiateIngressSpec(arguments *builtin_argument.ArgumentValuesSet) (builtin_argument.KurtosisValueType, *startosis_errors.InterpretationError) {
	kurtosisValueType, interpretationErr := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(IngressSpecTypeName, arguments)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return &IngressSpec{
		KurtosisValueTypeDefault: kurtosisValueType,
	}, nil
}

func (target *IngressSpec) Copy() (builtin_argument.KurtosisValueType, error) {
	copiedDefault, err := target.KurtosisValueTypeDefault.Copy()
	if err != nil {
		return nil, err
	}
	return &IngressSpec{
		KurtosisValueTypeDefault: copiedDefault,
	}, nil
}

func (target *IngressSpec) GetTlsConfig() (*IngressTLSConfig, *startosis_errors.InterpretationError) {
	tls, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*IngressTLSConfig](
		target.KurtosisValueTypeDefault, IngressTlsAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if !found {
		return nil, nil
	}

	return tls, nil
}

func (target *IngressSpec) GetAnnotations() (*starlark.Dict, *startosis_errors.InterpretationError) {
	annotations, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*starlark.Dict](
		target.KurtosisValueTypeDefault, AnnotationsAttr)

	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if !found {
		return nil, nil
	}
	return annotations, nil
}

func (target *IngressSpec) handleStringExtraction(attrName string) (string, error) {
	value, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](
		target.KurtosisValueTypeDefault, attrName)

	if interpretationErr != nil {
		return "", interpretationErr
	}
	if !found {
		return "", nil
	}
	return value.GoString(), nil
}

// func (target *IngressSpec) handleStringPtrExtraction(attrName string) (*string, error) {
// 	value, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*starlark.String](
// 		target.KurtosisValueTypeDefault, attrName)

// 	if interpretationErr != nil {
// 		return nil, interpretationErr
// 	}
// 	if !found || value == nil {
// 		return nil, nil
// 	}
// 	goValue := value.GoString()
// 	return &goValue, nil
// }

func (target *IngressSpec) GetHost() (string, error) {
	value, err := target.handleStringExtraction(HostAttr)
	if err != nil {
		return "", err
	}
	return value, err
}

func (target *IngressSpec) GetIngressClassName() (string, error) {
	value, err := target.handleStringExtraction(IngressClassNameAttr)
	if err != nil {
		return "", err
	}
	return value, err
}

func (target *IngressSpec) GetRules() ([]*IngressHttpRule, error) {
	ruleList, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*starlark.List](
		target.KurtosisValueTypeDefault, IngressHttpRuleAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if !found {
		return nil, nil
	}

	var rules []*IngressHttpRule
	for idx := 0; idx < ruleList.Len(); idx++ {
		item := ruleList.Index(idx)
		rule, ok := item.(*IngressHttpRule)
		if !ok {
			return nil, startosis_errors.NewInterpretationError(
				"Item number %d in '%s' list was not of type IngressHttpRule. Expecting to be a %s",
				idx, IngressHttpRuleAttr, ruleList.Type(),
			)
		}
		rules = append(rules, rule)
	}

	return rules, nil
}
