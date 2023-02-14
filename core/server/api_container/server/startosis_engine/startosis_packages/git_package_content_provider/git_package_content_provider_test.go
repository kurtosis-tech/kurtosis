package git_package_content_provider

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"testing"
)

const (
	packagesDirRelPath    = "startosis-packages"
	packagesTmpDirRelPath = "tmp-startosis-packages"
)

func TestGitPackageProvider_SucceedsForValidPackage(t *testing.T) {
	packageDir, err := os.MkdirTemp("", packagesDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageDir)
	packageTmpDir, err := os.MkdirTemp("", packagesTmpDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageTmpDir)

	provider := NewGitPackageContentProvider(packageDir, packageTmpDir)

	sampleStartosisModule := "github.com/kurtosis-tech/sample-startosis-load/sample.star"
	contents, err := provider.GetModuleContents(sampleStartosisModule)
	require.Nil(t, err)
	require.Equal(t, "a = \"World!\"\n", contents)
}

func TestGitPackageProvider_SucceedsForValidPackageWithExplicitMasterSet(t *testing.T) {
	packageDir, err := os.MkdirTemp("", packagesDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageDir)
	packageTmpDir, err := os.MkdirTemp("", packagesTmpDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageTmpDir)

	provider := NewGitPackageContentProvider(packageDir, packageTmpDir)

	sampleStartosisModule := "github.com/kurtosis-tech/sample-startosis-load/sample.star@master"
	contents, err := provider.GetModuleContents(sampleStartosisModule)
	require.Nil(t, err)
	require.Equal(t, "a = \"World!\"\n", contents)
}

func TestGitPackageProvider_SucceedsForValidPackageWithBranch(t *testing.T) {
	packageDir, err := os.MkdirTemp("", packagesDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageDir)
	packageTmpDir, err := os.MkdirTemp("", packagesTmpDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageTmpDir)

	provider := NewGitPackageContentProvider(packageDir, packageTmpDir)

	sampleStartosisModule := "github.com/kurtosis-tech/sample-startosis-load/sample.star@test-branch"
	contents, err := provider.GetModuleContents(sampleStartosisModule)
	require.Nil(t, err)
	require.Equal(t, "a = \"Maybe!\"\n", contents)
}

func TestGitPackageProvider_FailsForInvalidBranch(t *testing.T) {
	packageDir, err := os.MkdirTemp("", packagesDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageDir)
	packageTmpDir, err := os.MkdirTemp("", packagesTmpDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageTmpDir)

	provider := NewGitPackageContentProvider(packageDir, packageTmpDir)

	sampleStartosisModule := "github.com/kurtosis-tech/sample-startosis-load/sample.star@non-existent-branch"
	_, err = provider.GetModuleContents(sampleStartosisModule)
	require.NotNil(t, err)
}

func TestGitPackageProvider_SucceedsForValidPackageWithTag(t *testing.T) {
	packageDir, err := os.MkdirTemp("", packagesDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageDir)
	packageTmpDir, err := os.MkdirTemp("", packagesTmpDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageTmpDir)

	provider := NewGitPackageContentProvider(packageDir, packageTmpDir)

	sampleStartosisModule := "github.com/kurtosis-tech/sample-startosis-load/sample.star@0.1.1"
	contents, err := provider.GetModuleContents(sampleStartosisModule)
	require.Nil(t, err)
	require.Equal(t, "a = \"World!\"\n", contents)
}

func TestGitPackageProvider_SucceedsForValidPackageWithCommit(t *testing.T) {
	packageDir, err := os.MkdirTemp("", packagesDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageDir)
	packageTmpDir, err := os.MkdirTemp("", packagesTmpDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageTmpDir)

	provider := NewGitPackageContentProvider(packageDir, packageTmpDir)

	sampleStartosisModule := "github.com/kurtosis-tech/sample-startosis-load/sample.star@ec9062828e1a687a5db7dfa750f754f88119e4c0"
	contents, err := provider.GetModuleContents(sampleStartosisModule)
	require.Nil(t, err)
	require.Equal(t, "a = \"World!\"\n", contents)
}

func TestGitPackageProvider_SucceedsForValidPackageWithCommitOnABranch(t *testing.T) {
	packageDir, err := os.MkdirTemp("", packagesDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageDir)
	packageTmpDir, err := os.MkdirTemp("", packagesTmpDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageTmpDir)

	provider := NewGitPackageContentProvider(packageDir, packageTmpDir)

	sampleStartosisModule := "github.com/kurtosis-tech/sample-startosis-load/sample.star@df88baf51caffbe7e8f66c0e54715f680f4482b2"
	contents, err := provider.GetModuleContents(sampleStartosisModule)
	require.Nil(t, err)
	require.Equal(t, "a = \"Maybe!\"\n", contents)
}

func TestGitPackageProvider_SucceedsForNonStarlarkFile(t *testing.T) {
	packageDir, err := os.MkdirTemp("", packagesDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageDir)
	packageTmpDir, err := os.MkdirTemp("", packagesTmpDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageTmpDir)

	provider := NewGitPackageContentProvider(packageDir, packageTmpDir)

	sampleStarlarkPackage := "github.com/kurtosis-tech/eth2-package/static_files/prometheus-config/prometheus.yml.tmpl"
	contents, err := provider.GetModuleContents(sampleStarlarkPackage)
	require.Nil(t, err)
	require.NotEmpty(t, contents)
}

func TestGitPackageProvider_FailsForNonExistentPackage(t *testing.T) {
	oackageDir, err := os.MkdirTemp("", packagesDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(oackageDir)
	packageTmpDir, err := os.MkdirTemp("", packagesTmpDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageTmpDir)

	provider := NewGitPackageContentProvider(oackageDir, packageTmpDir)
	nonExistentModulePath := "github.com/kurtosis-tech/non-existent-startosis-load/sample.star"

	_, err = provider.GetModuleContents(nonExistentModulePath)
	require.NotNil(t, err)
}

func TestGetAbsolutePathOnDisk_WorksForPureDirectories(t *testing.T) {
	packageDir, err := os.MkdirTemp("", packagesDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageDir)
	packageTmpDir, err := os.MkdirTemp("", packagesTmpDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageTmpDir)

	provider := NewGitPackageContentProvider(packageDir, packageTmpDir)

	packagePath := "github.com/kurtosis-tech/datastore-army-package/src/helpers.star"
	pathOnDisk, err := provider.GetOnDiskAbsoluteFilePath(packagePath)

	require.Nil(t, err, "This test depends on your internet working and the kurtosis-tech/datastore-army-package existing")
	require.Equal(t, path.Join(packageDir, "kurtosis-tech", "datastore-army-package", "src/helpers.star"), pathOnDisk)
}

func Test_getPathToPackageRoot(t *testing.T) {
	githubUrlWithKurtosisPackageInSubfolder := "github.com/sample/sample-package/folder/subpackage"
	parsedGitUrl, err := parseGitURL(githubUrlWithKurtosisPackageInSubfolder)
	require.Nil(t, err, "Unexpected error occurred while parsing git url")
	actual := getPathToPackageRoot(parsedGitUrl)
	require.Equal(t, "sample/sample-package/folder/subpackage", actual)

	githubUrlWithRootKurtosisPackage := "github.com/sample/sample-package"
	parsedGitUrl, err = parseGitURL(githubUrlWithRootKurtosisPackage)
	require.Nil(t, err, "Unexpected error occurred while parsing git url")
	actual = getPathToPackageRoot(parsedGitUrl)
	require.Equal(t, "sample/sample-package", actual)
}

func Test_checkIfKurtosisYamlExistsInThePathProvided_somewhereInTheMiddle(t *testing.T) {
	mockStatMethod := func(filePath string) (os.FileInfo, error) {
		filePathAndMockReturnMap := map[string]error{
			"/root/kurtosis.yml":                os.ErrNotExist,
			"/root/subdir/kurtosis.yml":         os.ErrNotExist,
			"/root/subdir/subdir1/kurtosis.yml": nil,
		}

		return nil, filePathAndMockReturnMap[filePath]
	}

	filePath := "/root/subdir/subdir1/folder/some_file.txt"
	actual, err := checkIfKurtosisYamlExistsInThePathProvidedInternal(filePath, mockStatMethod)
	require.Nil(t, err)
	require.Equal(t, "/root/subdir/subdir1/kurtosis.yml", actual)
}

func Test_checkIfKurtosisYamlExistsInThePathProvided_packageIsSameAsWhereTheFileIs(t *testing.T) {
	mockStatMethod := func(filePath string) (os.FileInfo, error) {
		filePathAndMockReturnMap := map[string]error{
			"/root/kurtosis.yml":                os.ErrNotExist,
			"/root/subdir/kurtosis.yml":         nil,
			"/root/subdir/subdir1/kurtosis.yml": os.ErrNotExist,
		}

		return nil, filePathAndMockReturnMap[filePath]
	}

	filePath := "/root/subdir/some_file.txt"
	actual, err := checkIfKurtosisYamlExistsInThePathProvidedInternal(filePath, mockStatMethod)
	require.Nil(t, err)
	require.Equal(t, "/root/subdir/kurtosis.yml", actual)
}

func Test_checkIfKurtosisYamlExistsInThePathProvided_fileNotFound(t *testing.T) {
	mockStatMethod := func(filePath string) (os.FileInfo, error) {
		filePathAndMockReturnMap := map[string]error{
			"/root/kurtosis.yml":                os.ErrNotExist,
			"/root/subdir/kurtosis.yml":         os.ErrNotExist,
			"/root/subdir/subdir1/kurtosis.yml": os.ErrNotExist,
		}

		return nil, filePathAndMockReturnMap[filePath]
	}

	filePath := "/root/subdir/some_file.txt"
	actual, err := checkIfKurtosisYamlExistsInThePathProvidedInternal(filePath, mockStatMethod)
	require.Nil(t, err)
	require.Equal(t, filePathToKurtosisYamlNotFound, actual)
}

func Test_checkIfKurtosisYamlExistsInThePathProvided_unknownErrorOccurred(t *testing.T) {
	mockStatMethod := func(filePath string) (os.FileInfo, error) {
		filePathAndMockReturnMap := map[string]error{
			"/root/kurtosis.yml":                os.ErrNotExist,
			"/root/subdir/kurtosis.yml":         os.ErrClosed,
			"/root/subdir/subdir1/kurtosis.yml": os.ErrNotExist,
		}

		return nil, filePathAndMockReturnMap[filePath]
	}

	filePath := "/root/subdir/some_file.txt"
	_, err := checkIfKurtosisYamlExistsInThePathProvidedInternal(filePath, mockStatMethod)
	require.NotNil(t, err)
	require.ErrorContains(t, err, fmt.Sprintf("An error occurred while locating kurtosis.yml in the path of '%v'", filePath))
}
