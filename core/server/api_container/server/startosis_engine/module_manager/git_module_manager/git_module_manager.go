package git_module_manager

import (
	"github.com/go-git/go-git/v5"
	"github.com/kurtosis-tech/stacktrace"
	"io"
	"os"
	"path"
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
	parsedURL, err := parseGitURL(packageURL)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while parsing URL")
	}

	pathToStartosisFile := path.Join(moduleManager.moduleDir, parsedURL.relativeFilePath)

	// Load the file if it already exists
	contents, err := os.ReadFile(pathToStartosisFile)
	if err == nil {
		return string(contents), nil
	}

	// Otherwise Clone It
	err = moduleManager.atomicClone(parsedURL)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while cloning the Git Repo")
	}

	// Load it after cloning
	contents, err = os.ReadFile(pathToStartosisFile)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred in reading contents of the StarLark file")
	}

	return string(contents), nil
}

// atomicClone This first clones to a temporary directory and then moves it
func (moduleManager *GitModuleManager) atomicClone(parsedURL *ParsedGitURL) error {
	tempRepoDirPath, err := os.MkdirTemp(moduleManager.moduleTmpDir, temporaryRepoDirPattern)
	if err != nil {
		return stacktrace.Propagate(err, "Error creating temporary directory for the repository to be cloned into")
	}
	defer os.RemoveAll(tempRepoDirPath)
	gitClonePath := path.Join(tempRepoDirPath, parsedURL.relativeModulePath)
	_, err = git.PlainClone(gitClonePath, false, &git.CloneOptions{URL: parsedURL.gitURL, Progress: io.Discard})
	if err != nil {
		return stacktrace.Propagate(err, "Error in cloning git repository '%v' to '%v'", parsedURL.gitURL, gitClonePath)
	}

	moduleAuthorPath := path.Join(moduleManager.moduleDir, parsedURL.moduleAuthor)
	modulePath := path.Join(moduleManager.moduleDir, parsedURL.relativeModulePath)
	_, err = os.Stat(moduleAuthorPath)
	if err != nil {
		if err = os.Mkdir(moduleAuthorPath, moduleDirPermission); err != nil {
			return stacktrace.Propagate(err, "An error occurred while creating directory '%v'", moduleAuthorPath)
		}
	}
	if err = os.Rename(gitClonePath, modulePath); err != nil {
		return stacktrace.Propagate(err, "An error occurred while moving module at temporary destination to final destination")
	}

	return nil
}