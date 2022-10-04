package startosis_modules

import (
	"go.starlark.net/starlark"
	"sync"
)

// ModuleCache Thread safe cache of Modules
type ModuleCache struct {
	cache map[string]*ModuleCacheEntry
	mutex sync.Mutex
}

func NewModuleCache() *ModuleCache {
	return &ModuleCache{
		cache: make(map[string]*ModuleCacheEntry),
	}
}

func (moduleCache *ModuleCache) Add(module string, entry *ModuleCacheEntry) {
	moduleCache.mutex.Lock()
	defer moduleCache.mutex.Unlock()
	moduleCache.cache[module] = entry
}

func (moduleCache *ModuleCache) Get(module string) (*ModuleCacheEntry, bool) {
	moduleCache.mutex.Lock()
	defer moduleCache.mutex.Unlock()
	entry, found := moduleCache.cache[module]
	return entry, found
}

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
