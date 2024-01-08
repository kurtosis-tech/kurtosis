package service_user

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
)

const (
	UserTypeName = "User"

	UIDAttr = "uid"
	GIDAttr = "gid"
)

func NewUserType() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return &kurtosis_type_constructor.KurtosisTypeConstructor{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: UserTypeName,
			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:       UIDAttr,
					IsOptional: false,
				},
			},
			Deprecation: nil,
		},
		Instantiate: nil,
	}
}
