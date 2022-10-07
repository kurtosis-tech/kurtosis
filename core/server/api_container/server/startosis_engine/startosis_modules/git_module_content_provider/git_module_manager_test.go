package git_module_content_provider

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

const (
	moduleDirRelPath    = "startosis-modules"
	moduleTmpDirRelPath = "tmp-startosis-modules"
)

func TestStartosisInterpreter_GitModuleManagerSucceedsForExistentModule(t *testing.T) {
	moduleDir, err := os.MkdirTemp("", moduleDirRelPath)
	require.NotNil(t, err)
	defer os.RemoveAll(moduleDir)
	moduleTmpDir, err := os.MkdirTemp("", moduleTmpDirRelPath)
	require.NotNil(t, err)
	defer os.RemoveAll(moduleTmpDir)

	gitModuleManager := NewGitModuleContentProvider(moduleDir, moduleTmpDir)

	sampleStartosisModule := "github.com/kurtosis-tech/sample-startosis-load/sample.star"
	contents, err := gitModuleManager.GetModuleContents(sampleStartosisModule)
	require.Nil(t, err)
	require.Equal(t, "a = \"World!\"\n", contents)
}

func TestStartosisInterpreter_GitModuleManagerFailsForNonExistentModule(t *testing.T) {
	moduleDir, err := os.MkdirTemp("", moduleDirRelPath)
	require.NotNil(t, err)
	defer os.RemoveAll(moduleDir)
	moduleTmpDir, err := os.MkdirTemp("", moduleTmpDirRelPath)
	require.NotNil(t, err)
	defer os.RemoveAll(moduleTmpDir)

	gitModuleManager := NewGitModuleContentProvider(moduleDir, moduleTmpDir)
	nonExistentModulePath := "github.com/kurtosis-tech/non-existent-startosis-load/sample.star"

	_, err = gitModuleManager.GetModuleContents(nonExistentModulePath)
	require.NotNil(t, err)
}
