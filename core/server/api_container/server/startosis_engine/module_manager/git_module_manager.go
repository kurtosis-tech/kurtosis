package module_manager

import (
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
)

type GitModuleManager struct {
	moduleTmpDir string
	moduleDir    string
}

func NewGitModuleManager(moduleDir string, tmpDir string) *GitModuleManager {
	return &GitModuleManager{
		moduleDir:    moduleDir,
		moduleTmpDir: tmpDir,
	}
}

func (moduleManager *GitModuleManager) GetModule(packageURL string) (string, error) {
	// TODO move validate URL to parse and return things
	err := validateURL(packageURL)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while validating URL")
	}

	parsedUrl, _ := url.Parse(packageURL)
	splitURLPath := removeEmpty(strings.Split(parsedUrl.Path, "/"))
	contents, err := os.ReadFile(moduleManager.getPathToStartosisFile(splitURLPath))
	if err == nil {
		return string(contents), nil
	}

	firstTwoSubPaths := strings.Join(splitURLPath[:2], "/")
	authorName := splitURLPath[0]
	gitRepo := "https://github.com/" + firstTwoSubPaths

	tempRepoDirPath, err := os.MkdirTemp(moduleManager.moduleTmpDir, temporaryRepoDirPattern)
	if err != nil {
		return "", stacktrace.Propagate(err, "Error creating temporary directory for the repository to be cloned into")
	}
	defer os.RemoveAll(tempRepoDirPath)
	gitClonePath := path.Join(tempRepoDirPath, firstTwoSubPaths)
	_, err = git.PlainClone(gitClonePath, false, &git.CloneOptions{URL: gitRepo, Progress: io.Discard})
	if err != nil {
		return "", stacktrace.Propagate(err, "Error in cloning git repository '%v'", gitRepo)
	}
	moduleAuthorPath := path.Join(moduleManager.moduleDir, authorName)
	modulePath := path.Join(moduleManager.moduleDir, firstTwoSubPaths)
	_, err = os.Stat(moduleAuthorPath)
	if err != nil {
		if err = os.Mkdir(moduleAuthorPath, moduleDirPermission); err != nil {
			stacktrace.Propagate(err, "An error occurred while creating directory '%v'", moduleAuthorPath)
		}
	}
	if err = os.Rename(gitClonePath, modulePath); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while moving module at temporary destination to final destination")
	}

	contents, err = os.ReadFile(moduleManager.getPathToStartosisFile(splitURLPath))
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred in reading contents of the StarLark file")
	}

	return string(contents), nil
}

func (moduleManager *GitModuleManager) getPathToStartosisFile(splitUrlPath []string) string {
	splitUrlPath = append([]string{moduleManager.moduleDir}, splitUrlPath...)
	filePath := path.Join(splitUrlPath...)
	return filePath
}

func validateURL(packageURL string) error {
	parsedUrl, err := url.Parse(packageURL)
	if err != nil {
		return stacktrace.Propagate(err, "Error parsing the url '%v'", packageURL)
	}
	if parsedUrl.Scheme != "https" {
		return stacktrace.NewError("Expected the scheme to be 'https' got '%v'", parsedUrl.Scheme)
	}
	if parsedUrl.Host != "github.com" {
		return stacktrace.NewError("We only support modules on Github for now")
	}

	splitURLPath := removeEmpty(strings.Split(parsedUrl.Path, "/"))

	if len(splitURLPath) < 2 {
		return stacktrace.NewError("URL path should contain at least 2 parts")
	}

	return nil
}

func removeEmpty(splitPath []string) []string {
	var splitWithoutEmpties []string
	for _, subPath := range splitPath {
		if subPath != "" {
			splitWithoutEmpties = append(splitWithoutEmpties, subPath)
		}
	}
	return splitWithoutEmpties
}
