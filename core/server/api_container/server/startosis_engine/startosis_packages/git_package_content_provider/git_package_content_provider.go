package git_package_content_provider

import (
	"errors"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_constants"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/mholt/archiver"
	"github.com/sirupsen/logrus"
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

	onlyOneReplacement       = 1
	replacedWithEmptyString  = ""
	relativeFilePathNotFound = ""
	isNotBareClone           = false

	// this doesn't get us the entire history, speeding us up
	defaultDepth = 1
	// this gets us the entire history - useful for fetching commits on a repo
	depthAssumingBranchTagsCommitsAreSpecified = 0
	howImportWorksLink                         = "https://docs.kurtosis.com/explanations/how-do-kurtosis-imports-work"
	filePathToKurtosisYamlNotFound             = ""
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

	relPackagePathToPackagesDir := getPathToPackageRoot(parsedURL)
	packageAbsolutePathOnDisk := path.Join(provider.packagesDir, relPackagePathToPackagesDir)

	interpretationError = provider.atomicClone(parsedURL)
	if interpretationError != nil {
		return "", interpretationError
	}
	return packageAbsolutePathOnDisk, nil
}

func (provider *GitPackageContentProvider) GetOnDiskAbsoluteFilePath(fileInsidePackageUrl string) (string, *startosis_errors.InterpretationError) {
	parsedURL, interpretationError := parseGitURL(fileInsidePackageUrl)
	if interpretationError != nil {
		return "", interpretationError
	}
	if parsedURL.relativeFilePath == "" {
		return "", startosis_errors.NewInterpretationError("Path to import '%v' needs to point to a specific file but didn't. Users can only read or import specific files and not entire packages.", fileInsidePackageUrl)
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

func (provider *GitPackageContentProvider) GetModuleContents(fileInsideModuleUrl string) (string, *startosis_errors.InterpretationError) {
	pathToFile, interpretationError := provider.GetOnDiskAbsoluteFilePath(fileInsideModuleUrl)
	if interpretationError != nil {
		return "", interpretationError
	}

	maybeKurtosisYamlPath, err := checkIfFileIsInAValidPackage(pathToFile, provider.packagesDir)
	if err != nil {
		return "", startosis_errors.WrapWithInterpretationError(err, "Error occurred while verifying whether '%v' belongs to a Kurtosis package.", fileInsideModuleUrl)
	}

	if maybeKurtosisYamlPath == filePathToKurtosisYamlNotFound {
		return "", startosis_errors.NewInterpretationError("%v is not found in the path of '%v'; files can only be imported or read from Kurtosis packages. For more information, go to: %v", startosis_constants.KurtosisYamlName, fileInsideModuleUrl, howImportWorksLink)
	}

	// Load the file content from its absolute path
	contents, errWhileReadingFile := os.ReadFile(pathToFile)
	if errWhileReadingFile != nil {
		return "", startosis_errors.WrapWithInterpretationError(err, "Loading module content for module '%s' failed. An error occurred in reading contents of the file '%v'", fileInsideModuleUrl, pathToFile)
	}

	return string(contents), nil
}

func (provider *GitPackageContentProvider) StorePackageContents(packageId string, moduleTar []byte, overwriteExisting bool) (string, *startosis_errors.InterpretationError) {
	parsedPackageId, interpretationError := parseGitURL(packageId)
	if interpretationError != nil {
		return "", interpretationError
	}

	relPackagePathToPackagesDir := getPathToPackageRoot(parsedPackageId)
	packageAbsolutePathOnDisk := path.Join(provider.packagesDir, relPackagePathToPackagesDir)

	if overwriteExisting {
		err := os.RemoveAll(packageAbsolutePathOnDisk)
		if err != nil {
			return "", startosis_errors.WrapWithInterpretationError(err, "An error occurred while removing the existing package '%v' from disk at '%v'", packageId, packageAbsolutePathOnDisk)
		}
	}

	_, err := os.Stat(packageAbsolutePathOnDisk)
	if err == nil {
		return "", startosis_errors.NewInterpretationError("Package '%v' already exists on disk, not overwriting", packageAbsolutePathOnDisk)
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
	err = archiver.Unarchive(tempFile.Name(), packageAbsolutePathOnDisk)
	if err != nil {
		return "", startosis_errors.WrapWithInterpretationError(err, "An error occurred while unarchiving '%v' to '%v'", tempFile.Name(), packageAbsolutePathOnDisk)
	}

	return packageAbsolutePathOnDisk, nil
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

	depth := defaultDepth
	if parsedURL.tagBranchOrCommit != emptyTagBranchOrCommit {
		depth = depthAssumingBranchTagsCommitsAreSpecified
	}

	repo, err := git.PlainClone(gitClonePath, isNotBareClone, &git.CloneOptions{
		URL:               parsedURL.gitURL,
		Auth:              nil,
		RemoteName:        "",
		ReferenceName:     "",
		SingleBranch:      false,
		NoCheckout:        false,
		Depth:             depth,
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

	if parsedURL.tagBranchOrCommit != emptyTagBranchOrCommit {
		referenceTagOrBranch, found, referenceErr := getReferenceName(repo, parsedURL)
		if err != nil {
			return referenceErr
		}

		checkoutOptions := &git.CheckoutOptions{
			Hash:   plumbing.Hash{},
			Branch: "",
			Create: false,
			Force:  false,
			Keep:   false,
		}
		if found {
			// if we have a tag or branch we set it
			checkoutOptions.Branch = referenceTagOrBranch
		} else {
			maybeHash := plumbing.NewHash(parsedURL.tagBranchOrCommit)
			// check whether the hash is a valid hash
			_, err = repo.CommitObject(maybeHash)
			if err != nil {
				return startosis_errors.NewInterpretationError("Tried using the passed version '%v' as commit as we couldn't find a tag/branch for it in the repo '%v' but failed", parsedURL.tagBranchOrCommit, parsedURL.gitURL)
			}
			checkoutOptions.Hash = maybeHash
		}

		workTree, err := repo.Worktree()
		if err != nil {
			return startosis_errors.NewInterpretationError("Tried getting worktree for cloned repo '%v' but failed", parsedURL.gitURL)
		}

		if err := workTree.Checkout(checkoutOptions); err != nil {
			return startosis_errors.NewInterpretationError("Tried checking out '%v' on repository '%v' but failed", parsedURL.tagBranchOrCommit, parsedURL.gitURL)
		}
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

// methods checks whether the root of the package is same as repository root
// or it is a sub-folder under it
func getPathToPackageRoot(parsedPackagePath *ParsedGitURL) string {
	packagePathOnDisk := parsedPackagePath.relativeRepoPath
	if parsedPackagePath.relativeFilePath != relativeFilePathNotFound {
		packagePathOnDisk = parsedPackagePath.relativeFilePath
	}
	return packagePathOnDisk
}

func getReferenceName(repo *git.Repository, parsedURL *ParsedGitURL) (plumbing.ReferenceName, bool, *startosis_errors.InterpretationError) {
	tag, err := repo.Tag(parsedURL.tagBranchOrCommit)
	if err == nil {
		return tag.Name(), true, nil
	}
	if err != nil && err != git.ErrTagNotFound {
		return "", false, startosis_errors.NewInterpretationError("An error occurred while checking whether '%v' is a tag and exists in '%v'", parsedURL.tagBranchOrCommit, parsedURL.gitURL)
	}

	// for branches there is repo.Branch() but that doesn't work with remote branches
	refs, err := repo.References()
	if err != nil {
		return "", false, startosis_errors.NewInterpretationError("An error occurred while fetching references for repository '%v'", parsedURL.gitURL)
	}

	var branchReference *plumbing.Reference
	err = refs.ForEach(func(r *plumbing.Reference) error {
		if r.Name() == plumbing.NewRemoteReferenceName(git.DefaultRemoteName, parsedURL.tagBranchOrCommit) {
			branchReference = r
			return nil
		}
		return nil
	})

	// checking the error even though the above can't ever error
	if err != nil {
		return "", false, startosis_errors.NewInterpretationError("An error occurred while iterating through references in repository '%v'; This is a bug in Kurtosis", parsedURL.gitURL)
	}

	if branchReference != nil {
		return branchReference.Name(), true, nil
	}

	return "", false, nil
}

/**
While importing/reading a file we are currently cloning the repository, and trying to find whether kurtosis.yml exists in the path;
this is being done as part of interpretation step of starlark.
TODO: we should clean this up and have a dependency management system; all the dependencies should be stated kurtosis.yml upfront
TODO: this will simplify our validation process, and enable customers to use local packages like go.
TODO: in my opinion - we should eventually clone and validate the packages even before we start the interpretation process, maybe inside
api_container_service
*/
func checkIfFileIsInAValidPackage(absPathToFile string, packagesDir string) (string, *startosis_errors.InterpretationError) {
	return checkIfFileIsInAValidPackageInternal(absPathToFile, packagesDir, os.Stat)
}

/**
This method walks along the path of the file and determines whether kurtosis.yml is found in any directory. If the path is found, it returns
the absolute path of kurtosis.yml, otherwise it returns an empty string when the kurtosis.yml is not found.

For example, the path to the file is /kurtosis-data/startosis-packages/some-repo/some-folder/some-file-to-be-read.star
This method will start the walk from some-repo, then go to some-folder and so on.
It will continue the search for kurtosis.yml until either kurtosis.yml is found or the path is fully transversed.
*/
func checkIfFileIsInAValidPackageInternal(absPathToFile string, packagesDir string, stat func(string) (os.FileInfo, error)) (string, *startosis_errors.InterpretationError) {
	// it will remove /kurtosis-data/startosis-package from absPathToFile and start the search from repo itself.
	// we can be sure that kurtosis.yml will never be found in those folders.
	beginSearchForKurtosisYmlFromRepo := strings.TrimPrefix(absPathToFile, packagesDir)
	if beginSearchForKurtosisYmlFromRepo == absPathToFile {
		return filePathToKurtosisYamlNotFound, startosis_errors.NewInterpretationError("Absolute path to file: %v must start with following prefix %v", absPathToFile, packagesDir)
	}

	removeTrailingPathSeparator := strings.Trim(beginSearchForKurtosisYmlFromRepo, string(os.PathSeparator))
	dirs := strings.Split(removeTrailingPathSeparator, string(os.PathSeparator))
	logrus.Debugf("Found directories: %v", dirs)

	maybePackageRootPath := packagesDir
	for _, dir := range dirs[:len(dirs)-1] {
		maybePackageRootPath = path.Join(maybePackageRootPath, dir)
		pathToKurtosisYaml := path.Join(maybePackageRootPath, startosis_constants.KurtosisYamlName)
		if _, err := stat(pathToKurtosisYaml); err == nil {
			logrus.Debugf("Found root path: %v", maybePackageRootPath)
			// the method should return the absolute path to minimize the confusion
			return pathToKurtosisYaml, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return filePathToKurtosisYamlNotFound, startosis_errors.WrapWithInterpretationError(err, "An error occurred while locating %v in the path of '%v'", startosis_constants.KurtosisYamlName, absPathToFile)
		}
	}
	return filePathToKurtosisYamlNotFound, nil
}
