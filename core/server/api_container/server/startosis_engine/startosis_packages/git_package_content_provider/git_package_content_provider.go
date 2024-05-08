package git_package_content_provider

import (
	"errors"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/shared_utils"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/user_support_constants"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/docker_compose_transpiler"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_constants"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"github.com/kurtosis-tech/kurtosis/core/server/commons/yaml_parser"
	"github.com/mholt/archiver"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const (
	moduleDirPermission         = 0755
	temporaryRepoDirPattern     = "tmp-repo-dir-*"
	temporaryArchiveFilePattern = "temp-module-archive-*.tgz"
	defaultTmpDir               = ""
	emptyTagBranchOrCommit      = ""

	onlyOneReplacement       = 1
	replacedWithEmptyString  = ""
	relativeFilePathNotFound = ""
	isNotBareClone           = false

	// this doesn't get us the entire history, speeding us up
	defaultDepth = 1
	// this gets us the entire history - useful for fetching commits on a repo
	depthAssumingBranchTagsCommitsAreSpecified = 0

	filePathToKurtosisOrComposeYamlNotFound  = ""
	replaceCountPackageDirWithGithubConstant = 1

	osPathSeparatorString = string(os.PathSeparator)

	onlyOneReplace = 1

	defaultMainBranch = ""

	maxRetries           = 3
	retryDelayStartValue = 1 * time.Second
)

type GitPackageContentProvider struct {
	// Where to temporarily store repositories while
	repositoriesTmpDir              string
	repositoriesDir                 string
	packageReplaceOptionsRepository *packageReplaceOptionsRepository
	githubAuthProvider              *GitHubPackageAuthProvider
}

func NewGitPackageContentProvider(repositoriesDir, tmpDir string, githubAuthProvider *GitHubPackageAuthProvider, enclaveDb *enclave_db.EnclaveDB) *GitPackageContentProvider {
	return &GitPackageContentProvider{
		repositoriesDir:                 repositoriesDir,
		repositoriesTmpDir:              tmpDir,
		githubAuthProvider:              githubAuthProvider,
		packageReplaceOptionsRepository: newPackageReplaceOptionsRepository(enclaveDb),
	}
}

func (provider *GitPackageContentProvider) ClonePackage(packageId string) (string, *startosis_errors.InterpretationError) {
	parsedURL, err := shared_utils.ParseGitURL(packageId)
	if err != nil {
		return "", startosis_errors.WrapWithInterpretationError(err, "An error occurred parsing Git URL for package ID '%s'", packageId)
	}

	if interpretationError := provider.atomicClone(parsedURL, provider.getGitHubAuthToken(packageId)); interpretationError != nil {
		return "", interpretationError
	}

	relPackagePathToPackagesDir := getPathToPackageRoot(parsedURL)
	packageAbsolutePathOnDisk := path.Join(provider.repositoriesDir, relPackagePathToPackagesDir)

	return packageAbsolutePathOnDisk, nil
}

func (provider *GitPackageContentProvider) GetKurtosisYaml(packageAbsolutePathOnDisk string) (*yaml_parser.KurtosisYaml, *startosis_errors.InterpretationError) {
	pathToKurtosisYaml := path.Join(packageAbsolutePathOnDisk, startosis_constants.KurtosisYamlName)
	if _, err := os.Stat(pathToKurtosisYaml); err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Couldn't find a '%v' in the root of the package: '%v'. Packages are expected to have a '%v' at root; for more information have a look at %v",
			startosis_constants.KurtosisYamlName, packageAbsolutePathOnDisk, startosis_constants.KurtosisYamlName, user_support_constants.StarlarkPackagesReferenceURL)
	}

	kurtosisYaml, interpretationError := validateAndGetKurtosisYaml(pathToKurtosisYaml, provider.repositoriesDir)
	if interpretationError != nil {
		return nil, interpretationError
	}

	return kurtosisYaml, nil
}

func (provider *GitPackageContentProvider) GetOnDiskAbsolutePath(absoluteModuleLocator *startosis_packages.PackageAbsoluteLocator) (string, *startosis_errors.InterpretationError) {
	shouldOnlyAcceptsPackagePath := false
	return provider.getOnDiskAbsolutePath(absoluteModuleLocator, shouldOnlyAcceptsPackagePath)
}

func (provider *GitPackageContentProvider) GetOnDiskAbsolutePackageFilePath(absoluteLocator *startosis_packages.PackageAbsoluteLocator) (string, *startosis_errors.InterpretationError) {
	shouldOnlyAcceptsPackagePath := true
	return provider.getOnDiskAbsolutePath(absoluteLocator, shouldOnlyAcceptsPackagePath)
}

func (provider *GitPackageContentProvider) getOnDiskAbsolutePath(absoluteLocator *startosis_packages.PackageAbsoluteLocator, shouldOnlyAcceptsPackageFilePath bool) (string, *startosis_errors.InterpretationError) {
	repositoryPathURL := absoluteLocator.GetGitURL()
	parsedURL, err := shared_utils.ParseGitURL(repositoryPathURL)
	if err != nil {
		return "", startosis_errors.WrapWithInterpretationError(err, "An error occurred parsing Git URL for absolute file locator '%s'", repositoryPathURL)
	}
	pathToFileOnDisk := path.Join(provider.repositoriesDir, parsedURL.GetRelativeFilePath())
	pathToPackageOnDisk := path.Join(provider.repositoriesDir, parsedURL.GetRelativeRepoPath())

	// TODO(tedi): see if its safe to adjust ParsedGitURL api instead to prevent having to make this check
	// parsedURL.GetRelativeFilePath() is empty when repositoryPathURL does not refer to a specific file
	// In this case, assume caller wants to get on base of repository and adjust the pathToFileOnDisk to point to base of repository on disk
	if parsedURL.GetRelativeFilePath() == "" {
		pathToFileOnDisk = pathToPackageOnDisk
	}

	// Return the file path straight if it exists
	if _, err := os.Stat(pathToFileOnDisk); err == nil {
		return pathToFileOnDisk, nil
	}

	// Check if the repo exists
	// If the repo exists but the `pathToFileOnDisk` doesn't exist, the locator is invalid
	if _, err := os.Stat(pathToPackageOnDisk); err == nil {
		relativeFilePathWithoutPackageName := strings.Replace(parsedURL.GetRelativeFilePath(), parsedURL.GetRelativeRepoPath(), replacedWithEmptyString, onlyOneReplacement)
		return "", startosis_errors.NewInterpretationError("'%v' doesn't exist in the package '%v'", relativeFilePathWithoutPackageName, parsedURL.GetRelativeRepoPath())
	}

	// Otherwise clone the repo and return the absolute path of the requested file
	emptyPackageId := ""
	if interpretationError := provider.atomicClone(parsedURL, provider.getGitHubAuthToken(emptyPackageId)); interpretationError != nil {
		return "", interpretationError
	}

	if !shouldOnlyAcceptsPackageFilePath {
		return pathToFileOnDisk, nil
	}

	// check whether kurtosis yaml exists in the path
	maybeKurtosisOrComposeYamlPath, interpretationError := getKurtosisOrComposeYamlPathForFile(pathToFileOnDisk, provider.repositoriesDir)
	if interpretationError != nil {
		return "", startosis_errors.WrapWithInterpretationError(err, "Error occurred while verifying whether '%v' belongs to a Kurtosis package.", repositoryPathURL)
	}

	if maybeKurtosisOrComposeYamlPath == filePathToKurtosisOrComposeYamlNotFound {
		return "", startosis_errors.NewInterpretationError("%v or valid Docker Compose yaml not found in the path of '%v'; files can only be accessed from Kurtosis packages. For more information, go to: %v", startosis_constants.KurtosisYamlName, repositoryPathURL, user_support_constants.HowImportWorksLink)
	}

	if containsKurtosisYaml(maybeKurtosisOrComposeYamlPath) {
		if _, interpretationError = validateAndGetKurtosisYaml(maybeKurtosisOrComposeYamlPath, provider.repositoriesDir); interpretationError != nil {
			return "", interpretationError
		}
	}

	return pathToFileOnDisk, nil
}

func (provider *GitPackageContentProvider) GetModuleContents(absoluteLocator *startosis_packages.PackageAbsoluteLocator) (string, *startosis_errors.InterpretationError) {
	pathToFile, interpretationError := provider.GetOnDiskAbsolutePackageFilePath(absoluteLocator)
	if interpretationError != nil {
		return "", interpretationError
	}

	// if pathToFile contains compose yaml, assume Docker Compose Package
	if containsComposeYaml(pathToFile) {
		contents, err := docker_compose_transpiler.TranspileDockerComposePackageToStarlark(filepath.Dir(pathToFile), filepath.Base(pathToFile))
		if err != nil {
			return "", startosis_errors.WrapWithInterpretationError(err, "Loading module content for module '%s' failed. An error occurred in transpiling the Docker Compose Package to Starlark at path '%v'", absoluteLocator.GetLocator(), pathToFile)
		}
		return contents, nil
	} else {
		contentsBytes, err := os.ReadFile(pathToFile)
		if err != nil {
			return "", startosis_errors.WrapWithInterpretationError(err, "Loading module content for module '%s' failed. An error occurred in reading contents of the file '%v'", absoluteLocator.GetLocator(), pathToFile)
		}
		return string(contentsBytes), nil
	}
}

func (provider *GitPackageContentProvider) GetOnDiskAbsolutePackagePath(packageId string) (string, *startosis_errors.InterpretationError) {
	parsedPackageId, err := shared_utils.ParseGitURL(packageId)
	if err != nil {
		return "", startosis_errors.WrapWithInterpretationError(err, "An error occurred parsing Git URL for package ID '%s'", packageId)
	}

	relPackagePathToPackagesDir := getPathToPackageRoot(parsedPackageId)
	packageAbsolutePathOnDisk := path.Join(provider.repositoriesDir, relPackagePathToPackagesDir)

	_, err = os.Stat(packageAbsolutePathOnDisk)
	if err != nil {
		return "", startosis_errors.NewInterpretationError("Package '%v' does not exist on disk at '%s'", packageId, packageAbsolutePathOnDisk)
	}
	return packageAbsolutePathOnDisk, nil
}

func (provider *GitPackageContentProvider) StorePackageContents(packageId string, moduleTar io.Reader, overwriteExisting bool) (string, *startosis_errors.InterpretationError) {
	parsedPackageId, err := shared_utils.ParseGitURL(packageId)
	if err != nil {
		return "", startosis_errors.WrapWithInterpretationError(err, "An error occurred parsing Git URL for package ID '%s'", packageId)
	}

	relPackagePathToPackagesDir := getPathToPackageRoot(parsedPackageId)
	packageAbsolutePathOnDisk := path.Join(provider.repositoriesDir, relPackagePathToPackagesDir)

	if overwriteExisting {
		err := os.RemoveAll(packageAbsolutePathOnDisk)
		if err != nil {
			return "", startosis_errors.WrapWithInterpretationError(err, "An error occurred while removing the existing package '%v' from disk at '%v'", packageId, packageAbsolutePathOnDisk)
		}
	}

	_, err = os.Stat(packageAbsolutePathOnDisk)
	if err == nil {
		return "", startosis_errors.NewInterpretationError("Package '%v' already exists on disk, not overwriting", packageAbsolutePathOnDisk)
	}

	tempFile, err := os.CreateTemp(defaultTmpDir, temporaryArchiveFilePattern)
	if err != nil {
		return "", startosis_errors.NewInterpretationError("An error occurred while creating temporary file to write compressed '%v' to", packageId)
	}
	defer os.Remove(tempFile.Name())

	_, err = io.Copy(tempFile, moduleTar)
	if err != nil {
		return "", startosis_errors.WrapWithInterpretationError(err, "An error occurred while writing contents of '%v' to '%v'", packageId, tempFile.Name())
	}
	err = archiver.Unarchive(tempFile.Name(), packageAbsolutePathOnDisk)
	if err != nil {
		return "", startosis_errors.WrapWithInterpretationError(err, "An error occurred while unarchiving '%v' to '%v'", tempFile.Name(), packageAbsolutePathOnDisk)
	}

	return packageAbsolutePathOnDisk, nil
}

func (provider *GitPackageContentProvider) GetAbsoluteLocator(
	packageId string,
	sourceModuleLocator string,
	relativeOrAbsoluteLocator string,
	packageReplaceOptions map[string]string,
) (*startosis_packages.PackageAbsoluteLocator, *startosis_errors.InterpretationError) {
	var absoluteLocatorStr string

	if shouldBlockAbsoluteLocatorBecauseIsInTheSameSourceModuleLocatorPackage(relativeOrAbsoluteLocator, sourceModuleLocator, packageId) {
		return nil, startosis_errors.NewInterpretationError("Locator '%s' is referencing a file within the same package using absolute import syntax, but only relative import syntax (path starting with '/' or '.') is allowed for within-package imports", relativeOrAbsoluteLocator)
	}

	// maybe it's not a relative url in which case we return the url
	_, errorParsingUrl := shared_utils.ParseGitURL(relativeOrAbsoluteLocator)
	if errorParsingUrl == nil {
		// Parsing succeeded, meaning this is already an absolute locator and no relative -> absolute translation is needed
		absoluteLocatorStr = relativeOrAbsoluteLocator
	} else {
		// Parsing did not succeed, meaning this is a relative locator
		sourceModuleParsedGitUrl, errorParsingPackageId := shared_utils.ParseGitURL(sourceModuleLocator)
		if errorParsingPackageId != nil {
			return nil, startosis_errors.NewInterpretationError("Source module locator '%v' isn't a valid locator; relative URLs don't work with standalone scripts", sourceModuleLocator)
		}

		absoluteLocatorStr = sourceModuleParsedGitUrl.GetAbsoluteLocatorRelativeToThisURL(relativeOrAbsoluteLocator)
	}

	absoluteLocator := startosis_packages.NewPackageAbsoluteLocator(absoluteLocatorStr, defaultMainBranch)

	replacedAbsoluteLocator := replaceAbsoluteLocator(absoluteLocator, packageReplaceOptions)

	return replacedAbsoluteLocator, nil
}

func (provider *GitPackageContentProvider) CloneReplacedPackagesIfNeeded(currentPackageReplaceOptions map[string]string) *startosis_errors.InterpretationError {

	existingPackageReplaceOptions, err := provider.packageReplaceOptionsRepository.Get()
	if err != nil {
		return startosis_errors.WrapWithInterpretationError(err, "An error occurred getting the existing package replace options from the repository")
	}

	for packageId, existingReplace := range existingPackageReplaceOptions {

		shouldClonePackage := false

		isExistingLocalReplace := isLocalDependencyReplace(existingReplace)
		logrus.Debugf("existingReplace '%v' isExistingLocalReplace? '%v', ", existingReplace, isExistingLocalReplace)

		currentReplace, isCurrentReplace := currentPackageReplaceOptions[packageId]
		if isCurrentReplace {
			// the package will be cloned if the current replace is remote and the existing is local
			isCurrentRemoteReplace := !isLocalLocator(currentReplace)
			logrus.Debugf("currentReplace '%v' isCurrentRemoteReplace? '%v', ", isCurrentRemoteReplace, currentReplace)
			if isCurrentRemoteReplace && isExistingLocalReplace {
				shouldClonePackage = true
			}
		}

		// there is no current replace for this dependency but the version in the cache is local
		if !isCurrentReplace && isExistingLocalReplace {
			shouldClonePackage = true
		}

		if shouldClonePackage {
			if _, err := provider.ClonePackage(packageId); err != nil {
				return startosis_errors.WrapWithInterpretationError(err, "An error occurred cloning package '%v'", packageId)
			}
		}
	}

	// upgrade the existing-replace list with the new values
	for packageId, currentReplace := range currentPackageReplaceOptions {
		existingPackageReplaceOptions[packageId] = currentReplace
	}

	if err = provider.packageReplaceOptionsRepository.Save(existingPackageReplaceOptions); err != nil {
		return startosis_errors.WrapWithInterpretationError(err, "An error occurred saving the existing package replace options from the repository")
	}
	return nil
}

// atomicClone This first clones to a temporary directory and then moves it into the package file system
// TODO make this support versioning via tags, commit hashes or branches
func (provider *GitPackageContentProvider) atomicClone(parsedURL *shared_utils.ParsedGitURL, githubAuthToken string) *startosis_errors.InterpretationError {
	// First we clone into a temporary directory
	tempRepoDirPath, err := os.MkdirTemp(provider.repositoriesTmpDir, temporaryRepoDirPattern)
	if err != nil {
		return startosis_errors.WrapWithInterpretationError(err, "Cloning the repository '%s' failed. Error creating temporary directory for the repository to be cloned into", parsedURL.GetGitURL())
	}
	defer os.RemoveAll(tempRepoDirPath)
	gitClonePath := path.Join(tempRepoDirPath, parsedURL.GetRelativeRepoPath())

	depth := defaultDepth
	if parsedURL.GetTagBranchOrCommit() != emptyTagBranchOrCommit {
		depth = depthAssumingBranchTagsCommitsAreSpecified
	}

	var githubAuth *http.BasicAuth
	if githubAuthToken != "" {
		githubAuth = &http.BasicAuth{
			Username: "token",
			Password: githubAuthToken,
		}
	}

	repo, interpretationError := provider.cloneWithRetries(parsedURL, gitClonePath, githubAuth, depth)
	if interpretationError != nil {
		return interpretationError
	}

	if parsedURL.GetTagBranchOrCommit() != emptyTagBranchOrCommit {
		referenceTagOrBranch, found, referenceErr := getReferenceName(repo, parsedURL)
		if err != nil {
			return referenceErr
		}

		checkoutOptions := &git.CheckoutOptions{
			Hash:                      plumbing.Hash{},
			Branch:                    "",
			Create:                    false,
			Force:                     false,
			Keep:                      false,
			SparseCheckoutDirectories: []string{},
		}
		if found {
			// if we have a tag or branch we set it
			checkoutOptions.Branch = referenceTagOrBranch
		} else {
			maybeHash := plumbing.NewHash(parsedURL.GetTagBranchOrCommit())
			// check whether the hash is a valid hash
			_, err = repo.CommitObject(maybeHash)
			if err != nil {
				return startosis_errors.NewInterpretationError("Tried using the passed version '%v' as commit as we couldn't find a tag/branch for it in the repo '%v' but failed", parsedURL.GetTagBranchOrCommit(), parsedURL.GetGitURL())
			}
			checkoutOptions.Hash = maybeHash
		}

		workTree, err := repo.Worktree()
		if err != nil {
			return startosis_errors.NewInterpretationError("Tried getting worktree for cloned repo '%v' but failed", parsedURL.GetGitURL())
		}

		if err := workTree.Checkout(checkoutOptions); err != nil {
			return startosis_errors.NewInterpretationError("Tried checking out '%v' on repository '%v' but failed", parsedURL.GetTagBranchOrCommit(), parsedURL.GetGitURL())
		}
	}

	// Then we move it into the target directory
	packageAuthorPath := path.Join(provider.repositoriesDir, parsedURL.GetRepositoryAuthor())
	packagePath := path.Join(provider.repositoriesDir, parsedURL.GetRelativeRepoPath())
	fileMode, err := os.Stat(packageAuthorPath)
	if err == nil && !fileMode.IsDir() {
		return startosis_errors.WrapWithInterpretationError(err, "Expected '%s' to be a directory but it is something else", packageAuthorPath)
	}
	if err != nil {
		if err = os.Mkdir(packageAuthorPath, moduleDirPermission); err != nil {
			return startosis_errors.WrapWithInterpretationError(err, "Cloning the repository '%s' failed. An error occurred while creating the directory '%s'.", parsedURL.GetGitURL(), packageAuthorPath)
		}
	}
	if _, err = os.Stat(packagePath); !os.IsNotExist(err) {
		logrus.Debugf("Package '%s' already exists in enclave at path '%s'. Going to remove previous version.", parsedURL.GetGitURL(), packagePath)
		if err = os.RemoveAll(packagePath); err != nil {
			return startosis_errors.NewInterpretationError("Unable to remove a previous version of repository '%s' existing inside Kurtosis enclave at '%s'.", parsedURL.GetGitURL(), packagePath)
		}
	}
	if err = os.Rename(gitClonePath, packagePath); err != nil {
		return startosis_errors.NewInterpretationError("Cloning the repository '%s' failed. An error occurred while moving repository at temporary destination '%s' to final destination '%s'", parsedURL.GetGitURL(), gitClonePath, packagePath)
	}
	return nil
}

func (provider *GitPackageContentProvider) cloneWithRetries(parsedURL *shared_utils.ParsedGitURL, gitClonePath string, githubAuth *http.BasicAuth, depth int) (*git.Repository, *startosis_errors.InterpretationError) {
	retryDelay := retryDelayStartValue

	var repo *git.Repository
	var err error

	for retry := 0; retry < maxRetries; retry++ {

		//TODO evaluate to use the GitHub client GetContents call instead, because we are cloning the entire repository's workspace with this approach
		//TODO and the Starlark package could be just a small sub-folder inside a giant mono-repository
		//TODO and even now, in the upload_files instruction, we are allowing to upload files or a folder for any repository, but we are cloning the entire repository for this
		repo, err = git.PlainClone(gitClonePath, isNotBareClone, &git.CloneOptions{
			URL:               parsedURL.GetGitURL(),
			Auth:              githubAuth,
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
			Mirror:            false,
			ShallowSubmodules: false,
			ProxyOptions: transport.ProxyOptions{
				URL:      "",
				Username: "",
				Password: "",
			},
			Shared: false,
		})

		if err == nil {
			break
		}

		if retry < maxRetries-1 {
			logrus.Infof("Cloning failed with error '%v'. Retrying in '%v'", err, retryDelay)
			time.Sleep(retryDelay)
			retryDelay *= 2
		}
	}
	if err != nil {
		// We silence the underlying error here as it can be confusing to the user. For example, when there's a typo in
		// the repo name, pointing to a non existing repo, the underlying error is: "authentication required"
		logrus.Errorf("Error cloning git repository: '%s' to '%s'. Error was: \n%s", parsedURL.GetGitURL(), gitClonePath, err.Error())
		return nil, startosis_errors.NewInterpretationError("Error in cloning git repository '%s' to '%s'. Make sure that '%v' exists or if it's a private repository, that you are logged into GitHub via `kurtosis github login`.", parsedURL.GetGitURL(), gitClonePath, parsedURL.GetGitURL())
	}
	return repo, nil
}

// Returns empty string if no token found by [githubAuthProvider]
// If packageId is empty string, only checks for and returns github token for the user if it exists
func (provider *GitPackageContentProvider) getGitHubAuthToken(packageId string) string {
	var githubAuthToken string
	githubAuthToken = provider.githubAuthProvider.GetGitHubTokenForPackage(packageId)
	if githubAuthToken != "" {
		return githubAuthToken
	}
	githubAuthToken = provider.githubAuthProvider.GetGitHubTokenForUser()
	return githubAuthToken
}

// methods checks whether the root of the package is same as repository root
// or it is a sub-folder under it
func getPathToPackageRoot(parsedPackagePath *shared_utils.ParsedGitURL) string {
	packagePathOnDisk := parsedPackagePath.GetRelativeRepoPath()
	if parsedPackagePath.GetRelativeFilePath() != relativeFilePathNotFound {
		packagePathOnDisk = parsedPackagePath.GetRelativeFilePath()
	}
	return packagePathOnDisk
}

func getReferenceName(repo *git.Repository, parsedURL *shared_utils.ParsedGitURL) (plumbing.ReferenceName, bool, *startosis_errors.InterpretationError) {
	tag, err := repo.Tag(parsedURL.GetTagBranchOrCommit())
	if err == nil {
		return tag.Name(), true, nil
	}
	if err != nil && err != git.ErrTagNotFound {
		return "", false, startosis_errors.NewInterpretationError("An error occurred while checking whether '%v' is a tag and exists in '%v'", parsedURL.GetTagBranchOrCommit(), parsedURL.GetGitURL())
	}

	// for branches there is repo.Branch() but that doesn't work with remote branches
	refs, err := repo.References()
	if err != nil {
		return "", false, startosis_errors.NewInterpretationError("An error occurred while fetching references for repository '%v'", parsedURL.GetGitURL())
	}

	var branchReference *plumbing.Reference
	err = refs.ForEach(func(r *plumbing.Reference) error {
		if r.Name() == plumbing.NewRemoteReferenceName(git.DefaultRemoteName, parsedURL.GetTagBranchOrCommit()) {
			branchReference = r
			return nil
		}
		return nil
	})

	// checking the error even though the above can't ever error
	if err != nil {
		return "", false, startosis_errors.NewInterpretationError("An error occurred while iterating through references in repository '%v'; This is a bug in Kurtosis", parsedURL.GetGitURL())
	}

	if branchReference != nil {
		return branchReference.Name(), true, nil
	}

	return "", false, nil
}

// this method validates the contents of the kurtosis.yml found at path identified by the absPathToKurtosisYmlInThePackage
func validateAndGetKurtosisYaml(absPathToKurtosisYmlInThePackage string, packageDir string) (*yaml_parser.KurtosisYaml, *startosis_errors.InterpretationError) {
	kurtosisYaml, errWhileParsing := yaml_parser.ParseKurtosisYaml(absPathToKurtosisYmlInThePackage)
	if errWhileParsing != nil {
		return nil, startosis_errors.WrapWithInterpretationError(errWhileParsing, "Error occurred while parsing %v", absPathToKurtosisYmlInThePackage)
	}

	// this method validates whether the package name is also the locator - it should be the location where kurtosis.yml exists
	if err := validatePackageNameMatchesKurtosisYamlLocation(kurtosisYaml, absPathToKurtosisYmlInThePackage, packageDir); err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Error occurred while validating %v", absPathToKurtosisYmlInThePackage)
	}

	return kurtosisYaml, nil
}

// this method validates whether the package name found in kurtosis yml is same as the location where kurtosis.yml is found
func validatePackageNameMatchesKurtosisYamlLocation(kurtosisYaml *yaml_parser.KurtosisYaml, absPathToKurtosisYmlInThePackage string, packageDir string) *startosis_errors.InterpretationError {
	// get package name from absolute path to package
	packageNameFromAbsPackagePath := strings.Replace(absPathToKurtosisYmlInThePackage, packageDir, shared_utils.GithubDomainPrefix, replaceCountPackageDirWithGithubConstant)
	packageName := kurtosisYaml.GetPackageName()

	if strings.HasSuffix(packageName, osPathSeparatorString) {
		return startosis_errors.NewInterpretationError("Kurtosis package name cannot have trailing %q; package name: %v and kurtosis.yml is found at: %v", osPathSeparatorString, packageName, packageNameFromAbsPackagePath)
	}

	// re-using ParseGitURL with packageName found from kurtosis.yml as it already does some validations
	_, err := shared_utils.ParseGitURL(packageName)
	if err != nil {
		return startosis_errors.WrapWithInterpretationError(err, "Error occurred while validating package name: %v which is found in kurtosis.yml at: '%v'", kurtosisYaml.GetPackageName(), packageNameFromAbsPackagePath)
	}

	removeKurtosisYmlFromPackageName := path.Dir(packageNameFromAbsPackagePath)

	// wrapping the strings with trim - so that we can ignore `/` mismatches
	if packageName != removeKurtosisYmlFromPackageName {
		return startosis_errors.NewInterpretationError("The package name in %v must match the location it is in. Package name is '%v' and kurtosis.yml is found here: '%v'", startosis_constants.KurtosisYamlName, kurtosisYaml.GetPackageName(), removeKurtosisYmlFromPackageName)
	}
	return nil
}

// While importing/reading a file we are currently cloning the repository, and trying to find whether kurtosis.yml exists in the path;
// this is being done as part of interpretation step of starlark.
// TODO: we should clean this up and have a dependency management system; all the dependencies should be stated kurtosis.yml upfront
// TODO: this will simplify our validation process, and enable customers to use local packages like go.
// TODO: in my opinion - we should eventually clone and validate the packages even before we start the interpretation process, maybe inside
// TODO: api_container_service
func getKurtosisOrComposeYamlPathForFile(absPathToFile string, packagesDir string) (string, *startosis_errors.InterpretationError) {
	return getKurtosisOrComposeYamlPathForFileUrlInternal(absPathToFile, packagesDir, os.Stat)
}

// This method walks along the path of the file and determines whether kurtosis.yml or a valid compose.yml is found in any directory. If the path is found, it returns
// the absolute path of the .yml, otherwise it returns an empty string when either kurtosis.yml or valid compose.yml not found.
//
// For example, the path to the file is /kurtosis-data/startosis-packages/some-repo/some-folder/some-file-to-be-read.star
// This method will start the walk from some-repo, then go to some-folder and so on.
// It will continue the search for kurtosis.yml until either kurtosis.yml is found or the path is fully transversed.
func getKurtosisOrComposeYamlPathForFileUrlInternal(absPathToFile string, packagesDir string, stat func(string) (os.FileInfo, error)) (string, *startosis_errors.InterpretationError) {
	// it will remove /kurtosis-data/startosis-package from absPathToFile and start the search from repo itself.
	// we can be sure that kurtosis.yml will never be found in those folders.
	beginSearchForKurtosisYamlFromRepo := strings.TrimPrefix(absPathToFile, packagesDir)
	if beginSearchForKurtosisYamlFromRepo == absPathToFile {
		return filePathToKurtosisOrComposeYamlNotFound, startosis_errors.NewInterpretationError("Absolute path to file: %v must start with following prefix %v", absPathToFile, packagesDir)
	}

	removeTrailingPathSeparator := strings.Trim(beginSearchForKurtosisYamlFromRepo, osPathSeparatorString)
	dirs := strings.Split(removeTrailingPathSeparator, osPathSeparatorString)
	logrus.Debugf("Found directories: %v", dirs)

	var validYamlFilenames []string
	validYamlFilenames = append(validYamlFilenames, startosis_constants.KurtosisYamlName)
	validYamlFilenames = append(validYamlFilenames, docker_compose_transpiler.DefaultComposeFilenames...)

	maybePackageRootPath := packagesDir
	for _, dir := range dirs[:len(dirs)-1] {
		maybePackageRootPath = path.Join(maybePackageRootPath, dir)
		for _, validYamlFilename := range validYamlFilenames {
			pathToYaml := path.Join(maybePackageRootPath, validYamlFilename)
			if _, err := stat(pathToYaml); err == nil {
				logrus.Debugf("Found root path: %v", maybePackageRootPath)
				// the method should return the absolute path to minimize the confusion
				return pathToYaml, nil
			} else if !errors.Is(err, os.ErrNotExist) {
				return filePathToKurtosisOrComposeYamlNotFound, startosis_errors.WrapWithInterpretationError(err, "An error occurred while locating %v in the path of '%v'", validYamlFilename, absPathToFile)
			}
		}
	}

	return filePathToKurtosisOrComposeYamlNotFound, nil
}

func isLocalDependencyReplace(replace string) bool {
	if strings.HasPrefix(replace, osPathSeparatorString) || strings.HasPrefix(replace, dotRelativePathIndicatorString) {
		return true
	}
	return false
}

// Returns if kurtosisOrComposeYamlPath contains the kurtosis.yml substring
func containsKurtosisYaml(kurtosisOrComposeYamlPath string) bool {
	return strings.Contains(kurtosisOrComposeYamlPath, startosis_constants.KurtosisYamlName)
}

func containsComposeYaml(filepath string) bool {
	for _, composeYaml := range docker_compose_transpiler.DefaultComposeFilenames {
		if strings.Contains(filepath, composeYaml) {
			return true
		}
	}
	return false
}
