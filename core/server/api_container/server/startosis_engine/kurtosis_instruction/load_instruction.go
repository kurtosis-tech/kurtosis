package kurtosis_instruction

import (
	"fmt"
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
	moduleManager *module_manager.ModuleManager
}

func NewLoadInstruction(moduleManager *module_manager.ModuleManager) *LoadInstruction {
	return &LoadInstruction{
		moduleManager: moduleManager,
	}
}

func (load *LoadInstruction) Load(thread *starlark.Thread, module string) (starlark.StringDict, error) {
	e, ok := load.moduleCache[module]
	if e == nil {
		if ok {
			// request for package whose loading is in progress
			return nil, fmt.Errorf("cycle in load graph")
		}

		// Add a placeholder to indicate "load in progress".
		load.moduleCache[module] = nil

		// Load it.
		moduleManager := *load.moduleManager
		contents, err := moduleManager.GetModule(module)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred while fetching contents of the module '%v'", module)
		}

		thread := &starlark.Thread{Name: "exec " + module, Load: thread.Load}
		globals, err := starlark.ExecFile(thread, module, contents, nil)
		e = &CacheEntry{globals, err}

		// Update the cache.
		load.moduleCache[module] = e
	}
	return e.globals, e.err
}

