package startosis_packages

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"io"
)

// PackageContentProvider A package content provider allows you to get a Startosis package given a URL
// It fetches the contents of the package for you
//
// Regenerate mock with the following command from core/server directory:
// mockery -r --name=PackageContentProvider --filename=mock_package_content_provider.go --structname=MockPackageContentProvider --with-expecter --inpackage
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
	StorePackageContents(packageId string, packageContent io.Reader, overwriteExisting bool) (string, *startosis_errors.InterpretationError)

	// ClonePackage clones the package with the given id and returns the absolute path on disk
	ClonePackage(packageId string) (string, *startosis_errors.InterpretationError)

	// GetAbsoluteLocatorForRelativeModuleLocator returns the absolute package path for a relative module path
	GetAbsoluteLocatorForRelativeModuleLocator(packageId string, relativeOrAbsoluteModulePath string) (string, *startosis_errors.InterpretationError)
}
