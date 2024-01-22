package service_config

import (
	"path"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_load"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_constants"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"go.starlark.net/starlark"
)

const (
	ImageLoadTypeName = "ImageLoad"
	LoadImageAttr     = "image_file_path"
)

func NewImageLoadType() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return &kurtosis_type_constructor.KurtosisTypeConstructor{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: ImageLoadTypeName,
			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              LoadImageAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, LoadImageAttr)
					},
				},
				{
					Name:              TargetStageAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, TargetStageAttr)
					},
				},
			},
		},
		Instantiate: instantiateImageLoad,
	}
}

func instantiateImageLoad(arguments *builtin_argument.ArgumentValuesSet) (builtin_argument.KurtosisValueType, *startosis_errors.InterpretationError) {
	kurtosisValueType, err := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(ImageLoadTypeName, arguments)
	if err != nil {
		return nil, err
	}
	return &ImageLoad{
		KurtosisValueTypeDefault: kurtosisValueType,
	}, nil
}

// ImageLoad is a starlark.Value that holds all the information for the startosis_engine to initiate an image Load
type ImageLoad struct {
	*kurtosis_type_constructor.KurtosisValueTypeDefault
}

func (imageLoad *ImageLoad) Copy() (builtin_argument.KurtosisValueType, error) {
	copiedValueType, err := imageLoad.KurtosisValueTypeDefault.Copy()
	if err != nil {
		return nil, err
	}
	return &ImageLoad{
		KurtosisValueTypeDefault: copiedValueType,
	}, nil
}

// Name to give image built from ImageLoad
func (imageLoad *ImageLoad) GetImagePathOnDisk() (string, *startosis_errors.InterpretationError) {
	imageName, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](imageLoad.KurtosisValueTypeDefault, LoadImageAttr)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	if !found {
		return "", startosis_errors.NewInterpretationError("Required attribute '%s' could not be found on type '%s'",
			LoadImageAttr, ImageLoadTypeName)
	}
	imageNameStr := imageName.GoString()
	return imageNameStr, nil
}

func (imageLoad *ImageLoad) ToKurtosisType(
	locatorOfModuleInWhichThisBuiltInIsBeingCalled string,
	packageId string,
	packageContentProvider startosis_packages.PackageContentProvider,
	packageReplaceOptions map[string]string) (*image_load.ImageLoad, *startosis_errors.InterpretationError) {
	// get locator of context directory (relative or absolute)
	imagePathOnDisk, interpretationErr := imageLoad.GetImagePathOnDisk()
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	containerImageFilePathOnDisk, interpretationErr := getOnDiskImageLoadPath(
		imagePathOnDisk,
		packageId,
		locatorOfModuleInWhichThisBuiltInIsBeingCalled,
		packageContentProvider,
		packageReplaceOptions)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	return image_load.NewImageLoad(containerImageFilePathOnDisk), nil
}

// Returns the filepath of the Load context directory and container image on APIC based on package info
func getOnDiskImageLoadPath(
	imageFilePathOnDisk string,
	packageId string,
	locatorOfModuleInWhichThisBuiltInIsBeingCalled string,
	packageContentProvider startosis_packages.PackageContentProvider,
	packageReplaceOptions map[string]string) (string, *startosis_errors.InterpretationError) {
	if packageId == startosis_constants.PackageIdPlaceholderForStandaloneScript {
		return "", startosis_errors.NewInterpretationError("Cannot use ImageLoad in a standalone script; create a package and rerun to use ImageLoad.")
	}

	// get absolute locator of context directory
	contextDirAbsoluteLocator, interpretationErr := packageContentProvider.GetAbsoluteLocator(packageId, locatorOfModuleInWhichThisBuiltInIsBeingCalled, imageFilePathOnDisk, packageReplaceOptions)
	if interpretationErr != nil {
		return "", interpretationErr
	}

	// get on disk directory path of Dockerfile
	containerImageAbsoluteLocator := path.Join(contextDirAbsoluteLocator, defaultContainerImageFileName)
	containerImagePathOnDisk, interpretationErr := packageContentProvider.GetOnDiskAbsoluteFilePath(containerImageAbsoluteLocator)
	if interpretationErr != nil {
		return "", interpretationErr
	}

	return containerImagePathOnDisk, nil
}
