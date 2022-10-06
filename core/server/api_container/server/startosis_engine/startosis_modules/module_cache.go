package startosis_modules

import (
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

func (moduleCache *ModuleCache) SetLoadInProgress(moduleID string) {
	moduleCache.mutex.Lock()
	defer moduleCache.mutex.Unlock()
	moduleCache.cache[moduleID] = loadInProgress
}

func (moduleCache *ModuleCache) IsLoadInProgress(moduleID string) bool {
	moduleCache.mutex.Lock()
	defer moduleCache.mutex.Unlock()
	entry, found := moduleCache.cache[moduleID]
	if found && entry == loadInProgress {
		return true
	}
	return false
}

func (moduleCache *ModuleCache) Add(moduleID string, entry *ModuleCacheEntry) {
	moduleCache.mutex.Lock()
	defer moduleCache.mutex.Unlock()
	moduleCache.cache[moduleID] = entry
}

func (moduleCache *ModuleCache) LoadFinished(moduleID string) {
	moduleCache.mutex.Lock()
	defer moduleCache.mutex.Unlock()
	entry, found := moduleCache.cache[moduleID]
	if found && entry == loadInProgress {
		delete(moduleCache.cache, moduleID)
	}
}

func (moduleCache *ModuleCache) Get(module string) (*ModuleCacheEntry, bool) {
	moduleCache.mutex.Lock()
	defer moduleCache.mutex.Unlock()
	entry, found := moduleCache.cache[module]
	return entry, found
}
