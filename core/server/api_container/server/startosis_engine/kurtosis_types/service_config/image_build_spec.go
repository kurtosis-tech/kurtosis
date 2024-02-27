package service_config

import (
	"path"
	"path/filepath"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_build_spec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_constants"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"go.starlark.net/starlark"
)

const (
	ImageBuildSpecTypeName = "ImageBuildSpec"

	BuiltImageNameAttr = "image_name"
	BuildContextAttr   = "build_context_dir"
	BuildFileAttr      = "build_file"
	TargetStageAttr    = "target_stage"

	defaultContainerImageFileName = "Dockerfile"
)

func NewImageBuildSpecType() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return &kurtosis_type_constructor.KurtosisTypeConstructor{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: ImageBuildSpecTypeName,
			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              BuiltImageNameAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, BuiltImageNameAttr)
					},
				},
				{
					Name:              BuildContextAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, BuildContextAttr)
					},
				},
				{
					Name:              BuildFileAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator:         nil,
				},
				{
					Name:              TargetStageAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator:         nil,
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

// ImageBuildSpec is a starlark.Value that holds all the information for the startosis_engine to initiate an image build
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

// Name to give image built from ImageBuildSpec
func (imageBuildSpec *ImageBuildSpec) GetImageName() (string, *startosis_errors.InterpretationError) {
	imageName, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](imageBuildSpec.KurtosisValueTypeDefault, BuiltImageNameAttr)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	if !found {
		return "", startosis_errors.NewInterpretationError("Required attribute '%s' could not be found on type '%s'",
			BuiltImageNameAttr, ImageBuildSpecTypeName)
	}
	imageNameStr := imageName.GoString()
	return imageNameStr, nil
}

// Relative locator of build context directory
func (imageBuildSpec *ImageBuildSpec) GetBuildContextLocator() (string, *startosis_errors.InterpretationError) {
	buildContextLocator, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](imageBuildSpec.KurtosisValueTypeDefault, BuildContextAttr)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	if !found {
		return "", startosis_errors.NewInterpretationError("Required attribute '%s' could not be found on type '%s'",
			BuildContextAttr, ImageBuildSpecTypeName)
	}
	buildContextLocatorStr := buildContextLocator.GoString()
	return buildContextLocatorStr, nil
}

// Name of the build file (Dockerfile)
func (imageBuildSpec *ImageBuildSpec) GetBuildFile() (string, *startosis_errors.InterpretationError) {
	buildFile, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](imageBuildSpec.KurtosisValueTypeDefault, BuildFileAttr)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	if !found {
		return "", startosis_errors.NewInterpretationError("Required attribute '%s' could not be found on type '%s'",
			BuildFileAttr, ImageBuildSpecTypeName)
	}
	buildFileStr := buildFile.GoString()
	return buildFileStr, nil
}

// GetTargetStage is used for specifying which stage of a multi-stage container image build to execute
// Default value is the empty string for single stage image builds (common case)
// Info on target stage and multi-stag builds for Docker images: https://docs.docker.com/build/building/multi-stage/
func (imageBuildSpec *ImageBuildSpec) GetTargetStage() (string, *startosis_errors.InterpretationError) {
	targetStage, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](imageBuildSpec.KurtosisValueTypeDefault, TargetStageAttr)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	if !found {
		return "", nil
	}
	targetStageStr := targetStage.GoString()
	return targetStageStr, nil
}

func (imageBuildSpec *ImageBuildSpec) ToKurtosisType(
	locatorOfModuleInWhichThisBuiltInIsBeingCalled string,
	packageId string,
	packageContentProvider startosis_packages.PackageContentProvider,
	packageReplaceOptions map[string]string) (*image_build_spec.ImageBuildSpec, *startosis_errors.InterpretationError) {
	// get locator of context directory (relative or absolute)
	buildContextLocator, interpretationErr := imageBuildSpec.GetBuildContextLocator()
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	buildFile, interpretationErr := imageBuildSpec.GetBuildFile()
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	buildContextDirPathOnDisk, containerImageFilePathOnDisk, interpretationErr := getOnDiskImageBuildSpecPaths(
		buildContextLocator,
		buildFile,
		packageId,
		locatorOfModuleInWhichThisBuiltInIsBeingCalled,
		packageContentProvider,
		packageReplaceOptions)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	targetStageStr, interpretationErr := imageBuildSpec.GetTargetStage()
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	return image_build_spec.NewImageBuildSpec(buildContextDirPathOnDisk, containerImageFilePathOnDisk, targetStageStr), nil
}

// Returns the filepath of the build context directory and container image on APIC based on package info
func getOnDiskImageBuildSpecPaths(
	buildContextLocator string,
	buildFile string,
	packageId string,
	locatorOfModuleInWhichThisBuiltInIsBeingCalled string,
	packageContentProvider startosis_packages.PackageContentProvider,
	packageReplaceOptions map[string]string) (string, string, *startosis_errors.InterpretationError) {
	if packageId == startosis_constants.PackageIdPlaceholderForStandaloneScript {
		return "", "", startosis_errors.NewInterpretationError("Cannot use ImageBuildSpec in a standalone script; create a package and rerun to use ImageBuildSpec.")
	}

	// get absolute locator of context directory
	contextDirAbsoluteLocator, interpretationErr := packageContentProvider.GetAbsoluteLocator(packageId, locatorOfModuleInWhichThisBuiltInIsBeingCalled, buildContextLocator, packageReplaceOptions)
	if interpretationErr != nil {
		return "", "", interpretationErr
	}

	// get on disk directory path of Dockerfile
	containerImageAbsoluteLocator := path.Join(contextDirAbsoluteLocator, defaultContainerImageFileName)
	if buildFile != "" {
		containerImageAbsoluteLocator = path.Join(contextDirAbsoluteLocator, buildFile)
	}

	containerImagePathOnDisk, interpretationErr := packageContentProvider.GetOnDiskAbsolutePackageFilePath(containerImageAbsoluteLocator)
	if interpretationErr != nil {
		return "", "", interpretationErr
	}

	// Assume, that container image sits at the same level as context directory to get context dir path on disk
	contextDirPathOnDisk := filepath.Dir(containerImagePathOnDisk)

	return contextDirPathOnDisk, containerImagePathOnDisk, nil
}
