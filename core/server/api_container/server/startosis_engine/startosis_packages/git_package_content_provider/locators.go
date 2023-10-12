package git_package_content_provider

import (
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

func isSamePackageLocalAbsoluteLocator(locator string, parentPackageId string) bool {
	return strings.HasPrefix(locator, parentPackageId)
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
		replacedAbsoluteLocator := strings.Replace(absoluteLocator, packageToBeReplaced, replaceWithPackage, onlyOneReplace)
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
		numberSlashes := strings.Count(pathToAnalyze, urlPathSeparator)

		// check for the minimal path e.g.: github.com/org/package
		if numberSlashes < minimumSubPathsForValidGitURL {
			break
		}
		lastIndex := strings.LastIndex(pathToAnalyze, urlPathSeparator)

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
