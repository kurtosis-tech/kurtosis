//package kubernetes
//
//import (
//	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
//	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
//	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
//	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
//	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
//	"go.starlark.net/starlark"
//)
//
//const (
//	IngressPathConfigTypeName = "IngressPathConfig"
//
//	BackendAttr  = "backend"
//	PathAttr     = "path"
//	PathTypeAttr = "path_type"
//)
//
//func NewIngressPathConfigType() *kurtosis_type_constructor.KurtosisTypeConstructor {
//	return &kurtosis_type_constructor.KurtosisTypeConstructor{
//		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
//			Name: IngressPathConfigTypeName,
//			Arguments: []*builtin_argument.BuiltinArgument{
//				{
//					Name:              BackendAttr,
//					IsOptional:        false,
//					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Value],
//					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
//						if _, ok := value.(*IngressBackendConfig); !ok {
//							return startosis_errors.NewInterpretationError("Expected '%s' to be an IngressBackendConfig", BackendAttr)
//						}
//						return nil
//					},
//				},
//				{
//					Name:              PathAttr,
//					IsOptional:        false,
//					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
//					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
//						return builtin_argument.NonEmptyString(value, PathAttr)
//					},
//				},
//				{
//					Name:              PathTypeAttr,
//					IsOptional:        false,
//					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
//					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
//						return builtin_argument.NonEmptyString(value, PathTypeAttr)
//					},
//				},
//			},
//			Deprecation: nil,
//		},
//		Instantiate: instantiateIngressPathConfig,
//	}
//}
//
//func instantiateIngressPathConfig(arguments *builtin_argument.ArgumentValuesSet) (builtin_argument.KurtosisValueType, *startosis_errors.InterpretationError) {
//	kurtosisValueType, interpretationErr := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(IngressPathConfigTypeName, arguments)
//	if interpretationErr != nil {
//		return nil, interpretationErr
//	}
//	return &IngressPathConfig{
//		KurtosisValueTypeDefault: kurtosisValueType,
//	}, nil
//}
//
//type IngressPathConfig struct {
//	*kurtosis_type_constructor.KurtosisValueTypeDefault
//}
//
//func (config *IngressPathConfig) Copy() (builtin_argument.KurtosisValueType, error) {
//	copiedValueType, err := config.KurtosisValueTypeDefault.Copy()
//	if err != nil {
//		return nil, err
//	}
//	return &IngressPathConfig{
//		KurtosisValueTypeDefault: copiedValueType,
//	}, nil
//}
//
//func (config *IngressPathConfig) GetBackend() (*IngressBackendConfig, *startosis_errors.InterpretationError) {
//	backend, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*IngressBackendConfig](
//		config.KurtosisValueTypeDefault, BackendAttr)
//	if interpretationErr != nil {
//		return nil, interpretationErr
//	}
//	if !found {
//		return nil, startosis_errors.NewInterpretationError("Required attribute '%v' couldn't be found on '%v' type",
//			BackendAttr, IngressPathConfigTypeName)
//	}
//	return backend, nil
//}
//
//func (config *IngressPathConfig) GetPath() (string, *startosis_errors.InterpretationError) {
//	path, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](
//		config.KurtosisValueTypeDefault, PathAttr)
//	if interpretationErr != nil {
//		return "", interpretationErr
//	}
//	if !found {
//		return "", startosis_errors.NewInterpretationError("Required attribute '%v' couldn't be found on '%v' type",
//			PathAttr, IngressPathConfigTypeName)
//	}
//	return path.GoString(), nil
//}
//
//func (config *IngressPathConfig) GetPathType() (string, *startosis_errors.InterpretationError) {
//	pathType, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](
//		config.KurtosisValueTypeDefault, PathTypeAttr)
//	if interpretationErr != nil {
//		return "", interpretationErr
//	}
//	if !found {
//		return "", startosis_errors.NewInterpretationError("Required attribute '%v' couldn't be found on '%v' type",
//			PathTypeAttr, IngressPathConfigTypeName)
//	}
//	return pathType.GoString(), nil
//}
//
//func (config *IngressPathConfig) ToKurtosisType() (*service.IngressPathConfig, *startosis_errors.InterpretationError) {
//	backend, interpretationErr := config.GetBackend()
//	if interpretationErr != nil {
//		return nil, interpretationErr
//	}
//
//	kurtosisBackend, interpretationErr := backend.ToKurtosisType()
//	if interpretationErr != nil {
//		return nil, startosis_errors.WrapWithInterpretationError(interpretationErr, "An error occurred converting backend to Kurtosis type")
//	}
//
//	path, interpretationErr := config.GetPath()
//	if interpretationErr != nil {
//		return nil, interpretationErr
//	}
//
//	pathType, interpretationErr := config.GetPathType()
//	if interpretationErr != nil {
//		return nil, interpretationErr
//	}
//
//	return &service.IngressPathConfig{
//		Backend:  kurtosisBackend,
//		Path:     path,
//		PathType: pathType,
//	}, nil
//}
