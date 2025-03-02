package kubernetes

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
)

const (
	IngressTargetTypeName = "IngressTarget"
	TargetAttr            = "targetPort"
	PrefixPathAttr        = "prefixPath"
	PathTypeAttr          = "pathType"
	AnnotationsAttr       = "annotations"
)

type IngressTarget struct {
	*kurtosis_type_constructor.KurtosisValueTypeDefault
}

func NewIngressTargetType() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return &kurtosis_type_constructor.KurtosisTypeConstructor{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: IngressTargetTypeName,
			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              TargetAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						// TODO: Use k8s validator if exists?
						return builtin_argument.NonEmptyString(value, TargetAttr)
					},
				},
				{
					Name:              PrefixPathAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						// TODO: Use k8s validator if exists?
						return builtin_argument.NonEmptyString(value, PrefixPathAttr)
					},
				},
				{
					Name:              PathTypeAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						str, ok := value.(starlark.String)
						if !ok {
							return startosis_errors.NewInterpretationError("Expected '%s' to be a string", PathTypeAttr)
						}
						pathType := str.GoString()
						if pathType != "Prefix" && pathType != "Exact" && pathType != "ImplementationSpecific" {
							return startosis_errors.NewInterpretationError("PathType must be one of: Prefix, Exact, ImplementationSpecific")
						}
						return nil
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
			},
			Deprecation: nil,
		},
		Instantiate: instantiateIngressTarget,
	}
}

func instantiateIngressTarget(arguments *builtin_argument.ArgumentValuesSet) (builtin_argument.KurtosisValueType, *startosis_errors.InterpretationError) {
	kurtosisValueType, interpretationErr := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(IngressTargetTypeName, arguments)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return &IngressTarget{
		KurtosisValueTypeDefault: kurtosisValueType,
	}, nil
}

func (target *IngressTarget) Copy() (builtin_argument.KurtosisValueType, error) {
	copiedDefault, err := target.KurtosisValueTypeDefault.Copy()
	if err != nil {
		return nil, err
	}
	return &IngressTarget{
		KurtosisValueTypeDefault: copiedDefault,
	}, nil
}

func (target *IngressTarget) GetTarget() (string, *startosis_errors.InterpretationError) {
	targetStr, _, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](
		target.KurtosisValueTypeDefault, TargetAttr)

	if interpretationErr != nil {
		return "", interpretationErr
	}
	return targetStr.GoString(), nil
}

func (target *IngressTarget) GetPrefixPath() (string, *startosis_errors.InterpretationError) {
	prefixPath, _, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](
		target.KurtosisValueTypeDefault, PrefixPathAttr)

	if interpretationErr != nil {
		return "", interpretationErr
	}
	return prefixPath.GoString(), nil
}

func (target *IngressTarget) GetPathType() (string, *startosis_errors.InterpretationError) {
	pathType, _, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](
		target.KurtosisValueTypeDefault, PathTypeAttr)

	if interpretationErr != nil {
		return "", interpretationErr
	}
	return pathType.GoString(), nil
}

func (target *IngressTarget) GetAnnotations() (map[string]string, *startosis_errors.InterpretationError) {
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

	return dict, nil
}

func (target *IngressTarget) ToKurtosisType() (*service.IngressTarget, *startosis_errors.InterpretationError) {
	targetPort, err := target.GetTarget()
	if err != nil {
		return nil, err
	}

	prefixPath, err := target.GetPrefixPath()
	if err != nil {
		return nil, err
	}

	pathType, err := target.GetPathType()
	if err != nil {
		return nil, err
	}

	annotations, err := target.GetAnnotations()
	if err != nil {
		return nil, err
	}

	result := &service.IngressTarget{
		Annotations: annotations,
		PathType:    pathType,
		PrefixPath:  prefixPath,
		Target:      targetPort,
	}

	return result, nil
}
