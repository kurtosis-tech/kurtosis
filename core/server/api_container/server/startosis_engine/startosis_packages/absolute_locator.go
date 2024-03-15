package startosis_packages

import "fmt"

const (
	defaultMainBranch = ""
)

// PackageAbsoluteLocator represents the absolute locator for a file in a Kurtosis package
type PackageAbsoluteLocator struct {
	// it is the file locator value
	locator string
	// indicates if the absolute locator correspond to a specific tag, branch or commit
	// the zero value (empty string) correspond to the main (or master) branch
	tagBranchOrCommit string
}

func NewPackageAbsoluteLocator(locator string, tagBranchOrCommit string) *PackageAbsoluteLocator {
	return &PackageAbsoluteLocator{locator: locator, tagBranchOrCommit: tagBranchOrCommit}
}

func (packageAbsoluteLocator *PackageAbsoluteLocator) GetLocator() string {
	return packageAbsoluteLocator.locator
}

func (packageAbsoluteLocator *PackageAbsoluteLocator) GetTagBranchOrCommit() string {
	return packageAbsoluteLocator.tagBranchOrCommit
}

func (packageAbsoluteLocator *PackageAbsoluteLocator) GetGitURL() string {
	if packageAbsoluteLocator.tagBranchOrCommit == defaultMainBranch {
		return packageAbsoluteLocator.locator
	}
	return fmt.Sprintf("%s@%s", packageAbsoluteLocator.locator, packageAbsoluteLocator.tagBranchOrCommit)
}
