package git_package_content_provider

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
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

func replaceAbsoluteLocator(absoluteLocator string, packageReplaceOptions map[string]string) (string, *startosis_errors.InterpretationError) {
	if absoluteLocator == "" {
		return absoluteLocator, nil
	}

	found, packageToBeReplaced, replaceWithPackage, interpretationErr := findPackageReplace(absoluteLocator, packageReplaceOptions)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	if found {
		// we skip if it's a local replace because we will use the same absolute locator
		// due the file was already uploaded in the enclave's package cache
		if isLocalLocator(replaceWithPackage) {
			return absoluteLocator, nil
		}
		replacedAbsoluteLocator := strings.Replace(absoluteLocator, packageToBeReplaced, replaceWithPackage, onlyOneReplace)
		logrus.Debugf("absoluteLocator '%s' replaced with '%s'", absoluteLocator, replacedAbsoluteLocator)
		return replacedAbsoluteLocator, nil
	}

	return absoluteLocator, nil
}

func findPackageReplace(absoluteLocator string, packageReplaceOptions map[string]string) (bool, string, string, *startosis_errors.InterpretationError) {
	if len(packageReplaceOptions) == 0 {
		return false, "", "", nil
	}

	urlToAnalyze, interpretationErr := parseGitURL(absoluteLocator)
	if interpretationErr != nil {
		return false, "", "", interpretationErr
	}
	gitUrl := urlToAnalyze.gitURL

	for {

		lastIndex := strings.LastIndex(gitUrl, urlPathSeparator)

		if len(gitUrl) <= lastIndex || lastIndex == subStrNotPresentIndicator {
			break
		}
		packageToBeReplaced := gitUrl[:lastIndex]
		replaceWithPackage, ok := packageReplaceOptions[packageToBeReplaced]
		if ok {
			logrus.Debugf("dependency replace found for '%s', package '%s' will be replaced with '%s'", absoluteLocator, packageToBeReplaced, replaceWithPackage)
			return true, packageToBeReplaced, replaceWithPackage, nil
		}

		gitUrl = packageToBeReplaced
	}

	return false, "", "", nil
}
