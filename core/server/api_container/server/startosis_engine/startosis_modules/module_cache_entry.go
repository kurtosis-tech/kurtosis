package startosis_modules

import "go.starlark.net/starlark"

// ModuleCacheEntry The module cache entry
type ModuleCacheEntry struct {
	globalVariables starlark.StringDict
	err             error
}

func NewModuleCacheEntry(globalVariables starlark.StringDict, err error) *ModuleCacheEntry {
	return &ModuleCacheEntry{
		globalVariables: globalVariables,
		err:             err,
	}
}

func (moduleCacheEntry *ModuleCacheEntry) GetGlobalVariables() starlark.StringDict {
	return moduleCacheEntry.globalVariables
}

func (moduleCacheEntry *ModuleCacheEntry) GetError() error {
	return moduleCacheEntry.err
}
