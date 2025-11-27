package service_config

import (
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
	"math"
)

const (
	UserTypeName = "User"

	UIDAttr = "uid"
	GIDAttr = "gid"

	idIsAtLeast0 = 0
)

func NewUserType() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return &kurtosis_type_constructor.KurtosisTypeConstructor{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: UserTypeName,
			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              UIDAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Int],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.Int64InRange(value, UIDAttr, idIsAtLeast0, math.MaxInt64)
					},
				},
				{
					Name:              GIDAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Int],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.Int64InRange(value, GIDAttr, idIsAtLeast0, math.MaxInt64)
					},
				},
			},
			Deprecation: nil,
		},
		Instantiate: instantiate,
	}
}

func instantiate(arguments *builtin_argument.ArgumentValuesSet) (builtin_argument.KurtosisValueType, *startosis_errors.InterpretationError) {
	kurtosisValueType, interpretationErr := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(UserTypeName, arguments)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return &User{
		kurtosisValueType,
	}, nil
}

type User struct {
	*kurtosis_type_constructor.KurtosisValueTypeDefault
}

func (user *User) Copy() (builtin_argument.KurtosisValueType, error) {
	copiedValueType, err := user.KurtosisValueTypeDefault.Copy()
	if err != nil {
		return nil, err
	}
	return &User{
		KurtosisValueTypeDefault: copiedValueType,
	}, nil
}

func (user *User) GetUID() (int64, *startosis_errors.InterpretationError) {
	uidValue, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.Int](
		user.KurtosisValueTypeDefault, UIDAttr)
	if interpretationErr != nil {
		return 0, interpretationErr
	}
	if !found {
		return 0, startosis_errors.NewInterpretationError("Required attribute '%v' couldn't be found on '%v' type", UIDAttr, UserTypeName)
	}
	uid, ok := uidValue.Int64()
	if !ok {
		return 0, startosis_errors.NewInterpretationError("Couldn't convert uid '%v' to int64", uidValue)
	}
	return uid, nil
}

func (user *User) GetGIDIfSet() (int64, bool, *startosis_errors.InterpretationError) {
	gidValue, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.Int](
		user.KurtosisValueTypeDefault, GIDAttr)
	if interpretationErr != nil {
		return 0, false, interpretationErr
	}
	if !found {
		return 0, false, nil
	}
	gid, ok := gidValue.Int64()
	if !ok {
		return 0, false, startosis_errors.NewInterpretationError("Couldn't convert gid '%v' to int64", gidValue)
	}
	return gid, true, nil
}
