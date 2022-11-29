package git_package_content_provider

import (
	"github.com/go-git/go-git/v5"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/mholt/archiver"
	"io"
	"os"
	"path"
	"strings"
)

const (
	moduleDirPermission         = 0755
	temporaryRepoDirPattern     = "tmp-repo-dir-*"
	temporaryArchiveFilePattern = "temp-module-archive-*.tgz"
	defaultTmpDir               = ""

	onlyOneReplacement      = 1
	replacedWithEmptyString = ""
)

type GitPackageContentProvider struct {
	packagesTmpDir string
	packagesDir    string
}

func NewGitPackageContentProvider(moduleDir string, tmpDir string) *GitPackageContentProvider {
	return &GitPackageContentProvider{
		packagesDir:    moduleDir,
		packagesTmpDir: tmpDir,
	}
}

func (provider *GitPackageContentProvider) ClonePackage(moduleId string) (string, *startosis_errors.InterpretationError) {
	parsedURL, interpretationError := parseGitURL(moduleId)
	if interpretationError != nil {
		return "", interpretationError
	}

	moduleAbsolutePathOnDisk := path.Join(provider.packagesDir, parsedURL.relativeRepoPath)

	interpretationError = provider.atomicClone(parsedURL)
	if interpretationError != nil {
		return "", interpretationError
	}
	return moduleAbsolutePathOnDisk, nil
}

func (provider *GitPackageContentProvider) GetOnDiskAbsoluteFilePath(fileInsideModuleUrl string) (string, *startosis_errors.InterpretationError) {
	parsedURL, interpretationError := parseGitURL(fileInsideModuleUrl)
	if interpretationError != nil {
		return "", interpretationError
	}
	if parsedURL.relativeFilePath == "" {
		return "", startosis_errors.NewInterpretationError("The relative path to file is empty for '%v'", fileInsideModuleUrl)
	}
	pathToFile := path.Join(provider.packagesDir, parsedURL.relativeFilePath)
	packagePath := path.Join(provider.packagesDir, parsedURL.relativeRepoPath)

	// Return the file path straight if it exists
	if _, err := os.Stat(pathToFile); err == nil {
		return pathToFile, nil
	}

	// Check if the repo exists
	// If the repo exists but the `pathToFile` doesn't that means there's a mistake in the locator
	if _, err := os.Stat(packagePath); err == nil {
		relativeFilePathWithoutPackageName := strings.Replace(parsedURL.relativeFilePath, parsedURL.relativeRepoPath, replacedWithEmptyString, onlyOneReplacement)
		return "", startosis_errors.NewInterpretationError("'%v' doesn't exist in the package '%v'", relativeFilePathWithoutPackageName, parsedURL.relativeRepoPath)
	}

	// Otherwise clone the repo and return the absolute path of the requested file
	interpretationError = provider.atomicClone(parsedURL)
	if interpretationError != nil {
		return "", interpretationError
	}
	return pathToFile, nil
}

func (provider *GitPackageContentProvider) GetPackageContents(fileInsideModuleUrl string) (string, *startosis_errors.InterpretationError) {
	pathToFile, interpretationError := provider.GetOnDiskAbsoluteFilePath(fileInsideModuleUrl)
	if interpretationError != nil {
		return "", interpretationError
	}

	// Load the file content from its absolute path
	contents, err := os.ReadFile(pathToFile)
	if err != nil {
		return "", startosis_errors.WrapWithInterpretationError(err, "Loading module content for module '%s' failed. An error occurred in reading contents of the file '%v'", fileInsideModuleUrl, pathToFile)
	}

	return string(contents), nil
}

func (provider *GitPackageContentProvider) StorePackageContents(packageId string, moduleTar []byte, overwriteExisting bool) (string, *startosis_errors.InterpretationError) {
	parsedPackageId, interpretationError := parseGitURL(packageId)
	if interpretationError != nil {
		return "", interpretationError
	}
	packagePathOnDisk := path.Join(provider.packagesDir, parsedPackageId.relativeRepoPath)

	if overwriteExisting {
		err := os.RemoveAll(packagePathOnDisk)
		if err != nil {
			return "", startosis_errors.WrapWithInterpretationError(err, "An error occurred while removing the existing module '%v' from disk at '%v'", packageId, packagePathOnDisk)
		}
	}

	_, err := os.Stat(packagePathOnDisk)
	if err == nil {
		return "", startosis_errors.NewInterpretationError("Module '%v' already exists on disk, not overwriting", packagePathOnDisk)
	}

	tempFile, err := os.CreateTemp(defaultTmpDir, temporaryArchiveFilePattern)
	if err != nil {
		return "", startosis_errors.NewInterpretationError("An error occurred while creating temporary file to write compressed '%v' to", packageId)
	}
	defer os.Remove(tempFile.Name())

	bytesWritten, err := tempFile.Write(moduleTar)
	if err != nil {
		return "", startosis_errors.WrapWithInterpretationError(err, "An error occurred while writing contents of '%v' to '%v'", packageId, tempFile.Name())
	}
	if bytesWritten != len(moduleTar) {
		return "", startosis_errors.NewInterpretationError("Expected to write '%v' bytes but wrote '%v'", len(moduleTar), bytesWritten)
	}
	err = archiver.Unarchive(tempFile.Name(), packagePathOnDisk)
	if err != nil {
		return "", startosis_errors.WrapWithInterpretationError(err, "An error occurred while unarchiving '%v' to '%v'", tempFile.Name(), packagePathOnDisk)
	}

	return packagePathOnDisk, nil
}

// atomicClone This first clones to a temporary directory and then moves it
// TODO make this support versioning via tags, commit hashes or branches
func (provider *GitPackageContentProvider) atomicClone(parsedURL *ParsedGitURL) *startosis_errors.InterpretationError {
	// First we clone into a temporary directory
	tempRepoDirPath, err := os.MkdirTemp(provider.packagesTmpDir, temporaryRepoDirPattern)
	if err != nil {
		return startosis_errors.WrapWithInterpretationError(err, "Cloning the module '%s' failed. Error creating temporary directory for the repository to be cloned into", parsedURL.gitURL)
	}
	defer os.RemoveAll(tempRepoDirPath)
	gitClonePath := path.Join(tempRepoDirPath, parsedURL.relativeRepoPath)
	_, err = git.PlainClone(gitClonePath, false, &git.CloneOptions{
		URL:               parsedURL.gitURL,
		Auth:              nil,
		RemoteName:        "",
		ReferenceName:     "",
		SingleBranch:      false,
		NoCheckout:        false,
		Depth:             0,
		RecurseSubmodules: 0,
		Progress:          io.Discard,
		Tags:              0,
		InsecureSkipTLS:   false,
		CABundle:          nil,
	})
	if err != nil {
		// TODO remove public repository from error after we support private repositories
		return startosis_errors.WrapWithInterpretationError(err, "Error in cloning git repository '%s' to '%s'. Make sure that '%v' exists and is a public repository.", parsedURL.gitURL, gitClonePath, parsedURL.gitURL)
	}

	// Then we move it into the target directory
	packageAuthorPath := path.Join(provider.packagesDir, parsedURL.moduleAuthor)
	packagePath := path.Join(provider.packagesDir, parsedURL.relativeRepoPath)
	fileMode, err := os.Stat(packageAuthorPath)
	if err == nil && !fileMode.IsDir() {
		return startosis_errors.WrapWithInterpretationError(err, "Expected '%s' to be a directory but it is something else", packageAuthorPath)
	}
	if err != nil {
		if err = os.Mkdir(packageAuthorPath, moduleDirPermission); err != nil {
			return startosis_errors.WrapWithInterpretationError(err, "Cloning the package '%s' failed. An error occurred while creating the directory '%s'.", parsedURL.gitURL, packageAuthorPath)
		}
	}
	if err = os.Rename(gitClonePath, packagePath); err != nil {
		return startosis_errors.NewInterpretationError("Cloning the package '%s' failed. An error occurred while moving package at temporary destination '%s' to final destination '%s'", parsedURL.gitURL, gitClonePath, packagePath)
	}
	return nil
}
