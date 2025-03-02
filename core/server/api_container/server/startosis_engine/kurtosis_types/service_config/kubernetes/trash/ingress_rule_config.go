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
	IngressRuleConfigTypeName = "IngressRuleConfig"

	HostAttr  = "host"
	PathsAttr = "paths"
)

func NewIngressRuleConfigType() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return &kurtosis_type_constructor.KurtosisTypeConstructor{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: IngressRuleConfigTypeName,
			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              HostAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, HostAttr)
					},
				},
				{
					Name:              PathsAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.List],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						if _, ok := value.(*starlark.List); !ok {
							return startosis_errors.NewInterpretationError("Expected '%s' to be a list of IngressPathConfig", PathsAttr)
						}
						return nil
					},
				},
			},
			Deprecation: nil,
		},
		Instantiate: instantiateIngressRuleConfig,
	}
}

func instantiateIngressRuleConfig(arguments *builtin_argument.ArgumentValuesSet) (builtin_argument.KurtosisValueType, *startosis_errors.InterpretationError) {
	kurtosisValueType, interpretationErr := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(IngressRuleConfigTypeName, arguments)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return &IngressRuleConfig{
		KurtosisValueTypeDefault: kurtosisValueType,
	}, nil
}

type IngressRuleConfig struct {
	*kurtosis_type_constructor.KurtosisValueTypeDefault
}

func (config *IngressRuleConfig) Copy() (builtin_argument.KurtosisValueType, error) {
	copiedValueType, err := config.KurtosisValueTypeDefault.Copy()
	if err != nil {
		return nil, err
	}
	return &IngressRuleConfig{
		KurtosisValueTypeDefault: copiedValueType,
	}, nil
}

func (config *IngressRuleConfig) GetHost() (string, *startosis_errors.InterpretationError) {
	host, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](
		config.KurtosisValueTypeDefault, HostAttr)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	if !found {
		return "", startosis_errors.NewInterpretationError("Required attribute '%v' couldn't be found on '%v' type",
			HostAttr, IngressRuleConfigTypeName)
	}
	return host.GoString(), nil
}

func (config *IngressRuleConfig) GetPaths() ([]*IngressPathConfig, *startosis_errors.InterpretationError) {
	pathsList, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*starlark.List](
		config.KurtosisValueTypeDefault, PathsAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if !found {
		return nil, startosis_errors.NewInterpretationError("Required attribute '%v' couldn't be found on '%v' type",
			PathsAttr, IngressRuleConfigTypeName)
	}

	paths := make([]*IngressPathConfig, 0, pathsList.Len())
	iter := pathsList.Iterate()
	defer iter.Done()
	var item starlark.Value
	for iter.Next(&item) {
		pathConfig, ok := item.(*IngressPathConfig)
		if !ok {
			return nil, startosis_errors.NewInterpretationError("Expected path item to be an IngressPathConfig")
		}
		paths = append(paths, pathConfig)
	}
	return paths, nil
}

func (config *IngressRuleConfig) ToKurtosisType() (*service.IngressRuleConfig, *startosis_errors.InterpretationError) {
	host, interpretationErr := config.GetHost()
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	paths, interpretationErr := config.GetPaths()
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	kurtosisPaths := make([]*service.IngressPathConfig, 0, len(paths))
	for _, path := range paths {
		kurtosisPath, interpretationErr := path.ToKurtosisType()
		if interpretationErr != nil {
			return nil, startosis_errors.WrapWithInterpretationError(interpretationErr, "An error occurred converting path to Kurtosis type")
		}
		kurtosisPaths = append(kurtosisPaths, kurtosisPath)
	}

	return &service.IngressRuleConfig{
		Host:  host,
		Paths: kurtosisPaths,
	}, nil
}
