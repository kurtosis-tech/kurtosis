package import_module

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_modules"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const (
	ImportModuleBuiltinName = "import_module"

	moduleFileArgName = "module_file"
)

/*
	GenerateImportBuiltin returns a sequential (not parallel) implementation of an equivalent or `load` in Starlark
	This function returns a starlarkstruct.Module object that can then me used to get variables and call functions from the loaded module.

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
func GenerateImportBuiltin(recursiveInterpret func(moduleId string, scriptContent string) (starlark.StringDict, error), moduleContentProvider startosis_modules.ModuleContentProvider, moduleGlobalCache map[string]*startosis_modules.ModuleCacheEntry) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		fileInModule, interpretationError := parseStartosisArgs(args, kwargs)
		if interpretationError != nil {
			return nil, interpretationError
		}

		var loadInProgress *startosis_modules.ModuleCacheEntry
		cacheEntry, found := moduleGlobalCache[fileInModule]
		if found && cacheEntry == loadInProgress {
			return nil, startosis_errors.NewInterpretationError("There's a cycle in the import_module calls")
		}
		if found {
			return cacheEntry.GetModule(), cacheEntry.GetError()
		}

		moduleGlobalCache[fileInModule] = loadInProgress
		shouldUnsetLoadInProgress := true
		defer func() {
			if shouldUnsetLoadInProgress {
				delete(moduleGlobalCache, fileInModule)
			}
		}()

		// Load it.
		contents, interpretationError := moduleContentProvider.GetModuleContents(fileInModule)
		if interpretationError != nil {
			return nil, startosis_errors.WrapWithInterpretationError(interpretationError, "An error occurred while loading the module '%v'", fileInModule)
		}

		globalVariables, err := recursiveInterpret(fileInModule, contents)
		// the above error goes unchecked as it needs to be persisted to the cache and then returned to the parent loader

		// Update the cache.
		var newModule *starlarkstruct.Module
		if err == nil {
			newModule = &starlarkstruct.Module{
				Name:    fileInModule,
				Members: globalVariables,
			}
		}
		cacheEntry = startosis_modules.NewModuleCacheEntry(newModule, err)
		moduleGlobalCache[fileInModule] = cacheEntry

		shouldUnsetLoadInProgress = false
		return cacheEntry.GetModule(), cacheEntry.GetError()
		// this error isn't propagated as its returned to the interpreter & persisted in the cache
	}
}

func parseStartosisArgs(args starlark.Tuple, kwargs []starlark.Tuple) (string, *startosis_errors.InterpretationError) {
	var moduleFileArg starlark.String
	if err := starlark.UnpackArgs(ImportModuleBuiltinName, args, kwargs, moduleFileArgName, &moduleFileArg); err != nil {
		return "", explicitInterpretationError(err)
	}

	moduleFilePath, interpretationErr := kurtosis_instruction.ParseNonEmptyString(moduleFileArgName, moduleFileArg)
	if interpretationErr != nil {
		return "", explicitInterpretationError(interpretationErr)
	}

	return moduleFilePath, nil
}

func explicitInterpretationError(err error) *startosis_errors.InterpretationError {
	return startosis_errors.WrapWithInterpretationError(
		err,
		"Unable to parse arguments of command '%s'. It should be a non empty string argument pointing to the fully qualified .star file (i.e. \"github.com/kurtosis/module/helpers.star\")",
		ImportModuleBuiltinName)
}
