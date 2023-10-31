package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/builtins/read_file"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_helper"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"go.starlark.net/starlark"
	"testing"
)

const (
	readFileWithLocalAbsoluteLocatorExpectedErrorMsg = "Cannot use absolute locators"
)

type readFileWithLocalAbsoluteLocatorTestCase struct {
	*testing.T

	packageContentProvider startosis_packages.PackageContentProvider
}

func (suite *KurtosisHelperTestSuite) TestReadFileWithLocalAbsoluteLocatorShouldNotBeValid() {
	suite.packageContentProvider.EXPECT().GetAbsoluteLocatorForRelativeLocator(testModulePackageId, testModuleMainFileLocator, testModuleFileName, testNoPackageReplaceOptions).Return("", startosis_errors.NewInterpretationError(readFileWithLocalAbsoluteLocatorExpectedErrorMsg))

	suite.runShouldFail(
		testModuleMainFileLocator,
		&readFileWithLocalAbsoluteLocatorTestCase{
			T:                      suite.T(),
			packageContentProvider: suite.packageContentProvider,
		},
		readFileWithLocalAbsoluteLocatorExpectedErrorMsg,
	)
}

func (t *readFileWithLocalAbsoluteLocatorTestCase) GetHelper() *kurtosis_helper.KurtosisHelper {
	return read_file.NewReadFileHelper(testModulePackageId, t.packageContentProvider, testNoPackageReplaceOptions)
}

func (t *readFileWithLocalAbsoluteLocatorTestCase) GetStarlarkCode() string {
	return fmt.Sprintf(`%s(%s=%q)`, read_file.ReadFileBuiltinName, read_file.SrcArgName, testModuleFileName)
}

func (t *readFileWithLocalAbsoluteLocatorTestCase) GetStarlarkCodeForAssertion() string {
	return ""
}

func (t *readFileWithLocalAbsoluteLocatorTestCase) Assert(_ starlark.Value) {

}
