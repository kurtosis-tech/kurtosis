package git_module_content_provider

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

func (moduleManager *GitModuleManager) GetModuleContentProvider(moduleURL string) (string, error) {
	parsedURL, err := parseGitURL(moduleURL)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while parsing URL '%v'", moduleURL)
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
		return "", stacktrace.Propagate(err, "An error occurred while cloning the Git Repo '%v'", parsedURL)
	}

	// Load it after cloning
	contents, err = os.ReadFile(pathToStartosisFile)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred in reading contents of the Startosis file '%v'", pathToStartosisFile)
	}

	return string(contents), nil
}

// atomicClone This first clones to a temporary directory and then moves it
// TODO make this support versioning via tags, commit hashes or branches
func (moduleManager *GitModuleManager) atomicClone(parsedURL *ParsedGitURL) error {
	// First we clone into a temporary directory
	tempRepoDirPath, err := os.MkdirTemp(moduleManager.moduleTmpDir, temporaryRepoDirPattern)
	if err != nil {
		return stacktrace.Propagate(err, "Error creating temporary directory for the repository to be cloned into")
	}
	defer os.RemoveAll(tempRepoDirPath)
	gitClonePath := path.Join(tempRepoDirPath, parsedURL.relativeRepoPath)
	_, err = git.PlainClone(gitClonePath, false, &git.CloneOptions{URL: parsedURL.gitURL, Progress: io.Discard})
	if err != nil {
		return stacktrace.Propagate(err, "Error in cloning git repository '%v' to '%v'", parsedURL.gitURL, gitClonePath)
	}

	// Then we move it into the target directory
	moduleAuthorPath := path.Join(moduleManager.moduleDir, parsedURL.moduleAuthor)
	modulePath := path.Join(moduleManager.moduleDir, parsedURL.relativeRepoPath)
	fileMode, err := os.Stat(moduleAuthorPath)
	if err == nil && !fileMode.IsDir() {
		return stacktrace.Propagate(err, "Expected '%v' to be a directory but it is something else", moduleAuthorPath)
	}
	if err != nil {
		if err = os.Mkdir(moduleAuthorPath, moduleDirPermission); err != nil {
			return stacktrace.Propagate(err, "An error occurred while creating the directory '%v'", moduleAuthorPath)
		}
	}
	if err = os.Rename(gitClonePath, modulePath); err != nil {
		return stacktrace.Propagate(err, "An error occurred while moving module at temporary destination '%v' to final destination '%v'", gitClonePath, modulePath)
	}

	return nil
}
