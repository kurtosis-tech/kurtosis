package service_config

import (
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/image_registry_spec"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
)

const (
	ImageSpecTypeName = "ImageSpec"

	ImageSpecImageAttr        = "image"
	ImageRegistryAttr         = "registry"
	ImageRegistryUsernameAttr = "username"
	ImageRegistryPasswordAttr = "password"
)

func NewImageSpec() *kurtosis_type_constructor.KurtosisTypeConstructor {

	return &kurtosis_type_constructor.KurtosisTypeConstructor{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: ImageSpecTypeName,
			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              ImageSpecImageAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, ImageSpecImageAttr)
					},
				},
				{
					Name:              ImageRegistryAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, ImageRegistryAttr)
					},
				},
				{
					Name:              ImageRegistryUsernameAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, ImageRegistryUsernameAttr)
					},
				},
				{
					Name:              ImageRegistryPasswordAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, ImageRegistryPasswordAttr)
					},
				},
			},
		},
		Instantiate: instantiateImageRegistrySpec,
	}
}

func instantiateImageRegistrySpec(arguments *builtin_argument.ArgumentValuesSet) (builtin_argument.KurtosisValueType, *startosis_errors.InterpretationError) {
	kurtosisValueType, err := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(ImageSpecTypeName, arguments)
	if err != nil {
		return nil, err
	}
	return &ImageSpec{
		KurtosisValueTypeDefault: kurtosisValueType,
	}, nil
}

// ImageSpec is a starlark.Value that holds all the information to log in to a Docker registry
type ImageSpec struct {
	*kurtosis_type_constructor.KurtosisValueTypeDefault
}

func (irs *ImageSpec) Copy() (builtin_argument.KurtosisValueType, error) {
	copiedValueType, err := irs.KurtosisValueTypeDefault.Copy()
	if err != nil {
		return nil, err
	}
	return &ImageSpec{
		KurtosisValueTypeDefault: copiedValueType,
	}, nil
}

// GetImage returns the image that needs to be pulled
func (irs *ImageSpec) GetImage() (string, *startosis_errors.InterpretationError) {
	image, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](irs.KurtosisValueTypeDefault, ImageAttr)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	if !found {
		return "", startosis_errors.NewInterpretationError("Required attribute '%s' could not be found on type '%s'",
			ImageAttr, ImageSpecTypeName)
	}
	imageStr := image.GoString()
	return imageStr, nil
}

// GetPasswordIfSet returns the password of the account for the registry
func (irs *ImageSpec) GetPasswordIfSet() (string, *startosis_errors.InterpretationError) {
	password, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](irs.KurtosisValueTypeDefault, ImageRegistryPasswordAttr)
	if !found {
		return "", nil
	}
	if interpretationErr != nil {
		return "", interpretationErr
	}
	passwordStr := password.GoString()
	return passwordStr, nil
}

// GetRegistryAddrIfSet returns the address of the registry from which the image has to be pulled
func (irs *ImageSpec) GetRegistryAddrIfSet() (string, *startosis_errors.InterpretationError) {
	registryAddr, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](irs.KurtosisValueTypeDefault, ImageRegistryAttr)
	if !found {
		return "", nil
	}
	if interpretationErr != nil {
		return "", interpretationErr
	}
	registryAddrStr := registryAddr.GoString()
	return registryAddrStr, nil
}

// GetUsernameIfSet returns the address of the registry from which the image has to be pulled
func (irs *ImageSpec) GetUsernameIfSet() (string, *startosis_errors.InterpretationError) {
	registryAddr, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](irs.KurtosisValueTypeDefault, ImageRegistryUsernameAttr)
	if !found {
		return "", nil
	}
	if interpretationErr != nil {
		return "", interpretationErr
	}
	registryAddrStr := registryAddr.GoString()
	return registryAddrStr, nil
}

func (irs *ImageSpec) ToKurtosisType() (*image_registry_spec.ImageRegistrySpec, *startosis_errors.InterpretationError) {
	image, err := irs.GetImage()
	if err != nil {
		return nil, err
	}

	username, err := irs.GetUsernameIfSet()
	if err != nil {
		return nil, err
	}

	password, err := irs.GetPasswordIfSet()
	if err != nil {
		return nil, err
	}

	registryAddr, err := irs.GetRegistryAddrIfSet()
	if err != nil {
		return nil, err
	}

	return image_registry_spec.NewImageRegistrySpec(image, username, password, registryAddr), nil
}
