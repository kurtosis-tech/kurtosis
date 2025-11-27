package read_file

import (
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_helper"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"go.starlark.net/starlark"
)

const (
	ReadFileBuiltinName = "read_file"

	SrcArgName = "src"
)

func NewReadFileHelper(
	packageId string,
	packageContentProvider startosis_packages.PackageContentProvider,
	packageReplaceOptions map[string]string,
) *kurtosis_helper.KurtosisHelper {
	return &kurtosis_helper.KurtosisHelper{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: ReadFileBuiltinName,
			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              SrcArgName,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, SrcArgName)
					},
				},
			},
		},

		Capabilities: &readFileCapabilities{
			packageId:              packageId,
			packageContentProvider: packageContentProvider,
			packageReplaceOptions:  packageReplaceOptions,
		},
	}
}

type readFileCapabilities struct {
	packageId              string
	packageContentProvider startosis_packages.PackageContentProvider
	packageReplaceOptions  map[string]string
}

func (builtin *readFileCapabilities) Interpret(locatorOfModuleInWhichThisBuiltInIsBeingCalled string, arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError) {
	srcValue, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, SrcArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for arg '%s'", srcValue)
	}
	fileToReadStr := srcValue.GoString()
	absoluteLocator, relativePathParsingInterpretationErr := builtin.packageContentProvider.GetAbsoluteLocator(builtin.packageId, locatorOfModuleInWhichThisBuiltInIsBeingCalled, fileToReadStr, builtin.packageReplaceOptions)
	if relativePathParsingInterpretationErr != nil {
		return nil, relativePathParsingInterpretationErr
	}
	packageContent, interpretationErr := builtin.packageContentProvider.GetModuleContents(absoluteLocator)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return starlark.String(packageContent), nil
}
