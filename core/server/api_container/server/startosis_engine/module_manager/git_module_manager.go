package module_manager

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/kurtosis-tech/stacktrace"
	"io"
	"net/url"
	"os"
	"path"
	"strings"
)

const (
	moduleDirPermission     = 0755
	temporaryRepoDirPattern = "tmp-repo-dir-*"
	githubDomain            = "github.com"
	httpsSchema             = "https"
)

type GitModuleManager struct {
	moduleTmpDir string
	moduleDir    string
	gitURL       string
}

type ParsedGitURL struct {
	moduleAuthor       string
	moduleName         string
	gitURL             string
	relativeModulePath string
	relativeFilePath   string
}

func NewParsedGITURL(moduleAuthor, moduleName, gitURL, relativeModulePath, relativeFilePath string) *ParsedGitURL {
	return &ParsedGitURL{
		moduleAuthor:       moduleAuthor,
		moduleName:         moduleName,
		gitURL:             gitURL,
		relativeModulePath: relativeModulePath,
		relativeFilePath:   relativeFilePath,
	}
}

func NewGitModuleManager(moduleDir string, tmpDir string) *GitModuleManager {
	return &GitModuleManager{
		moduleDir:    moduleDir,
		moduleTmpDir: tmpDir,
	}
}

func (moduleManager *GitModuleManager) GetModule(packageURL string) (string, error) {
	parsedURL, err := parseGitURL(packageURL)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while parsing URL")
	}

	pathToStartosisFile := path.Join(moduleManager.moduleDir, parsedURL.relativeFilePath)

	contents, err := os.ReadFile(pathToStartosisFile)
	if err == nil {
		return string(contents), nil
	}

	tempRepoDirPath, err := os.MkdirTemp(moduleManager.moduleTmpDir, temporaryRepoDirPattern)
	if err != nil {
		return "", stacktrace.Propagate(err, "Error creating temporary directory for the repository to be cloned into")
	}
	defer os.RemoveAll(tempRepoDirPath)
	gitClonePath := path.Join(tempRepoDirPath, parsedURL.relativeModulePath)
	_, err = git.PlainClone(gitClonePath, false, &git.CloneOptions{URL: parsedURL.gitURL, Progress: io.Discard})
	if err != nil {
		return "", stacktrace.Propagate(err, "Error in cloning git repository '%v' to '%v'", parsedURL.gitURL, gitClonePath)
	}
	moduleAuthorPath := path.Join(moduleManager.moduleDir, parsedURL.moduleAuthor)
	modulePath := path.Join(moduleManager.moduleDir, parsedURL.relativeModulePath)
	_, err = os.Stat(moduleAuthorPath)
	if err != nil {
		if err = os.Mkdir(moduleAuthorPath, moduleDirPermission); err != nil {
			stacktrace.Propagate(err, "An error occurred while creating directory '%v'", moduleAuthorPath)
		}
	}
	if err = os.Rename(gitClonePath, modulePath); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while moving module at temporary destination to final destination")
	}

	contents, err = os.ReadFile(pathToStartosisFile)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred in reading contents of the StarLark file")
	}

	return string(contents), nil
}

func parseGitURL(packageURL string) (*ParsedGitURL, error) {
	parsedUrl, err := url.Parse(packageURL)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error parsing the url '%v'", packageURL)
	}
	if parsedUrl.Scheme != httpsSchema {
		return nil, stacktrace.NewError("Expected the scheme to be 'https' got '%v'", parsedUrl.Scheme)
	}
	if parsedUrl.Host != githubDomain {
		return nil, stacktrace.NewError("We only support modules on Github for now")
	}

	splitURLPath := removeEmptyStringsFromSlice(strings.Split(parsedUrl.Path, "/"))

	if len(splitURLPath) < 2 {
		return nil, stacktrace.NewError("URL path should contain at least 2 parts")
	}

	moduleAuthor := splitURLPath[0]
	moduleName := splitURLPath[1]
	gitURL := fmt.Sprintf("%v://%v/%v/%v.git", httpsSchema, githubDomain, moduleAuthor, moduleName)
	relativeModulePath := path.Join(moduleAuthor, moduleName)
	relativeFilePath := path.Join(splitURLPath...)

	parsedGitURL := NewParsedGITURL(
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
