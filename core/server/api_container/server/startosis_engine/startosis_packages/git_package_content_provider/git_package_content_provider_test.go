package git_package_content_provider

import (
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

	sampleStartosisModule := "github.com/kurtosis-tech/sample-startosis-load/sample.star@0.1.0"
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

	sampleStartosisModule := "github.com/kurtosis-tech/sample-startosis-load/sample.star@758f55f902416e94f7a956e9a14938e39833df55"
	contents, err := provider.GetModuleContents(sampleStartosisModule)
	require.Nil(t, err)
	require.Equal(t, "a = \"Hello!\"\n", contents)
}

func TestGitPackageProvider_SucceedsForValidPackageWithCommitOnABranch(t *testing.T) {
	packageDir, err := os.MkdirTemp("", packagesDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageDir)
	packageTmpDir, err := os.MkdirTemp("", packagesTmpDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageTmpDir)

	provider := NewGitPackageContentProvider(packageDir, packageTmpDir)

	sampleStartosisModule := "github.com/kurtosis-tech/sample-startosis-load/sample.star@507eecb57dff8e054f83640ba5953507d83a81c5"
	contents, err := provider.GetModuleContents(sampleStartosisModule)
	require.Nil(t, err)
	require.Equal(t, "a = \"Test!\"\n", contents)
}

func TestGitPackageProvider_SucceedsForNonStarlarkFile(t *testing.T) {
	packageDir, err := os.MkdirTemp("", packagesDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageDir)
	packageTmpDir, err := os.MkdirTemp("", packagesTmpDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(packageTmpDir)

	provider := NewGitPackageContentProvider(packageDir, packageTmpDir)

	sampleStarlarkPackage := "github.com/kurtosis-tech/eth2-merge-kurtosis-module/kurtosis-module/static_files/prometheus-config/prometheus.yml.tmpl"
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
