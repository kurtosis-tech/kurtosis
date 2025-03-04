package kubernetes

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
)

const (
	ExtraIngressConfigTypeName = "ExtraIngressConfig"
	ExtraIngressConfigAttr     = "extraIngressConfig"
	//IngressTargetsTypeName = "ingressTargets"
	//IngressClassConfigName     = "ingress_class_name"
	//HostConfigsAttr            = "host_configs"
	IngressesAttr = "ingresses"
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

func (extraIngressConfig *ExtraIngressConfig) GetIngresses() ([]*KtIngressSpec, error) {
	ingressTargetsList, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*starlark.List](
		extraIngressConfig.KurtosisValueTypeDefault, Ingresses,
	)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if !found {
		return nil, nil
	}

	var ingressTargets []*KtIngressSpec
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

//func (extraIngressConfig *ExtraIngressConfig) GetStarlarkMultiIngressClassConfigs() (*MultiIngressClassConfigs, *startosis_errors.InterpretationError) {
//	multiIngressClassConfigs, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*MultiIngressClassConfigs](
//		extraIngressConfig.KurtosisValueTypeDefault, MultiIngressClassConfigsTypeName)
//
//	if interpretationErr != nil {
//		return nil, interpretationErr
//	}
//
//	if found && multiIngressClassConfigs == nil {
//		logrus.Debug("Ingress class configs found but were nil and without error. This should never happen.")
//		return nil, nil
//	}
//
//	return multiIngressClassConfigs, nil
//}
//
//// Apparently this can't be done inline, you have to declare a method to wrap the
//// type conversion, wtf is wrong with this god forsaken language
//func castMiccToDict(ev starlark.Value) (*starlark.Dict, error) {
//	dict, ok := ev.(*starlark.Dict)
//	if !ok {
//		return nil, fmt.Errorf("Unexpected type, expected a dict")
//	}
//	return dict, nil
//}
//
//// func (extraIngressConfig *ExtraIngressConfig) GetMultiIngressClassConfigs() (*KtMultiClassConfig, *startosis_errors.InterpretationError) {
//func (extraIngressConfig *ExtraIngressConfig) GetMultiIngressClassConfigs() (*KtMultiClassConfig, error) {
//	micc, interpretationErr := extraIngressConfig.GetStarlarkMultiIngressClassConfigs()
//
//	//buildArgsStarlark, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*starlark.Dict](imageBuildSpec.KurtosisValueTypeDefault, BuildArgsAttr)
//	if interpretationErr != nil {
//		return nil, interpretationErr
//	}
//
//	dictValue, err := castMiccToDict(micc)
//	if err != nil {
//		return nil, startosis_errors.NewInterpretationError("'%s' is not a dict", MultiIngressClassConfigsTypeName)
//	}
//
//	kmcc := KtMultiClassConfig{}
//	for _, key := range dictValue.Keys() {
//		v, found, err := dictValue.Get(key)
//		if !found {
//			return nil, fmt.Errorf("key '%v' not found in dictionary", key)
//		}
//		if err != nil {
//			return nil, fmt.Errorf("error reading ingress class config for %s", key)
//		}
//		slValues, ok := v.(IngressClassConfig)
//		if !ok {
//			return nil, fmt.Errorf("error casting ingress class %s", key)
//		}
//		classConfig, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*IngressClassConfig](
//			slValues.KurtosisValueTypeDefault, IngressConfigTypeName)
//		if interpretationErr != nil {
//			return nil, interpretationErr
//		}
//
//		stringKey, err := kurtosis_types.SafeCastToString(key, "Ingress class name")
//		kmcc[stringKey] = classConfig.ToKurtosisType()
//	}
//	return &kmcc, nil
//}
//
//
//type MultiIngressClassConfigs struct {
//	*kurtosis_type_constructor.KurtosisValueTypeDefault
//}

//type KtIngressConfig struct {
//	Target     string
//	PrefixPath string
//	Type       string
//}
//
//type KtTlsConfig struct {
//	secretName string
//}
//
//type KtHostConfig struct {
//	Host      string
//	TlsConfig *KtTlsConfig
//	Ingresses []KtIngressConfig
//}
//
//type KtIngressClassConfig struct {
//	KtIngressClassName string
//	KtIngressConfigs   []KtIngressConfig
//}
//
//// KtMultiClassConfig mapping of Ingress class names onto config for that ingress
//type KtMultiClassConfig map[string]KtIngressClassConfig
//type KtExtraIngressConfig struct {
//	MultiIngressClassConfigs *KtMultiClassConfig
//}

//func (micc *MultiIngressClassConfigs) GetMultiIngressClassConfigs() *KtMultiClassConfig {
//	extraIngressConfig, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*ExtraIngressConfig](micc.KurtosisValueTypeDefault)
//
//	if interpretationErr != nil {
//		return nil, interpretationErr
//	}
//	if !found || extraIngressConfig == nil {
//		return nil, nil
//	}
//	return extraIngressConfig, nil
//}

//func (micc *MultiIngressClassConfigs) ToKurtosisType() (KtMultiClassConfig, error) {
//	value, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*starlark.Dict](
//		micc.KurtosisValueTypeDefault)
//
//	//value, ok := micc.(*starlark.Dict)
//	value, ok := micc.(*starlark.Dict)
//	if !ok {
//		return nil, error("Error reading MultiIngressClassConfigs as dict")
//	}
//
//	//value
//}
//
//const (
//	MultiIngressClassConfigsTypeName = "MultiIngressClassConfigs"
//)
//
//func NewMultiIngressClassConfig() *kurtosis_type_constructor.KurtosisTypeConstructor {
//	return &kurtosis_type_constructor.KurtosisTypeConstructor{
//		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
//			Name: MultiIngressClassConfigsTypeName,
//			Arguments: []*builtin_argument.BuiltinArgument{
//				{
//					Name:              "MultiIngressClassConfigs",
//					IsOptional:        false,
//					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.Dict],
//					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
//						if _, ok := value.(*starlark.Dict); !ok {
//							return startosis_errors.NewInterpretationError("Expected '%s' to contain a dict of ingress class configurations reading %s", ExtraIngressConfigAttr, value.String())
//						}
//						return nil
//					},
//				},
//			},
//			Deprecation: nil,
//		},
//		Instantiate: instantiateExtraIngressConfig,
//	}
//}
//
////const (
////	IngressClassConfigAttr = ""
////)
//
//const (
//	IngressClassConfigsTypeName = "IngressClassConfig"
//	IngressItemsListFieldAttr   = "hosts"
//)
//
//type IngressClassConfig struct {
//	*kurtosis_type_constructor.KurtosisValueTypeDefault
//}
//
////func (c *IngressClassConfig) GetIngressName() (*KtIngressClassConfig, error) {
//
//func (c *IngressClassConfig) ToKurtosisType() (*KtIngressClassConfig, error) {
//	ingressClassConfig, interpretationError := c.GetStarlarkIngressClassConfig()
//
//	if interpretationError != nil {
//		return nil, interpretationError
//	}
//	if ingressClassConfig == nil {
//		return nil, nil
//	}
//
//	//type KtIngressClassConfig struct {
//	//	KtIngressClassName string
//	//	KtIngressConfigs   []KtIngressConfig
//	//}
//
//	//c.getIngressClassName()
//	ingressClassList, interpretationError := kurtosis_type_constructor.ExtractAttrValue[*starlark.Dict](
//		c.KurtosisValueTypeDefault, IngressItemsListFieldAttr
//		)
//
//	for _, ingressClass
//
//	return &KtIngressClassConfig{}, nil
//}
//
//func (c *IngressClassConfig) GetStarlarkIngressClassConfig() (*IngressClassConfig, *startosis_errors.InterpretationError) {
//	ingressClassConfig, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*IngressClassConfig](c.KurtosisValueTypeDefault, IngressClassConfigsTypeName)
//
//	if interpretationErr != nil {
//		return nil, interpretationErr
//	}
//	if !found || ingressClassConfig == nil {
//		return nil, nil
//	}
//	return ingressClassConfig, nil
//}
//
//func (ic *IngressClassConfig) GetIngressClassConfig() (*IngressClassConfig, error) {
//	ingressClassConfig, err := ic.GetStarlarkIngressClassConfig()
//	if err != nil {
//		return nil, err
//	}
//
//	if interpretationErr != nil {
//		return "", false, startosis_errors.WrapWithInterpretationError(err, "An error occurred getting the ingress class name")
//	}
//	if value == nil || value == starlark.None {
//		return "", false, nil
//	}
//	strValue, ok := value.(starlark.String)
//	if !ok {
//		return "", false, startosis_errors.NewInterpretationError("Expected ingress class name to be a string but was '%v'", value.Type())
//	}
//	return strValue.GoString(), true, nil
//}
//
//func (extraIngressConfig *ExtraIngressConfig) GetHostConfigs() (map[string]*HostConfig, *startosis_errors.InterpretationError) {
//	value, err := extraIngressConfig.GetAttributeValue(HostConfigsAttr)
//	if err != nil {
//		return nil, startosis_errors.WrapWithInterpretationError(err, "An error occurred getting the host configs")
//	}
//	dict, ok := value.(*starlark.Dict)
//	if !ok {
//		return nil, startosis_errors.NewInterpretationError("Expected host configs to be a dict but was '%v'", value.Type())
//	}
//
//	result := make(map[string]*HostConfig)
//	for _, item := range dict.Items() {
//		key, ok := item[0].(starlark.String)
//		if !ok {
//			return nil, startosis_errors.NewInterpretationError("Expected host extraIngressConfig key to be a string but was '%v'", item[0].Type())
//		}
//		value, ok := item[1].(*HostConfig)
//		if !ok {
//			return nil, startosis_errors.NewInterpretationError("Expected host extraIngressConfig value to be a HostConfig but was '%v'", item[1].Type())
//		}
//		result[key.GoString()] = value
//	}
//	return result, nil
//}
//
//
//func (extraIngressConfig *ExtraIngressConfig) ToKurtosisType() (*KtExtraIngressConfig, error) {
//	micc, err := extraIngressConfig.GetStarlarkMultiIngressClassConfigs()
//	if err != nil {
//		return nil, err
//	}
//
//	if micc == nil {
//		return nil, nil
//	}
//
//	ktMigType, err := micc.ToKurtosisType()
//	if err != nil {
//		return nil, err
//	}
//	return &KtExtraIngressConfig{
//		MultiIngressClassConfigs: ktMigType,
//	}, nil
//}

//
//func (config *ExtraIngressConfig) ToKurtosisType() (*KtExtraIngressConfig, *startosis_errors.InterpretationError) {
//	hostConfigs, err := config.GetHostConfigs()
//	if err != nil {
//		return nil, err
//	}
//
//	if len(hostConfigs) == 0 {
//		return nil, startosis_errors.NewInterpretationError("Extra ingress config must have at least one host config")
//	}
//
//	// Convert to service.IngressConfig
//	ingressClassName, hasIngressClassName, err := config.GetIngressClassName()
//	if err != nil {
//		return nil, err
//	}
//
//	result := &service.IngressConfig{
//		Name:    "extra-ingress", // Fixed name since we only support one
//		Enabled: true,
//	}
//
//	if hasIngressClassName {
//		result.ClassName = ingressClassName
//	}
//
//	// Convert host configs to rules
//	for host, hostConfig := range hostConfigs {
//		if host == "" {
//			return nil, startosis_errors.NewInterpretationError("Host name cannot be empty")
//		}
//
//		targets, err := hostConfig.GetIngressTargets()
//		if err != nil {
//			return nil, err
//		}
//
//		if len(targets) == 0 {
//			return nil, startosis_errors.NewInterpretationError("Host config must have at least one ingress target")
//		}
//
//		rule := service.IngressRuleConfig{
//			Host:  host,
//			Paths: make([]service.IngressPathConfig, 0, len(targets)),
//		}
//
//		// Convert targets to paths
//		for _, target := range targets {
//			targetPort, err := target.GetTarget()
//			if err != nil {
//				return nil, err
//			}
//			if targetPort == "" {
//				return nil, startosis_errors.NewInterpretationError("Target port name cannot be empty")
//			}
//
//			prefixPath, err := target.GetPrefixPath()
//			if err != nil {
//				return nil, err
//			}
//			if prefixPath == "" {
//				return nil, startosis_errors.NewInterpretationError("Prefix path cannot be empty")
//			}
//
//			pathType, hasPathType, err := target.GetPathType()
//			if err != nil {
//				return nil, err
//			}
//
//			annotations, hasAnnotations, err := target.GetAnnotations()
//			if err != nil {
//				return nil, err
//			}
//
//			path := service.IngressPathConfig{
//				Path: prefixPath,
//				Backend: service.IngressBackendConfig{
//					PortName: targetPort,
//				},
//			}
//
//			if hasPathType {
//				path.PathType = pathType
//			}
//
//			if hasAnnotations {
//				if result.Annotations == nil {
//					result.Annotations = make(map[string]string)
//				}
//				for k, v := range annotations {
//					result.Annotations[k] = v
//				}
//			}
//
//			rule.Paths = append(rule.Paths, path)
//		}
//
//		// Handle TLS config
//		tlsConfig, hasTLS, err := hostConfig.GetTLSConfig()
//		if err != nil {
//			return nil, err
//		}
//
//		if hasTLS {
//			secretName, err := tlsConfig.GetSecretName()
//			if err != nil {
//				return nil, err
//			}
//
//			result.TLS = append(result.TLS, service.IngressTLSConfig{
//				Hosts:      []string{host},
//				SecretName: secretName,
//			})
//		}
//
//		result.HttpRules = append(result.HttpRules, rule)
//	}
//
//	return result, nil
//}
