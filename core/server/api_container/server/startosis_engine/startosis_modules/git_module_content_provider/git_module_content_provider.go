package git_module_content_provider

import (
	"github.com/go-git/go-git/v5"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/mholt/archiver"
	"io"
	"os"
	"path"
)

const (
	moduleDirPermission         = 0755
	temporaryRepoDirPattern     = "tmp-repo-dir-*"
	temporaryArchiveFilePattern = "temp-module-archive-*.tgz"
	defaultTmpDir               = ""
)

type GitModuleContentProvider struct {
	modulesTmpDir string
	modulesDir    string
}

func NewGitModuleContentProvider(moduleDir string, tmpDir string) *GitModuleContentProvider {
	return &GitModuleContentProvider{
		modulesDir:    moduleDir,
		modulesTmpDir: tmpDir,
	}
}

func (provider *GitModuleContentProvider) GetOnDiskAbsoluteFilePath(fileInsideModuleUrl string) (string, *startosis_errors.InterpretationError) {
	parsedURL, interpretationError := parseGitURL(fileInsideModuleUrl)
	if interpretationError != nil {
		return "", interpretationError
	}
	if parsedURL.relativeFilePath == "" {
		return "", startosis_errors.NewInterpretationError("The relative path to file is empty for '%v'", fileInsideModuleUrl)
	}
	pathToFile := path.Join(provider.modulesDir, parsedURL.relativeFilePath)

	// Return the file path straight if it exists
	if _, err := os.Stat(pathToFile); err == nil {
		return pathToFile, nil
	}

	// Otherwise clone the repo and return the absolute path of the requested file
	interpretationError = provider.atomicClone(parsedURL)
	if interpretationError != nil {
		return "", interpretationError
	}
	return pathToFile, nil
}

func (provider *GitModuleContentProvider) GetModuleContents(fileInsideModuleUrl string) (string, *startosis_errors.InterpretationError) {
	pathToFile, interpretationError := provider.GetOnDiskAbsoluteFilePath(fileInsideModuleUrl)
	if interpretationError != nil {
		return "", interpretationError
	}

	// Load the file content from its absolute path
	contents, err := os.ReadFile(pathToFile)
	if err != nil {
		return "", startosis_errors.WrapError(err, "Loading module content for module '%s' failed. An error occurred in reading contents of the file '%v'", fileInsideModuleUrl, pathToFile)
	}

	return string(contents), nil
}

func (provider *GitModuleContentProvider) StoreModuleContents(moduleId string, moduleTar []byte, overwriteExisting bool) (string, *startosis_errors.InterpretationError) {
	parsedModuleId, interpretationError := parseGitURL(moduleId)
	if interpretationError != nil {
		return "", interpretationError
	}
	modulePathOnDisk := path.Join(provider.modulesDir, parsedModuleId.relativeRepoPath)

	if overwriteExisting {
		err := os.RemoveAll(modulePathOnDisk)
		if err != nil {
			return "", startosis_errors.WrapError(err, "An error occurred while removing the existing module '%v' from disk at '%v'", moduleId, modulePathOnDisk)
		}
	}

	_, err := os.Stat(modulePathOnDisk)
	if err == nil {
		return "", startosis_errors.NewInterpretationError("Module '%v' already exists on disk, not overwriting", modulePathOnDisk)
	}

	tempFile, err := os.CreateTemp(defaultTmpDir, temporaryArchiveFilePattern)
	if err != nil {
		return "", startosis_errors.NewInterpretationError("An error occurred while creating temporary file to write compressed '%v' to", moduleId)
	}
	defer os.Remove(tempFile.Name())

	bytesWritten, err := tempFile.Write(moduleTar)
	if err != nil {
		return "", startosis_errors.WrapError(err, "An error occurred while writing contents of '%v' to '%v'", moduleId, tempFile.Name())
	}
	if bytesWritten != len(moduleTar) {
		return "", startosis_errors.NewInterpretationError("Expected to write '%v' bytes but wrote '%v'", len(moduleTar), bytesWritten)
	}
	err = archiver.Unarchive(tempFile.Name(), modulePathOnDisk)
	if err != nil {
		return "", startosis_errors.WrapError(err, "An error occurred while unarchiving '%v' to '%v'", tempFile.Name(), modulePathOnDisk)
	}

	return modulePathOnDisk, nil
}

// atomicClone This first clones to a temporary directory and then moves it
// TODO make this support versioning via tags, commit hashes or branches
func (provider *GitModuleContentProvider) atomicClone(parsedURL *ParsedGitURL) *startosis_errors.InterpretationError {
	// First we clone into a temporary directory
	tempRepoDirPath, err := os.MkdirTemp(provider.modulesTmpDir, temporaryRepoDirPattern)
	if err != nil {
		return startosis_errors.WrapError(err, "Cloning the module '%s' failed. Error creating temporary directory for the repository to be cloned into", parsedURL.gitURL)
	}
	defer os.RemoveAll(tempRepoDirPath)
	gitClonePath := path.Join(tempRepoDirPath, parsedURL.relativeRepoPath)
	_, err = git.PlainClone(gitClonePath, false, &git.CloneOptions{URL: parsedURL.gitURL, Progress: io.Discard})
	if err != nil {
		return startosis_errors.WrapError(err, "Error in cloning git repository '%s' to '%s'", parsedURL.gitURL, gitClonePath)
	}

	// Then we move it into the target directory
	moduleAuthorPath := path.Join(provider.modulesDir, parsedURL.moduleAuthor)
	modulePath := path.Join(provider.modulesDir, parsedURL.relativeRepoPath)
	fileMode, err := os.Stat(moduleAuthorPath)
	if err == nil && !fileMode.IsDir() {
		return startosis_errors.WrapError(err, "Expected '%s' to be a directory but it is something else", moduleAuthorPath)
	}
	if err != nil {
		if err = os.Mkdir(moduleAuthorPath, moduleDirPermission); err != nil {
			return startosis_errors.WrapError(err, "Cloning the module '%s' failed. An error occurred while creating the directory '%s'.", parsedURL.gitURL, moduleAuthorPath)
		}
	}
	if err = os.Rename(gitClonePath, modulePath); err != nil {
		return startosis_errors.NewInterpretationError("Cloning the module '%s' failed. An error occurred while moving module at temporary destination '%s' to final destination '%s'", parsedURL.gitURL, gitClonePath, modulePath)
	}
	return nil
}
