package startosis_modules

import (
	"go.starlark.net/starlarkstruct"
)

// ModuleCacheEntry The module cache entry
type ModuleCacheEntry struct {
	module *starlarkstruct.Module
	err    error
}

func NewModuleCacheEntry(module *starlarkstruct.Module, err error) *ModuleCacheEntry {
	return &ModuleCacheEntry{
		module: module,
		err:    err,
	}
}

func (moduleCacheEntry *ModuleCacheEntry) GetModule() *starlarkstruct.Module {
	return moduleCacheEntry.module
}

func (moduleCacheEntry *ModuleCacheEntry) GetError() error {
	return moduleCacheEntry.err
}
