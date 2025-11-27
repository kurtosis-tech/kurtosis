package import_module

import (
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_helper"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"github.com/sirupsen/logrus"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const (
	ImportModuleBuiltinName = "import_module"

	ModuleFileArgName = "module_file"
)

/*
	NewImportModule returns a sequential (not parallel) implementation of an equivalent or `load` in Starlark
	This function returns a starlarkstruct.Module object that can then be used to get variables and call functions from the loaded module.

How does the returned function work?
1. The function first checks whether a module is currently loading. If so then there's cycle and it errors immediately,
2. The function checks then interpreter.modulesGlobalCache for preloaded symbols or previous interpretation errors, if there is either of them it returns
3. At this point this is a new module for this instance of the interpreter, we set load to in progress (this is useful for cycle detection).
5. We defer undo the loading in case there is a failure loading the contents of the module. We don't want it to be the loading state as the next call to load the module would incorrectly return a cycle error.
6. We then load the contents of the module file using a custom provider which fetches Git repos.
7. After we have the contents of the module, we execute it using the `recursiveInterpret` function provided by the interpreter
8. At this point we cache the symbols from the loaded module
9. We now return the contents of the module and any interpretation errors
This function is recursive in the sense, to load a module that loads modules we call the same function
*/
func NewImportModule(
	packageId string,
	recursiveInterpret func(moduleId string, scriptContent string) (starlark.StringDict, *startosis_errors.InterpretationError),
	packageContentProvider startosis_packages.PackageContentProvider,
	moduleGlobalCache map[string]*startosis_packages.ModuleCacheEntry,
	packageReplaceOptions map[string]string,
) *kurtosis_helper.KurtosisHelper {
	return &kurtosis_helper.KurtosisHelper{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: ImportModuleBuiltinName,
			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              ModuleFileArgName,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, ModuleFileArgName)
					},
				},
			},
		},

		Capabilities: &importModuleCapabilities{
			packageId:              packageId,
			packageContentProvider: packageContentProvider,
			recursiveInterpret:     recursiveInterpret,
			moduleGlobalCache:      moduleGlobalCache,
			packageReplaceOptions:  packageReplaceOptions,
		},
	}
}

type importModuleCapabilities struct {
	packageId              string
	packageContentProvider startosis_packages.PackageContentProvider
	recursiveInterpret     func(moduleId string, scriptContent string) (starlark.StringDict, *startosis_errors.InterpretationError)
	moduleGlobalCache      map[string]*startosis_packages.ModuleCacheEntry
	packageReplaceOptions  map[string]string
}

func (builtin *importModuleCapabilities) Interpret(locatorOfModuleInWhichThisBuiltInIsBeingCalled string, arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError) {
	moduleLocatorArgValue, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, ModuleFileArgName)
	if err != nil {
		return nil, explicitInterpretationError(err)
	}
	relativeOrAbsoluteModuleLocator := moduleLocatorArgValue.GoString()
	absoluteModuleLocatorObj, relativePathParsingInterpretationErr := builtin.packageContentProvider.GetAbsoluteLocator(builtin.packageId, locatorOfModuleInWhichThisBuiltInIsBeingCalled, relativeOrAbsoluteModuleLocator, builtin.packageReplaceOptions)
	if relativePathParsingInterpretationErr != nil {
		return nil, relativePathParsingInterpretationErr
	}
	absoluteModuleLocator := absoluteModuleLocatorObj.GetLocator()
	logrus.Debugf("importing module from absolute locator '%s' with tag, branch or commit %s", absoluteModuleLocator, absoluteModuleLocatorObj.GetTagBranchOrCommit())

	var loadInProgress *startosis_packages.ModuleCacheEntry
	cacheEntry, found := builtin.moduleGlobalCache[absoluteModuleLocator]
	if found && cacheEntry == loadInProgress {
		return nil, startosis_errors.NewInterpretationError("There's a cycle in the import_module calls")
	}
	if found {
		return cacheEntry.GetModule(), cacheEntry.GetError()
	}

	builtin.moduleGlobalCache[absoluteModuleLocator] = loadInProgress
	shouldUnsetLoadInProgress := true
	defer func() {
		if shouldUnsetLoadInProgress {
			delete(builtin.moduleGlobalCache, absoluteModuleLocator)
		}
	}()

	// Load it.
	contents, interpretationError := builtin.packageContentProvider.GetModuleContents(absoluteModuleLocatorObj)
	if interpretationError != nil {
		return nil, startosis_errors.WrapWithInterpretationError(interpretationError, "An error occurred while loading the module '%v'", absoluteModuleLocator)
	}

	globalVariables, interpretationErr := builtin.recursiveInterpret(absoluteModuleLocator, contents)
	// the above error goes unchecked as it needs to be persisted to the cache and then returned to the parent loader

	// Update the cache.
	if interpretationErr == nil {
		newModule := &starlarkstruct.Module{
			Name:    absoluteModuleLocator,
			Members: globalVariables,
		}
		cacheEntry = startosis_packages.NewModuleCacheEntry(newModule, nil)
	} else {
		cacheEntry = startosis_packages.NewModuleCacheEntry(nil, interpretationErr)
	}
	builtin.moduleGlobalCache[absoluteModuleLocator] = cacheEntry

	shouldUnsetLoadInProgress = false
	if cacheEntry.GetError() != nil {
		return nil, cacheEntry.GetError()
	}
	return cacheEntry.GetModule(), nil
	// this error isn't propagated as it is returned to the interpreter & persisted in the cache
}

func explicitInterpretationError(err error) *startosis_errors.InterpretationError {
	return startosis_errors.WrapWithInterpretationError(
		err,
		"Unable to parse arguments of command '%s'. It should be a non empty string argument pointing to the fully qualified .star file (i.e. \"github.com/kurtosis/package/helpers.star\")",
		ImportModuleBuiltinName)
}
