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

// TODO: Remove skip once PortSpec is available is remote package
func TestGetAbsolutePathOnDisk_WorksForPureDirectories(t *testing.T) {
	t.Skip()
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
