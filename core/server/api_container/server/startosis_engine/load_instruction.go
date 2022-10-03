package startosis_engine

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/module_manager"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
)

type CacheEntry struct {
	globals starlark.StringDict
	err     error
}

type LoadInstruction struct {
	moduleCache   map[string]*CacheEntry
	moduleManager module_manager.ModuleManager
}

func NewLoadInstruction(moduleManager module_manager.ModuleManager) *LoadInstruction {
	return &LoadInstruction{
		moduleManager: moduleManager,
		moduleCache:   make(map[string]*CacheEntry),
	}
}

func (load *LoadInstruction) Load(thread *starlark.Thread, module string) (starlark.StringDict, error) {
	entries, ok := load.moduleCache[module]
	if entries == nil {
		if ok {
			return nil, stacktrace.NewError("There is a cycle in the load graph")
		}

		// Add a placeholder to indicate "load in progress".
		load.moduleCache[module] = nil

		// Load it.
		contents, err := load.moduleManager.GetModule(module)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred while fetching contents of the module '%v'", module)
		}

		thread := &starlark.Thread{Name: "exec " + module, Load: thread.Load}
		globals, err := starlark.ExecFile(thread, module, contents, nil)
		entries = &CacheEntry{globals, err}

		// Update the cache.
		load.moduleCache[module] = entries
	}
	return entries.globals, entries.err
}

