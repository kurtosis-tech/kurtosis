package startosis_packages

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlarkstruct"
)

// ModuleCacheEntry The module cache entry
type ModuleCacheEntry struct {
	module            *starlarkstruct.Module
	interpretationErr *startosis_errors.InterpretationError
}

func NewModuleCacheEntry(module *starlarkstruct.Module, interpretationErr *startosis_errors.InterpretationError) *ModuleCacheEntry {
	return &ModuleCacheEntry{
		module:            module,
		interpretationErr: interpretationErr,
	}
}

func (moduleCacheEntry *ModuleCacheEntry) GetModule() *starlarkstruct.Module {
	return moduleCacheEntry.module
}

func (moduleCacheEntry *ModuleCacheEntry) GetError() *startosis_errors.InterpretationError {
	return moduleCacheEntry.interpretationErr
}
