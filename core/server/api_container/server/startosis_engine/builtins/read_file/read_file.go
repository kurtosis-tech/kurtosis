package read_file

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_helper"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/starlark_warning"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"go.starlark.net/starlark"
)

const (
	ReadFileBuiltinName = "read_file"

	SrcArgName = "src"
)

func NewReadFileHelper(packageId string, packageContentProvider startosis_packages.PackageContentProvider) *kurtosis_helper.KurtosisHelper {
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
					//TODO remove this deprecation warning when the local absolute locators block is implemented
					Deprecation: starlark_warning.Deprecation(
						starlark_warning.DeprecationDate{
							Day: 0, Year: 0, Month: 0, //nolint:gomnd
						},
						"Local 'absolute locators' are being deprecated in favor of 'relative locators' to normalize when a locator is pointing to inside or outside the package. e.g.: if your package name is 'github.com/sample/sample-kurtosis-package' and the package contains a 'local absolute locator' for example 'github.com/sample/sample-kurtosis-package/component/component.star' it should be modified to a relative version like this '/component/component.star' or './component/component.star', or, if you are referencing it in a sub-folder, you can use a 'relative locator' like this '../component/component.star'.",
						func(value starlark.Value) bool {
							if err := builtin_argument.RelativeOrRemoteAbsoluteLocator(value, packageId, SrcArgName); err != nil {
								return true
							}
							return false
						},
					),
				},
			},
		},

		Capabilities: &readFileCapabilities{
			packageContentProvider: packageContentProvider,
		},
	}
}

type readFileCapabilities struct {
	packageContentProvider startosis_packages.PackageContentProvider
}

func (builtin *readFileCapabilities) Interpret(locatorOfModuleInWhichThisBuiltInIsBeingCalled string, arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError) {
	srcValue, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, SrcArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for arg '%s'", srcValue)
	}
	fileToReadStr := srcValue.GoString()
	fileToReadStr, relativePathParsingInterpretationErr := builtin.packageContentProvider.GetAbsoluteLocatorForRelativeModuleLocator(locatorOfModuleInWhichThisBuiltInIsBeingCalled, fileToReadStr)
	if relativePathParsingInterpretationErr != nil {
		return nil, relativePathParsingInterpretationErr
	}
	packageContent, interpretationErr := builtin.packageContentProvider.GetModuleContents(fileToReadStr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return starlark.String(packageContent), nil
}
