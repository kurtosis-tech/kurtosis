package service_config

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_registry_spec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
)

const (
	ImageRegistrySpecTypeName = "ImageRegistrySpec"

	RegistryImageAttr    = "image"
	RegistryAddrAttr     = "registry"
	RegistryUsernameAttr = "username"
	RegistryPasswordAttr = "password"
	RegistryEmailAttr    = "email"
)

func NewImageRegistrySpec() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return &kurtosis_type_constructor.KurtosisTypeConstructor{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: ImageRegistrySpecTypeName,
			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              RegistryImageAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, RegistryImageAttr)
					},
				},
				{
					Name:              RegistryAddrAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, RegistryAddrAttr)
					},
				},
				{
					Name:              RegistryEmailAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.ValidEmailString(value, RegistryEmailAttr)
					},
				},
				{
					Name:              RegistryUsernameAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, RegistryUsernameAttr)
					},
				},
				{
					Name:              RegistryPasswordAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, RegistryUsernameAttr)
					},
				},
			},
		},
		Instantiate: instantiateImageRegistrySpec,
	}
}

func instantiateImageRegistrySpec(arguments *builtin_argument.ArgumentValuesSet) (builtin_argument.KurtosisValueType, *startosis_errors.InterpretationError) {
	kurtosisValueType, err := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(ImageRegistrySpecTypeName, arguments)
	if err != nil {
		return nil, err
	}
	return &ImageRegistrySpec{
		KurtosisValueTypeDefault: kurtosisValueType,
	}, nil
}

// ImageRegistrySpec is a starlark.Value that holds all the information to log in to a Docker registry
type ImageRegistrySpec struct {
	*kurtosis_type_constructor.KurtosisValueTypeDefault
}

func (irs *ImageRegistrySpec) Copy() (builtin_argument.KurtosisValueType, error) {
	copiedValueType, err := irs.KurtosisValueTypeDefault.Copy()
	if err != nil {
		return nil, err
	}
	return &ImageRegistrySpec{
		KurtosisValueTypeDefault: copiedValueType,
	}, nil
}

// GetImage returns the image that needs to be pulled
func (irs *ImageRegistrySpec) GetImage() (string, *startosis_errors.InterpretationError) {
	image, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](irs.KurtosisValueTypeDefault, RegistryImageAttr)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	if !found {
		return "", startosis_errors.NewInterpretationError("Required attribute '%s' could not be found on type '%s'",
			RegistryImageAttr, ImageRegistrySpecTypeName)
	}
	imageStr := image.GoString()
	return imageStr, nil
}

// GetEmail returns the email address of the account for the registry
func (irs *ImageRegistrySpec) GetEmail() (string, *startosis_errors.InterpretationError) {
	email, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](irs.KurtosisValueTypeDefault, RegistryEmailAttr)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	if !found {
		return "", startosis_errors.NewInterpretationError("Required attribute '%s' could not be found on type '%s'",
			RegistryEmailAttr, ImageRegistrySpecTypeName)
	}
	emailStr := email.GoString()
	return emailStr, nil
}

// GetPassword returns the password of the account for the registry
func (irs *ImageRegistrySpec) GetPassword() (string, *startosis_errors.InterpretationError) {
	password, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](irs.KurtosisValueTypeDefault, RegistryPasswordAttr)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	if !found {
		return "", startosis_errors.NewInterpretationError("Required attribute '%s' could not be found on type '%s'",
			RegistryPasswordAttr, ImageRegistrySpecTypeName)
	}
	passwordStr := password.GoString()
	return passwordStr, nil
}

// GetRegistryAddr returns the address of the registry from which the image has to be pulled
func (irs *ImageRegistrySpec) GetRegistryAddr() (string, *startosis_errors.InterpretationError) {
	registryAddr, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](irs.KurtosisValueTypeDefault, RegistryAddrAttr)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	if !found {
		return "", startosis_errors.NewInterpretationError("Required attribute '%s' could not be found on type '%s'",
			RegistryAddrAttr, ImageRegistrySpecTypeName)
	}
	registryAddrStr := registryAddr.GoString()
	return registryAddrStr, nil
}

// GetUsername returns the address of the registry from which the image has to be pulled
func (irs *ImageRegistrySpec) GetUsername() (string, *startosis_errors.InterpretationError) {
	registryAddr, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](irs.KurtosisValueTypeDefault, RegistryUsernameAttr)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	if !found {
		return "", startosis_errors.NewInterpretationError("Required attribute '%s' could not be found on type '%s'",
			RegistryUsernameAttr, ImageRegistrySpecTypeName)
	}
	registryAddrStr := registryAddr.GoString()
	return registryAddrStr, nil
}

func (irs *ImageRegistrySpec) ToKurtosisType() (*image_registry_spec.ImageRegistrySpec, *startosis_errors.InterpretationError) {
	image, err := irs.GetImage()
	if err != nil {
		return nil, err
	}

	email, err := irs.GetEmail()
	if err != nil {
		return nil, err
	}

	username, err := irs.GetUsername()
	if err != nil {
		return nil, err
	}

	password, err := irs.GetPassword()
	if err != nil {
		return nil, err
	}

	registryAddr, err := irs.GetRegistryAddr()
	if err != nil {
		return nil, err
	}

	return image_registry_spec.NewImageRegistrySpec(image, email, username, password, registryAddr), nil
}
