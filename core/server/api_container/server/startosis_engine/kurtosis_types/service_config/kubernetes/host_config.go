package kubernetes

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
)

const (
	HostConfigTypeName = "HostConfig"
	TLSConfigAttr      = "tls_config"
	Ingresses          = "ingress_targets"
)

type HostConfig struct {
	*kurtosis_type_constructor.KurtosisValueTypeDefault
}

func NewHostConfigType() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return &kurtosis_type_constructor.KurtosisTypeConstructor{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: HostConfigTypeName,
			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              TLSConfigAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*TLSConfig],
				},
				{
					Name:              Ingresses,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.List],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						if _, ok := value.(*starlark.List); !ok {
							return startosis_errors.NewInterpretationError("Expected '%s' to be a list of ingress targets", Ingresses)
						}
						return nil
					},
				},
			},
			Deprecation: nil,
		},
		Instantiate: instantiateHostConfig,
	}
}

func instantiateHostConfig(arguments *builtin_argument.ArgumentValuesSet) (builtin_argument.KurtosisValueType, *startosis_errors.InterpretationError) {
	kurtosisValueType, interpretationErr := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(HostConfigTypeName, arguments)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return &HostConfig{
		KurtosisValueTypeDefault: kurtosisValueType,
	}, nil
}

func (config *HostConfig) Copy() (builtin_argument.KurtosisValueType, error) {
	copiedDefault, err := config.KurtosisValueTypeDefault.Copy()
	if err != nil {
		return nil, err
	}
	return &HostConfig{
		KurtosisValueTypeDefault: copiedDefault,
	}, nil
}

func (config *HostConfig) GetTLSConfig() (*TLSConfig, bool, *startosis_errors.InterpretationError) {
	value, err := config.GetAttributeValue(TLSConfigAttr)
	if err != nil {
		return nil, false, startosis_errors.WrapWithInterpretationError(err, "An error occurred getting the TLS config")
	}
	if value == nil || value == starlark.None {
		return nil, false, nil
	}
	tlsConfig, ok := value.(*TLSConfig)
	if !ok {
		return nil, false, startosis_errors.NewInterpretationError("Expected TLS config to be a TLSConfig but was '%v'", value.Type())
	}
	return tlsConfig, true, nil
}

func (config *HostConfig) GetIngressTargets() ([]*IngressSpec, *startosis_errors.InterpretationError) {
	value, err := config.GetAttributeValue(Ingresses)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "An error occurred getting the ingress targets")
	}
	list, ok := value.(*starlark.List)
	if !ok {
		return nil, startosis_errors.NewInterpretationError("Expected ingress targets to be a list but was '%v'", value.Type())
	}

	var result []*IngressSpec
	for i := 0; i < list.Len(); i++ {
		item := list.Index(i)
		target, ok := item.(*IngressSpec)
		if !ok {
			return nil, startosis_errors.NewInterpretationError("Expected ingress target to be an IngressSpec but was '%v'", item.Type())
		}
		result = append(result, target)
	}

	// Validate no path conflicts
	seenPaths := make(map[string]bool)
	for _, target := range result {
		path, err := target.GetPrefixPath()
		if err != nil {
			return nil, err
		}
		if seenPaths[path] {
			return nil, startosis_errors.NewInterpretationError("Duplicate path '%s' found in ingress targets", path)
		}
		seenPaths[path] = true
	}

	return result, nil
}

func (config *HostConfig) ToKurtosisType() (*service.HostConfig, *startosis_errors.InterpretationError) {
	tlsConfig, hasTLSConfig, err := config.GetTLSConfig()
	if err != nil {
		return nil, err
	}

	targets, err := config.GetIngressTargets()
	if err != nil {
		return nil, err
	}

	var convertedTLSConfig *service.TLSConfig
	if hasTLSConfig {
		convertedTLSConfig, err = tlsConfig.ToKurtosisType()
		if err != nil {
			return nil, err
		}
	}

	var convertedTargets []*service.IngressTarget
	for _, target := range targets {
		converted, err := target.ToKurtosisType()
		if err != nil {
			return nil, err
		}
		convertedTargets = append(convertedTargets, converted)
	}

	return &service.HostConfig{
		TLSConfig:      convertedTLSConfig,
		IngressTargets: convertedTargets,
	}, nil
}
