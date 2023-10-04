package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/builtins/import_module"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_helper"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_constants"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"testing"
)

const (
	importModule_moduleConstKey = "CONST_STR"
)

var (
	importModule_mockStarlarkModule = &starlarkstruct.Module{
		Name: TestModuleFileName,
		Members: starlark.StringDict{
			importModule_moduleConstKey: starlark.String("Hello World!"),
		},
	}
)

type importModuleTestCase struct {
	*testing.T

	// store the cache inside the test object such that we can check its state in Assert()
	moduleGlobalCache      map[string]*startosis_packages.ModuleCacheEntry
	packageContentProvider startosis_packages.PackageContentProvider
}

func (suite *KurtosisHelperTestSuite) TestImportFile() {
	// start with an empty cache to validate it gets populated
	moduleGlobalCache := map[string]*startosis_packages.ModuleCacheEntry{}

	suite.packageContentProvider.EXPECT().GetModuleContents(TestModuleFileName).Return("Hello World!", nil)
	suite.packageContentProvider.EXPECT().GetAbsoluteLocatorForRelativeModuleLocator(startosis_constants.PackageIdPlaceholderForStandaloneScript, TestModuleRelativeLocator).Return(TestModuleFileName, nil)

	suite.run(&importModuleTestCase{
		T:                      suite.T(),
		moduleGlobalCache:      moduleGlobalCache,
		packageContentProvider: suite.packageContentProvider,
	})
}

func (t *importModuleTestCase) GetHelper() *kurtosis_helper.KurtosisHelper {
	recursiveInterpret := func(moduleId string, scriptContent string) (starlark.StringDict, *startosis_errors.InterpretationError) {
		return importModule_mockStarlarkModule.Members, nil
	}
	return import_module.NewImportModule(TestModulePackageId, recursiveInterpret, t.packageContentProvider, t.moduleGlobalCache)
}

func (t *importModuleTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%q)", import_module.ImportModuleBuiltinName, import_module.ModuleFileArgName, TestModuleRelativeLocator)
}

func (t *importModuleTestCase) GetStarlarkCodeForAssertion() string {
	return ""
}

func (t *importModuleTestCase) Assert(result starlark.Value) {
	loadedModule, ok := result.(*starlarkstruct.Module)
	require.True(t, ok, "object returned was not a starlark module")
	require.Equal(t, loadedModule.Name, TestModuleFileName)
	require.Len(t, loadedModule.Members, 1)
	require.Contains(t, loadedModule.Members, importModule_moduleConstKey)
}
