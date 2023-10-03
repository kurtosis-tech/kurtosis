package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/builtins/read_file"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_helper"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"go.starlark.net/starlark"
	"testing"
)

const (
	readFileWithLocalAbsoluteLocatorExpectedErrorMsg = "Cannot construct 'read_file' from the provided arguments.\n\tCaused by: The following argument(s) could not be parsed or did not pass validation: {\"src\":\"The locator '\\\"github.com/kurtosistech/test-package/helpers.star\\\"' set in attribute 'src' is not a 'local relative locator'. Local absolute locators are not allowed you should modified it to be a valid 'local relative locator'\"}"
)

type readFileWithLocalAbsoluteLocatorTestCase struct {
	*testing.T

	packageContentProvider *startosis_packages.MockPackageContentProvider
}

func (suite *KurtosisHelperTestSuite) TestReadFileWithLocalAbsoluteLocatorShouldNotBeValid() {

	suite.runShouldFail(
		&readFileWithLocalAbsoluteLocatorTestCase{
			T:                      suite.T(),
			packageContentProvider: suite.packageContentProvider,
		},
		readFileWithLocalAbsoluteLocatorExpectedErrorMsg,
	)
}

func (t *readFileWithLocalAbsoluteLocatorTestCase) GetHelper() *kurtosis_helper.KurtosisHelper {
	return read_file.NewReadFileHelper(TestModulePackageId, t.packageContentProvider)
}

func (t *readFileWithLocalAbsoluteLocatorTestCase) GetStarlarkCode() string {
	return fmt.Sprintf(`%s(%s=%q)`, read_file.ReadFileBuiltinName, read_file.SrcArgName, TestModuleFileName)
}

func (t *readFileWithLocalAbsoluteLocatorTestCase) GetStarlarkCodeForAssertion() string {
	return ""
}

func (t *readFileWithLocalAbsoluteLocatorTestCase) Assert(_ starlark.Value) {

}
