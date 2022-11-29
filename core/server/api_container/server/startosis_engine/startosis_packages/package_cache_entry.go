package startosis_packages

import (
	"go.starlark.net/starlarkstruct"
)

// PackageCacheEntry The module cache entry
type PackageCacheEntry struct {
	starlarkPackage *starlarkstruct.Module
	err             error
}

func NewPackageCacheEntry(starlarkPackage *starlarkstruct.Module, err error) *PackageCacheEntry {
	return &PackageCacheEntry{
		starlarkPackage: starlarkPackage,
		err:             err,
	}
}

func (packageCacheEntry *PackageCacheEntry) GetPackage() *starlarkstruct.Module {
	return packageCacheEntry.starlarkPackage
}

func (packageCacheEntry *PackageCacheEntry) GetError() error {
	return packageCacheEntry.err
}
