package git_module_content_provider

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

const (
	modulesDirRelPath    = "startosis-modules"
	modulesTmpDirRelPath = "tmp-startosis-modules"
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
