package git_package_content_provider

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/shared_utils"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"github.com/sirupsen/logrus"
	"strings"
)

const (
	dotRelativePathIndicatorString = "."
	subStrNotPresentIndicator      = -1
)

func isLocalLocator(locator string) bool {
	if strings.HasPrefix(locator, osPathSeparatorString) || strings.HasPrefix(locator, dotRelativePathIndicatorString) {
		return true
	}
	return false
}

func shouldBlockAbsoluteLocatorBecauseIsInTheSameSourceModuleLocatorPackage(relativeOrAbsoluteLocator string, sourceModuleLocator string, rootPackageId string) bool {
	// Make sure the root package id ends with a trailing slash.
	rootPackageId = strings.TrimPrefix(rootPackageId, "/") + "/"
	isSourceModuleInRootPackage := strings.HasPrefix(sourceModuleLocator, rootPackageId)
	isAbsoluteLocatorInRootPackage := strings.HasPrefix(relativeOrAbsoluteLocator, rootPackageId)
	return isSourceModuleInRootPackage && isAbsoluteLocatorInRootPackage
}

func replaceAbsoluteLocator(absoluteLocator *startosis_packages.PackageAbsoluteLocator, packageReplaceOptions map[string]string) *startosis_packages.PackageAbsoluteLocator {
	if absoluteLocator.GetLocator() == "" {
		return absoluteLocator
	}

	found, packageToBeReplaced, replaceWithPackage, maybeTagBranchOrCommit := findPackageReplace(absoluteLocator, packageReplaceOptions)

	if found {
		// we skip if it's a local replace because we will use the same absolute locator
		// due the file was already uploaded in the enclave's package cache
		if isLocalLocator(replaceWithPackage) {
			return absoluteLocator
		}
		replacedAbsoluteLocatorStr := strings.Replace(absoluteLocator.GetLocator(), packageToBeReplaced, replaceWithPackage, onlyOneReplace)
		replacedAbsoluteLocator := startosis_packages.NewPackageAbsoluteLocator(replacedAbsoluteLocatorStr, maybeTagBranchOrCommit)
		logrus.Debugf("absoluteLocator '%s' replaced with '%s' with tag, branch or commit %s", absoluteLocator.GetLocator(), replacedAbsoluteLocator.GetLocator(), replacedAbsoluteLocator.GetTagBranchOrCommit())

		return replacedAbsoluteLocator
	}

	return absoluteLocator
}

func findPackageReplace(absoluteLocator *startosis_packages.PackageAbsoluteLocator, packageReplaceOptions map[string]string) (bool, string, string, string) {
	if len(packageReplaceOptions) == 0 {
		return false, "", "", ""
	}

	pathToAnalyze := absoluteLocator.GetLocator()
	for {
		numberSlashes := strings.Count(pathToAnalyze, shared_utils.UrlPathSeparator)

		// check for the minimal path e.g.: github.com/org/package
		if numberSlashes < shared_utils.MinimumSubPathsForValidGitURL {
			break
		}
		lastIndex := strings.LastIndex(pathToAnalyze, shared_utils.UrlPathSeparator)

		packageToBeReplaced := pathToAnalyze[:lastIndex]
		replacePackageWithMaybeWitBranchOrCommit, ok := packageReplaceOptions[packageToBeReplaced]
		if ok {
			replaceWithPackage, maybeTagBranchOrCommit := shared_utils.ParseOutTagBranchOrCommit(replacePackageWithMaybeWitBranchOrCommit)
			logrus.Debugf("dependency replace found for '%s', package '%s' will be replaced with '%s'", absoluteLocator, packageToBeReplaced, replaceWithPackage)
			return true, packageToBeReplaced, replaceWithPackage, maybeTagBranchOrCommit
		}

		pathToAnalyze = packageToBeReplaced
	}

	return false, "", "", ""
}
