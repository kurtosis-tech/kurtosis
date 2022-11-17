package startosis_modules

import "github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"

// ModuleContentProvider A module content provider allows you to get a Startosis module given a url
// It fetches the contents of the module for you
type ModuleContentProvider interface {
	// GetOnDiskAbsoluteFilePath returns the absolute file path of a file inside a module.
	// The corresponding Github repo will be cloned if necessary
	GetOnDiskAbsoluteFilePath(string) (string, *startosis_errors.InterpretationError)

	// GetModuleContents returns the stringified content of a file inside a module
	GetModuleContents(string) (string, *startosis_errors.InterpretationError)

	// StoreModuleContents writes on disk the content of the module passed as params
	StoreModuleContents(string, []byte, bool) (string, *startosis_errors.InterpretationError)

	// CloneModule clones the module with the given id and returns the absolute path on disk
	CloneModule(moduleId string) (string, error)
}
