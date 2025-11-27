package store_spec

import (
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/store_spec"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
)

const (
	StoreSpecTypeName = "StoreSpec"

	SrcAttr  = "src"
	NameAttr = "name"
)

func NewStoreSpecType() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return &kurtosis_type_constructor.KurtosisTypeConstructor{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: StoreSpecTypeName,
			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              SrcAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, kurtosis_types.ServiceNameAttr)
					},
				},
				{
					Name:              NameAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, kurtosis_types.ServiceNameAttr)
					},
				},
			},
			Deprecation: nil,
		},
		Instantiate: instantiate,
	}
}

func instantiate(arguments *builtin_argument.ArgumentValuesSet) (builtin_argument.KurtosisValueType, *startosis_errors.InterpretationError) {
	kurtosisValueType, interpretationErr := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(StoreSpecTypeName, arguments)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return &StoreSpec{
		KurtosisValueTypeDefault: kurtosisValueType,
	}, nil
}

type StoreSpec struct {
	*kurtosis_type_constructor.KurtosisValueTypeDefault
}

func (storeSpecObj *StoreSpec) Copy() (builtin_argument.KurtosisValueType, error) {
	copiedValueType, err := storeSpecObj.KurtosisValueTypeDefault.Copy()
	if err != nil {
		return nil, err
	}
	return &StoreSpec{
		KurtosisValueTypeDefault: copiedValueType,
	}, nil
}

func (storeSpecObj *StoreSpec) GetName() (string, *startosis_errors.InterpretationError) {
	name, _, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](
		storeSpecObj.KurtosisValueTypeDefault, NameAttr)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	return name.GoString(), nil
}

func (storeSpecObj *StoreSpec) GetSrc() (string, *startosis_errors.InterpretationError) {
	src, _, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](
		storeSpecObj.KurtosisValueTypeDefault, SrcAttr)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	return src.GoString(), nil
}

func (storeSpecObj *StoreSpec) ToKurtosisType() (*store_spec.StoreSpec, *startosis_errors.InterpretationError) {
	name, interpretationErr := storeSpecObj.GetName()
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	src, interpretationErr := storeSpecObj.GetSrc()
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	return store_spec.NewStoreSpec(src, name), nil
}
