package git_module_content_provider

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"testing"
)

const (
	modulesDirRelPath    = "startosis-modules"
	modulesTmpDirRelPath = "tmp-startosis-modules"
	testModulesDir       = "/kurtosis-data/startosis-modules"
	testModulesTmpDir    = "/kurtosis-data/tmp-startosis-modules"
)

func TestGitModuleProvider_SucceedsForValidModule(t *testing.T) {
	moduleDir, err := os.MkdirTemp("", modulesDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(moduleDir)
	moduleTmpDir, err := os.MkdirTemp("", modulesTmpDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(moduleTmpDir)

	provider := NewGitModuleContentProvider(moduleDir, moduleTmpDir)

	sampleStartosisModule := "github.com/kurtosis-tech/sample-startosis-load/sample.star"
	contents, err := provider.GetModuleContents(sampleStartosisModule)
	require.Nil(t, err)
	require.Equal(t, "a = \"World!\"\n", contents)
}

func TestGitModuleProvider_SucceedsForNonStartosisFile(t *testing.T) {
	moduleDir, err := os.MkdirTemp("", modulesDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(moduleDir)
	moduleTmpDir, err := os.MkdirTemp("", modulesTmpDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(moduleTmpDir)

	provider := NewGitModuleContentProvider(moduleDir, moduleTmpDir)

	sampleStartosisModule := "github.com/kurtosis-tech/eth2-merge-kurtosis-module/kurtosis-module/static_files/prometheus-config/prometheus.yml.tmpl"
	contents, err := provider.GetModuleContents(sampleStartosisModule)
	require.Nil(t, err)
	require.NotEmpty(t, contents)
}

func TestGitModuleProvider_FailsForNonExistentModule(t *testing.T) {
	moduleDir, err := os.MkdirTemp("", modulesDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(moduleDir)
	moduleTmpDir, err := os.MkdirTemp("", modulesTmpDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(moduleTmpDir)

	provider := NewGitModuleContentProvider(moduleDir, moduleTmpDir)
	nonExistentModulePath := "github.com/kurtosis-tech/non-existent-startosis-load/sample.star"

	_, err = provider.GetModuleContents(nonExistentModulePath)
	require.NotNil(t, err)
}

func TestGetAbsolutePath_ValidRelativePath(t *testing.T) {
	provider := NewGitModuleContentProvider(testModulesDir, testModulesTmpDir)

	fileBeingInterpreted := path.Join(testModulesDir, "fizz", "buzz", "main.star")
	relativeFilePathToLoad := "./lib/lib.star"

	expectedAbsolutePath := path.Join(testModulesDir, "fizz", "buzz", relativeFilePathToLoad)

	result, err := provider.getAbsolutePath(fileBeingInterpreted, relativeFilePathToLoad)
	require.Nil(t, err)
	require.Equal(t, expectedAbsolutePath, result)
}

func TestGetAbsolutePath_ValidRelativePathWithDirectoryChange(t *testing.T) {
	provider := NewGitModuleContentProvider(testModulesDir, testModulesTmpDir)

	fileBeingInterpreted := path.Join(testModulesDir, "fizz", "buzz", "foo", "main.star")
	relativeFilePathToLoad := "../lib/lib.star"

	expectedAbsolutePath := path.Join(testModulesDir, "fizz", "buzz", "lib/lib.star")

	result, err := provider.getAbsolutePath(fileBeingInterpreted, relativeFilePathToLoad)
	require.Nil(t, err)
	require.Equal(t, expectedAbsolutePath, result)
}

func TestGetAbsolutePath_UnsafePathsLeadToErrors(t *testing.T) {
	provider := NewGitModuleContentProvider(testModulesDir, testModulesTmpDir)

	fileBeingInterpreted := path.Join(testModulesDir, "fizz", "buzz", "main.star")
	pathThatEscapesOutOfModule := "./../../lib.star"

	_, err := provider.getAbsolutePath(fileBeingInterpreted, pathThatEscapesOutOfModule)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "which is unsafe.")
}

func TestGetAbsolutePath_RelativeLoadWithInvalidFileBeingInterpretedPathFails(t *testing.T) {
	provider := NewGitModuleContentProvider(testModulesDir, testModulesTmpDir)

	invalidFileBeingInterpreted := "fileNameNotInUse"
	relativeFilePathToLoad := "./lib/lib.star"

	_, err := provider.getAbsolutePath(invalidFileBeingInterpreted, relativeFilePathToLoad)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("File being interpreted '%v' seems to have an illegal path. This is a bug in Kurtosis.", invalidFileBeingInterpreted))
}

func TestIsGithubPath_WorksForPathThatStartsWithTheGithubDomain(t *testing.T) {
	provider := NewGitModuleContentProvider(testModulesDir, testModulesTmpDir)

	validGitHubPath := "github.com/fizz/buzz/main.star"
	require.True(t, provider.IsGithubPath(validGitHubPath))
}

func TestIsGithubPath_FalseForPathWithoutGithubDomain(t *testing.T) {
	provider := NewGitModuleContentProvider(testModulesDir, testModulesTmpDir)

	invalidGitlabPath := "gitlab.com/fizz/buzz/main.star"
	require.False(t, provider.IsGithubPath(invalidGitlabPath))
}

func TestGetFileAtRelativePath_FailsForAbsolutePath(t *testing.T) {
	provider := NewGitModuleContentProvider(testModulesDir, testModulesTmpDir)

	inputPath := "/absolute/path/main.star"
	_, err := provider.GetFileAtRelativePath("/doesnt/matter/main.star", inputPath)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Expected a relative path but got absolute path")
}

func TestGetFileAtRelativePath_SucceedsForValidRelativePath(t *testing.T) {
	moduleDir, err := os.MkdirTemp("", modulesDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(moduleDir)
	staticFileDir := fmt.Sprintf("%v/testAuthor/testModule/static_files", moduleDir)
	err = os.MkdirAll(staticFileDir, moduleDirPermission)
	require.Nil(t, err)
	fileContents := "this should work"
	err = os.WriteFile(fmt.Sprintf("%v/main.txt", staticFileDir), []byte(fileContents), moduleDirPermission)
	require.Nil(t, err)
	moduleTmpDir, err := os.MkdirTemp("", modulesTmpDirRelPath)
	require.Nil(t, err)
	defer os.RemoveAll(moduleTmpDir)

	provider := NewGitModuleContentProvider(moduleDir, moduleTmpDir)

	inputPath := "./static_files/main.txt"
	fileBeingInterpreted := fmt.Sprintf("%v/testAuthor/testModule/test.star", moduleDir)
	contents, err := provider.GetFileAtRelativePath(fileBeingInterpreted, inputPath)
	require.Nil(t, err)
	require.Equal(t, fileContents, contents)
}
