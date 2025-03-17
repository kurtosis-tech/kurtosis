package kubernetes

import (
	//"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/kubernetes"
	//"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/kubernetes"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
)

const (
	KubernetesConfigTypeName       = "KubernetesConfig"
	ExtraIngressConfigAttrTypeName = "extraIngressConfig"
)

func NewKubernetesConfigType() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return &kurtosis_type_constructor.KurtosisTypeConstructor{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: KubernetesConfigTypeName,
			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              ExtraIngressConfigAttrTypeName,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*ExtraIngressConfig],
				},
			},
			Deprecation: nil,
		},
		Instantiate: instantiateKubernetesConfig,
	}
}

func instantiateKubernetesConfig(arguments *builtin_argument.ArgumentValuesSet) (builtin_argument.KurtosisValueType, *startosis_errors.InterpretationError) {
	kurtosisValueType, interpretationErr := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(KubernetesConfigTypeName, arguments)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return &KubernetesConfig{
		KurtosisValueTypeDefault: kurtosisValueType,
	}, nil
}

type KubernetesConfig struct {
	*kurtosis_type_constructor.KurtosisValueTypeDefault
}

func (config *KubernetesConfig) Copy() (builtin_argument.KurtosisValueType, error) {
	copiedValueType, err := config.KurtosisValueTypeDefault.Copy()
	if err != nil {
		return nil, err
	}
	return &KubernetesConfig{
		KurtosisValueTypeDefault: copiedValueType,
	}, nil
}

//func (config *KubernetesConfig) GetStarlarkMultiIngressClassConfigs() (*MultiIngressClassConfigs, *startosis_errors.InterpretationError) {
//	ingressClassConfigs, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*MultiIngressClassConfigs](config.KurtosisValueTypeDefault, ExtraIngressConfigAttr)
//
//	if interpretationErr != nil {
//		return nil, startosis_errors.WrapWithInterpretationError(interpretationErr, "An error occurred getting the multi ingress class configs")
//	}
//
//	if found && ingressClassConfigs == nil {
//		logrus.Debug("Ingress class configs found but were nil and withour error. This should never happen.")
//		return nil, nil
//	}
//
//	return ingressClassConfigs, nil
//}

//multiIngressClassConfig, err := extraIngressConfig.GetStarlarkMultiIngressClassConfigs()
//if err != nil {
//	return nil, err
//}
//mutliIngressClassConfigNative := multiIngressClassConfig.convertTlsConfig()
//
//extraConfig, hasExtraConfig, err := multiIngressClassConfig.Get()
//if err != nil {
//	return nil, err
//}
//
//result := &service.KubernetesConfig{}
//
//if hasIngresses {
//	var convertedIngresses []*service.IngressConfig
//	for _, ingress := range ingresses {
//		converted, err := ingress.convertTlsConfig()
//		if err != nil {
//			return nil, err
//		}
//		convertedIngresses = append(convertedIngresses, converted)
//	}
//	result.Ingresses = convertedIngresses
//}
//
//if hasExtraConfig {
//	converted, err := extraConfig.convertTlsConfig()
//	if err != nil {
//		return nil, err
//	}
//	result.ExtraIngressConfig = converted
//}
//
//return result, nil
//}
