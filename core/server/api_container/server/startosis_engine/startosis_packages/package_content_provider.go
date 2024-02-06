package startosis_packages

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/commons/yaml_parser"
	"io"
)

// PackageContentProvider A package content provider allows you to get a Startosis package given a URL
// It fetches the contents of the package for you
//
// Regenerate mock with the following command from core/server directory:
// mockery -r --name=PackageContentProvider --filename=mock_package_content_provider.go --structname=MockPackageContentProvider --with-expecter --inpackage
type PackageContentProvider interface {
	// GetOnDiskAbsolutePackageFilePath returns the absolute file path of a file inside a module.
	// The corresponding GitHub repo will be cloned if necessary
	GetOnDiskAbsolutePackageFilePath(fileInsidePackageUrl string) (string, *startosis_errors.InterpretationError)

	// GetOnDiskAbsolutePath returns the absolute path (it can be a filer or a folder and, it can be from a package or not) on APIC's disk.
	// The corresponding GitHub repo will be cloned if necessary
	GetOnDiskAbsolutePath(repositoryPathURL string) (string, *startosis_errors.InterpretationError)

	// GetModuleContents returns the stringifies content of a file inside a module
	GetModuleContents(fileInsidePackageUrl string) (string, *startosis_errors.InterpretationError)

	// GetOnDiskAbsolutePackagePath returns the absolute folder path containing this package
	// It throws an error if the package does not exist on disk
	GetOnDiskAbsolutePackagePath(packageId string) (string, *startosis_errors.InterpretationError)

	// StorePackageContents writes on disk the content of the package passed as params
	StorePackageContents(packageId string, packageContent io.Reader, overwriteExisting bool) (string, *startosis_errors.InterpretationError)

	// ClonePackage clones the package with the given id and returns the absolute path on disk
	ClonePackage(packageId string) (string, *startosis_errors.InterpretationError)

	// GetAbsoluteLocator does:
	// 1. if the given locator is relative, translates it to absolute locator using sourceModuleLocator (if it's already absolute, does nothing)
	// 2. applies any replace rules, if they match the now-absolute locator
	GetAbsoluteLocator(packageId string, locatorOfModuleInWhichThisBuiltInIsBeingCalled string, relativeOrAbsoluteLocator string, packageReplaceOptions map[string]string) (string, *startosis_errors.InterpretationError)

	// GetKurtosisYaml returns the package kurtosis.yml file content
	GetKurtosisYaml(packageAbsolutePathOnDisk string) (*yaml_parser.KurtosisYaml, *startosis_errors.InterpretationError)

	// CloneReplacedPackagesIfIsNeeded will compare the received currentPackageReplaceOptions with the historical replace options (from previous run)
	// and will clone the packages depending on the comparison result
	CloneReplacedPackagesIfNeeded(currentPackageReplaceOptions map[string]string) *startosis_errors.InterpretationError
}
