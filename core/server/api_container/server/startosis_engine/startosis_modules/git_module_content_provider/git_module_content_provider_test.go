package git_module_content_provider

import (
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"testing"
)

const (
	moduleDirRelPath    = "startosis-modules"
	moduleTmpDirRelPath = "tmp-startosis-modules"
)

func TestGitModuleProvider_SucceedsForValidModule(t *testing.T) {
	moduleDir, err := os.MkdirTemp("", moduleDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(moduleDir)
	moduleTmpDir, err := os.MkdirTemp("", moduleTmpDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(moduleTmpDir)

	provider := NewGitModuleContentProvider(moduleDir, moduleTmpDir)

	sampleStartosisModule := "github.com/kurtosis-tech/sample-startosis-load/sample.star"
	contents, err := provider.GetModuleContents(sampleStartosisModule)
	require.Nil(t, err)
	require.Equal(t, "a = \"World!\"\n", contents)
}

func TestGitModuleProvider_FailsForNonExistentModule(t *testing.T) {
	moduleDir, err := os.MkdirTemp("", moduleDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(moduleDir)
	moduleTmpDir, err := os.MkdirTemp("", moduleTmpDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(moduleTmpDir)

	provider := NewGitModuleContentProvider(moduleDir, moduleTmpDir)
	nonExistentModulePath := "github.com/kurtosis-tech/non-existent-startosis-load/sample.star"

	_, err = provider.GetModuleContents(nonExistentModulePath)
	require.NotNil(t, err)
}

func TestGitModuleProvider_ValidRelativePath(t *testing.T) {
	moduleDir := "/kurtosis-data/startosis-modules"
	provider := NewGitModuleContentProvider(moduleDir, "/kurtosis-data/tmp-startosis-modules")

	fileBeingInterpreted := path.Join(moduleDir, "fizz", "buzz", "main.star")
	relativeFilePathToLoad := "./lib/lib.star"

	expectedAbsolutePath := path.Join(moduleDir, "fizz", "buzz", relativeFilePathToLoad)

	result, err := provider.getAbsolutePath(fileBeingInterpreted, relativeFilePathToLoad)
	require.Nil(t, err)
	require.Equal(t, expectedAbsolutePath, result)
}

func TestGitModuleProvider_UnsafePathsLeadToErrors(t *testing.T) {
	moduleDir := "/kurtosis-data/startosis-modules"
	provider := NewGitModuleContentProvider(moduleDir, "/kurtosis-data/tmp-startosis-modules")

	fileBeingInterpreted := path.Join(moduleDir, "fizz", "buzz", "main.star")
	pathThatEscapesOutOfModule := "./../../lib.star"

	_, err := provider.getAbsolutePath(fileBeingInterpreted, pathThatEscapesOutOfModule)
	require.NotNil(t, err)
}

func TestGitModuleProvider_RelativeLoadWithInvalidFilePathFails(t *testing.T) {
	moduleDir := "/kurtosis-data/startosis-modules"
	provider := NewGitModuleContentProvider(moduleDir, "/kurtosis-data/tmp-startosis-modules")

	fileBeingInterpreted := "fileNameNotInUse"
	relativeFilePathToLoad := "./lib/lib.star"

	_, err := provider.getAbsolutePath(fileBeingInterpreted, relativeFilePathToLoad)
	require.NotNil(t, err)
}
