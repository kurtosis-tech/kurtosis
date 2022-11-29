package git_package_content_provider

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"net/url"
	"path"
	"strings"
)

const (
	githubDomain     = "github.com"
	httpsSchema      = "https"
	urlPathSeparator = "/"
	// for a valid GitURl we need it to look like github.com/author/moduleName
	// the last two are the minimum requirements for a valid Startosis URL
	minimumSubPathsForValidGitURL = 2
)

// ParsedGitURL an object representing a parsed moduleURL
type ParsedGitURL struct {
	// moduleAuthor the git of the module (GitHub user or org)
	moduleAuthor string
	// moduleName the name of the module
	moduleName string
	// gitURL the url ending with `.git` where the module lives
	gitURL string
	// relativeRepoPath the relative path to the repo this would be moduleAuthor/moduleName/
	relativeRepoPath string
	// relativeFilePath the full path of the file relative to the module store relativeRepoPath/path/to/file.star
	// empty if there is no file
	relativeFilePath string
}

func newParsedGitURL(moduleAuthor, moduleName, gitURL, relativeRepoPath, relativeFilePath string) *ParsedGitURL {
	return &ParsedGitURL{
		moduleAuthor:     moduleAuthor,
		moduleName:       moduleName,
		gitURL:           gitURL,
		relativeRepoPath: relativeRepoPath,
		relativeFilePath: relativeFilePath,
	}
}

// parseGitURL this takes a Git url (GitHub) for now and converts it into the struct ParsedGitURL
// This can in the future be extended to GitLab or BitBucket or any other Git Host
func parseGitURL(packageURL string) (*ParsedGitURL, *startosis_errors.InterpretationError) {
	// we expect something like github.com/author/module/path.star
	// we don't want schemas
	parsedURL, err := url.Parse(packageURL)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Error parsing the URL of module '%v'", packageURL)
	}
	if parsedURL.Scheme != "" {
		return nil, startosis_errors.NewInterpretationError("Error parsing the URL of module '%v'. Expected schema to be empty got '%v'", packageURL, parsedURL.Scheme)
	}

	// we prefix schema and make sure that the URL still parses
	packageURLPrefixedWithHttps := httpsSchema + "://" + packageURL
	parsedURL, err = url.Parse(packageURLPrefixedWithHttps)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Error parsing the URL with scheme for module '%v'", packageURLPrefixedWithHttps)
	}
	if parsedURL.Host != githubDomain {
		return nil, startosis_errors.NewInterpretationError("Error parsing the URL of module. We only support modules on Github for now but got '%v'", packageURL)
	}

	splitURLPath := cleanPathAndSplit(parsedURL.Path)

	if len(splitURLPath) < minimumSubPathsForValidGitURL {
		return nil, startosis_errors.NewInterpretationError("Error parsing the URL of module: '%v'. The path should contain at least %d subpaths got '%v'", packageURL, minimumSubPathsForValidGitURL, splitURLPath)
	}

	moduleAuthor := splitURLPath[0]
	moduleName := splitURLPath[1]
	gitURL := fmt.Sprintf("%v://%v/%v/%v.git", httpsSchema, githubDomain, moduleAuthor, moduleName)
	relativeModulePath := path.Join(moduleAuthor, moduleName)

	relativeFilePath := ""
	if len(splitURLPath) > minimumSubPathsForValidGitURL {
		relativeFilePath = path.Join(splitURLPath...)
	}

	parsedGitURL := newParsedGitURL(
		moduleAuthor,
		moduleName,
		gitURL,
		relativeModulePath,
		relativeFilePath,
	)

	return parsedGitURL, nil
}

// cleanPath removes empty "" from the string slice
func cleanPathAndSplit(urlPath string) []string {
	cleanPath := path.Clean(urlPath)
	splitPath := strings.Split(cleanPath, urlPathSeparator)
	var sliceWithoutEmptyStrings []string
	for _, subPath := range splitPath {
		if subPath != "" {
			sliceWithoutEmptyStrings = append(sliceWithoutEmptyStrings, subPath)
		}
	}
	return sliceWithoutEmptyStrings
}
