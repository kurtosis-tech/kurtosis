package trash

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
)

const (
	IngressConfigTypeName = "IngressConfig"

	RulesAttr = "rules"
	TLSAttr   = "tls"
)

func NewIngressConfigType() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return &kurtosis_type_constructor.KurtosisTypeConstructor{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: IngressConfigTypeName,
			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              RulesAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.List],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						if _, ok := value.(*starlark.List); !ok {
							return startosis_errors.NewInterpretationError("Expected '%s' to be a list of IngressRuleConfig", RulesAttr)
						}
						return nil
					},
				},
				{
					Name:              TLSAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.List],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						if _, ok := value.(*starlark.List); !ok {
							return startosis_errors.NewInterpretationError("Expected '%s' to be a list of IngressTLSConfig", TLSAttr)
						}
						return nil
					},
				},
			},
			Deprecation: nil,
		},
		Instantiate: instantiateIngressConfig,
	}
}

func instantiateIngressConfig(arguments *builtin_argument.ArgumentValuesSet) (builtin_argument.KurtosisValueType, *startosis_errors.InterpretationError) {
	kurtosisValueType, interpretationErr := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(IngressConfigTypeName, arguments)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return &IngressConfig{
		KurtosisValueTypeDefault: kurtosisValueType,
	}, nil
}

type IngressConfig struct {
	*kurtosis_type_constructor.KurtosisValueTypeDefault
}

func (config *IngressConfig) Copy() (builtin_argument.KurtosisValueType, error) {
	copiedValueType, err := config.KurtosisValueTypeDefault.Copy()
	if err != nil {
		return nil, err
	}
	return &IngressConfig{
		KurtosisValueTypeDefault: copiedValueType,
	}, nil
}

func (config *IngressConfig) GetRules() ([]*IngressRuleConfig, *startosis_errors.InterpretationError) {
	rulesList, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*starlark.List](
		config.KurtosisValueTypeDefault, RulesAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if !found {
		return nil, startosis_errors.NewInterpretationError("Required attribute '%v' couldn't be found on '%v' type",
			RulesAttr, IngressConfigTypeName)
	}

	rules := make([]*IngressRuleConfig, 0, rulesList.Len())
	iter := rulesList.Iterate()
	defer iter.Done()
	var item starlark.Value
	for iter.Next(&item) {
		ruleConfig, ok := item.(*IngressRuleConfig)
		if !ok {
			return nil, startosis_errors.NewInterpretationError("Expected rule item to be an IngressRuleConfig")
		}
		rules = append(rules, ruleConfig)
	}
	return rules, nil
}

func (config *IngressConfig) GetTLS() ([]*IngressTLSConfig, *startosis_errors.InterpretationError) {
	tlsList, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*starlark.List](
		config.KurtosisValueTypeDefault, TLSAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if !found {
		return nil, nil
	}

	tls := make([]*IngressTLSConfig, 0, tlsList.Len())
	iter := tlsList.Iterate()
	defer iter.Done()
	var item starlark.Value
	for iter.Next(&item) {
		tlsConfig, ok := item.(*IngressTLSConfig)
		if !ok {
			return nil, startosis_errors.NewInterpretationError("Expected TLS item to be an IngressTLSConfig")
		}
		tls = append(tls, tlsConfig)
	}
	return tls, nil
}

func (config *IngressConfig) ToKurtosisType() (*service.IngressConfig, *startosis_errors.InterpretationError) {
	rules, interpretationErr := config.GetRules()
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	kurtosisRules := make([]*service.IngressRuleConfig, 0, len(rules))
	for _, rule := range rules {
		kurtosisRule, interpretationErr := rule.ToKurtosisType()
		if interpretationErr != nil {
			return nil, startosis_errors.WrapWithInterpretationError(interpretationErr, "An error occurred converting rule to Kurtosis type")
		}
		kurtosisRules = append(kurtosisRules, kurtosisRule)
	}

	tls, interpretationErr := config.GetTLS()
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	kurtosisTLS := make([]*service.IngressTLSConfig, 0, len(tls))
	for _, tlsConfig := range tls {
		kurtosisTLSConfig, interpretationErr := tlsConfig.ToKurtosisType()
		if interpretationErr != nil {
			return nil, startosis_errors.WrapWithInterpretationError(interpretationErr, "An error occurred converting TLS config to Kurtosis type")
		}
		kurtosisTLS = append(kurtosisTLS, kurtosisTLSConfig)
	}

	return &service.IngressConfig{
		Rules: kurtosisRules,
		TLS:   kurtosisTLS,
	}, nil
}
