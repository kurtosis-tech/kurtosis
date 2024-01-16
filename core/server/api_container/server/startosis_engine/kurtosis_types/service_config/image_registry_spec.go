package service_config

import (
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

// ImageRegistrySpec is a starlark.Value that holds all the information to login to a Docker registry
type ImageRegistrySpec struct {
	*kurtosis_type_constructor.KurtosisValueTypeDefault
}

func (imageRegistrySpec *ImageRegistrySpec) Copy() (builtin_argument.KurtosisValueType, error) {
	copiedValueType, err := imageRegistrySpec.KurtosisValueTypeDefault.Copy()
	if err != nil {
		return nil, err
	}
	return &ImageRegistrySpec{
		KurtosisValueTypeDefault: copiedValueType,
	}, nil
}

// GetImage returns the image that needs to be pulled
func (imageBuildSpec *ImageBuildSpec) GetImage() (string, *startosis_errors.InterpretationError) {
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
func (imageBuildSpec *ImageBuildSpec) GetEmail() (string, *startosis_errors.InterpretationError) {
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
func (imageBuildSpec *ImageBuildSpec) GetPassword() (string, *startosis_errors.InterpretationError) {
	email, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](ImageRegistrySpec{}.KurtosisValueTypeDefault, PasswordAttr)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	if !found {
		return "", startosis_errors.NewInterpretationError("Required attribute '%s' could not be found on type '%s'",
			PasswordAttr, ImageRegistrySpecType)
	}
	emailStr := email.GoString()
	return emailStr, nil
}

// GetRegistryAddr returns the address of the registry from which the image has to be pulled
func (imageBuildSpec *ImageBuildSpec) GetRegistryAddr() (string, *startosis_errors.InterpretationError) {
	email, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](ImageRegistrySpec{}.KurtosisValueTypeDefault, RegistryAddrAttr)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	if !found {
		return "", startosis_errors.NewInterpretationError("Required attribute '%s' could not be found on type '%s'",
			RegistryAddrAttr, ImageRegistrySpecType)
	}
	emailStr := email.GoString()
	return emailStr, nil
}
