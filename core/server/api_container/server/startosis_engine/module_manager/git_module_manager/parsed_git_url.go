package git_module_manager

import (
	"fmt"
	"github.com/kurtosis-tech/stacktrace"
	"net/url"
	"path"
	"strings"
)

const (
	githubDomain           = "github.com"
	httpsSchema            = "https"
	startosisFileExtension = ".star"
)

type ParsedGitURL struct {
	moduleAuthor       string
	moduleName         string
	gitURL             string
	relativeModulePath string
	relativeFilePath   string
}

func newParsedGitURL(moduleAuthor, moduleName, gitURL, relativeModulePath, relativeFilePath string) *ParsedGitURL {
	return &ParsedGitURL{
		moduleAuthor:       moduleAuthor,
		moduleName:         moduleName,
		gitURL:             gitURL,
		relativeModulePath: relativeModulePath,
		relativeFilePath:   relativeFilePath,
	}
}

func parseGitURL(packageURL string) (*ParsedGitURL, error) {
	parsedURL, err := url.Parse(packageURL)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error parsing the url '%v'", packageURL)
	}
	if parsedURL.Scheme != "" {
		return nil, stacktrace.NewError("Expected schema to be empty got '%v'", parsedURL.Scheme)
	}

	packageURLPrefixedWithHttps := httpsSchema + "://" + packageURL
	parsedURL, err = url.Parse(packageURLPrefixedWithHttps)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error parsing the url '%v'", packageURL)
	}
	if parsedURL.Host != githubDomain {
		return nil, stacktrace.NewError("We only support modules on Github for now")
	}

	splitURLPath := removeEmptyStringsFromSlice(strings.Split(parsedURL.Path, "/"))

	if len(splitURLPath) < 3 {
		return nil, stacktrace.NewError("URL path should contain at least 3 subpaths")
	}

	lastItem := splitURLPath[len(splitURLPath)-1]
	if !strings.HasSuffix(lastItem, startosisFileExtension) {
		return nil, stacktrace.NewError("Expected last subpath to be a '%v' file but it wasn't", startosisFileExtension)
	}

	moduleAuthor := splitURLPath[0]
	moduleName := splitURLPath[1]
	gitURL := fmt.Sprintf("%v://%v/%v/%v.git", httpsSchema, githubDomain, moduleAuthor, moduleName)
	relativeModulePath := path.Join(moduleAuthor, moduleName)
	relativeFilePath := path.Join(splitURLPath...)

	parsedGitURL := newParsedGitURL(
		moduleAuthor,
		moduleName,
		gitURL,
		relativeModulePath,
		relativeFilePath,
	)

	return parsedGitURL, nil
}

func removeEmptyStringsFromSlice(stringSlice []string) []string {
	var sliceWithoutEmptyStrings []string
	for _, subPath := range stringSlice {
		if subPath != "" {
			sliceWithoutEmptyStrings = append(sliceWithoutEmptyStrings, subPath)
		}
	}
	return sliceWithoutEmptyStrings
}
