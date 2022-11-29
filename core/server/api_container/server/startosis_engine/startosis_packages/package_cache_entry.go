package startosis_packages

import (
	"go.starlark.net/starlarkstruct"
)

// ModuleCacheEntry The module cache entry
type ModuleCacheEntry struct {
	module *starlarkstruct.Module
	err    error
}

func NewPackageCacheEntry(module *starlarkstruct.Module, err error) *ModuleCacheEntry {
	return &ModuleCacheEntry{
		module: module,
		err:    err,
	}
}

func (packageCacheEntry *ModuleCacheEntry) GetModule() *starlarkstruct.Module {
	return packageCacheEntry.module
}

func (packageCacheEntry *ModuleCacheEntry) GetError() error {
	return packageCacheEntry.err
}
