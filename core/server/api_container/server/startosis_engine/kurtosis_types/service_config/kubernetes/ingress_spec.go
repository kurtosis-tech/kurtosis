package kubernetes

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
)

const (
	IngressSpecTypeName = "IngressSpec"

	AnnotationsAttr      = "annotations"
	IngressClassNameAttr = "ingressClassName"
	HostAttr             = "host"
	IngressTlsAttr       = "tls"
	IngressHttpRuleAttr  = "httpRules"
)

type IngressSpec struct {
	*kurtosis_type_constructor.KurtosisValueTypeDefault
}

func NewIngressTargetType() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return &kurtosis_type_constructor.KurtosisTypeConstructor{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: IngressSpecTypeName,
			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              HostAttr,
					IsOptional:        true,
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
						return nil
					},
				},
				{
					Name:              TLSConfigAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*IngressTLSConfig],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						if _, ok := value.(*IngressTLSConfig); !ok {
							return startosis_errors.NewInterpretationError("Expected %s to be of type IngressTlsConfig", TLSConfigAttr)
						}
						return nil
					},
				},
				{
					Name:              IngressHttpRuleAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.List],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						if _, ok := value.(*IngressHttpRule); !ok {
							return startosis_errors.NewInterpretationError("Error expected %s to be of type IngressHttpRule", IngressHttpRuleAttr)
						}
						return nil
					},
				},
			},
			Deprecation: nil,
		},
		Instantiate: instantiateIngressTarget,
	}
}

func instantiateIngressTarget(arguments *builtin_argument.ArgumentValuesSet) (builtin_argument.KurtosisValueType, *startosis_errors.InterpretationError) {
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

func (target *IngressSpec) GetTlsConfig() (*KtTlsConfig, *startosis_errors.InterpretationError) {
	tls, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*IngressTLSConfig](
		target.KurtosisValueTypeDefault, IngressTlsAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if !found {
		return nil, nil
	}

	ktTls, err := tls.ToKurtosisType()
	if err != nil {
		return nil, err
	}
	return ktTls, nil
}

func (target *IngressSpec) GetAnnotations() (*KtAnnotations, *startosis_errors.InterpretationError) {
	annotations, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*starlark.Dict](
		target.KurtosisValueTypeDefault, AnnotationsAttr)

	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if !found {
		return nil, nil
	}
	dict, err := kurtosis_types.SafeCastToMapStringString(annotations, "ingressTargetAnnotations")
	if err != nil {
		return nil, err
	}
	return &dict, nil
}

func (target *IngressSpec) handleStringPtrExtraction(attrName string) (*string, error) {
	value, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*starlark.String](
		target.KurtosisValueTypeDefault, attrName)

	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if !found || value == nil {
		return nil, nil
	}
	goValue := value.GoString()
	return &goValue, nil
}

func (target *IngressSpec) GetHost() (*string, error) {
	return target.handleStringPtrExtraction(HostAttr)
}

func (target *IngressSpec) GetIngressClassName() (*string, error) {
	return target.handleStringPtrExtraction(IngressClassNameAttr)
}

func (target *IngressSpec) GetRules() ([]*KtHttpRule, error) {
	ruleList, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*starlark.List](
		target.KurtosisValueTypeDefault, IngressHttpRuleAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if !found {
		return nil, nil
	}

	var rules []*KtHttpRule
	for idx := 0; idx < ruleList.Len(); idx++ {
		item := ruleList.Index(idx)
		rule, ok := item.(*IngressHttpRule)
		if !ok {
			return nil, startosis_errors.NewInterpretationError(
				"Item number %d in '%s' list was not of type IngressHttpRule. Expecting '%s' to be a %s",
				idx, IngressHttpRuleAttr, ruleList.Type(),
			)
		}
		r, err := rule.ToKurtosisType()
		if err != nil {
			return nil, err
		}
		rules = append(rules, r)
	}

	return rules, nil
}

func (target *IngressSpec) ToKurtosisType() (*KtIngressSpec, *startosis_errors.InterpretationError) {
	handleError := func(err error, attr string) *startosis_errors.InterpretationError {
		return startosis_errors.WrapWithInterpretationError(
			err, "Error interpreting %s", attr,
		)
	}

	ingressClassName, err := target.GetIngressClassName()
	if err != nil {
		return nil, handleError(err, IngressClassNameAttr)
	}

	host, err := target.handleStringPtrExtraction(HostAttr)
	if err != nil {
		return nil, handleError(err, HostAttr)
	}

	annotations, err := target.GetAnnotations()
	if err != nil {
		return nil, handleError(err, AnnotationsAttr)
	}

	tlsConfig, err := target.GetTlsConfig()
	if err != nil {
		return nil, handleError(err, TLSConfigAttr)
	}

	rules, err := target.GetRules()
	if err != nil {
		return nil, handleError(err, IngressHttpRuleAttr)
	}

	result := &KtIngressSpec{
		Annotations:      annotations,
		Host:             host,
		TlsConfig:        tlsConfig,
		HttpRules:        rules,
		IngressClassName: ingressClassName,
	}

	return result, nil
}
