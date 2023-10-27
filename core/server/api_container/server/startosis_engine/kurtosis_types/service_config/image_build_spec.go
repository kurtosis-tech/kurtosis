package service_config

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_build_spec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
	"path"
)

const (
	ImageBuildSpecTypeName = "ImageBuildSpec"

	ContextDirAttr  = "context_dir"
	TargetStageAttr = "target_stage"

	defaultContainerImageFileName = "Dockerfile"
)

func NewImageBuildSpecType() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return &kurtosis_type_constructor.KurtosisTypeConstructor{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: ImageBuildSpecTypeName,
			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              ContextDirAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					// TODO: add a validator
					Validator: nil,
				},
				{
					Name:              TargetStageAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					// TODO: add a validator
					Validator: nil,
				},
			},
		},
		Instantiate: instantiateImageBuildSpec,
	}
}

func instantiateImageBuildSpec(arguments *builtin_argument.ArgumentValuesSet) (builtin_argument.KurtosisValueType, *startosis_errors.InterpretationError) {
	kurtosisValueType, err := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(ImageBuildSpecTypeName, arguments)
	if err != nil {
		return nil, err
	}
	return &ImageBuildSpec{
		KurtosisValueTypeDefault: kurtosisValueType,
	}, nil
}

// ReadyCondition is a starlark.Value that holds all the information needed for ensuring service readiness
type ImageBuildSpec struct {
	*kurtosis_type_constructor.KurtosisValueTypeDefault
}

func (imageBuildSpec *ImageBuildSpec) Copy() (builtin_argument.KurtosisValueType, error) {
	copiedValueType, err := imageBuildSpec.KurtosisValueTypeDefault.Copy()
	if err != nil {
		return nil, err
	}
	return &ImageBuildSpec{
		KurtosisValueTypeDefault: copiedValueType,
	}, nil
}

// Relative locator of context directory
func (imageBuildSpec *ImageBuildSpec) GetContextDir() (string, *startosis_errors.InterpretationError) {
	contextDir, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](imageBuildSpec.KurtosisValueTypeDefault, ContextDirAttr)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	if !found {
		return "", startosis_errors.NewInterpretationError("Required attribute '%s' could not be found on type '%s'",
			ContextDirAttr, ImageBuildSpecTypeName)
	}
	contextDirStr := contextDir.GoString()
	return contextDirStr, nil
}

func (imageBuildSpec *ImageBuildSpec) GetTargetStage() (string, *startosis_errors.InterpretationError) {
	targetStage, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](imageBuildSpec.KurtosisValueTypeDefault, TargetStageAttr)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	if !found {
		return "", nil
	}
	targetStageStr := targetStage.String()
	return targetStageStr, nil
}

func (imageBuildSpec *ImageBuildSpec) ToKurtosisType(contextDirAbsFilePath string) (*image_build_spec.ImageBuildSpec, *startosis_errors.InterpretationError) {
	targetStageStr, interpretationErr := imageBuildSpec.GetTargetStage()
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	containerImageFilePath := path.Join(contextDirAbsFilePath, defaultContainerImageFileName)
	return image_build_spec.NewImageBuildSpec(contextDirAbsFilePath, containerImageFilePath, targetStageStr), nil
}
