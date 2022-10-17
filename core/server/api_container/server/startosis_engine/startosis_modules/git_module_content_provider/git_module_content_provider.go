package git_module_content_provider

import (
	"github.com/go-git/go-git/v5"
	"github.com/kurtosis-tech/stacktrace"
	"io"
	"os"
	"path"
	"strings"
)

const (
	moduleDirPermission                      = 0755
	temporaryRepoDirPattern                  = "tmp-repo-dir-*"
	minimumNumberOfSubPathsInInterpretedFile = 4
	moduleSubPathPosition                    = 4
)

type GitModuleContentProvider struct {
	moduleTmpDir string
	moduleDir    string
}

func NewGitModuleContentProvider(moduleDir string, tmpDir string) *GitModuleContentProvider {
	return &GitModuleContentProvider{
		moduleDir:    moduleDir,
		moduleTmpDir: tmpDir,
	}
}

func (provider *GitModuleContentProvider) GetModuleContents(moduleURL string) (string, error) {
	parsedURL, err := parseGitURL(moduleURL)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while parsing URL '%v'", moduleURL)
	}

	pathToStartosisFile := path.Join(provider.moduleDir, parsedURL.relativeFilePath)

	// Load the file if it already exists
	contents, err := os.ReadFile(pathToStartosisFile)
	if err == nil {
		return string(contents), nil
	}

	// Otherwise Clone It
	err = provider.atomicClone(parsedURL)
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

func (provider *GitModuleContentProvider) GetFileAtRelativePath(fileBeingInterpreted string, relFilepathOfFileToRead string) (string, error) {
	absoluteFilePath, err := provider.getAbsolutePath(fileBeingInterpreted, relFilepathOfFileToRead)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while getting the absolute path for file '%v' relative to file being interpreted '%v'", relFilepathOfFileToRead, fileBeingInterpreted)
	}
	fileContents, err := os.ReadFile(absoluteFilePath)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while reading the file '%v'", absoluteFilePath)
	}
	return string(fileContents), nil
}

func (provider *GitModuleContentProvider) IsGithubPath(path string) bool {
	if strings.HasPrefix(path, githubDomain) {
		return true
	}
	return false
}

func (provider *GitModuleContentProvider) getAbsolutePath(fileBeingInterpreted string, relFilepathOfFileToRead string) (string, error) {
	if !strings.HasPrefix(fileBeingInterpreted, provider.moduleDir) {
		return "", stacktrace.NewError("File being interpreted '%v' seems to have an illegal path. This is a bug in Kurtosis.", fileBeingInterpreted)
	}
	fileBeingInterpretedSplit := cleanPathAndSplit(fileBeingInterpreted)
	if len(fileBeingInterpretedSplit) < minimumNumberOfSubPathsInInterpretedFile {
		return "", stacktrace.NewError("File being interpreted '%v' seems to have an illegal path. This is a bug in Kurtosis.", fileBeingInterpreted)
	}
	dirNameOfFileBeingInterpreted := path.Dir(fileBeingInterpreted)
	absPathOfFileToRead := path.Join(dirNameOfFileBeingInterpreted, relFilepathOfFileToRead)
	absPathOfFileToRead = path.Clean(absPathOfFileToRead)

	moduleDirOfInterpretedFile := string(os.PathSeparator) + path.Join(fileBeingInterpretedSplit[0:moduleSubPathPosition]...)
	if !strings.HasPrefix(absPathOfFileToRead, moduleDirOfInterpretedFile) {
		return "", stacktrace.NewError("Final path of file '%v' seems to be outside of module '%v', which in unsafe.", absPathOfFileToRead, dirNameOfFileBeingInterpreted)
	}
	return absPathOfFileToRead, nil
}

// atomicClone This first clones to a temporary directory and then moves it
// TODO make this support versioning via tags, commit hashes or branches
func (provider *GitModuleContentProvider) atomicClone(parsedURL *ParsedGitURL) error {
	// First we clone into a temporary directory
	tempRepoDirPath, err := os.MkdirTemp(provider.moduleTmpDir, temporaryRepoDirPattern)
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
	moduleAuthorPath := path.Join(provider.moduleDir, parsedURL.moduleAuthor)
	modulePath := path.Join(provider.moduleDir, parsedURL.relativeRepoPath)
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
