package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/builtins/import_module"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_helper"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"testing"
)

const (
	importModuleWithLocalAbsoluteLocatorExpectedErrorMsg = "Cannot use absolute locators"
)

var (
	importModuleWithLocalAbsoluteLocator_mockStarlarkModule = &starlarkstruct.Module{
		Name: testModuleFileName,
		Members: starlark.StringDict{
			importModule_moduleConstKey: starlark.String("Hello World!"),
		},
	}
)

type importModuleWithLocalAbsoluteLocatorTestCase struct {
	*testing.T

	// store the cache inside the test object such that we can check its state in Assert()
	moduleGlobalCache      map[string]*startosis_packages.ModuleCacheEntry
	packageContentProvider startosis_packages.PackageContentProvider
}

func (suite *KurtosisHelperTestSuite) TestImportFileWithLocalAbsoluteLocatorShouldNotBeValid() {
	suite.packageContentProvider.EXPECT().GetAbsoluteLocatorForRelativeLocator(testModulePackageId, testModuleMainFileLocator, testModuleFileName, testNoPackageReplaceOptions).Return("", startosis_errors.NewInterpretationError(importModuleWithLocalAbsoluteLocatorExpectedErrorMsg))

	// start with an empty cache to validate it gets populated
	moduleGlobalCache := map[string]*startosis_packages.ModuleCacheEntry{}

	suite.runShouldFail(
		testModuleMainFileLocator,
		&importModuleWithLocalAbsoluteLocatorTestCase{
			T:                      suite.T(),
			moduleGlobalCache:      moduleGlobalCache,
			packageContentProvider: suite.packageContentProvider,
		},
		importModuleWithLocalAbsoluteLocatorExpectedErrorMsg,
	)
}

func (t *importModuleWithLocalAbsoluteLocatorTestCase) GetHelper() *kurtosis_helper.KurtosisHelper {
	recursiveInterpret := func(moduleId string, scriptContent string) (starlark.StringDict, *startosis_errors.InterpretationError) {
		return importModuleWithLocalAbsoluteLocator_mockStarlarkModule.Members, nil
	}
	return import_module.NewImportModule(testModulePackageId, recursiveInterpret, t.packageContentProvider, t.moduleGlobalCache, testNoPackageReplaceOptions)
}

func (t *importModuleWithLocalAbsoluteLocatorTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%q)", import_module.ImportModuleBuiltinName, import_module.ModuleFileArgName, testModuleFileName)
}

func (t *importModuleWithLocalAbsoluteLocatorTestCase) GetStarlarkCodeForAssertion() string {
	return ""
}

func (t *importModuleWithLocalAbsoluteLocatorTestCase) Assert(_ starlark.Value) {
}
