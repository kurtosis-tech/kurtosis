package startosis_packages

import "github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"

// PackageContentProvider A package content provider allows you to get a Startosis package given a URL
// It fetches the contents of the package for you
type PackageContentProvider interface {
	// GetOnDiskAbsoluteFilePath returns the absolute file path of a file inside a module.
	// The corresponding GitHub repo will be cloned if necessary
	GetOnDiskAbsoluteFilePath(fileInsidePackageUrl string) (string, *startosis_errors.InterpretationError)

	// GetModuleContents returns the stringified content of a file inside a module
	GetModuleContents(fileInsidePackageUrl string) (string, *startosis_errors.InterpretationError)

	// GetOnDiskAbsolutePackagePath returns the absolute folder path containing this package
	// It throws an error if the package does not exist on disk
	GetOnDiskAbsolutePackagePath(packageId string) (string, *startosis_errors.InterpretationError)

	// StorePackageContents writes on disk the content of the package passed as params
	StorePackageContents(packageId string, packageContent []byte, overwriteExisting bool) (string, *startosis_errors.InterpretationError)

	// ClonePackage clones the package with the given id and returns the absolute path on disk
	ClonePackage(packageId string) (string, *startosis_errors.InterpretationError)
}
