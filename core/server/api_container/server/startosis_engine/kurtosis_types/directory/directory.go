package directory

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
)

const (
	DirectoryTypeName = "Directory"

	PersistentKeyAttr = "persistent_key"
	ArtifactNameAttr  = "artifact_name"
	DescriptionAttr   = "desc"
)

func NewDirectoryType() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return &kurtosis_type_constructor.KurtosisTypeConstructor{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: DirectoryTypeName,

			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              PersistentKeyAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, PersistentKeyAttr)
					},
				},
				{
					Name:              ArtifactNameAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, ArtifactNameAttr)
					},
				},
				{
					Name:              DescriptionAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, DescriptionAttr)
					},
				},
			},
		},

		Instantiate: instantiate,
	}
}

func instantiate(arguments *builtin_argument.ArgumentValuesSet) (builtin_argument.KurtosisValueType, *startosis_errors.InterpretationError) {
	kurtosisValueType, interpretationErr := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(DirectoryTypeName, arguments)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return &Directory{
		KurtosisValueTypeDefault: kurtosisValueType,
	}, nil
}

type Directory struct {
	*kurtosis_type_constructor.KurtosisValueTypeDefault
}

func (directory *Directory) Copy() (builtin_argument.KurtosisValueType, error) {
	copiedValueType, err := directory.KurtosisValueTypeDefault.Copy()
	if err != nil {
		return nil, err
	}
	return &Directory{
		KurtosisValueTypeDefault: copiedValueType,
	}, nil
}

func (directory *Directory) GetPersistentKeyIfSet() (string, bool, *startosis_errors.InterpretationError) {
	persistentKey, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](
		directory.KurtosisValueTypeDefault, PersistentKeyAttr)
	if interpretationErr != nil {
		return "", false, interpretationErr
	}
	if !found {
		return "", false, nil
	}
	return persistentKey.GoString(), true, nil
}

func (directory *Directory) GetArtifactNameIfSet() (string, bool, *startosis_errors.InterpretationError) {
	fileArtifact, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](
		directory.KurtosisValueTypeDefault, ArtifactNameAttr)
	if interpretationErr != nil {
		return "", false, interpretationErr
	}
	if !found {
		return "", false, nil
	}
	return fileArtifact.GoString(), true, nil
}
