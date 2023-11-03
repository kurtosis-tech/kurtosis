package shared_utils

import (
	"fmt"
	"github.com/kurtosis-tech/stacktrace"
	"net/url"
	"path"
	"strings"
)

const (
	GithubDomainPrefix = "github.com"
	httpsSchema        = "https"
	UrlPathSeparator   = "/"
	//MinimumSubPathsForValidGitURL for a valid GitURl we need it to look like github.com/author/moduleName
	// the last two are the minimum requirements for a valid Startosis URL
	MinimumSubPathsForValidGitURL = 2

	tagBranchOrCommitDelimiter = "@"
	emptyTagBranchOrCommit     = ""

	packageRootPrefixIndicatorInRelativeLocators = "/"
	substrNotPresent                             = -1
	extensionCharacter                           = "."
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

	// if the URL contains an @ then we treat anything after that as a tag, branch or commit
	// in that order
	tagBranchOrCommit string
}

func newParsedGitURL(moduleAuthor, moduleName, gitURL, relativeRepoPath, relativeFilePath string, tagBranchOrCommit string) *ParsedGitURL {
	return &ParsedGitURL{
		moduleAuthor:      moduleAuthor,
		moduleName:        moduleName,
		gitURL:            gitURL,
		relativeRepoPath:  relativeRepoPath,
		relativeFilePath:  relativeFilePath,
		tagBranchOrCommit: tagBranchOrCommit,
	}
}

func (parsedUrl *ParsedGitURL) GetModuleAuthor() string {
	return parsedUrl.moduleAuthor
}

func (parsedUrl *ParsedGitURL) GetModuleName() string {
	return parsedUrl.moduleName
}

func (parsedUrl *ParsedGitURL) GetGitURL() string {
	return parsedUrl.gitURL
}

func (parsedUrl *ParsedGitURL) GetRelativeRepoPath() string {
	return parsedUrl.relativeRepoPath
}

func (parsedUrl *ParsedGitURL) GetRelativeFilePath() string {
	return parsedUrl.relativeFilePath
}

func (parsedUrl *ParsedGitURL) GetTagBranchOrCommit() string {
	return parsedUrl.tagBranchOrCommit
}

func (parsedUrl *ParsedGitURL) GetAbsoluteLocatorRelativeToThisURL(relativeUrl string) string {
	if strings.HasPrefix(relativeUrl, packageRootPrefixIndicatorInRelativeLocators) {
		return path.Join(GithubDomainPrefix, parsedUrl.relativeRepoPath, relativeUrl)
	}
	return path.Join(GithubDomainPrefix, path.Dir(parsedUrl.relativeFilePath), relativeUrl)
}

// ParseGitURL this takes a Git url (GitHub) for now and converts it into the struct ParsedGitURL
// This can in the future be extended to GitLab or BitBucket or any other Git Host
func ParseGitURL(packageURL string) (*ParsedGitURL, error) {
	// we expect something like github.com/author/module/path.star
	// we don't want schemas
	parsedURL, err := url.Parse(packageURL)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error parsing the URL of module '%v'", packageURL)
	}
	if parsedURL.Scheme != "" {
		return nil, stacktrace.NewError("Error parsing the URL of module '%v'. Expected schema to be empty got '%v'", packageURL, parsedURL.Scheme)
	}

	// we prefix schema and make sure that the URL still parses
	packageURLPrefixedWithHttps := httpsSchema + "://" + packageURL
	parsedURL, err = url.Parse(packageURLPrefixedWithHttps)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error parsing the URL with scheme for module '%v'", packageURLPrefixedWithHttps)
	}
	if strings.ToLower(parsedURL.Host) != GithubDomainPrefix {
		return nil, stacktrace.NewError("Error parsing the URL of module. We only support modules on Github for now but got '%v'", packageURL)
	}

	pathWithoutVersion, maybeTagBranchOrCommit := parseOutTagBranchOrCommit(parsedURL.Path)

	splitURLPath := cleanPathAndSplit(pathWithoutVersion)

	if len(splitURLPath) < MinimumSubPathsForValidGitURL {
		return nil, stacktrace.NewError("Error parsing the URL of module: '%v'. The path should contain at least %d subpaths got '%v'", packageURL, MinimumSubPathsForValidGitURL, splitURLPath)
	}

	moduleAuthor := splitURLPath[0]
	moduleName := splitURLPath[1]
	gitURL := fmt.Sprintf("%v://%v/%v/%v.git", httpsSchema, GithubDomainPrefix, moduleAuthor, moduleName)
	relativeModulePath := path.Join(moduleAuthor, moduleName)

	relativeFilePath := ""
	if len(splitURLPath) > MinimumSubPathsForValidGitURL {
		relativeFilePath = path.Join(splitURLPath...)
	}

	parsedGitURL := newParsedGitURL(
		moduleAuthor,
		moduleName,
		gitURL,
		relativeModulePath,
		relativeFilePath,
		maybeTagBranchOrCommit,
	)

	return parsedGitURL, nil
}

// cleanPath removes empty "" from the string slice
func cleanPathAndSplit(urlPath string) []string {
	cleanPath := path.Clean(urlPath)
	splitPath := strings.Split(cleanPath, UrlPathSeparator)
	var sliceWithoutEmptyStrings []string
	for _, subPath := range splitPath {
		if subPath != "" {
			sliceWithoutEmptyStrings = append(sliceWithoutEmptyStrings, subPath)
		}
	}
	return sliceWithoutEmptyStrings
}

// parseOutTagBranchOrCommit splits the string around "@" and then split the after string around "/"
func parseOutTagBranchOrCommit(input string) (string, string) {
	cleanInput := path.Clean(input)
	pathWithoutVersion, maybeTagBranchOrCommitWithFile, _ := strings.Cut(cleanInput, tagBranchOrCommitDelimiter)

	// input can have been set with version in few diff ways
	// 1- github.com/kurtosis-tech/sample-dependency-package/main.star@branch-or-version (when is called from cli run command)
	// 2- github.com/kurtosis-tech/sample-dependency-package@branch-or-version/main.star (when is declared in the replace section of the kurtosis.yml file)
	// 3- github.com/kurtosis-tech/sample-dependency-package/main.star@foo/bar - here the tag is foo/bar;
	// 4- github.com/kurtosis-tech/sample-dependency-package@foo/bar/mains.tar - here the tag is foo/bar; while file is /kurtosis-tech/sample-dependency-package/main.star
	// we check if there is a file in maybeTagBranchOrCommitWithFile and then add it to pathWithoutVersion
	maybeTagBranchOrCommit, lastSectionOfTagBranchCommitWithFile, _ := cutLast(maybeTagBranchOrCommitWithFile, UrlPathSeparator)

	if lastSectionOfTagBranchCommitWithFile != "" && strings.Contains(lastSectionOfTagBranchCommitWithFile, extensionCharacter) {
		// we assume pathWithoutVersion does not contain a file inside yet
		pathWithoutVersion = path.Join(pathWithoutVersion, lastSectionOfTagBranchCommitWithFile)
	} else if lastSectionOfTagBranchCommitWithFile != "" && !strings.Contains(lastSectionOfTagBranchCommitWithFile, extensionCharacter) {
		maybeTagBranchOrCommit = path.Join(maybeTagBranchOrCommit, lastSectionOfTagBranchCommitWithFile)
	}

	return pathWithoutVersion, maybeTagBranchOrCommit
}

func cutLast(pathToCut string, separator string) (string, string, bool) {
	lastIndex := strings.LastIndex(pathToCut, separator)
	if lastIndex == substrNotPresent {
		return pathToCut, "", false
	}
	return pathToCut[:lastIndex], pathToCut[lastIndex+1:], false
}
