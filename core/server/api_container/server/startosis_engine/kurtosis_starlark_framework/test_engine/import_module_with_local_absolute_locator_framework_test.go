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
	importModuleWithLocalAbsoluteLocatorExpectedErrorMsg = "Cannot construct 'import_module' from the provided arguments.\n\tCaused by: The following argument(s) could not be parsed or did not pass validation: {\"module_file\":\"The locator '\\\"github.com/kurtosistech/test-package/helpers.star\\\"' set in attribute 'module_file' is not a 'local relative locator'. Local absolute locators are not allowed you should modified it to be a valid 'local relative locator'\"}"
)

var (
	importModuleWithLocalAbsoluteLocator_mockStarlarkModule = &starlarkstruct.Module{
		Name: TestModuleFileName,
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
	// start with an empty cache to validate it gets populated
	moduleGlobalCache := map[string]*startosis_packages.ModuleCacheEntry{}

	suite.runShouldFail(
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
	return import_module.NewImportModule(TestModulePackageId, recursiveInterpret, t.packageContentProvider, t.moduleGlobalCache, TestNoPackageReplaceOptions)
}

func (t *importModuleWithLocalAbsoluteLocatorTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%q)", import_module.ImportModuleBuiltinName, import_module.ModuleFileArgName, TestModuleFileName)
}

func (t *importModuleWithLocalAbsoluteLocatorTestCase) GetStarlarkCodeForAssertion() string {
	return ""
}

func (t *importModuleWithLocalAbsoluteLocatorTestCase) Assert(_ starlark.Value) {
}
