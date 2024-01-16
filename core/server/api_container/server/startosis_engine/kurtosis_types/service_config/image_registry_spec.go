package service_config

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_registry_spec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
	"net/mail"
)

const (
	ImageRegistrySpecType = "ImageRegistrySpec"

	RegistryAddrAttr = "registry"
	UsernameAttr     = "username"
	PasswordAttr     = "password"
	EmailAttr        = "email"
)

func NewImageRegistrySpec() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return &kurtosis_type_constructor.KurtosisTypeConstructor{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: ImageRegistrySpecType,
			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              ImageAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, ImageAttr)
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
					Name:              EmailAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						interpretationErr := builtin_argument.NonEmptyString(value, EmailAttr)
						if interpretationErr != nil {
							return interpretationErr
						}
						emailAddressAsStr := value.String()
						if _, err := mail.ParseAddress(emailAddressAsStr); err != nil {
							return startosis_errors.WrapWithInterpretationError(err, "An error occurred while validating email address '%v' passed via attr '%v'", emailAddressAsStr, EmailAttr)
						}
						return nil
					},
				},
				{
					Name:              UsernameAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, UsernameAttr)
					},
				},
				{
					Name:              PasswordAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, UsernameAttr)
					},
				},
			},
		},
		Instantiate: instantiateImageRegistrySpec,
	}
}

func instantiateImageRegistrySpec(arguments *builtin_argument.ArgumentValuesSet) (builtin_argument.KurtosisValueType, *startosis_errors.InterpretationError) {
	kurtosisValueType, err := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(ImageRegistrySpecType, arguments)
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
	image, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](ImageRegistrySpec{}.KurtosisValueTypeDefault, ImageAttr)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	if !found {
		return "", startosis_errors.NewInterpretationError("Required attribute '%s' could not be found on type '%s'",
			ImageAttr, ImageRegistrySpecType)
	}
	imageStr := image.GoString()
	return imageStr, nil
}

// GetEmail returns the email address of the account for the registry
func (irs *ImageRegistrySpec) GetEmail() (string, *startosis_errors.InterpretationError) {
	email, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](ImageRegistrySpec{}.KurtosisValueTypeDefault, EmailAttr)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	if !found {
		return "", startosis_errors.NewInterpretationError("Required attribute '%s' could not be found on type '%s'",
			EmailAttr, ImageRegistrySpecType)
	}
	emailStr := email.GoString()
	return emailStr, nil
}

// GetPassword returns the password of the account for the registry
func (irs *ImageRegistrySpec) GetPassword() (string, *startosis_errors.InterpretationError) {
	password, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](ImageRegistrySpec{}.KurtosisValueTypeDefault, PasswordAttr)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	if !found {
		return "", startosis_errors.NewInterpretationError("Required attribute '%s' could not be found on type '%s'",
			PasswordAttr, ImageRegistrySpecType)
	}
	passwordStr := password.GoString()
	return passwordStr, nil
}

// GetRegistryAddr returns the address of the registry from which the image has to be pulled
func (irs *ImageRegistrySpec) GetRegistryAddr() (string, *startosis_errors.InterpretationError) {
	registryAddr, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](ImageRegistrySpec{}.KurtosisValueTypeDefault, RegistryAddrAttr)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	if !found {
		return "", startosis_errors.NewInterpretationError("Required attribute '%s' could not be found on type '%s'",
			RegistryAddrAttr, ImageRegistrySpecType)
	}
	registryAddrStr := registryAddr.GoString()
	return registryAddrStr, nil
}

// GetUsername returns the address of the registry from which the image has to be pulled
func (irs *ImageRegistrySpec) GetUsername() (string, *startosis_errors.InterpretationError) {
	registryAddr, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](ImageRegistrySpec{}.KurtosisValueTypeDefault, UsernameAttr)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	if !found {
		return "", startosis_errors.NewInterpretationError("Required attribute '%s' could not be found on type '%s'",
			UsernameAttr, ImageRegistrySpecType)
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
