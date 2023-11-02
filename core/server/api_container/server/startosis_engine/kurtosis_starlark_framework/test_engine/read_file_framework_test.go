package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/builtins/read_file"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_helper"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_constants"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

type readFileTestCase struct {
	*testing.T

	packageContentProvider *startosis_packages.MockPackageContentProvider
}

func (suite *KurtosisHelperTestSuite) TestReadFile() {
	suite.packageContentProvider.EXPECT().GetAbsoluteLocatorForRelativeLocator(testModulePackageId, startosis_constants.PackageIdPlaceholderForStandaloneScript, testModuleRelativeLocator, testNoPackageReplaceOptions).Return(testModuleFileName, nil)
	suite.packageContentProvider.EXPECT().GetModuleContents(testModuleFileName).Return("Hello World!", nil)

	suite.run(&readFileTestCase{
		T:                      suite.T(),
		packageContentProvider: suite.packageContentProvider,
	})
}

func (t *readFileTestCase) GetHelper() *kurtosis_helper.KurtosisHelper {
	return read_file.NewReadFileHelper(testModulePackageId, t.packageContentProvider, testNoPackageReplaceOptions)
}

func (t *readFileTestCase) GetStarlarkCode() string {
	return fmt.Sprintf(`%s(%s=%q)`, read_file.ReadFileBuiltinName, read_file.SrcArgName, testModuleRelativeLocator)
}

func (t *readFileTestCase) GetStarlarkCodeForAssertion() string {
	return ""
}

func (t *readFileTestCase) Assert(result starlark.Value) {
	t.packageContentProvider.AssertCalled(t, "GetAbsoluteLocator", testModulePackageId, startosis_constants.PackageIdPlaceholderForStandaloneScript, testModuleRelativeLocator, testNoPackageReplaceOptions)
	t.packageContentProvider.AssertCalled(t, "GetModuleContents", testModuleFileName)
	require.Equal(t, result, starlark.String("Hello World!"))
}
