package git_package_content_provider

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/shared_utils"
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

	isSourceModuleInRootPackage := strings.HasPrefix(sourceModuleLocator, rootPackageId)
	isAbsoluteLocatorInRootPackage := strings.HasPrefix(relativeOrAbsoluteLocator, rootPackageId)

	return isSourceModuleInRootPackage && isAbsoluteLocatorInRootPackage
}

func replaceAbsoluteLocator(absoluteLocator string, packageReplaceOptions map[string]string) string {
	if absoluteLocator == "" {
		return absoluteLocator
	}

	found, packageToBeReplaced, replaceWithPackage := findPackageReplace(absoluteLocator, packageReplaceOptions)

	if found {
		// we skip if it's a local replace because we will use the same absolute locator
		// due the file was already uploaded in the enclave's package cache
		if isLocalLocator(replaceWithPackage) {
			return absoluteLocator
		}
		replacedAbsoluteLocatorMaybeWitBranchOrCommit := strings.Replace(absoluteLocator, packageToBeReplaced, replaceWithPackage, onlyOneReplace)

		replacedAbsoluteLocator, _ := shared_utils.ParseOutTagBranchOrCommit(replacedAbsoluteLocatorMaybeWitBranchOrCommit)

		logrus.Debugf("absoluteLocator '%s' replaced with '%s'", absoluteLocator, replacedAbsoluteLocator)
		return replacedAbsoluteLocator
	}

	return absoluteLocator
}

func findPackageReplace(absoluteLocator string, packageReplaceOptions map[string]string) (bool, string, string) {
	if len(packageReplaceOptions) == 0 {
		return false, "", ""
	}

	pathToAnalyze := absoluteLocator
	for {
		numberSlashes := strings.Count(pathToAnalyze, shared_utils.UrlPathSeparator)

		// check for the minimal path e.g.: github.com/org/package
		if numberSlashes < shared_utils.MinimumSubPathsForValidGitURL {
			break
		}
		lastIndex := strings.LastIndex(pathToAnalyze, shared_utils.UrlPathSeparator)

		packageToBeReplaced := pathToAnalyze[:lastIndex]
		replaceWithPackage, ok := packageReplaceOptions[packageToBeReplaced]
		if ok {
			logrus.Debugf("dependency replace found for '%s', package '%s' will be replaced with '%s'", absoluteLocator, packageToBeReplaced, replaceWithPackage)
			return true, packageToBeReplaced, replaceWithPackage
		}

		pathToAnalyze = packageToBeReplaced
	}

	return false, "", ""
}
